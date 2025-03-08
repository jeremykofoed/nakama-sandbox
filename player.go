package main

import (
	"fmt"
	"context"
	"time"
	"encoding/json"
	"github.com/heroiclabs/nakama-common/runtime"
)

var playerDataStorageCollection = "playerData" //This is the collection name for all of the player related data.

// Player data structure.
type Player struct {
	ID string `json:"id"` //Nakama user id.
	DisplayName string `json:"display_name"`
	Level int `json:"level"`
	Experience int64 `json:"experience"`
	Health int `json:"health"`
	Currencies []Currency `json:"currency"` //Nakama supports a wallet that can be implemented at a later time.
	StatusEffects []StatusEffect `json:"status_effects"` //Used to store player state modifiers.
	BattleState map[string]interface{} `json:"battle_state"` //Used to store the battle game state.
	Attributes map[string]interface{} `json:"attributes"` //Key-Value map for addional data as needed.
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

// Used to setup the player data when one isn't found for the user in storage.
func NewPlayer(userID, displayerName string) *Player {
	return &Player{
		ID: userID,
		DisplayName: displayerName,
		Level: 1,
		Experience: 0,
		Health: 100,
		Currencies: []Currency{},
		StatusEffects: []StatusEffect{},
		BattleState: make(map[string]interface{}),
		Attributes: make(map[string]interface{}),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
}

// This function saves the player data to nakama storage.
// See https://heroiclabs.com/docs/nakama/concepts/storage/permissions/ for information on public read/write permissions or other storage information.
func SavePlayerData(nk runtime.NakamaModule, player *Player) error {
	//Json-ify the player struct in prepartion for storage.
	data, err := json.Marshal(player)
	if err != nil {
		return err
	}
	wObj := []*runtime.StorageWrite{
		{
			Collection: playerDataStorageCollection,
			Key: player.ID,
			Value: string(data),
			PermissionRead: 1, // Owner and runtime can read.
			PermissionWrite: 0, // No one can write save the runtime.
		},
	}
	//Write to the storage engine.
	if _, err := nk.StorageWrite(context.Background(), wObj); err != nil {
		return fmt.Errorf("failed to write player data to storage: %v", err)
	}
	return nil
}

// This function gets the player data from nakama storage.
func LoadPlayerData(nk runtime.NakamaModule, userID string) (*Player, error) {
	//Read from the storage engine.
	rObj, err := nk.StorageRead(context.Background(), []*runtime.StorageRead{
		{
			Collection: playerDataStorageCollection,
			Key: userID,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(rObj) == 0 {
		return nil, fmt.Errorf("player not found") //@JWK TODO: Call NewPlayer() and SavePlayerData().
	}
	var player Player
	if err = json.Unmarshal([]byte(rObj[0].Value), &player); err != nil {
		return nil, err
	}
	return &player, nil
}