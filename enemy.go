package main

import (
	"fmt"
	"sync"
	"context"
	"encoding/json"
	"github.com/heroiclabs/nakama-common/runtime"
)

var enemyDataStorageKey = "enemies"

// Enemy types
type EnemyType string
const ( //Building it this way avoids using string values on maps but allows the json to bear the string value.
	Zombie EnemyType = "zombie"
	Mutant EnemyType = "mutant"
	Beast EnemyType = "beast"
)

// Enemy data structure.
type Enemy struct {
	Type EnemyType `json:"type"`
	Health int `json:"health"`
	AttackModifier float64 `json:"attack_modifier"` //This is used to adjust attack type damage values.
	StatusEffects []*StatusEffect `json:"status_effects"` //Used to store player state modifiers.
	Rewards []RewardInfo `json:"rewards"` //Rewards assigned at time of enemy selection.
}

// Registry to hold all of the definitions.  Using a mutex here since the data could be live-ops driven meaning it could change after nakama init.
// **NOTE: If the plan is to not update this information after nakama init then this paradigm can be change to a simple read-only map instead.
var EnemyRegistry = struct {
	sync.RWMutex //Read/write mutex to help with concurrent access allowing mulitple readers or a single writer.
	Enemies map[EnemyType]Enemy
}{
	Enemies: make(map[EnemyType]Enemy),
}

// This function will initialize the Enemy Registry.
func InitEnemyRegistry(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) error {
	//Read from the storage engine.
	rObj, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{
			Collection: configDataStorageCollection,
			Key: enemyDataStorageKey,
		},
	})
	if err != nil {
		logger.Error("Error getting enemy configuration data: %v", err)
		return err
	}
	//Load defaults if nothing was found in storage and save them into storage.
	if len(rObj) == 0 {
		EnemyRegistry.Lock()  //Call lock on the mutex in preparation for writing.
		EnemyRegistry.Enemies[Zombie] = Enemy{
			Type: Zombie,
			Health: 50,
			AttackModifier: 1.5,
			StatusEffects: []*StatusEffect{},
			Rewards: []RewardInfo{},
		}
		EnemyRegistry.Enemies[Mutant] = Enemy{
			Type: Mutant,
			Health: 75,
			AttackModifier: 1.1,
			StatusEffects: []*StatusEffect{},
			Rewards: []RewardInfo{},
		}
		EnemyRegistry.Enemies[Beast] = Enemy{
			Type: Beast,
			Health: 25,
			AttackModifier: 2,
			StatusEffects: []*StatusEffect{},
			Rewards: []RewardInfo{},
		}
		EnemyRegistry.Unlock() //Don't forget to release the mutex lock.
		return SaveEnemyRegistry(nk)
	}

	var enemies map[EnemyType]Enemy
	if err := json.Unmarshal([]byte(rObj[0].Value), &enemies); err != nil {
		logger.Error("Failed to unmarshal enemy data: %v", err)
		return err
	}
	EnemyRegistry.Lock()  //Call lock on the mutex in preparation for writing.
	EnemyRegistry.Enemies = enemies
	EnemyRegistry.Unlock() //Don't forget to release the mutex lock.

	return nil
}

// This function will save the Enemy Registry to storage.
func SaveEnemyRegistry(nk runtime.NakamaModule) error {
	EnemyRegistry.RLock() //Read lock.
	//Json-ify the enemy registry in prepartion for storage.
	data, err := json.Marshal(EnemyRegistry.Enemies)
	EnemyRegistry.RUnlock() //Don't forget to release the lock.
	if err != nil {
		return err
	}
	wObj := []*runtime.StorageWrite{
		{
			Collection: configDataStorageCollection,
			Key: enemyDataStorageKey,
			Value: string(data),
			PermissionRead: 1, // Owner and runtime can read.
			PermissionWrite: 0, // No one can write save the runtime.
		},
	}
	//Write to the storage engine.
	if _, err := nk.StorageWrite(context.Background(), wObj); err != nil {
		return fmt.Errorf("failed to write enemy data to storage: %v", err)
	}
	return nil
}

// Interface function to get health.
func (e *Enemy) GetHealth() int {
	return e.Health
}

// Interface function to set health.
func (e *Enemy) SetHealth(health int) {
	e.Health = health
}