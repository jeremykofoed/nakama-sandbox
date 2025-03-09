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
func LoadPlayerData(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) (*Player, error) {
	//Get the user id from the runtime.
	userID, err := UtilGetUserId(ctx)
	if err != nil {
		logger.Error("Unable to extract user id from context due to error: %v", err)
		return nil, err
	}

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
		//Get the users account data to get a display name.
		accounts, err := nk.AccountsGetId(ctx, []string{userID})
		if err != nil {
			//More robust logging to get more info.
			logger.WithFields(map[string]interface{}{
				"userID": userID,
			}).Error("Unable to lookup account due to error: %v.", err)
			return nil, err
		}
		if len(accounts) != 1 {
			//More robust logging to get more info.
			logger.WithFields(map[string]interface{}{
				"userID": userID,
			}).Error("Found either no account or more than one.")
			return nil, err
		}
		//Set the display name from the users account data.
		displayName := accounts[0].User.DisplayName
		//Create new player object.
		player := NewPlayer(userID, displayName)
		//Save to storage.
		if err = SavePlayerData(nk, player); err != nil {
			return nil, err
		}
		return player, nil
	}
	var player Player
	//Unmarshal json data to player object.
	if err = json.Unmarshal([]byte(rObj[0].Value), &player); err != nil {
		return nil, err
	}
	return &player, nil
}