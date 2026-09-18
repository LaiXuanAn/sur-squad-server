package entity

import (
	"math/rand"
	"testing"
)

func TestDefaultWeaponCatalogContainsAllWeaponTypes(t *testing.T) {
	weapons := DefaultWeaponCatalog()
	if len(weapons) != 8 {
		t.Fatalf("expected 8 weapons, got %d", len(weapons))
	}

	found := make(map[WeaponType]bool, len(weapons))
	for _, weapon := range weapons {
		found[weapon.Type] = true
		if !weapon.RangeClass.Valid() {
			t.Fatalf("weapon %q has invalid range class %q", weapon.Type, weapon.RangeClass)
		}
	}
	for _, weaponType := range []WeaponType{
		WeaponDagger, WeaponBow, WeaponStaff, WeaponSpear,
		WeaponSword, WeaponWand, WeaponAxe, WeaponBlunt,
	} {
		if !found[weaponType] {
			t.Fatalf("missing weapon type %q", weaponType)
		}
	}
}

func TestWeaponReplacesAllCharacterStats(t *testing.T) {
	weapon := Weapon{
		Type: WeaponAxe, RangeClass: RangeMelee, Health: 120, Damage: 15,
		MoveSpeed: 4, AttackSpeed: 0.8, AttackRange: 2,
		RegenRate: 0.5, DamageRatio: 1.2,
	}
	character := NewCharacter()
	character.ApplyWeapon(weapon)

	if character.Weapon != weapon {
		t.Fatalf("unexpected weapon: %+v", character.Weapon)
	}
	if character.RangeClass != RangeMelee {
		t.Fatalf("expected melee character, got %q", character.RangeClass)
	}
	if character.Health != 120 || character.MaxHealth != 120 || character.Damage != 15 || character.MoveSpeed != 4 ||
		character.AttackSpeed != 0.8 || character.AttackRange != 2 ||
		character.RegenRate != 0.5 || character.DamageRatio != 1.2 {
		t.Fatalf("weapon stats were not fully applied: %+v", character)
	}
}

func TestParseWeaponCatalogRejectsInvalidRangeClass(t *testing.T) {
	_, err := ParseWeaponCatalog([]byte(`{"weapons":[{"type":"sword","range_class":"short"}]}`))
	if err == nil {
		t.Fatal("expected invalid range class to be rejected")
	}
}

func TestCreateCharacterSelectsWeaponFromCatalog(t *testing.T) {
	weapons := DefaultWeaponCatalog()
	character := CreateCharacter(rand.New(rand.NewSource(1)), weapons)

	found := false
	for _, weapon := range weapons {
		if character.Weapon == weapon {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("character received unknown weapon: %+v", character.Weapon)
	}
}
