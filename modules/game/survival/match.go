package survival

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"time"

	"squad-survival-be/modules/game/core/entity"
	"squad-survival-be/modules/game/core/spatial"
	"squad-survival-be/modules/game/core/system"
	"squad-survival-be/modules/game/core/world"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

// Config của mode survival, có thể thay đổi được thông qua params khi tạo match.
const (
	ModuleName            = "survival"
	DefaultMode           = "survival"
	MaxPlayers            = 32
	tickRate              = entity.TickRate
	reservationTTLSeconds = 10
	emptyMatchTTLSeconds  = 60
	spatialCellSize       = 20.0
)

type Match struct{}

type State struct {
	Mode                string
	AllowJoinInProgress bool
	Players             map[string]*entity.Player
	Reservations        map[string]int64
	SpatialGrid         *spatial.Grid
	EmptyTicks          int64
	random              *rand.Rand
}

type Label struct {
	Mode        string `json:"mode"`
	Status      string `json:"status"`
	PlayerCount int    `json:"player_count"`
	MaxPlayers  int    `json:"max_players"`
	Joinable    bool   `json:"joinable"`
}

func NewMatch(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule) (runtime.Match, error) {
	return &Match{}, nil
}

func (m *Match) MatchInit(_ context.Context, logger runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &State{
		Mode:                stringParam(params, "mode", DefaultMode),
		AllowJoinInProgress: boolParam(params, "allow_join_in_progress", true),
		Players:             make(map[string]*entity.Player),
		Reservations:        make(map[string]int64),
		SpatialGrid:         spatial.NewGrid(spatialCellSize),
		random:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	logger.Info("Survival match initialized: mode=%s max_players=%d", state.Mode, MaxPlayers)
	return state, tickRate, state.label()
}

func (m *Match) MatchJoinAttempt(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, _ runtime.MatchDispatcher, tick int64, rawState interface{}, presence runtime.Presence, _ map[string]string) (interface{}, bool, string) {
	state := rawState.(*State)
	state.expireReservations(tick)

	if _, exists := state.Players[presence.GetSessionId()]; exists {
		return state, true, ""
	}
	if !state.AllowJoinInProgress && len(state.Players) > 0 {
		return state, false, "match already started"
	}
	if state.occupiedSlots() >= MaxPlayers {
		return state, false, "match is full"
	}

	state.Reservations[presence.GetSessionId()] = tick + reservationTTLSeconds*tickRate
	return state, true, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, rawState interface{}, presences []runtime.Presence) interface{} {
	state := rawState.(*State)
	playerCount := len(state.Players)
	displayNames, err := resolveDisplayNames(ctx, nk, presences)
	if err != nil && logger != nil {
		logger.Error("Could not load player display names: %v", err)
	}
	for _, presence := range presences {
		delete(state.Reservations, presence.GetSessionId())
		if _, exists := state.Players[presence.GetSessionId()]; exists {
			continue
		}
		player := entity.NewPlayer(
			presence.GetUserId(),
			presence.GetSessionId(),
			displayNames[presence.GetUserId()],
			world.RandomSpawn(state.random),
			state.random,
		)
		if err = state.SpatialGrid.Insert(player); err != nil {
			if logger != nil {
				logger.Error("Could not insert player into spatial grid: session_id=%s error=%v", presence.GetSessionId(), err)
			}
			continue
		}
		state.Players[presence.GetSessionId()] = player
	}
	state.EmptyTicks = 0
	state.updateLabel(dispatcher)
	if len(state.Players) != playerCount {
		state.broadcastSnapshot(logger, dispatcher, tick)
	}
	return state
}

type userLookup interface {
	UsersGetId(ctx context.Context, userIDs []string, facebookIDs []string) ([]*api.User, error)
}

func resolveDisplayNames(ctx context.Context, lookup userLookup, presences []runtime.Presence) (map[string]string, error) {
	displayNames := make(map[string]string, len(presences))
	userIDs := make([]string, 0, len(presences))
	for _, presence := range presences {
		displayNames[presence.GetUserId()] = presence.GetUsername()
		userIDs = append(userIDs, presence.GetUserId())
	}
	if lookup == nil || len(userIDs) == 0 {
		return displayNames, nil
	}

	users, err := lookup.UsersGetId(ctx, userIDs, nil)
	if err != nil {
		return displayNames, err
	}
	for _, user := range users {
		if user.GetDisplayName() != "" {
			displayNames[user.GetId()] = user.GetDisplayName()
		}
	}
	return displayNames, nil
}

func (m *Match) MatchLeave(_ context.Context, logger runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, rawState interface{}, presences []runtime.Presence) interface{} {
	state := rawState.(*State)
	playerCount := len(state.Players)
	for _, presence := range presences {
		delete(state.Reservations, presence.GetSessionId())
		if _, exists := state.Players[presence.GetSessionId()]; exists && !state.SpatialGrid.Remove(presence.GetSessionId()) && logger != nil {
			logger.Error("Could not remove player from spatial grid: session_id=%s", presence.GetSessionId())
		}
		delete(state.Players, presence.GetSessionId())
	}
	state.updateLabel(dispatcher)
	if len(state.Players) != playerCount {
		state.broadcastSnapshot(logger, dispatcher, tick)
	}
	return state
}

func (m *Match) MatchLoop(_ context.Context, logger runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, rawState interface{}, messages []runtime.MatchData) interface{} {
	state := rawState.(*State)
	if state.expireReservations(tick) {
		state.updateLabel(dispatcher)
	}

	for _, message := range messages {
		if message.GetOpCode() != system.OpMovementInput {
			continue
		}
		player, ok := state.Players[message.GetSessionId()]
		if !ok {
			continue
		}
		input, err := entity.DecodeMovementInput(message.GetData())
		if err != nil {
			continue
		}
		entity.ApplyMovementInput(player, input, tick)
	}

	for _, player := range state.Players {
		entity.StepMovement(player, tick)
		if err := state.SpatialGrid.Move(player); err != nil && logger != nil {
			logger.Error("Could not move player in spatial grid: session_id=%s error=%v", player.SessionID, err)
		}
	}

	if len(state.Players) == 0 && len(state.Reservations) == 0 {
		state.EmptyTicks++
		if state.EmptyTicks >= emptyMatchTTLSeconds*tickRate {
			logger.Info("Stopping empty survival match")
			return nil
		}
	} else {
		state.EmptyTicks = 0
	}

	return state
}

func (m *Match) MatchTerminate(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, _ runtime.MatchDispatcher, _ int64, _ interface{}, _ int) interface{} {
	return nil
}

func (m *Match) MatchSignal(_ context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, _ runtime.MatchDispatcher, _ int64, rawState interface{}, _ string) (interface{}, string) {
	state := rawState.(*State)
	return state, state.label()
}

func (s *State) occupiedSlots() int {
	return len(s.Players) + len(s.Reservations)
}

func (s *State) expireReservations(tick int64) bool {
	expired := false
	for sessionID, expiresAt := range s.Reservations {
		if tick >= expiresAt {
			delete(s.Reservations, sessionID)
			expired = true
		}
	}
	return expired
}

func (s *State) updateLabel(dispatcher runtime.MatchDispatcher) {
	_ = dispatcher.MatchLabelUpdate(s.label())
}

func (s *State) broadcastSnapshot(logger runtime.Logger, dispatcher runtime.MatchDispatcher, tick int64) {
	snapshot, err := system.EncodeStateSnapshot(tick, len(s.Players))
	if err == nil {
		err = dispatcher.BroadcastMessage(system.OpStateSnapshot, snapshot, nil, nil, true)
	}
	if err != nil && logger != nil {
		logger.Error("Could not broadcast state snapshot: %v", err)
	}
}

func (s *State) label() string {
	label, _ := json.Marshal(Label{
		Mode:        s.Mode,
		Status:      "playing",
		PlayerCount: len(s.Players),
		MaxPlayers:  MaxPlayers,
		Joinable:    s.AllowJoinInProgress && s.occupiedSlots() < MaxPlayers,
	})
	return string(label)
}

func stringParam(params map[string]interface{}, key, fallback string) string {
	if value, ok := params[key].(string); ok && value != "" {
		return value
	}
	return fallback
}

func boolParam(params map[string]interface{}, key string, fallback bool) bool {
	if value, ok := params[key].(bool); ok {
		return value
	}
	return fallback
}
