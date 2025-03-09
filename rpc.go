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

		//Save the changes to player object.
		err = player.SavePlayerData(nk)
		if err != nil {
			logger.Error("Unable to save player data: %v", err)
			return "", err
		}

		//Limited scope response struct
		response := struct {
			PlayerData *Player `json:"player_data"`
		}{
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

func AttackTargetRPC()func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		//Get the user id from the runtime.
		userID, err := UtilGetUserId(ctx)
		if err != nil {
			logger.Error("Unable to extract user id from context due to error: %v", err)
			return "", err
		}

		//Client payload structure
		var attackRequest = struct {
			TargetID string `json:"target_id"`
			Attack AttackType `json:"attack"`
		}{
			TargetID: "",
			Attack: "",
		}
		if err := json.Unmarshal([]byte(payload), &attackRequest); err != nil {
			return "", runtime.NewError("unable to unmarshal payload", 13)
		}

		//Validate client input.

		//Get Player object.
		player, err := LoadPlayerData(ctx, logger, nk, userID)
		if err != nil {
			logger.Error("Unable to load player data: %v", err)
			return "", err
		}

		//Perform the attack.

		//Save any changes to player object.
		err = player.SavePlayerData(nk)
		if err != nil {
			logger.Error("Unable to save player data: %v", err)
			return "", err
		}

		//Limited scope response struct
		response := struct {
			PlayerData *Player `json:"player_data"`
		}{
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
		response := struct {
			PlayerHealth int `json:"player_health"`
			StatusEffects []StatusEffect `json:"status_effects"`
			BattleStats map[EnemyType]int `json:"battle_stats"`
		}{
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