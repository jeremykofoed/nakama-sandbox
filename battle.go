package main

import (
	"math/rand"
	"time"
	"fmt"
	"context"

	"github.com/heroiclabs/nakama-common/runtime"
)

// Battle data structure.
type BattleState struct {
	Enemies []Enemy `json:"enemies"` //Plan for more than one possible target.
	//@JWK What else is needed???
}

//This function will attempt to get an on-going battle or create one.
func (p *Player) LoadBattleState(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) error {
	if p.BattleState.Enemies != nil {
		if len(p.BattleState.Enemies) > 0 {
			return nil
		}
	}
	err := p.CreateBattle(ctx, logger, nk)
	if err != nil {
		return  err
	}
	return nil
}

// This will setup a battle.
func (p *Player) CreateBattle(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) error {
	enemy, exists := GetEnemy()
	if !exists {
		return fmt.Errorf("unable to get an enemy, scope out of bounds possibly.")
	}
	enemy.ID = UtilMakeUUID()
	enemy.Rewards = GetRewards()
	var enemies []Enemy
	enemies = append(enemies, enemy)
	p.BattleState.Enemies = enemies
	return nil
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

// 
func GetRewards() []RewardInfo {
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