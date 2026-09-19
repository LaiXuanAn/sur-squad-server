package system

import (
	"testing"

	corecombat "squad-survival-be/modules/game/core/combat"
	"squad-survival-be/modules/game/core/entity"

	"google.golang.org/protobuf/proto"
)

func TestEncodeCombatEventBatch(t *testing.T) {
	events := []corecombat.Event{
		{
			Type: corecombat.EventAttackStarted, AttackID: "a:1:1",
			AttackerUserID: "a", AttackerCharacterID: "a:1", TargetUserID: "b", TargetCharacterID: "b:1",
			WeaponType: entity.WeaponSword, StartTick: 10, ImpactTick: 12, CompleteTick: 15,
		},
		{
			Type: corecombat.EventDamageApplied, AttackID: "a:1:1",
			AttackerUserID: "a", AttackerCharacterID: "a:1", TargetUserID: "b", TargetCharacterID: "b:1",
			Damage: 11, RemainingHealth: 89, Tick: 12,
		},
		{
			Type: corecombat.EventCharacterDied, AttackID: "a:1:1",
			AttackerUserID: "a", AttackerCharacterID: "a:1", TargetUserID: "b", TargetCharacterID: "b:1", Tick: 12,
		},
	}

	data, err := EncodeCombatEventBatch(12, events)
	if err != nil {
		t.Fatal(err)
	}
	var batch CombatEventBatch
	if err = proto.Unmarshal(data, &batch); err != nil {
		t.Fatal(err)
	}
	if batch.Tick != 12 || len(batch.Events) != 3 {
		t.Fatalf("unexpected combat batch: %+v", &batch)
	}
	if started := batch.Events[0].GetAttackStarted(); started == nil || started.AttackId != "a:1:1" || started.ImpactTick != 12 {
		t.Fatalf("unexpected attack started event: %+v", started)
	}
	if damage := batch.Events[1].GetDamageApplied(); damage == nil || damage.Damage != 11 || damage.RemainingHealth != 89 {
		t.Fatalf("unexpected damage event: %+v", damage)
	}
	if died := batch.Events[2].GetCharacterDied(); died == nil || died.KillerCharacterId != "a:1" || died.TargetCharacterId != "b:1" {
		t.Fatalf("unexpected death event: %+v", died)
	}
}

func TestEncodeCombatEventBatchRejectsUnknownEvent(t *testing.T) {
	if _, err := EncodeCombatEventBatch(1, []corecombat.Event{{}}); err == nil {
		t.Fatal("expected unknown combat event to fail encoding")
	}
}
