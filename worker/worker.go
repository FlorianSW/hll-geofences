package worker

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/floriansw/go-hll-rcon/rconv2"
	"github.com/floriansw/hll-geofences/data"
	"github.com/floriansw/hll-geofences/sync"
)

type worker struct {
	pool               *rconv2.ConnectionPool
	l                  *slog.Logger
	c                  data.Server
	axisFences         []data.Fence
	alliesFences       []data.Fence
	punishAfterSeconds time.Duration

	sessionTicker *time.Ticker
	playerTicker  *time.Ticker
	punishTicker  *time.Ticker

	current        *rconv2.GetSessionInfoResponse
	outsidePlayers sync.Map[string, outsidePlayer]
	firstCoord     sync.Map[string, *rconv2.GetPlayersPosition]
}

type outsidePlayer struct {
	Name         string
	LastGrid     rconv2.Grid
	FirstOutside time.Time
}

var alliedTeams = []rconv2.GetPlayersTeam{
	rconv2.GetPlayersTeamB8A,
	rconv2.GetPlayersTeamDAK,
	rconv2.GetPlayersTeamCW,
	rconv2.GetPlayersTeamSOV,
	rconv2.GetPlayersTeamME,
}

var axisTeams = []rconv2.GetPlayersTeam{
	rconv2.GetPlayersTeamGER,
}

func NewWorker(l *slog.Logger, pool *rconv2.ConnectionPool, c data.Server) *worker {
	punishAfterSeconds := 10
	if c.PunishAfterSeconds != nil {
		punishAfterSeconds = *c.PunishAfterSeconds
	}
	return &worker{
		l:                  l,
		pool:               pool,
		punishAfterSeconds: time.Duration(punishAfterSeconds) * time.Second,
		c:                  c,

		sessionTicker:  time.NewTicker(1 * time.Second),
		playerTicker:   time.NewTicker(500 * time.Millisecond),
		punishTicker:   time.NewTicker(time.Second),
		outsidePlayers: sync.Map[string, outsidePlayer]{},
		firstCoord:     sync.Map[string, *rconv2.GetPlayersPosition]{},
	}
}

func (w *worker) Run(ctx context.Context) {
	if err := w.populateSession(ctx); err != nil {
		w.l.Error("fetch-session", "error", err)
		return
	}

	go w.pollSession(ctx)
	go w.pollPlayers(ctx)
	go w.punishPlayers(ctx)
}

func (w *worker) populateSession(ctx context.Context) error {
	return w.pool.WithConnection(ctx, func(c *rconv2.Connection) error {
		si, err := c.GetSessionInfo(ctx)
		if err != nil {
			return err
		}
		w.current = si
		w.axisFences = w.applicableFences(w.c.AxisFence)
		w.alliesFences = w.applicableFences(w.c.AlliesFence)
		if len(w.alliesFences) == 0 && len(w.axisFences) == 0 {
			w.l.Debug("no-applicable-fences", "map_name", si.MapName, "game_mode", si.GameMode, "player_count", si.PlayerCount)
		}
		return nil
	})
}

func (w *worker) punishPlayers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.punishTicker.Stop()
			return
		case <-w.punishTicker.C:
			w.outsidePlayers.Range(func(id string, o outsidePlayer) bool {
				if time.Since(o.FirstOutside) > w.punishAfterSeconds && time.Since(o.FirstOutside) < w.punishAfterSeconds+5*time.Second {
					go w.punishPlayer(ctx, id, o)
				}
				return true
			})
		}
	}
}

func (w *worker) punishPlayer(ctx context.Context, id string, o outsidePlayer) {
	err := w.pool.WithConnection(ctx, func(c *rconv2.Connection) error {
		return c.PunishPlayer(ctx, id, fmt.Sprintf(w.c.PunishMessage(), w.punishAfterSeconds.String()))
	})
	if err != nil {
		w.l.Error("punish-player", "player_id", id, "error", err)
		return
	}
	w.l.Info("punish-player", "player", o.Name, "grid", o.LastGrid.String())

	time.Sleep(5 * time.Second)
	w.outsidePlayers.Delete(id)
}

func (w *worker) pollSession(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.sessionTicker.Stop()
			return
		case <-w.sessionTicker.C:
			if err := w.populateSession(ctx); err != nil {
				w.l.Error("poll-session", "error", err)
			}
		}
	}
}

func (w *worker) pollPlayers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.playerTicker.Stop()
			return
		case <-w.playerTicker.C:
			if len(w.alliesFences) == 0 && len(w.axisFences) == 0 {
				continue
			}

			var players *rconv2.GetPlayersResponse
			err := w.pool.WithConnection(ctx, func(c *rconv2.Connection) error {
				var err error
				players, err = c.GetPlayers(ctx)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				w.l.Error("poll-players", "error", err)
				continue
			}
			for _, player := range players.Players {
				go w.checkPlayer(ctx, player)
			}
			w.firstCoord.Range(func(id string, p *rconv2.GetPlayersPosition) bool {
				for _, player := range players.Players {
					if player.Id == id {
						return true
					}
				}
				w.firstCoord.Delete(id)
				return true
			})
		}
	}
}

func (w *worker) checkPlayer(ctx context.Context, p rconv2.GetPlayersPlayer) {
	l := w.l.WithGroup("check-player").With("player_id", p.Id, "player_name", p.Name)
	if !p.Position.ToGetPlayerPosition().IsSpawned() {
		l.Debug("player-not-spawned")
		w.firstCoord.Store(p.Id, nil)
		return
	}

	// If this is the first time we've seen this player, or if it is still the same position (spawn screen e.g.)
	// ignore them. The game engine returns the position of a random HQ for players first joining the server, which
	// might trigger an out-of-fence warning when we do not ignore that here.
	if fp, loaded := w.firstCoord.LoadOrStore(p.Id, &p.Position); !loaded {
		l.Debug("player-first-time")
		w.firstCoord.Store(p.Id, &p.Position)
		return
	} else if fp != nil && p.Position.ToGetPlayerPosition().Equal(fp.ToGetPlayerPosition()) {
		l.Debug("player-not-loaded")
		return
	}

	// the player moved (e.g., spawned somewhere), makes sure we start tracking
	// the position of this player and evaluate them against fences.
	w.firstCoord.Store(p.Id, nil)

	var fences []data.Fence
	if slices.Contains(alliedTeams, p.Team) {
		fences = w.alliesFences
	} else if slices.Contains(axisTeams, p.Team) {
		fences = w.axisFences
	}
	if len(fences) == 0 {
		l.Debug("no-applicable-fences", "team", p.Team, "allied_teams", alliedTeams, "axis_teams", axisTeams)
		return
	}

	g := p.Position.ToGetPlayerPosition().Grid(w.current)
	for _, f := range fences {
		if f.Includes(g) {
			l.Debug("player-within-fence", "grid", g.String(), "fence", f)
			w.outsidePlayers.Delete(p.Id)
			return
		}
	}
	if o, ok := w.outsidePlayers.Load(p.Id); ok {
		l.Debug("player-still-outside", "last_grid", o.LastGrid.String(), "grid", g.String(), "first_outside", o.FirstOutside)
		o.LastGrid = g
		w.outsidePlayers.Store(p.Id, o)
		return
	}

	w.outsidePlayers.Store(p.Id, outsidePlayer{FirstOutside: time.Now(), Name: p.Name, LastGrid: g})
	l.Info("player-outside-fence", "grid", g)
	err := w.pool.WithConnection(ctx, func(c *rconv2.Connection) error {
		return c.MessagePlayer(ctx, p.Name, fmt.Sprintf(w.c.WarningMessage(), w.punishAfterSeconds.String()))
	})
	if err != nil {
		l.Error("message-player-outside-fence", "grid", g, "error", err)
	}
}

func (w *worker) applicableFences(f []data.Fence) (v []data.Fence) {
	for _, fence := range f {
		if fence.Matches(w.current) {
			v = append(v, fence)
		}
	}
	return
}
