package system

import (
	"fmt"

	corecombat "squad-survival-be/modules/game/core/combat"

	"google.golang.org/protobuf/proto"
)

func EncodeCombatEventBatch(tick int64, events []corecombat.Event) ([]byte, error) {
	batch := &CombatEventBatch{
		Tick:   tick,
		Events: make([]*CombatEvent, 0, len(events)),
	}
	for _, event := range events {
		encoded, err := combatEventSnapshot(event)
		if err != nil {
			return nil, err
		}
		batch.Events = append(batch.Events, encoded)
	}
	return proto.Marshal(batch)
}

func combatEventSnapshot(event corecombat.Event) (*CombatEvent, error) {
	switch event.Type {
	case corecombat.EventAttackStarted:
		return &CombatEvent{Event: &CombatEvent_AttackStarted{AttackStarted: &AttackStarted{
			AttackId: event.AttackID, AttackerUserId: event.AttackerUserID, AttackerCharacterId: event.AttackerCharacterID,
			TargetUserId: event.TargetUserID, TargetCharacterId: event.TargetCharacterID, WeaponType: string(event.WeaponType),
			StartTick: event.StartTick, ImpactTick: event.ImpactTick, CompleteTick: event.CompleteTick,
		}}}, nil
	case corecombat.EventDamageApplied:
		return &CombatEvent{Event: &CombatEvent_DamageApplied{DamageApplied: &DamageApplied{
			AttackId: event.AttackID, AttackerUserId: event.AttackerUserID, AttackerCharacterId: event.AttackerCharacterID,
			TargetUserId: event.TargetUserID, TargetCharacterId: event.TargetCharacterID,
			Damage: event.Damage, RemainingHealth: event.RemainingHealth, Tick: event.Tick,
		}}}, nil
	case corecombat.EventCharacterDied:
		return &CombatEvent{Event: &CombatEvent_CharacterDied{CharacterDied: &CharacterDied{
			AttackId: event.AttackID, KillerUserId: event.AttackerUserID, KillerCharacterId: event.AttackerCharacterID,
			TargetUserId: event.TargetUserID, TargetCharacterId: event.TargetCharacterID, Tick: event.Tick,
		}}}, nil
	default:
		return nil, fmt.Errorf("unknown combat event type %d", event.Type)
	}
}
