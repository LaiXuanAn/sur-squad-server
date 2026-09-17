package core

import (
	"math"
	"math/rand"
)

// Thiết lập map
const (
	MoveSpeed         = 5.0
	PlayAreaRadius    = 10.0
	SpawnRadius       = 10.0
	InputTimeoutTicks = int64(3)
	TickRate          = 10
)

type Vector2 struct {
	X float64
	Y float64
}

type Player struct {
	UserID        string
	SessionID     string
	DisplayName   string
	Position      Vector2
	Facing        Vector2
	Direction     Vector2
	LastSequence  uint64
	HasSequence   bool
	LastInputTick int64
}

func NewPlayer(userID, sessionID, displayName string, position Vector2) *Player {
	return &Player{
		UserID:      userID,
		SessionID:   sessionID,
		DisplayName: displayName,
		Position:    ClampToPlayArea(position),
		Facing:      Vector2{X: 1},
	}
}

func ApplyMovementInput(player *Player, input *MovementInput, tick int64) bool {
	if player.HasSequence && input.Sequence <= player.LastSequence {
		return false
	}
	if !isFinite(input.X) || !isFinite(input.Y) {
		return false
	}

	direction := NormalizeDirection(Vector2{X: input.X, Y: input.Y})
	player.Direction = direction
	if direction.X != 0 || direction.Y != 0 {
		player.Facing = direction
	}
	player.LastSequence = input.Sequence
	player.HasSequence = true
	player.LastInputTick = tick
	return true
}

func StepMovement(player *Player, tick int64) {
	if player.HasSequence && tick-player.LastInputTick >= InputTimeoutTicks {
		player.Direction = Vector2{}
	}

	deltaSeconds := 1.0 / float64(TickRate)
	player.Position.X += player.Direction.X * MoveSpeed * deltaSeconds
	player.Position.Y += player.Direction.Y * MoveSpeed * deltaSeconds
	player.Position = ClampToPlayArea(player.Position)
}

func NormalizeDirection(direction Vector2) Vector2 {
	lengthSquared := direction.X*direction.X + direction.Y*direction.Y
	if lengthSquared == 0 || lengthSquared <= 1 {
		return direction
	}

	length := math.Sqrt(lengthSquared)
	return Vector2{X: direction.X / length, Y: direction.Y / length}
}

func ClampToPlayArea(position Vector2) Vector2 {
	distanceSquared := position.X*position.X + position.Y*position.Y
	radiusSquared := PlayAreaRadius * PlayAreaRadius
	if distanceSquared <= radiusSquared {
		return position
	}

	distance := math.Sqrt(distanceSquared)
	return Vector2{
		X: position.X / distance * PlayAreaRadius,
		Y: position.Y / distance * PlayAreaRadius,
	}
}

func RandomSpawn(random *rand.Rand) Vector2 {
	angle := random.Float64() * 2 * math.Pi
	radius := math.Sqrt(random.Float64()) * SpawnRadius
	return Vector2{
		X: math.Cos(angle) * radius,
		Y: math.Sin(angle) * radius,
	}
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
