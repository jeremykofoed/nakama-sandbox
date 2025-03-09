package main

import (
	"fmt"
	"context"
	"time"
	"encoding/json"
	"github.com/heroiclabs/nakama-common/runtime"
)

var playerDataStorageCollection = "data" //This is the collection name for all of the player related data.
var PlayerDataStorageKey = "player"

// Player data structure.
type Player struct {
	ID string `json:"id"` //Nakama user id.
	DisplayName string `json:"display_name"`
	Level int `json:"level"`
	Experience int64 `json:"experience"`
	Health int `json:"health"`
	Currencies []Currency `json:"currency"` //Nakama supports a wallet that can be implemented at a later time.
	StatusEffects []StatusEffect `json:"status_effects"` //Used to store player state modifiers.
	BattleState BattleState `json:"battle_state"` //Used to store the battle game state.
	BattleStats map[EnemyType]int `json:"battle_stats"` //Used to store the number of enemies vanquished.  Expound upon this to include other stats.
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
		Currencies: []Currency{
			{Type: Gold, Amount: 0, },
			{Type: Gems, Amount: 0, },
		},
		StatusEffects: []StatusEffect{},
		BattleState: BattleState{},
		BattleStats: make(map[EnemyType]int),
		Attributes: make(map[string]interface{}),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
}

// This function saves the player data to nakama storage.
// See https://heroiclabs.com/docs/nakama/concepts/storage/permissions/ for information on public read/write permissions or other storage information.
func (p *Player) SavePlayerData(nk runtime.NakamaModule) error {
	p.UpdatedAt = time.Now().Unix()
	//Json-ify the player struct in prepartion for storage.
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	wObj := []*runtime.StorageWrite{
		{
			Collection: playerDataStorageCollection,
			Key: PlayerDataStorageKey,
			UserID: p.ID, 
			Value: string(data),
			PermissionRead: 1, // Owner and runtime can read.
			PermissionWrite: 1, // Owner and runtime can read.
		},
	}
	//Write to the storage engine.
	if _, err := nk.StorageWrite(context.Background(), wObj); err != nil {
		return fmt.Errorf("failed to write player data to storage: %v", err)
	}
	return nil
}

// This function gets the player data from nakama storage.
func LoadPlayerData(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, userID string) (*Player, error) {
	//Read from the storage engine.
	rObj, err := nk.StorageRead(context.Background(), []*runtime.StorageRead{
		{
			Collection: playerDataStorageCollection,
			Key: PlayerDataStorageKey,
			UserID: userID,
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
		if displayName == "" {  //This is often empty as it requies a separate call to set.
			displayName = accounts[0].User.Username
		}
		//Create new player object.
		player := NewPlayer(userID, displayName)
		return player, nil
	}
	var player Player
	//Unmarshal json data to player object.
	if err = json.Unmarshal([]byte(rObj[0].Value), &player); err != nil {
		return nil, err
	}
	return &player, nil
}
