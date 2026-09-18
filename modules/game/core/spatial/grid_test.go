package spatial

import (
	"errors"
	"math/rand"
	"testing"

	"squad-survival-be/modules/game/core/entity"
)

func TestQueryPlayersFiltersRadiusSelfAndSorts(t *testing.T) {
	grid := NewGrid(20)
	origin := player("origin", 0, 0)
	nearB := player("b", 3, 4)
	nearA := player("a", -3, -4)
	edge := player("edge", 10, 0)
	outside := player("outside", 10.01, 0)
	for _, candidate := range []*entity.Player{origin, nearB, nearA, edge, outside} {
		if err := grid.Insert(candidate); err != nil {
			t.Fatal(err)
		}
	}

	got := grid.QueryPlayers(origin)
	want := []string{"a", "b", "edge"}
	if len(got) != len(want) {
		t.Fatalf("expected %d players, got %d", len(want), len(got))
	}
	for index, sessionID := range want {
		if got[index].SessionID != sessionID {
			t.Fatalf("result %d: expected %q, got %q", index, sessionID, got[index].SessionID)
		}
	}
}

func TestQueryPlayersFindsPlayersAcrossNegativeCells(t *testing.T) {
	grid := NewGrid(20)
	left := player("left", -0.1, 0)
	right := player("right", 0.1, 0)
	if err := grid.Insert(left); err != nil {
		t.Fatal(err)
	}
	if err := grid.Insert(right); err != nil {
		t.Fatal(err)
	}

	got := grid.QueryPlayers(left)
	if len(got) != 1 || got[0] != right {
		t.Fatalf("unexpected players across negative cell boundary: %+v", got)
	}
}

func TestMoveUpdatesPlayerCell(t *testing.T) {
	grid := NewGrid(20)
	origin := player("origin", 0, 0)
	moving := player("moving", 30, 0)
	if err := grid.Insert(origin); err != nil {
		t.Fatal(err)
	}
	if err := grid.Insert(moving); err != nil {
		t.Fatal(err)
	}
	if got := grid.QueryPlayers(origin); len(got) != 0 {
		t.Fatalf("expected moving player outside radius, got %+v", got)
	}

	moving.Position = entity.Vector2{X: 5}
	if err := grid.Move(moving); err != nil {
		t.Fatal(err)
	}
	if got := grid.QueryPlayers(origin); len(got) != 1 || got[0] != moving {
		t.Fatalf("expected moved player in radius, got %+v", got)
	}
}

func TestRemoveAndLifecycleErrors(t *testing.T) {
	grid := NewGrid(20)
	tracked := player("tracked", 0, 0)
	if err := grid.Insert(tracked); err != nil {
		t.Fatal(err)
	}
	if err := grid.Insert(tracked); !errors.Is(err, ErrPlayerExists) {
		t.Fatalf("expected duplicate insert error, got %v", err)
	}
	if err := grid.Move(player("missing", 0, 0)); !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf("expected missing move error, got %v", err)
	}
	if !grid.Remove(tracked.SessionID) {
		t.Fatal("expected tracked player to be removed")
	}
	if grid.Remove(tracked.SessionID) {
		t.Fatal("expected second remove to return false")
	}
}

func player(sessionID string, x, y float64) *entity.Player {
	return entity.NewPlayer(sessionID, sessionID, sessionID, entity.Vector2{X: x, Y: y}, rand.New(rand.NewSource(1)))
}
