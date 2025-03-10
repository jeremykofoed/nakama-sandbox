package main

import (
	"fmt"
	"sync"

	"github.com/heroiclabs/nakama-common/runtime"
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
// @JWK TODO: Change Damange to a range like min, max to use in RNG Fx instead of static damage.
type AttackInfo struct {
	Type AttackType `json:"type"`
	Damage int `json:"damage"` //Damage potential that could end up being less or none if say it were a glancing blow or parried.
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
func (p *Player) PlayerAttack(logger runtime.Logger, targetID string, attackRequest AttackType) error {
	//@JWK TODO: RNG for success.
	//@JWK TODO: Store battle data updates keeping track of health, status', etc.
	//@JWK TODO: Check health of each player and return if anyone is dead.
	//@JWK TODO: Add to battle stats when enemy is killed.
	//@JWK TODO: Bonus, implement battle log.
	
	//Look for the target
	targetEnemy := p.GetEnemy(targetID)
	if targetEnemy.Type == "" {
		return runtime.NewError(fmt.Sprintf("Enemy not found by supplied ID: %s", targetID), 5) //Not found
	}
	logger.Debug("Target found: %+v", targetEnemy)

	//Check attack.
	AttackRegistry.RLock() //Read lock.
	attackAction := AttackRegistry.Attacks[attackRequest]
	AttackRegistry.RUnlock() //Release read lock.
	//Did we find the attack?
	if attackAction.Type == "" {
		return runtime.NewError(fmt.Sprintf("Attack action not found: %s", attackRequest), 5) //Not found
	}
	logger.Debug("Attack action found: %+v", attackAction)

	//Check if anyone was dead befor attack action / status effects.
	if p.IsPlayerDead() == true {
		return runtime.NewError("Player is deceased.", 5) //Not found
	}
	if targetEnemy.IsEnemyDead() == true {
		return runtime.NewError("Enemy is deceased.", 5) //Not found
	}

	//Check for status effects that affect combat for the player.
	var hitChance float64 = attackAction.BaseHitChance
	logger.Debug("hitChance: %f", hitChance)
	for _, effect := range p.StatusEffects {
		//@JWK TODO: Make sure the effects aren't expired.
		if effect.Type == Dazed || effect.Type == Blind { //Status effects that affect hit chance.
			hitChance -= effect.Modifier
			logger.Debug("hitChance: %f", hitChance)
		}
	}

	//Perform attack.
	//@JWK TODO: Handle this!
	if ActionSuceeded(logger, hitChance) == true {
		logger.Debug("Performing attack: %+v", attackAction)
		// @JWK TODO: Change Damange to a range like min, max to use in RNG Fx instead of static damage.
		dmg := (attackAction.Damage) * -1 //Damage subtracts from pool, flip the sign.
		logger.Debug("Dmg: %d", dmg)
		//Adjust health.
		targetEnemy.EnemyHealth(dmg)
		logger.Debug("targetEnemy: %+v", targetEnemy)
		//Apply status effects if the attack lands.
		for _, effect := range attackAction.ApplicableStatusEffect {
			logger.Debug("Status effect: %+v", effect)
			if ActionSuceeded(logger, effect.Chance) == true {
				logger.Debug("Apply status effect: %+v", effect)
				//Add status effect.
				AddStatusEffect(logger, effect.Type, targetEnemy)
			}
		}
	}
	
	//Tick status effects.
	TickStatusEffect(logger, targetEnemy)
	TickStatusEffect(logger, p)

	//Check if anyone died after attack action / status effects.
	if p.IsPlayerDead() == true {
		//@JWK TODO: Handle this (Game is over?  Clear stats?  New character and fresh start?).
	}
	if targetEnemy.IsEnemyDead() == true {
		//@JWK TODO: Handle this (Get rewards, update stats, get new enemies, etc).
		//Update battle stats.
		logger.Debug("Enemy died, running clean up.")
		p.CleanUpSuccessfulBattle(logger, targetID)
	}

	//@JWK TODO: Once the player has done an attack, the enemy gets a turn to attack.

	return nil
}

//
func (p *Player) GetEnemy(targetID string) *Enemy {
	var enemy *Enemy
	//Get battle state.
	battleState := p.BattleState
	//Check for enemies.
	if battleState.Enemies == nil {
		return enemy
	}
	//Check for enemies.
	enemies := battleState.Enemies
	if len(enemies) <= 0 {
		return enemy
	}
	//Find the target.
	_, exists := enemies[targetID]
	if exists {
		return enemies[targetID]
	}
	return enemy
}

//
func (p *Player) IsPlayerDead() bool {
	return p.Health <= 0
}

//
func (e *Enemy) IsEnemyDead() bool {
	return e.Health <= 0
}

// This function will change health of the enemy.
func (e *Enemy) EnemyHealth(delta int) {
	e.Health += delta
}

// This function will perform an attack on the player.
func (e *Enemy) EnemyAttack() {
	return
}

// This function determines if an action succeeds.
func ActionSuceeded(logger runtime.Logger, hitChance float64) bool {
	if hitChance <= 0 { //Effects can drop it below
		return false
	}
	if hitChance >= 1 { //Effects can raise it above.
		return true
	}
	maxRange := 100
	//Scale change to RNG Fx rance.
	chanceThreshold := int(hitChance * float64(maxRange)) // Ex: hitChance:0.6 -> chanceThreshold:60.
	logger.Debug("chanceThreshold: %d", chanceThreshold)
	//Get RNG
	diceRoll := BattleDiceRoll(0, maxRange)
	logger.Debug("diceRoll: %d", diceRoll)
	success := diceRoll <= chanceThreshold
	logger.Debug("actionSuceeded: %t", success)
	return success
}

