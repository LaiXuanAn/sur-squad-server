package world

import (
	"math"
	"math/rand"
	"testing"
)

func TestClampToPlayAreaLeavesInsidePositionUnchanged(t *testing.T) {
	position := Vector2{X: 10, Y: -20}
	if got := ClampToPlayArea(position); got != position {
		t.Fatalf("expected position unchanged, got %+v", got)
	}
}

func TestClampToPlayAreaClampsToCircularBoundary(t *testing.T) {
	position := ClampToPlayArea(Vector2{X: PlayAreaRadius + 10})
	if math.Hypot(position.X, position.Y) != PlayAreaRadius {
		t.Fatalf("expected position on play-area edge, got %+v", position)
	}
}

func TestRandomSpawnStaysInsideSpawnArea(t *testing.T) {
	random := rand.New(rand.NewSource(1))
	for range 1000 {
		position := RandomSpawn(random)
		if math.Hypot(position.X, position.Y) > SpawnRadius {
			t.Fatalf("spawn outside radius: %+v", position)
		}
	}
}
