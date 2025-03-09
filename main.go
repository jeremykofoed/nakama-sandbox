package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/heroiclabs/nakama-common/runtime"
)

type BuildEnvironment int32
// Build / Deployed Environment
const (
	Local BuildEnvironment = iota
	Development
	QA
	Production
)

var configDataStorageCollection = "config"

// See https://heroiclabs.com/docs/nakama/server-framework/go-runtime/ for more details on the Go Runtime and how to use it.
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("InitModule loaded!")

	//Make sure we can get ENV from the runtime.
	env, ok := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	if !ok {
		return fmt.Errorf("'runtime.env' YML could not be cast")
	}
	//Make sure we can get the configured environment.  This will facilitate a seperation of logic as/if needed by environment.
	configuredEnvironment, ok := env["ConfiguredEnvironement"]
	var environment BuildEnvironment
	if !ok {
		environment = Production
		logger.Error("ConfiguredEnvironement not found, defaulting to production.")
	}
	logger.Info("ConfiguredEnvironement: %v", configuredEnvironment)
	switch configuredEnvironment {
	case "local":
		environment = Local
	case "development":
		environment = Development
	case "qa":
		environment = QA
	case "production":
		environment = Production
	}
	logger.Info("Environment: %d", environment)

	//Initialize registries.
	InitStatusEffectsRegistry()
	logger.Debug("Loaded StatusEffectsRegistry: %+v", StatusEffectsRegistry)
	InitAttackRegistry()
	logger.Debug("Loaded AttackRegistry: %+v", AttackRegistry)
	err := InitEnemyRegistry(ctx, logger, nk)
	if err != nil {
		logger.Error("Error processing InitEnemyRegistry(): %v", err)
	}
	logger.Debug("Loaded EnemyRegistry: %+v", EnemyRegistry)

	//Before/After hooks if any.

	//Custom RPCs if any.
	//@JWK TODO: Implement RPC to load game.  This will allow either enemy selection or finish a previous battle.
	if err := initializer.RegisterRpc("load_game", LoadGameRPC()); err != nil {
		return err
	}

	//@JWK TODO: Implement RPC to do attacks {player id, target id, attack info}.
	//@JWK TODO: Successful hits must return updated health values.
	//@JWK TODO: Implement RPC to get player health, status effects, and the number of enemy TYPES the player has killed.
	//@JWK TODO: Bonus, implement unit tests.

	return nil
}


//@JWK TODO: Make and move to util.go file
// Utility function to get the user id from the runtime.
func UtilGetUserId(ctx context.Context) (string, error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return "", fmt.Errorf("Could not extract the user id from the runtime context.")
	}
	return userId, nil
}

func UtilMakeUUID() string {
	id := uuid.New()
	return id.String()
}