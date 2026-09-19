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

	events := NewSimulation().Step(playerMap(attacker, far, near), 10, rand.New(rand.NewSource(1)))
	started := findEvent(events, EventAttackStarted, "a:1")
	if started == nil || started.TargetCharacterID != "b:1" {
		t.Fatalf("expected nearest target b:1, got %+v", started)
	}
	if started.StartTick != 10 || started.ImpactTick != 13 || started.CompleteTick != 15 {
		t.Fatalf("unexpected attack timing: %+v", started)
	}
}

func TestTargetTieBreaksByUserAndCharacterID(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeMelee, entity.Vector2{}, 3, 0.5)
	targetB2 := combatPlayer("b", "b:2", entity.RangeMelee, entity.Vector2{X: 1}, 1, 0.5)
	targetB1 := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: -1}, 1, 0.5)

	events := NewSimulation().Step(playerMap(attacker, targetB2, targetB1), 1, rand.New(rand.NewSource(1)))
	started := findEvent(events, EventAttackStarted, "a:1")
	if started == nil || started.TargetCharacterID != "b:1" {
		t.Fatalf("expected deterministic target b:1, got %+v", started)
	}
}

func TestRangedSpawnsAndHitsWithProjectile(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeRanged, entity.Vector2{}, 10, 0.5)
	target := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: 2}, 1, 0.5)
	attacker.Characters[0].AttackSpeed = 5
	attacker.Characters[0].Weapon.ProjectileSpeed = 20
	attacker.Characters[0].Weapon.Name = "bow"
	attacker.Characters[0].Weapon.Type = entity.WeaponBow
	random := rand.New(rand.NewSource(1))
	players := playerMap(attacker, target)
	simulation := NewSimulation()

	if event := findEvent(simulation.Step(players, 1, random), EventAttackStarted, "a:1"); event == nil {
		t.Fatal("expected ranged attack to start")
	}
	spawnEvents := simulation.Step(players, 2, random)
	if event := findEvent(spawnEvents, EventProjectileSpawned, "a:1"); event == nil {
		t.Fatalf("expected projectile spawn, got %+v", spawnEvents)
	}
	if len(simulation.Projectiles()) != 1 {
		t.Fatalf("expected active projectile, got %+v", simulation.Projectiles())
	}
	hitEvents := simulation.Step(players, 3, random)
	if findEvent(hitEvents, EventProjectileHit, "a:1") == nil || findEvent(hitEvents, EventDamageApplied, "a:1") == nil {
		t.Fatalf("expected projectile hit and damage, got %+v", hitEvents)
	}
	if len(simulation.Projectiles()) != 0 || target.Characters[0].Health >= 100 {
		t.Fatalf("projectile did not resolve: projectiles=%+v health=%f", simulation.Projectiles(), target.Characters[0].Health)
	}
}

func TestProjectileExpiresWhenTargetDiesBeforeHit(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeRanged, entity.Vector2{}, 10, 0.5)
	target := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: 8}, 1, 0.5)
	attacker.Characters[0].AttackSpeed = 5
	attacker.Characters[0].Weapon = entity.Weapon{Type: entity.WeaponBow, Name: "bow", ProjectileSpeed: 10}
	random := rand.New(rand.NewSource(1))
	players := playerMap(attacker, target)
	simulation := NewSimulation()
	simulation.Step(players, 1, random)
	simulation.Step(players, 2, random)
	target.Characters[0].Health = 0

	events := simulation.Step(players, 3, random)
	if findEvent(events, EventProjectileExpired, "a:1") == nil {
		t.Fatalf("expected projectile expired event, got %+v", events)
	}
	if len(simulation.Projectiles()) != 0 {
		t.Fatalf("expired projectile remained active: %+v", simulation.Projectiles())
	}
}

func TestLeavingRangeCancelsAndResetsAttack(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeMelee, entity.Vector2{}, 2, 0.5)
	target := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: 1}, 2, 0.5)
	random := rand.New(rand.NewSource(1))
	players := playerMap(attacker, target)
	simulation := NewSimulation()
	simulation.Step(players, 1, random)
	target.Characters[0].Position.X = 3

	events := simulation.Step(players, 2, random)
	if attacker.Characters[0].TargetCharacterID != "" || attacker.Characters[0].AttackStartTick != 0 {
		t.Fatalf("attack was not reset: %+v", attacker.Characters[0])
	}
	if event := findEvent(events, EventDamageApplied, "a:1"); event != nil {
		t.Fatalf("cancelled attack dealt damage: %+v", event)
	}
}

func TestMovementInputControlsAttackEligibility(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeMelee, entity.Vector2{}, 2, 0.5)
	target := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: 1}, 2, 0.5)
	random := rand.New(rand.NewSource(1))
	players := playerMap(attacker, target)
	simulation := NewSimulation()

	attacker.Direction = entity.Vector2{X: 0.19, Y: -0.19}
	if findEvent(simulation.Step(players, 1, random), EventAttackStarted, "a:1") == nil {
		t.Fatal("expected input below movement threshold to allow attack")
	}
	attacker.Direction = entity.Vector2{X: 0.2}
	events := simulation.Step(players, 2, random)
	if attacker.Characters[0].TargetCharacterID != "" {
		t.Fatalf("movement did not reset active attack: %+v", attacker.Characters[0])
	}
	if findEvent(events, EventDamageApplied, "a:1") != nil {
		t.Fatalf("movement above threshold allowed damage: %+v", events)
	}
	if findEvent(events, EventAttackStarted, "a:1") != nil {
		t.Fatalf("movement above threshold started another attack: %+v", events)
	}
}

func TestTargetDeathBeforeImpactCancelsAttack(t *testing.T) {
	attacker := combatPlayer("a", "a:1", entity.RangeMelee, entity.Vector2{}, 2, 0.5)
	target := combatPlayer("b", "b:1", entity.RangeMelee, entity.Vector2{X: 1}, 2, 0.5)
	random := rand.New(rand.NewSource(1))
	players := playerMap(attacker, target)
	simulation := NewSimulation()
	simulation.Step(players, 1, random)
	target.Characters[0].Health = 0

	events := simulation.Step(players, 2, random)
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
	simulation := NewSimulation()
	simulation.Step(players, 1, random)
	events := simulation.Step(players, 2, random)

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

	events := NewSimulation().Step(playerMap(attacker, target), 1, rand.New(rand.NewSource(1)))
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
			Damage: 10, DamageRatio: 1, Weapon: entity.Weapon{Type: entity.WeaponSword, Name: "sword"},
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
