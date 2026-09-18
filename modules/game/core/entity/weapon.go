package entity

import (
	_ "embed"
	"encoding/json"
	"errors"
	"math/rand"
)

type WeaponType string

type RangeClass string

const (
	WeaponDagger WeaponType = "dagger"
	WeaponBow    WeaponType = "bow"
	WeaponStaff  WeaponType = "staff"
	WeaponSpear  WeaponType = "spear"
	WeaponSword  WeaponType = "sword"
	WeaponWand   WeaponType = "wand"
	WeaponAxe    WeaponType = "axe"
	WeaponBlunt  WeaponType = "blunt"
)

const (
	RangeMelee  RangeClass = "melee"
	RangeMedium RangeClass = "medium"
	RangeRanged RangeClass = "ranged"
)

type Weapon struct {
	Type        WeaponType `json:"type"`
	RangeClass  RangeClass `json:"range_class"`
	Health      float64    `json:"health"`
	Damage      float64    `json:"damage"`
	MoveSpeed   float64    `json:"move_speed"` // World units per second.
	AttackSpeed float64    `json:"attack_speed"`
	AttackRange float64    `json:"attack_range"`
	RegenRate   float64    `json:"regen_rate"`
	DamageRatio float64    `json:"damage_ratio"`
}

type WeaponCatalog struct {
	Weapons []Weapon `json:"weapons"`
}

//go:embed weapons.json
var defaultWeaponJSON []byte

func ParseWeaponCatalog(data []byte) ([]Weapon, error) {
	var catalog WeaponCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, err
	}
	if len(catalog.Weapons) == 0 {
		return nil, errors.New("weapon catalog has no weapons")
	}
	for _, weapon := range catalog.Weapons {
		if weapon.Type == "" {
			return nil, errors.New("weapon type is required")
		}
		if !weapon.RangeClass.Valid() {
			return nil, errors.New("weapon range class must be melee, medium, or ranged")
		}
	}
	return catalog.Weapons, nil
}

func (r RangeClass) Valid() bool {
	return r == RangeMelee || r == RangeMedium || r == RangeRanged
}

func DefaultWeaponCatalog() []Weapon {
	weapons, err := ParseWeaponCatalog(defaultWeaponJSON)
	if err != nil {
		panic("invalid embedded weapon catalog: " + err.Error())
	}
	return weapons
}

func RandomWeapon(random *rand.Rand, weapons []Weapon) Weapon {
	if len(weapons) == 0 {
		weapons = DefaultWeaponCatalog()
	}
	return weapons[random.Intn(len(weapons))]
}
