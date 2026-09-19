package system

import (
	"squad-survival-be/modules/game/core/combat"
	"squad-survival-be/modules/game/core/entity"

	"google.golang.org/protobuf/proto"
)

func EncodePlayerDetectionSnapshot(tick int64, self *entity.Player, players []*entity.Player, projectiles []*combat.Projectile) ([]byte, error) {
	snapshot := &PlayerDetectionSnapshot{
		Tick:        tick,
		Self:        playerSnapshot(self),
		Players:     make([]*DetectedPlayer, 0, len(players)),
		Projectiles: make([]*ProjectileSnapshot, 0, len(projectiles)),
	}
	for _, projectile := range projectiles {
		if projectile == nil {
			continue
		}
		snapshot.Projectiles = append(snapshot.Projectiles, &ProjectileSnapshot{
			ProjectileId: projectile.ID, AttackId: projectile.AttackID,
			AttackerUserId: projectile.AttackerUserID, AttackerCharacterId: projectile.AttackerCharacterID,
			TargetUserId: projectile.TargetUserID, TargetCharacterId: projectile.TargetCharacterID,
			WeaponType: string(projectile.WeaponType), WeaponName: projectile.WeaponName,
			Position: vectorSnapshot(projectile.Position), Direction: vectorSnapshot(projectile.Direction), Speed: projectile.Speed,
		})
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
			WeaponName:  character.Weapon.Name,
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
