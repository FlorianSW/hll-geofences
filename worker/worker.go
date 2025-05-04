package worker

import (
	"context"
	"fmt"
	"github.com/floriansw/go-hll-rcon/rconv2"
	"github.com/floriansw/go-hll-rcon/rconv2/api"
	"github.com/floriansw/hll-geofences/data"
	"github.com/floriansw/hll-geofences/sync"
	"log/slog"
	"slices"
	"time"
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

	current        *api.GetSessionResponse
	outsidePlayers sync.Map[string, outsidePlayer]
	firstCoord     sync.Map[string, *api.WorldPosition]
}

type outsidePlayer struct {
	Name         string
	LastGrid     api.Grid
	FirstOutside time.Time
}

var alliedTeams = []api.PlayerTeam{
	api.PlayerTeamB8a,
	api.PlayerTeamDak,
	api.PlayerTeamGb,
	api.PlayerTeamRus,
	api.PlayerTeamUs,
}

var axisTeams = []api.PlayerTeam{
	api.PlayerTeamGer,
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
		firstCoord:     sync.Map[string, *api.WorldPosition]{},
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
		si, err := c.SessionInfo(ctx)
		if err != nil {
			return err
		}
		w.current = si
		w.axisFences = w.applicableFences(w.c.AxisFence)
		w.alliesFences = w.applicableFences(w.c.AlliesFence)
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

			err := w.pool.WithConnection(ctx, func(c *rconv2.Connection) error {
				players, err := c.Players(ctx)
				if err != nil {
					return err
				}
				for _, player := range players.Players {
					go w.checkPlayer(ctx, player)
				}
				w.firstCoord.Range(func(id string, p *api.WorldPosition) bool {
					for _, player := range players.Players {
						if player.Id == id {
							return true
						}
					}
					w.firstCoord.Delete(id)
					return true
				})
				w.outsidePlayers.Range(func(id string, _ outsidePlayer) bool {
					for _, player := range players.Players {
						if player.Id == id {
							return true
						}
					}
					w.outsidePlayers.Delete(id)
					return true
				})
				return nil
			})
			if err != nil {
				w.l.Error("poll-players", "error", err)
			}
		}
	}
}

func (w *worker) checkPlayer(ctx context.Context, p api.GetPlayerResponse) {
	if !p.Position.IsSpawned() {
		w.firstCoord.Store(p.Id, nil)
		return
	}

	// If this is the first time we've seen this player, or if it is still the same position (spawn screen e.g.)
	// ignore them. The game engine returns the position of a random HQ for players first joining the server, which
	// might trigger an out-of-fence warning when we do not ignore that here.
	if fp, ok := w.firstCoord.Load(p.Id); !ok {
		w.firstCoord.Store(p.Id, &p.Position)
		return
	} else if fp != nil && p.Position.Equal(*fp) {
		return
	} else {
		// the player moved (e.g., spawned somewhere), makes sure we start tracking
		// the position of this player and evaluate them against fences.
		w.firstCoord.Store(p.Id, nil)
	}

	g := p.Position.Grid(w.current)
	if o, ok := w.outsidePlayers.Load(p.Id); ok {
		o.LastGrid = g
		w.outsidePlayers.Store(p.Id, o)
		return
	}

	var fences []data.Fence
	if slices.Contains(alliedTeams, p.Team) {
		fences = w.alliesFences
	} else if slices.Contains(axisTeams, p.Team) {
		fences = w.axisFences
	}
	if len(fences) == 0 {
		return
	}

	for _, f := range fences {
		if f.Includes(g) {
			w.outsidePlayers.Delete(p.Id)
			return
		}
	}
	w.outsidePlayers.Store(p.Id, outsidePlayer{FirstOutside: time.Now(), Name: p.Name, LastGrid: g})
	w.l.Info("player-outside-fence", "player", p.Name, "grid", g)
	err := w.pool.WithConnection(ctx, func(c *rconv2.Connection) error {
		return c.MessagePlayer(ctx, p.Name, fmt.Sprintf(w.c.WarningMessage(), w.punishAfterSeconds.String()))
	})
	if err != nil {
		w.l.Error("message-player-outside-fence", "player", p.Name, "grid", g, "error", err)
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
