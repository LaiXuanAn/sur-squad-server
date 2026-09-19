package combat

import (
	"math"
	"math/rand"
	"sort"
	"strconv"

	"squad-survival-be/modules/game/core/entity"
)

type EventType uint8

const (
	EventAttackStarted EventType = iota + 1
	EventDamageApplied
	EventCharacterDied
)

type Event struct {
	Type                EventType
	AttackID            string
	AttackerUserID      string
	AttackerCharacterID string
	TargetUserID        string
	TargetCharacterID   string
	WeaponType          entity.WeaponType
	StartTick           int64
	ImpactTick          int64
	CompleteTick        int64
	Tick                int64
	Damage              float64
	RemainingHealth     float64
}

type ownedCharacter struct {
	owner     *entity.Player
	character *entity.Character
}

type damageIntent struct {
	event  Event
	target *entity.Character
}

func Step(players map[string]*entity.Player, tick int64, random *rand.Rand) []Event {
	if random == nil {
		random = rand.New(rand.NewSource(tick))
	}
	characters := sortedCharacters(players)
	lookup := characterLookup(characters)
	intents := make([]damageIntent, 0)

	for _, attacker := range characters {
		character := attacker.character
		if !canAttack(character) {
			character.ResetAttack()
			continue
		}
		if character.TargetCharacterID == "" {
			continue
		}

		target, ok := lookup[targetKey(character.TargetUserID, character.TargetCharacterID)]
		if !ok || !validTarget(attacker, target) || !withinRange(character, target.character) {
			character.ResetAttack()
			continue
		}

		attackID := attackID(character)
		if !character.AttackImpacted && tick >= character.AttackImpactTick {
			character.AttackImpacted = true
			intents = append(intents, damageIntent{
				target: target.character,
				event: Event{
					Type:                EventDamageApplied,
					AttackID:            attackID,
					AttackerUserID:      attacker.owner.UserID,
					AttackerCharacterID: character.ID,
					TargetUserID:        target.owner.UserID,
					TargetCharacterID:   target.character.ID,
					Tick:                tick,
					Damage:              character.RollDamage(random),
				},
			})
		}
		if tick >= character.AttackCompleteTick {
			character.ResetAttack()
		}
	}

	sort.Slice(intents, func(i, j int) bool {
		return intents[i].event.AttackerCharacterID < intents[j].event.AttackerCharacterID
	})
	events := applyDamage(intents, tick)

	for _, player := range players {
		player.RemoveDeadCharacters()
	}
	characters = sortedCharacters(players)
	lookup = characterLookup(characters)
	for _, attacker := range characters {
		character := attacker.character
		if character.TargetCharacterID != "" {
			if _, ok := lookup[targetKey(character.TargetUserID, character.TargetCharacterID)]; !ok {
				character.ResetAttack()
			}
		}
		if !canAttack(character) || character.TargetCharacterID != "" {
			continue
		}
		target, ok := nearestTarget(attacker, characters)
		if !ok {
			continue
		}
		events = append(events, startAttack(attacker, target, tick))
	}

	return events
}

func applyDamage(intents []damageIntent, tick int64) []Event {
	events := make([]Event, 0, len(intents)*2)
	dead := make(map[*entity.Character]bool)
	for _, intent := range intents {
		wasAlive := intent.target.Health > 0
		intent.target.Health -= intent.event.Damage
		intent.event.RemainingHealth = math.Max(0, intent.target.Health)
		events = append(events, intent.event)
		if wasAlive && intent.target.Health <= 0 && !dead[intent.target] {
			dead[intent.target] = true
			death := intent.event
			death.Type = EventCharacterDied
			death.Tick = tick
			death.Damage = 0
			death.RemainingHealth = 0
			events = append(events, death)
		}
	}
	return events
}

func startAttack(attacker, target ownedCharacter, tick int64) Event {
	character := attacker.character
	cycleTicks := int64(math.Ceil(float64(entity.TickRate) / character.AttackSpeed))
	if cycleTicks < 1 {
		cycleTicks = 1
	}
	impactOffset := int64(math.Ceil(float64(cycleTicks) * character.ImpactRatio))
	if impactOffset < 1 {
		impactOffset = 1
	}
	if impactOffset > cycleTicks {
		impactOffset = cycleTicks
	}

	character.AttackSequence++
	character.TargetUserID = target.owner.UserID
	character.TargetCharacterID = target.character.ID
	character.AttackStartTick = tick
	character.AttackImpactTick = tick + impactOffset
	character.AttackCompleteTick = tick + cycleTicks
	character.AttackImpacted = false

	return Event{
		Type:                EventAttackStarted,
		AttackID:            attackID(character),
		AttackerUserID:      attacker.owner.UserID,
		AttackerCharacterID: character.ID,
		TargetUserID:        target.owner.UserID,
		TargetCharacterID:   target.character.ID,
		WeaponType:          character.Weapon.Type,
		StartTick:           character.AttackStartTick,
		ImpactTick:          character.AttackImpactTick,
		CompleteTick:        character.AttackCompleteTick,
		Tick:                tick,
	}
}

func nearestTarget(attacker ownedCharacter, characters []ownedCharacter) (ownedCharacter, bool) {
	var selected ownedCharacter
	selectedDistance := math.Inf(1)
	found := false
	for _, candidate := range characters {
		if !validTarget(attacker, candidate) || !withinRange(attacker.character, candidate.character) {
			continue
		}
		distance := distanceSquared(attacker.character.Position, candidate.character.Position)
		if !found || distance < selectedDistance || distance == selectedDistance && targetLess(candidate, selected) {
			selected = candidate
			selectedDistance = distance
			found = true
		}
	}
	return selected, found
}

func canAttack(character *entity.Character) bool {
	if character == nil || character.ID == "" || character.Health <= 0 || character.AttackSpeed <= 0 || character.AttackRange < 0 || character.ImpactRatio <= 0 || character.ImpactRatio > 1 {
		return false
	}
	return character.RangeClass == entity.RangeMelee || character.RangeClass == entity.RangeReach
}

func validTarget(attacker, target ownedCharacter) bool {
	return target.character != nil && target.character.ID != "" && target.character.Health > 0 && attacker.owner != target.owner
}

func withinRange(attacker, target *entity.Character) bool {
	return distanceSquared(attacker.Position, target.Position) <= attacker.AttackRange*attacker.AttackRange
}

func distanceSquared(a, b entity.Vector2) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	return dx*dx + dy*dy
}

func targetLess(a, b ownedCharacter) bool {
	if a.owner.UserID == b.owner.UserID {
		return a.character.ID < b.character.ID
	}
	return a.owner.UserID < b.owner.UserID
}

func sortedCharacters(players map[string]*entity.Player) []ownedCharacter {
	owners := make([]*entity.Player, 0, len(players))
	for _, player := range players {
		if player != nil {
			owners = append(owners, player)
		}
	}
	sort.Slice(owners, func(i, j int) bool {
		if owners[i].UserID == owners[j].UserID {
			return owners[i].SessionID < owners[j].SessionID
		}
		return owners[i].UserID < owners[j].UserID
	})

	characters := make([]ownedCharacter, 0)
	for _, owner := range owners {
		for _, character := range owner.Characters {
			if character != nil {
				characters = append(characters, ownedCharacter{owner: owner, character: character})
			}
		}
	}
	sort.SliceStable(characters, func(i, j int) bool {
		if characters[i].owner.UserID == characters[j].owner.UserID {
			return characters[i].character.ID < characters[j].character.ID
		}
		return characters[i].owner.UserID < characters[j].owner.UserID
	})
	return characters
}

func characterLookup(characters []ownedCharacter) map[string]ownedCharacter {
	lookup := make(map[string]ownedCharacter, len(characters))
	for _, character := range characters {
		lookup[targetKey(character.owner.UserID, character.character.ID)] = character
	}
	return lookup
}

func targetKey(userID, characterID string) string {
	return userID + "\x00" + characterID
}

func attackID(character *entity.Character) string {
	return character.ID + ":" + strconv.FormatUint(character.AttackSequence, 10)
}
