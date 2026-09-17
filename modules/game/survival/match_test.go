package survival

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"squad-survival-be/modules/game/core"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/proto"
)

func TestJoinAttemptReservesAndExpiresSlot(t *testing.T) {
	match := &Match{}
	state := &State{
		Mode:                DefaultMode,
		AllowJoinInProgress: true,
		Players:             make(map[string]*core.Player),
		Reservations:        make(map[string]int64),
	}
	for i := 0; i < MaxPlayers; i++ {
		presence := testPresence{userID: fmt.Sprintf("user-%d", i), sessionID: fmt.Sprintf("session-%d", i)}
		_, allowed, _ := match.MatchJoinAttempt(nil, nil, nil, nil, nil, 0, state, presence, nil)
		if !allowed {
			t.Fatalf("expected player %d to reserve a slot", i+1)
		}
	}

	second := testPresence{userID: "overflow-user", sessionID: "overflow-session"}
	_, allowed, reason := match.MatchJoinAttempt(nil, nil, nil, nil, nil, 1, state, second, nil)
	if allowed || reason != "match is full" {
		t.Fatalf("expected full match rejection, got allowed=%v reason=%q", allowed, reason)
	}

	_, allowed, _ = match.MatchJoinAttempt(nil, nil, nil, nil, nil, reservationTTLSeconds*tickRate, state, second, nil)
	if !allowed {
		t.Fatal("expected expired reservation to release the slot")
	}
}

func TestJoinAndLeaveUpdateCapacityLabel(t *testing.T) {
	match := &Match{}
	dispatcher := &testDispatcher{}
	presence := testPresence{userID: "user-1", sessionID: "session-1"}
	state := &State{
		Mode:                DefaultMode,
		AllowJoinInProgress: true,
		Players:             make(map[string]*core.Player),
		Reservations:        map[string]int64{presence.sessionID: 100},
		random:              rand.New(rand.NewSource(1)),
	}

	match.MatchJoin(nil, nil, nil, nil, dispatcher, 0, state, []runtime.Presence{presence})
	if state.Players[presence.sessionID].DisplayName != presence.userID {
		t.Fatalf("expected username fallback, got %q", state.Players[presence.sessionID].DisplayName)
	}
	if dispatcher.label != `{"mode":"survival","status":"playing","player_count":1,"max_players":32,"joinable":true}` {
		t.Fatalf("unexpected full label: %s", dispatcher.label)
	}
	assertStateSnapshot(t, dispatcher, 0, 1)

	match.MatchLeave(nil, nil, nil, nil, dispatcher, 2, state, []runtime.Presence{presence})
	if dispatcher.label != `{"mode":"survival","status":"playing","player_count":0,"max_players":32,"joinable":true}` {
		t.Fatalf("unexpected empty label: %s", dispatcher.label)
	}
	assertStateSnapshot(t, dispatcher, 2, 0)
	if dispatcher.broadcastCount != 2 {
		t.Fatalf("expected one snapshot per join/leave, got %d", dispatcher.broadcastCount)
	}
}

func TestResolveDisplayNamesUsesProfileAndUsernameFallback(t *testing.T) {
	presences := []runtime.Presence{
		testPresence{userID: "user-1", sessionID: "session-1", username: "username-1"},
		testPresence{userID: "user-2", sessionID: "session-2", username: "username-2"},
	}
	lookup := testUserLookup{users: []*api.User{
		{Id: "user-1", DisplayName: "Commander One"},
		{Id: "user-2"},
	}}

	displayNames, err := resolveDisplayNames(nil, lookup, presences)
	if err != nil {
		t.Fatal(err)
	}
	if displayNames["user-1"] != "Commander One" {
		t.Fatalf("expected profile display name, got %q", displayNames["user-1"])
	}
	if displayNames["user-2"] != "username-2" {
		t.Fatalf("expected username fallback, got %q", displayNames["user-2"])
	}
}

func TestMatchLoopIgnoresInvalidMessagesWithoutBroadcastingSnapshot(t *testing.T) {
	match := &Match{}
	dispatcher := &testDispatcher{}
	player := core.NewPlayer("user-1", "session-1", "Player One", core.Vector2{})
	state := &State{
		Mode:                DefaultMode,
		AllowJoinInProgress: true,
		Players:             map[string]*core.Player{player.SessionID: player},
		Reservations:        make(map[string]int64),
	}
	messages := []runtime.MatchData{
		testMatchData{testPresence: testPresence{userID: player.UserID, sessionID: player.SessionID}, opCode: 999, data: []byte(`{}`)},
		testMatchData{testPresence: testPresence{userID: player.UserID, sessionID: player.SessionID}, opCode: core.OpMovementInput, data: []byte{0x0a, 0x01}},
	}

	result := match.MatchLoop(nil, nil, nil, nil, dispatcher, 1, state, messages)
	if result == nil {
		t.Fatal("invalid messages must not stop the match")
	}
	if player.Position != (core.Vector2{}) {
		t.Fatalf("invalid messages changed position: %+v", player.Position)
	}
	if dispatcher.broadcastCount != 0 {
		t.Fatalf("match loop unexpectedly broadcast %d messages", dispatcher.broadcastCount)
	}
}

func TestMatchLoopAppliesMovementWithoutBroadcastingSnapshot(t *testing.T) {
	match := &Match{}
	dispatcher := &testDispatcher{}
	player := core.NewPlayer("user-1", "session-1", "Player One", core.Vector2{})
	state := &State{
		Mode:                DefaultMode,
		AllowJoinInProgress: true,
		Players:             map[string]*core.Player{player.SessionID: player},
		Reservations:        make(map[string]int64),
	}
	message := testMatchData{
		testPresence: testPresence{userID: player.UserID, sessionID: player.SessionID},
		opCode:       core.OpMovementInput,
		data:         mustMarshalMovementInput(t, &core.MovementInput{X: 1, Sequence: 1}),
	}

	match.MatchLoop(nil, nil, nil, nil, dispatcher, 1, state, []runtime.MatchData{message})
	if player.Position != (core.Vector2{X: 0.5}) {
		t.Fatalf("unexpected player position: %+v", player.Position)
	}

	if dispatcher.broadcastCount != 0 {
		t.Fatalf("movement unexpectedly broadcast %d state snapshots", dispatcher.broadcastCount)
	}
}

func mustMarshalMovementInput(t *testing.T, input *core.MovementInput) []byte {
	t.Helper()
	data, err := proto.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertStateSnapshot(t *testing.T, dispatcher *testDispatcher, tick int64, playerCount int) {
	t.Helper()
	if dispatcher.broadcastOpCode != core.OpStateSnapshot || !dispatcher.broadcastReliable {
		t.Fatalf("unexpected broadcast: opcode=%d reliable=%v", dispatcher.broadcastOpCode, dispatcher.broadcastReliable)
	}

	var snapshot core.StateSnapshot
	if err := json.Unmarshal(dispatcher.broadcastData, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Tick != tick || snapshot.PlayerCount != playerCount {
		t.Fatalf("unexpected state snapshot: %+v", snapshot)
	}
}

type testPresence struct {
	userID    string
	sessionID string
	username  string
}

func (p testPresence) GetHidden() bool      { return false }
func (p testPresence) GetPersistence() bool { return false }
func (p testPresence) GetUsername() string {
	if p.username != "" {
		return p.username
	}
	return p.userID
}
func (p testPresence) GetStatus() string                 { return "" }
func (p testPresence) GetReason() runtime.PresenceReason { return runtime.PresenceReasonUnknown }
func (p testPresence) GetUserId() string                 { return p.userID }
func (p testPresence) GetSessionId() string              { return p.sessionID }
func (p testPresence) GetNodeId() string                 { return "node-1" }

type testUserLookup struct {
	users []*api.User
}

func (l testUserLookup) UsersGetId(context.Context, []string, []string) ([]*api.User, error) {
	return l.users, nil
}

type testMatchData struct {
	testPresence
	opCode int64
	data   []byte
}

func (m testMatchData) GetOpCode() int64      { return m.opCode }
func (m testMatchData) GetData() []byte       { return m.data }
func (m testMatchData) GetReliable() bool     { return false }
func (m testMatchData) GetReceiveTime() int64 { return 0 }

type testDispatcher struct {
	label             string
	broadcastOpCode   int64
	broadcastData     []byte
	broadcastReliable bool
	broadcastCount    int
}

func (d *testDispatcher) BroadcastMessage(opCode int64, data []byte, _ []runtime.Presence, _ runtime.Presence, reliable bool) error {
	d.broadcastOpCode = opCode
	d.broadcastData = data
	d.broadcastReliable = reliable
	d.broadcastCount++
	return nil
}
func (d *testDispatcher) BroadcastMessageDeferred(int64, []byte, []runtime.Presence, runtime.Presence, bool) error {
	return nil
}
func (d *testDispatcher) MatchKick([]runtime.Presence) error { return nil }
func (d *testDispatcher) MatchLabelUpdate(label string) error {
	d.label = label
	return nil
}
