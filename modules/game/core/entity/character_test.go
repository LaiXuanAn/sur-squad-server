package entity

import (
	"math/rand"
	"testing"
)

func TestNewCharacterUsesDefaultStats(t *testing.T) {
	character := NewCharacter()

	if character.Health != 100 || character.MaxHealth != 100 || character.Damage != 10 || character.MoveSpeed != 5 {
		t.Fatalf("unexpected base stats: %+v", character)
	}
	if character.AttackSpeed != 1.2 || character.AttackRange != 0 {
		t.Fatalf("unexpected attack stats: %+v", character)
	}
	if character.RegenRate != 0 || character.DamageRatio != 1.1 {
		t.Fatalf("unexpected recovery or damage ratio stats: %+v", character)
	}
}

func TestCharacterDamageRange(t *testing.T) {
	character := Character{Damage: 10, DamageRatio: 1.1}
	minimum, maximum := character.DamageRange()

	if !almostEqual(minimum, 9) || !almostEqual(maximum, 11) {
		t.Fatalf("expected damage range 9..11, got %f..%f", minimum, maximum)
	}
}

func TestCharacterDamageRangeDefaultsToFixedDamage(t *testing.T) {
	character := Character{Damage: 10}
	minimum, maximum := character.DamageRange()

	if minimum != 10 || maximum != 10 {
		t.Fatalf("expected fixed damage 10, got %f..%f", minimum, maximum)
	}
}

func TestCharacterRollDamageStaysInsideRange(t *testing.T) {
	character := Character{Damage: 10, DamageRatio: 1.1}
	random := rand.New(rand.NewSource(1))
	seenDifferentDamage := false
	previous := character.RollDamage(random)

	for range 1000 {
		damage := character.RollDamage(random)
		if damage < 9 || damage > 11 {
			t.Fatalf("damage outside 9..11: %f", damage)
		}
		if damage != previous {
			seenDifferentDamage = true
		}
		previous = damage
	}
	if !seenDifferentDamage {
		t.Fatal("expected damage rolls to vary")
	}
}

func TestCharacterRollDamageWithoutRatioIsFixed(t *testing.T) {
	character := Character{Damage: 10}
	random := rand.New(rand.NewSource(1))

	if damage := character.RollDamage(random); damage != 10 {
		t.Fatalf("expected fixed damage 10, got %f", damage)
	}
}
