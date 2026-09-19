package combat

import (
	"math/rand"
	"testing"

	"squad-survival-be/modules/game/core/entity"
)

func TestMeleeLocksNearestTargetAndSchedulesAttack(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeMelee, entity.Vector2{}, 2, 0.5)
	far := combatPlayer("c", "c:1", entity.RangeMelee, entity.Vector2{X: 1.5}, 2, 0.5)
	near := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: 1}, 2, 0.5)

	events := Step(playerMap(attacker, far, near), 10, rand.New(rand.NewSource(1)))
	started := findEvent(events, EventAttackStarted, "a:1")
	if started == nil || started.TargetCharacterID != "b:1" {
		t.Fatalf("expected nearest target b:1, got %+v", started)
	}
	if started.StartTick != 10 || started.ImpactTick != 13 || started.CompleteTick != 15 {
		t.Fatalf("unexpected attack timing: %+v", started)
	}
}

func TestTargetTieBreaksByUserAndCharacterID(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeReach, entity.Vector2{}, 3, 0.5)
	targetB2 := combatPlayer("b", "b:2", entity.RangeMelee, entity.Vector2{X: 1}, 1, 0.5)
	targetB1 := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: -1}, 1, 0.5)

	events := Step(playerMap(attacker, targetB2, targetB1), 1, rand.New(rand.NewSource(1)))
	started := findEvent(events, EventAttackStarted, "a:1")
	if started == nil || started.TargetCharacterID != "b:1" {
		t.Fatalf("expected deterministic target b:1, got %+v", started)
	}
}

func TestRangedDoesNotAttack(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeRanged, entity.Vector2{}, 10, 0.5)
	target := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{}, 1, 0.5)

	events := Step(playerMap(attacker, target), 1, rand.New(rand.NewSource(1)))
	if event := findEvent(events, EventAttackStarted, "a:1"); event != nil {
		t.Fatalf("ranged character unexpectedly attacked: %+v", event)
	}
}

func TestLeavingRangeCancelsAndResetsAttack(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeMelee, entity.Vector2{}, 2, 0.5)
	target := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: 1}, 2, 0.5)
	random := rand.New(rand.NewSource(1))
	players := playerMap(attacker, target)
	Step(players, 1, random)
	target.Characters[0].Position.X = 3

	events := Step(players, 2, random)
	if attacker.Characters[0].TargetCharacterID != "" || attacker.Characters[0].AttackStartTick != 0 {
		t.Fatalf("attack was not reset: %+v", attacker.Characters[0])
	}
	if event := findEvent(events, EventDamageApplied, "a:1"); event != nil {
		t.Fatalf("cancelled attack dealt damage: %+v", event)
	}
}

func TestTargetDeathBeforeImpactCancelsAttack(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeMelee, entity.Vector2{}, 2, 0.5)
	target := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: 1}, 2, 0.5)
	random := rand.New(rand.NewSource(1))
	players := playerMap(attacker, target)
	Step(players, 1, random)
	target.Characters[0].Health = 0

	events := Step(players, 2, random)
	if attacker.Characters[0].TargetCharacterID != "" {
		t.Fatalf("dead target remained locked: %+v", attacker.Characters[0])
	}
	if event := findEvent(events, EventDamageApplied, "a:1"); event != nil {
		t.Fatalf("dead target received damage: %+v", event)
	}
}

func TestSimultaneousImpactsCanKillBothCharacters(t *testing.T) {
	playerA := combatPlayer("a", "a:1", entity.RangeMelee, entity.Vector2{}, 2, 1)
	playerB := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: 1}, 2, 1)
	for _, player := range []*entity.Player{playerA, playerB} {
		character := player.Characters[0]
		character.Health = 5
		character.Damage = 10
		character.DamageRatio = 1
		character.AttackSpeed = 10
	}
	random := rand.New(rand.NewSource(1))
	players := playerMap(playerA, playerB)
	Step(players, 1, random)
	events := Step(players, 2, random)

	if len(playerA.Characters) != 0 || len(playerB.Characters) != 0 {
		t.Fatalf("expected both characters removed: a=%d b=%d", len(playerA.Characters), len(playerB.Characters))
	}
	if countEvents(events, EventDamageApplied) != 2 || countEvents(events, EventCharacterDied) != 2 {
		t.Fatalf("unexpected simultaneous combat events: %+v", events)
	}
}

func TestAttackRangeIncludesBoundary(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeMelee, entity.Vector2{}, 2, 0.5)
	target := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: 2}, 2, 0.5)

	events := Step(playerMap(attacker, target), 1, rand.New(rand.NewSource(1)))
	if findEvent(events, EventAttackStarted, "a:1") == nil {
		t.Fatal("expected target on attack boundary to be selected")
	}
}

func combatPlayer(userID, characterID string, rangeClass entity.RangeClass, position entity.Vector2, attackRange, impactRatio float64) *entity.Player {
	return &entity.Player{
		UserID: userID,
		Characters: []*entity.Character{{
			ID: characterID, RangeClass: rangeClass, Position: position,
			Health: 100, AttackSpeed: 2, AttackRange: attackRange, ImpactRatio: impactRatio,
			Damage: 10, DamageRatio: 1, Weapon: entity.Weapon{Type: entity.WeaponSword},
		}},
	}
}

func playerMap(players ...*entity.Player) map[string]*entity.Player {
	result := make(map[string]*entity.Player, len(players))
	for index, player := range players {
		result[player.UserID+string(rune(index))] = player
	}
	return result
}

func findEvent(events []Event, eventType EventType, attackerID string) *Event {
	for index := range events {
		if events[index].Type == eventType && events[index].AttackerCharacterID == attackerID {
			return &events[index]
		}
	}
	return nil
}

func countEvents(events []Event, eventType EventType) int {
	count := 0
	for _, event := range events {
		if event.Type == eventType {
			count++
		}
	}
	return count
}
