package main

import (
	"context"
	"database/sql"
	"fmt"

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

	//Before/After hooks if any.

	//Custom RPCs if any.
	//@JWK TODO: Implement RPC to do attacks {player id, target id, attack info}.
	//@JWK TODO: Successful hits must return updated health values.
	//@JWK TODO: Implement RPC to get player health, status effects, and the number of enemy TYPES the player has killed.
	//@JWK TODO: Bonus, implement unit tests.

	return nil
}
