package main

import (
	"math/rand"
	"time"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

type EntityProcessor interface {
	GetStatusEffects() []*StatusEffect
	SetStatusEffects([]*StatusEffect)
	GetHealth() int
	SetHealth(int)
}

// Battle data structure.
type BattleState struct {
	Enemies map[string]*Enemy `json:"enemies"` //Plan for more than one possible target.
	//@JWK What else is needed???
}

//This function will attempt to get an on-going battle or create one.
func (p *Player) LoadBattleState() error {
	if p.BattleState.Enemies != nil {
		if len(p.BattleState.Enemies) > 0 {
			return nil
		}
	}
	err := p.CreateBattle()
	if err != nil {
		return  err
	}
	return nil
}

// This will setup a battle.
func (p *Player) CreateBattle() error {
	enemy, exists := GetEnemy()
	if !exists {
		return fmt.Errorf("unable to get an enemy, scope out of bounds possibly.")
	}
	id := UtilMakeUUID()
	enemy.Rewards = CreateRewards()
	enemies := make(map[string]*Enemy)
	enemies[id] = &enemy
	p.BattleState.Enemies = enemies
	return nil
}

// This function will manage cleaning up successful battle.
func (p *Player) CleanUpSuccessfulBattle(logger runtime.Logger, targetID string) {
	//@JWK TODO: Do rewarding.
	//@JWK TODO: Get more enemies if none??
	//Record stats.
	logger.Debug("Record stats as part of clean up.")
	targetEnemy := p.GetEnemy(targetID)
	if targetEnemy.Type == "" {
		logger.Error("Unable to find the enemy when expected.")
	} else {
		p.RecordBattleStats(targetEnemy.Type)
	}
	//Clear enemy from battle state.
	logger.Debug("Remove dead enemy from battle state as part of clean up.")
	delete(p.BattleState.Enemies, targetID)
}

// This function will manage stats of battles.  For now it'll just increment types of enemies killed.
// @JWK TODO: Add in deaths, total enemies killed, etc.
func (p *Player) RecordBattleStats(enemyType EnemyType) {
	battleStats := p.BattleStats
	if battleStats == nil {
		battleStats = make(map[EnemyType]int)
	}
	battleStats[enemyType]++
}

// Inclusive RNG dice roll.
func BattleDiceRoll(min, max int) int {
	//If the numbers are inversed flip them.
	if min > max {
		max, min = min, max
	}
	//If the numbers are the same, just return one.
	if min == max {
		return min
	}
	//See the RNG.
	rand.Seed(time.Now().UnixNano())
	//Grab a number inclusively!
	return rand.Intn(max-min+1) + min
}

// This function uses the registry to randomly get an enemy.
func GetEnemy() (Enemy, bool) {
	EnemyRegistry.RLock() //Read lock.
	defer EnemyRegistry.RUnlock() //Don't forget to release the lock.
	//Check to make sure we have values in the registry.
	l := len(EnemyRegistry.Enemies)
	if l == 0 {
		return Enemy{}, false
	}
	//Form a slice with the keys.
	keys := make([]EnemyType, 0, l)
	for key := range EnemyRegistry.Enemies {
		keys = append(keys, key)	
	}
	//Grab a RNG number and use that to pick a key from the slice.
	num := BattleDiceRoll(0,(l-1)) //L - 1 to convert to index based max number
	if num >= l { //Make sure not to exceed the l(len).
		num = l
	}
	//Select the key.
	sKey := keys[num]
	//Grab the random enemy.
	enemy, exists := EnemyRegistry.Enemies[sKey]
	return enemy, exists
}

// This function creates reward information
func CreateRewards() []RewardInfo {
	var rewards []RewardInfo
	//Experience
	reward := RewardInfo{
		Type: Experience,
		Amount: int64(BattleDiceRoll(0, 50) + 25),
	}
	rewards = append(rewards, reward)
	//Gold
	reward = RewardInfo{
		Type: Gold,
		Amount: int64(BattleDiceRoll(10, 100)),
	}
	rewards = append(rewards, reward)
	if BattleDiceRoll(0, 1) == 1 {
		//Gems
		reward = RewardInfo{
			Type: Gems,
			Amount: int64(BattleDiceRoll(0, 5)),
		}
		rewards = append(rewards, reward)
	}
	return rewards
}