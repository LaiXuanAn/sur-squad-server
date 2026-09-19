package system

import (
	"squad-survival-be/modules/game/core/entity"

	"google.golang.org/protobuf/proto"
)

func EncodePlayerDetectionSnapshot(tick int64, self *entity.Player, players []*entity.Player) ([]byte, error) {
	snapshot := &PlayerDetectionSnapshot{
		Tick:    tick,
		Self:    playerSnapshot(self),
		Players: make([]*DetectedPlayer, 0, len(players)),
	}
	for _, player := range players {
		if detected := playerSnapshot(player); detected != nil {
			snapshot.Players = append(snapshot.Players, detected)
		}
	}
	return proto.Marshal(snapshot)
}

func playerSnapshot(player *entity.Player) *DetectedPlayer {
	if player == nil {
		return nil
	}

	characters := make([]*CharacterSnapshot, 0, len(player.Characters))
	for _, character := range player.Characters {
		if character == nil {
			continue
		}
		characters = append(characters, &CharacterSnapshot{
			CharacterId: character.ID,
			RangeClass:  string(character.RangeClass),
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

	return &DetectedPlayer{
		UserId:      player.UserID,
		SessionId:   player.SessionID,
		DisplayName: player.DisplayName,
		Position:    vectorSnapshot(player.Position),
		Facing:      vectorSnapshot(player.Facing),
		Direction:   vectorSnapshot(player.Direction),
		Characters:  characters,
	}
}

func vectorSnapshot(vector entity.Vector2) *Vector2 {
	return &Vector2{X: vector.X, Y: vector.Y}
}
