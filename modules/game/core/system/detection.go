package system

import (
	"squad-survival-be/modules/game/core/entity"

	"google.golang.org/protobuf/proto"
)

func EncodePlayerDetectionSnapshot(tick int64, players []*entity.Player) ([]byte, error) {
	snapshot := &PlayerDetectionSnapshot{
		Tick:    tick,
		Players: make([]*DetectedPlayer, 0, len(players)),
	}
	for _, player := range players {
		if player == nil {
			continue
		}

		characters := make([]*CharacterSnapshot, 0, len(player.Characters))
		for _, character := range player.Characters {
			if character == nil {
				continue
			}
			characters = append(characters, &CharacterSnapshot{
				Health:      character.Health,
				MaxHealth:   character.MaxHealth,
				Damage:      character.Damage,
				MoveSpeed:   character.MoveSpeed,
				AttackSpeed: character.AttackSpeed,
				AttackRange: character.AttackRange,
				RegenRate:   character.RegenRate,
				DamageRatio: character.DamageRatio,
				WeaponType:  string(character.Weapon.Type),
				Position:    vectorSnapshot(character.Position),
			})
		}

		snapshot.Players = append(snapshot.Players, &DetectedPlayer{
			UserId:      player.UserID,
			SessionId:   player.SessionID,
			DisplayName: player.DisplayName,
			Position:    vectorSnapshot(player.Position),
			Facing:      vectorSnapshot(player.Facing),
			Direction:   vectorSnapshot(player.Direction),
			Characters:  characters,
		})
	}
	return proto.Marshal(snapshot)
}

func vectorSnapshot(vector entity.Vector2) *Vector2 {
	return &Vector2{X: vector.X, Y: vector.Y}
}
