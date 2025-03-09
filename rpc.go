package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

func LoadGameRPC() func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		//Get the user id from the runtime.
		userID, err := UtilGetUserId(ctx)
		if err != nil {
			logger.Error("Unable to extract user id from context due to error: %v", err)
			return "", err
		}

		//Get Player object.
		player, err := LoadPlayerData(ctx, logger, nk, userID)
		if err != nil {
			logger.Error("Unable to load player data: %v", err)
			return "", err
		}

		//Get on-going battle state OR start a battle.
		err = player.LoadBattleState(ctx, logger, nk)
		if err != nil {
			logger.Error("Unable to get battle state: %v", err)
			return "", err
		}

		//@JWK TODO: Remove me when done
		logger.WithFields(map[string]interface{}{
			"player": player,
		}).Debug("LoadGameRPC()")

		//Save the changes to player object.
		err = player.SavePlayerData(nk)
		if err != nil {
			logger.Error("Unable to save player data: %v", err)
			return "", err
		}

		//Limited scope response struct
		type ClientResponse struct {
			PlayerData *Player `json:"player_data"`
		}
		response := ClientResponse{
			PlayerData: player,
		}

		//Return info to the client.
		jRes, err := json.Marshal(response)
		if err != nil {
			//More robust logging to get more info.
			logger.WithFields(map[string]interface{}{
				"response": response,
			}).Error("Unable to marshal client response: %v.", err)
			return "", err
		}

		return string(jRes), nil
	}
}

func RPCAttack(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) error {
	logger.Info("Payload: %s", payload)
	var value interface{}
	if err := json.Unmarshal([]byte(payload), &value); err != nil {
		return runtime.NewError("unable to unmarshal payload", 13)
	}
	response, err := json.Marshal(value)
	if err != nil {
		return runtime.NewError("unable to marshal payload", 13)
	}
	logger.Info("Response: %+v", response)
	return nil
}

func PlayerInfoRPC() func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		//Get the user id from the runtime.
		userID, err := UtilGetUserId(ctx)
		if err != nil {
			logger.Error("Unable to extract user id from context due to error: %v", err)
			return "", err
		}

		//Get Player object.
		player, err := LoadPlayerData(ctx, logger, nk, userID)
		if err != nil {
			logger.Error("Unable to load player data: %v", err)
			return "", err
		}

		//Limited scope response struct
		type ClientResponse struct {
			PlayerHealth int `json:"player_health"`
			StatusEffects []StatusEffect `json:"status_effects"`
			BattleStats map[EnemyType]int `json:"battle_stats"`
		}
		response := ClientResponse{
			PlayerHealth: player.Health,
			StatusEffects: player.StatusEffects,
			BattleStats: player.BattleStats,
		}

		//Return info to the client.
		jRes, err := json.Marshal(response)
		if err != nil {
			//More robust logging to get more info.
			logger.WithFields(map[string]interface{}{
				"response": response,
			}).Error("Unable to marshal client response: %v.", err)
			return "", err
		}

		return string(jRes), nil
	}
}