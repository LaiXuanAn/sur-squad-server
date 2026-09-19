package system

import (
	"math/rand"
	"testing"

	"squad-survival-be/modules/game/core/entity"

	"google.golang.org/protobuf/proto"
)

func TestEncodePlayerDetectionSnapshot(t *testing.T) {
	player := entity.NewPlayer(
		"user-1",
		"session-1",
		"Player One",
		entity.Vector2{X: 1, Y: 2},
		rand.New(rand.NewSource(1)),
	)
	player.Facing = entity.Vector2{X: 0, Y: 1}
	player.Direction = entity.Vector2{X: -1, Y: 0}
	player.Characters = append(player.Characters, nil)

	data, err := EncodePlayerDetectionSnapshot(42, player, []*entity.Player{player, nil})
	if err != nil {
		t.Fatal(err)
	}
	var snapshot PlayerDetectionSnapshot
	if err = proto.Unmarshal(data, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Tick != 42 || len(snapshot.Players) != 1 {
		t.Fatalf("unexpected snapshot envelope: %+v", &snapshot)
	}
	if snapshot.Self == nil || snapshot.Self.SessionId != player.SessionID {
		t.Fatalf("unexpected self snapshot: %+v", snapshot.Self)
	}

	detected := snapshot.Players[0]
	if detected.UserId != player.UserID || detected.SessionId != player.SessionID || detected.DisplayName != player.DisplayName {
		t.Fatalf("unexpected player identity: %+v", detected)
	}
	if detected.Position.X != 1 || detected.Position.Y != 2 || detected.Facing.Y != 1 || detected.Direction.X != -1 {
		t.Fatalf("unexpected vectors: %+v", detected)
	}
	if len(detected.Characters) != 1 {
		t.Fatalf("expected nil character to be skipped, got %d characters", len(detected.Characters))
	}
	character := detected.Characters[0]
	source := player.Characters[0]
	if character.CharacterId != source.ID || character.RangeClass != string(source.RangeClass) {
		t.Fatalf("unexpected character identity: %+v", character)
	}
	if character.Health != source.Health || character.MaxHealth != source.MaxHealth || character.Damage != source.Damage ||
		character.MoveSpeed != source.MoveSpeed || character.AttackSpeed != source.AttackSpeed ||
		character.AttackRange != source.AttackRange || character.RegenRate != source.RegenRate ||
		character.DamageRatio != source.DamageRatio || character.WeaponType != string(source.Weapon.Type) {
		t.Fatalf("unexpected character snapshot: %+v", character)
	}
	if character.Position.X != source.Position.X || character.Position.Y != source.Position.Y {
		t.Fatalf("unexpected character position: %+v", character.Position)
	}
}

func TestEncodeEmptyPlayerDetectionSnapshot(t *testing.T) {
	data, err := EncodePlayerDetectionSnapshot(7, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot PlayerDetectionSnapshot
	if err = proto.Unmarshal(data, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Tick != 7 || snapshot.Self != nil || len(snapshot.Players) != 0 {
		t.Fatalf("unexpected empty snapshot: %+v", &snapshot)
	}
}
