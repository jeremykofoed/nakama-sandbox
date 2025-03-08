package main

import (
	"sync"
)

// Attack types
type AttackType string
const ( //Building it this way avoids using string values on maps but allows the json to bear the string value.
	Jab AttackType = "jab"
	Punch AttackType = "punch"
	Kick AttackType = "kick"
	UpperCut AttackType = "uppercut"
	HeadButt AttackType = "headbutt"
	Bite AttackType = "bite"
	Scratch AttackType = "scratch"
)

// Information on single attack action.
type AttackInfo struct {
	Type AttackType `json:"type"`
	Damage float64 `json:"damage"` //Damage potential that could end up being less or none if say it were a glancing blow or parried.
	BaseHitChance float64 `json:"base_hit_chance"` //Hit chance on whether the attack connects or not to deal damage, or not.
	ApplicableStatusEffect []StatusEffectFromAttacks `json:"applicable_status_effects` //Effects that can be applied through attack actions.
}

type StatusEffectFromAttacks struct {
	Type StatusEffectType `json:"type"`
	Chance float64 `json:"chance"` //Chance to inflict.
}

// Registry to hold all of the definitions.  Using a mutex here since the data could be live-ops driven meaning it could change after nakama init.
// **NOTE: If the plan is to not update this information after nakama init then this paradigm can be change to a simple read-only map instead.
var AttackRegistry = struct {
	sync.RWMutex //Read/write mutex to help with concurrent access allowing mulitple readers or a single writer.
	Attacks map[AttackType]AttackInfo
}{
	Attacks: make(map[AttackType]AttackInfo),
}

// This function will initialize the Attack Registry.
// @JWK TODO: Move this information into storage instead of hard coded value to place it under the purvue of live-ops.
func InitAttackRegistry() {
	AttackRegistry.Lock() //Call lock on the mutex in preparation for writing.
	defer AttackRegistry.Unlock() //Don't forget to release the mutex lock.
	AttackRegistry.Attacks[Jab] = AttackInfo{
		Type: Jab,
		Damage: 2,
		BaseHitChance: 0.95,
		ApplicableStatusEffect: []StatusEffectFromAttacks{
			{
				Type: Dazed,
				Chance: 0.1,
			},
		},
	}
	AttackRegistry.Attacks[Punch] = AttackInfo{
		Type: Punch,
		Damage: 4,
		BaseHitChance: 0.9,
		ApplicableStatusEffect: []StatusEffectFromAttacks{
			{
				Type: Dazed,
				Chance: 0.2,
			},
			{
				Type: Bleed,
				Chance: 0.1,
			},
		},
	}
	AttackRegistry.Attacks[Kick] = AttackInfo{
		Type: Kick,
		Damage: 7,
		BaseHitChance: 0.75,
		ApplicableStatusEffect: []StatusEffectFromAttacks{
			{
				Type: Poison, //Dirty feet/shoes applying poison???
				Chance: 0.4,
			},
		},
	}
	AttackRegistry.Attacks[UpperCut] = AttackInfo{
		Type: UpperCut,
		Damage: 10,
		BaseHitChance: 0.5,
		ApplicableStatusEffect: []StatusEffectFromAttacks{},
	}
	AttackRegistry.Attacks[HeadButt] = AttackInfo{
		Type: HeadButt,
		Damage: 12,
		BaseHitChance: 0.35,
		ApplicableStatusEffect: []StatusEffectFromAttacks{
			{
				Type: Dazed,
				Chance: 0.9,
			},
			{
				Type: Bleed,
				Chance: 0.7,
			},
		},
	}
	AttackRegistry.Attacks[Bite] = AttackInfo{
		Type: Bite,
		Damage: 5,
		BaseHitChance: 0.9,
		ApplicableStatusEffect: []StatusEffectFromAttacks{
			{
				Type: Bleed,
				Chance: 0.8,
			},
		},
	}
	AttackRegistry.Attacks[Scratch] = AttackInfo{
		Type: Scratch,
		Damage: 4,
		BaseHitChance: 0.95,
		ApplicableStatusEffect: []StatusEffectFromAttacks{
			{
				Type: Poison,
				Chance: 0.3,
			},
		},
	}
}

//
func (p *Player) DoAttack(target string, attack AttackInfo) {
	//@JWK TODO: Implement target.
	//@JWK TODO: RNG for success.
	//@JWK TODO: Store battle data updates keeping track of health, status', etc.
	//@JWK TODO: Check health of each player and return if anyone is dead.
	//@JWK TODO: Bonus, implement battle log.
}

// This function calculates the hit chance of each attack performed.
func (p *Player) GetAttackHitChance(baseHitChance float64) float64 {
	return 0
}