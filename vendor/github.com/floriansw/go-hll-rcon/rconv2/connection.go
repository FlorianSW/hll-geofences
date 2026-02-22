package rconv2

import (
	"context"
	"errors"
	"strings"

	"github.com/floriansw/go-hll-rcon/rconv2/api"
)

// Connection represents a persistent connection to a HLL server using RCon. It can be used to issue commands against
// the HLL server and query data. The connection can either be utilised using the higher-level API methods, or by sending
// raw commands using ListCommand or Command.
//
// A Connection is not thread-safe by default. Do not attempt to run multiple commands (either of the higher-level or
// low-level API). Doing so may either run into non-expected indefinitely blocking execution (until the context.Context
// deadline exceeds) or to mixed up data (sending a command and getting back the response for another command).
// Instead, in goroutines use a ConnectionPool and request a new connection for each goroutine. The ConnectionPool will
// ensure that one Connection is only used once at the same time. It also speeds up processing by opening a number of
// Connections until the pool size is reached.
type Connection struct {
	id     string
	socket *socket
}

func (c *Connection) Players(ctx context.Context) (*api.GetPlayersResponse, error) {
	return execCommand[api.GetServerInformation, api.GetPlayersResponse](ctx, c.socket, api.GetServerInformation{
		Name: api.ServerInformationNamePlayers,
	})
}

func (c *Connection) Player(ctx context.Context, playerId string) (*api.GetPlayerResponse, error) {
	return execCommand[api.GetServerInformation, api.GetPlayerResponse](ctx, c.socket, api.GetServerInformation{
		Name:  api.ServerInformationNamePlayer,
		Value: playerId,
	})
}

func (c *Connection) ServerConfig(ctx context.Context) (*api.GetServerConfigResponse, error) {
	return execCommand[api.GetServerInformation, api.GetServerConfigResponse](ctx, c.socket, api.GetServerInformation{
		Name: api.ServerInformationNameServerConfig,
	})
}

func (c *Connection) SessionInfo(ctx context.Context) (*api.GetSessionResponse, error) {
	return execCommand[api.GetServerInformation, api.GetSessionResponse](ctx, c.socket, api.GetServerInformation{
		Name: api.ServerInformationNameSession,
	})
}

func (c *Connection) MapRotation(ctx context.Context) (*api.GetMapRotationResponse, error) {
	return execCommand[api.GetServerInformation, api.GetMapRotationResponse](ctx, c.socket, api.GetServerInformation{
		Name: api.ServerInformationNameMapRotation,
	})
}

func (c *Connection) MapSequence(ctx context.Context) (*api.GetMapSequenceResponse, error) {
	return execCommand[api.GetServerInformation, api.GetMapSequenceResponse](ctx, c.socket, api.GetServerInformation{
		Name: api.ServerInformationNameMapSequence,
	})
}

func (c *Connection) DisplayableCommands(ctx context.Context) (*api.GetDisplayableCommandsResponse, error) {
	return execCommand[api.GetDisplayableCommands, api.GetDisplayableCommandsResponse](ctx, c.socket, api.GetDisplayableCommands{})
}

func (c *Connection) AdminLog(ctx context.Context, timeSeconds int32, filter string) (*api.GetAdminLogResponse, error) {
	return execCommand[api.GetAdminLog, api.GetAdminLogResponse](ctx, c.socket, api.GetAdminLog{
		LogBackTrackTime: timeSeconds,
		Filters:          filter,
	})
}

func (c *Connection) ChangeMap(ctx context.Context, mapName string) error {
	_, err := execCommand[api.ChangeMap, any](ctx, c.socket, api.ChangeMap{
		MapName: mapName,
	})
	return err
}

func (c *Connection) SetSectorLayout(ctx context.Context, sectors []string) error {
	r := api.SetSectorLayout{}
	for i, sector := range sectors {
		if i == 0 {
			r.SectorOne = sector
		}
		if i == 1 {
			r.SectorTwo = sector
		}
		if i == 2 {
			r.SectorThree = sector
		}
		if i == 3 {
			r.SectorFour = sector
		}
		if i == 4 {
			r.SectorFive = sector
		}
	}
	_, err := execCommand[api.SetSectorLayout, any](ctx, c.socket, r)
	return err
}

func (c *Connection) AdminGroups(ctx context.Context) (*api.GetAdminGroupsResponse, error) {
	return execCommand[api.GetAdminGroups, api.GetAdminGroupsResponse](ctx, c.socket, api.GetAdminGroups{})
}

func (c *Connection) AdminUsers(ctx context.Context) (*api.GetAdminUsersResponse, error) {
	return execCommand[api.GetAdminUsers, api.GetAdminUsersResponse](ctx, c.socket, api.GetAdminUsers{})
}

func (c *Connection) AddAdmin(ctx context.Context, playerId, adminGroup, comment string) error {
	_, err := execCommand[api.AddAdmin, any](ctx, c.socket, api.AddAdmin{
		PlayerId:   playerId,
		AdminGroup: adminGroup,
		Comment:    comment,
	})
	return err
}

func (c *Connection) RemoveAdmin(ctx context.Context, playerId string) error {
	_, err := execCommand[api.RemoveAdmin, any](ctx, c.socket, api.RemoveAdmin{
		PlayerId: playerId,
	})
	return err
}

func (c *Connection) AddMapToRotation(ctx context.Context, mapName string, index int32) error {
	_, err := execCommand[api.AddMapToRotation, any](ctx, c.socket, api.AddMapToRotation{
		MapName: mapName,
		Index:   index,
	})
	return err
}

func (c *Connection) AddMapToSequence(ctx context.Context, mapName string, index int32) error {
	_, err := execCommand[api.AddMapToSequence, any](ctx, c.socket, api.AddMapToSequence{
		MapName: mapName,
		Index:   index,
	})
	return err
}

func (c *Connection) RemoveMapFromRotation(ctx context.Context, index int32) error {
	_, err := execCommand[api.RemoveMapFromRotation, any](ctx, c.socket, api.RemoveMapFromRotation{
		Index: index,
	})
	return err
}

func (c *Connection) RemoveMapToSequence(ctx context.Context, index int32) error {
	_, err := execCommand[api.RemoveMapFromSequence, any](ctx, c.socket, api.RemoveMapFromSequence{
		Index: index,
	})
	return err
}

func (c *Connection) SetMapShuffleEnabled(ctx context.Context, enable bool) error {
	_, err := execCommand[api.SetMapShuffleEnabled, any](ctx, c.socket, api.SetMapShuffleEnabled{
		Enable: enable,
	})
	return err
}

func (c *Connection) MoveMapInSequence(ctx context.Context, currentIndex, newIndex int32) error {
	_, err := execCommand[api.MoveMapInSequence, any](ctx, c.socket, api.MoveMapInSequence{
		CurrentIndex: currentIndex,
		NewIndex:     newIndex,
	})
	return err
}

func (c *Connection) SetTeamSwitchCooldown(ctx context.Context, timer int32) error {
	_, err := execCommand[api.SetTeamSwitchCooldown, any](ctx, c.socket, api.SetTeamSwitchCooldown{
		TeamSwitchTimer: timer,
	})
	return err
}

func (c *Connection) SetMatchTimer(ctx context.Context, gameMode string, timer int32) error {
	_, err := execCommand[api.SetMatchTimer, any](ctx, c.socket, api.SetMatchTimer{
		GameMode:    gameMode,
		MatchLength: timer,
	})
	return err
}

func (c *Connection) RemoveMatchTimer(ctx context.Context, gameMode string) error {
	_, err := execCommand[api.RemoveMatchTimer, any](ctx, c.socket, api.RemoveMatchTimer{
		GameMode: gameMode,
	})
	return err
}

func (c *Connection) SetWarmupTimer(ctx context.Context, gameMode string, timer int32) error {
	_, err := execCommand[api.SetWarmupTimer, any](ctx, c.socket, api.SetWarmupTimer{
		GameMode:     gameMode,
		WarmupLength: timer,
	})
	return err
}

func (c *Connection) RemoveWarmupTimer(ctx context.Context, gameMode string) error {
	_, err := execCommand[api.RemoveWarmupTimer, any](ctx, c.socket, api.RemoveWarmupTimer{
		GameMode: gameMode,
	})
	return err
}

func (c *Connection) SetDynamicWeatherEnabled(ctx context.Context, mapId string, enabled bool) error {
	_, err := execCommand[api.SetDynamicWeatherEnabled, any](ctx, c.socket, api.SetDynamicWeatherEnabled{
		MapId:  mapId,
		Enable: enabled,
	})
	return err
}

func (c *Connection) SetMaxQueuedPlayers(ctx context.Context, maxQueuedPlayers int32) error {
	_, err := execCommand[api.SetMaxQueuedPlayers, any](ctx, c.socket, api.SetMaxQueuedPlayers{
		MaxQueuedPlayers: maxQueuedPlayers,
	})
	return err
}

func (c *Connection) SetIdleKickDuration(ctx context.Context, idleTimeoutMinutes int32) error {
	_, err := execCommand[api.SetIdleKickDuration, any](ctx, c.socket, api.SetIdleKickDuration{
		IdleTimeoutMinutes: idleTimeoutMinutes,
	})
	return err
}

func (c *Connection) SendServerMessage(ctx context.Context, msg string) error {
	_, err := execCommand[api.SendServerMessage, any](ctx, c.socket, api.SendServerMessage{
		Message: msg,
	})
	return err
}

func (c *Connection) ServerBroadcast(ctx context.Context, msg string) error {
	_, err := execCommand[api.ServerBroadcast, any](ctx, c.socket, api.ServerBroadcast{
		Message: msg,
	})
	return err
}

func (c *Connection) SetHighPingThreshold(ctx context.Context, highPingMs int32) error {
	_, err := execCommand[api.SetHighPingThreshold, any](ctx, c.socket, api.SetHighPingThreshold{
		HighPingThresholdMs: highPingMs,
	})
	return err
}

func (c *Connection) MessagePlayer(ctx context.Context, playerId, message string) error {
	_, err := execCommand[api.MessagePlayer, any](ctx, c.socket, api.MessagePlayer{
		Message:  message,
		PlayerId: playerId,
	})
	return err
}

func (c *Connection) PunishPlayer(ctx context.Context, playerId, reason string) error {
	_, err := execCommand[api.PunishPlayer, any](ctx, c.socket, api.PunishPlayer{
		Reason:   reason,
		PlayerId: playerId,
	})
	return err
}

func (c *Connection) KickPlayer(ctx context.Context, playerId, reason string) error {
	_, err := execCommand[api.KickPlayer, any](ctx, c.socket, api.KickPlayer{
		Reason:   reason,
		PlayerId: playerId,
	})
	return err
}

func (c *Connection) TemporaryBanPlayer(ctx context.Context, playerId string, duration int32, reason, adminName string) error {
	_, err := execCommand[api.TemporaryBanPlayer, any](ctx, c.socket, api.TemporaryBanPlayer{
		Reason:    reason,
		PlayerId:  playerId,
		Duration:  duration,
		AdminName: adminName,
	})
	return err
}

func (c *Connection) TemporaryBans(ctx context.Context) (*api.GetTemporaryBansResponse, error) {
	return execCommand[api.GetTemporaryBans, api.GetTemporaryBansResponse](ctx, c.socket, api.GetTemporaryBans{})
}

func (c *Connection) RemoveTemporaryBan(ctx context.Context, playerId string) error {
	_, err := execCommand[api.RemoveTemporaryBan, any](ctx, c.socket, api.RemoveTemporaryBan{
		PlayerId: playerId,
	})
	return err
}

func (c *Connection) PermanentBanPlayer(ctx context.Context, playerId, reason, adminName string) error {
	_, err := execCommand[api.PermanentBanPlayer, any](ctx, c.socket, api.PermanentBanPlayer{
		Reason:    reason,
		PlayerId:  playerId,
		AdminName: adminName,
	})
	return err
}

func (c *Connection) PermanentBans(ctx context.Context) (*api.GetPermanentBansResponse, error) {
	return execCommand[api.GetPermanentBans, api.GetPermanentBansResponse](ctx, c.socket, api.GetPermanentBans{})
}

func (c *Connection) RemovePermanentBan(ctx context.Context, playerId string) error {
	_, err := execCommand[api.RemovePermanentBan, any](ctx, c.socket, api.RemovePermanentBan{
		PlayerId: playerId,
	})
	return err
}

func (c *Connection) SetAutoBalance(ctx context.Context, enable bool) error {
	_, err := execCommand[api.SetAutoBalance, any](ctx, c.socket, api.SetAutoBalance{
		EnableAutoBalance: enable,
	})
	return err
}

func (c *Connection) SetAutoBalanceThreshold(ctx context.Context, threshold int32) error {
	_, err := execCommand[api.SetAutoBalanceThreshold, any](ctx, c.socket, api.SetAutoBalanceThreshold{
		AutoBalanceThreshold: threshold,
	})
	return err
}

func (c *Connection) SetVoteKick(ctx context.Context, enabled bool) error {
	_, err := execCommand[api.SetVoteKick, any](ctx, c.socket, api.SetVoteKick{
		Enabled: enabled,
	})
	return err
}

func (c *Connection) ResetVoteKickThreshold(ctx context.Context) error {
	_, err := execCommand[api.ResetVoteKickThreshold, any](ctx, c.socket, api.ResetVoteKickThreshold{})
	return err
}

func (c *Connection) SetVoteKickThreshold(ctx context.Context, threshold string) error {
	_, err := execCommand[api.SetVoteKickThreshold, any](ctx, c.socket, api.SetVoteKickThreshold{
		ThresholdValue: threshold,
	})
	return err
}

func (c *Connection) SetWelcomeMessage(ctx context.Context, message string) error {
	_, err := execCommand[api.SetWelcomeMessage, any](ctx, c.socket, api.SetWelcomeMessage{
		Message: message,
	})
	return err
}

func (c *Connection) SetVipSlotCount(ctx context.Context, count int32) error {
	_, err := execCommand[api.SetVipSlotCount, any](ctx, c.socket, api.SetVipSlotCount{
		VipSlotCount: count,
	})
	return err
}

func (c *Connection) ForceTeamSwitch(ctx context.Context, playerId string, mode api.ForceMode) error {
	_, err := execCommand[api.ForceTeamSwitch, any](ctx, c.socket, api.ForceTeamSwitch{
		PlayerId:  playerId,
		ForceMode: mode,
	})
	return err
}

func (c *Connection) AddBannedWords(ctx context.Context, words []string) error {
	_, err := execCommand[api.AddBannedWords, any](ctx, c.socket, api.AddBannedWords{
		BannedWords: strings.Join(words, ","),
	})
	return err
}

func (c *Connection) RemoveBannedWords(ctx context.Context, words []string) error {
	_, err := execCommand[api.RemoveBannedWords, any](ctx, c.socket, api.RemoveBannedWords{
		BannedWords: strings.Join(words, ","),
	})
	return err
}

func (c *Connection) AddVip(ctx context.Context, playerId, comment string) error {
	_, err := execCommand[api.AddVip, any](ctx, c.socket, api.AddVip{
		PlayerId: playerId,
		Comment:  comment,
	})
	return err
}

func (c *Connection) RemoveVip(ctx context.Context, playerId string) error {
	_, err := execCommand[api.RemoveVip, any](ctx, c.socket, api.RemoveVip{
		PlayerId: playerId,
	})
	return err
}

func (c *Connection) RemovePlayerFromPlatoon(ctx context.Context, playerId, reason string) error {
	_, err := execCommand[api.RemovePlayerFromPlatoon, any](ctx, c.socket, api.RemovePlayerFromPlatoon{
		PlayerId: playerId,
		Reason:   reason,
	})
	return err
}

func (c *Connection) DisbandPlatoon(ctx context.Context, team, squad int32, reason string) error {
	_, err := execCommand[api.DisbandPlatoon, any](ctx, c.socket, api.DisbandPlatoon{
		TeamIndex:  team,
		SquadIndex: squad,
		Reason:     reason,
	})
	return err
}

// MapFilter A filter used in commands that return list of maps, e.g. Maps or MapRotation.
// The filter should return true, when the map should be included in the result set and false
// when the map should be skipped.
type MapFilter func(idx int, name string, result []string) bool

func (c *Connection) AvailableMaps(ctx context.Context, filters ...MapFilter) ([]string, error) {
	res, err := c.GetClientReferenceData(ctx, "AddMapToRotation")
	if err != nil {
		return nil, err
	}
	var param api.Parameter
	for _, p := range res.Parameters {
		if p.Id == "MapName" {
			param = p
			break
		}
	}
	if param.Id != "MapName" {
		return nil, errors.New("could not find map name parameter")
	}
	maps := strings.Split(param.ValueMember, ",")
	return filter(maps, filters...), nil
}

func filter(maps []string, filters ...MapFilter) []string {
	var result []string
	for i, m := range maps {
		add := true
		for _, filter := range filters {
			if !filter(i, m, result) {
				add = false
			}
		}
		if add {
			result = append(result, m)
		}
	}
	return result
}

func (c *Connection) GetClientReferenceData(ctx context.Context, command string) (*api.GetClientReferenceDataResponse, error) {
	return execCommand[api.GetClientReferenceData, api.GetClientReferenceDataResponse](ctx, c.socket, api.GetClientReferenceData(command))
}

func execCommand[T, U any](ctx context.Context, so *socket, req T) (result *U, err error) {
	err = so.SetContext(ctx)
	if err != nil {
		return nil, err
	}
	r := Request[T, U]{
		Body: req,
	}
	res, err := r.do(so)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, NewUnexpectedStatus(res.StatusCode, res.StatusMessage)
	}
	body := res.Body()
	return &body, nil
}
