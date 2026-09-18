package entity

import (
	"math"
	"math/rand"
	"testing"
)

func TestDiagonalMovementIsNormalized(t *testing.T) {
	player := NewPlayer("user-1", "session-1", "Player One", Vector2{})
	if !ApplyMovementInput(player, &MovementInput{X: 1, Y: 1, Sequence: 1}, 0) {
		t.Fatal("expected movement input to be accepted")
	}

	StepMovement(player, 0)
	if !almostEqual(math.Hypot(player.Position.X, player.Position.Y), 0.5) {
		t.Fatalf("expected 0.5 units per tick, got %+v", player.Position)
	}
	if !almostEqual(player.Facing.X, math.Sqrt(0.5)) || !almostEqual(player.Facing.Y, math.Sqrt(0.5)) {
		t.Fatalf("unexpected facing: %+v", player.Facing)
	}
}

func TestZeroInputStopsWithoutChangingFacing(t *testing.T) {
	player := NewPlayer("user-1", "session-1", "Player One", Vector2{})
	ApplyMovementInput(player, &MovementInput{Y: 1, Sequence: 1}, 0)
	ApplyMovementInput(player, &MovementInput{Sequence: 2}, 1)
	StepMovement(player, 1)

	if player.Position != (Vector2{}) {
		t.Fatalf("expected player to remain still, got %+v", player.Position)
	}
	if player.Facing != (Vector2{Y: 1}) {
		t.Fatalf("expected facing to remain unchanged, got %+v", player.Facing)
	}
}

func TestStaleSequenceIsIgnored(t *testing.T) {
	player := NewPlayer("user-1", "session-1", "Player One", Vector2{})
	ApplyMovementInput(player, &MovementInput{X: 1, Sequence: 5}, 0)

	if ApplyMovementInput(player, &MovementInput{Y: 1, Sequence: 5}, 1) {
		t.Fatal("expected duplicate sequence to be ignored")
	}
	if ApplyMovementInput(player, &MovementInput{Y: 1, Sequence: 4}, 1) {
		t.Fatal("expected stale sequence to be ignored")
	}
	if player.Direction != (Vector2{X: 1}) {
		t.Fatalf("stale input changed direction: %+v", player.Direction)
	}
}

func TestInputTimesOutAfterThreeTicks(t *testing.T) {
	player := NewPlayer("user-1", "session-1", "Player One", Vector2{})
	ApplyMovementInput(player, &MovementInput{X: 1, Sequence: 1}, 0)

	for tick := int64(0); tick <= InputTimeoutTicks; tick++ {
		StepMovement(player, tick)
	}
	if !almostEqual(player.Position.X, 1.5) {
		t.Fatalf("expected movement for three ticks, got x=%f", player.Position.X)
	}
	if player.Direction != (Vector2{}) {
		t.Fatalf("expected timed out input to stop, got %+v", player.Direction)
	}
}

func TestMovementCannotLeavePlayArea(t *testing.T) {
	player := NewPlayer("user-1", "session-1", "Player One", Vector2{X: 9.9})
	ApplyMovementInput(player, &MovementInput{X: 1, Sequence: 1}, 0)
	StepMovement(player, 0)

	if !almostEqual(math.Hypot(player.Position.X, player.Position.Y), PlayAreaRadius) {
		t.Fatalf("expected player on play-area edge, got %+v", player.Position)
	}
}

func TestRandomSpawnStaysInsidePlayArea(t *testing.T) {
	random := rand.New(rand.NewSource(1))
	for i := 0; i < 1000; i++ {
		position := RandomSpawn(random)
		if math.Hypot(position.X, position.Y) > SpawnRadius {
			t.Fatalf("spawn outside radius: %+v", position)
		}
	}
}

func almostEqual(left, right float64) bool {
	return math.Abs(left-right) < 1e-9
}
