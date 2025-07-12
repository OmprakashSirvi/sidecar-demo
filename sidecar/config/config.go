package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"sidecar/applogger"
	"sidecar/constants"
	"sidecar/globals"
	"sidecar/models"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Initialize configurations path and load configurations
func InitConfig() {
	logger := applogger.GetLogger()
	viper.SetConfigName("proxy")
	viper.SetConfigType("yaml")
	abs, _ := filepath.Abs(globals.Global.ConfigDir)
	viper.AddConfigPath(abs)
	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("not able to read proxy.yaml")
	}

	viper.AutomaticEnv()

	// Setting the default environment variables
	viper.SetDefault(constants.MY_ENV, "local")
	loadAuthzConfigs(&logger)
}

// Get the key name with the current env
func GetKeyName(key string) string {
	env := viper.GetString(constants.MY_ENV)
	keyName := fmt.Sprintf("%v.%v", env, key)
	if viper.IsSet(keyName) {
		// This means there is an override available for this key in current env configuration
		return keyName
	}
	// No override is available for this key, use the existing one..
	return key
}

// Loads and validates the authz-configs
func loadAuthzConfigs(logger *zerolog.Logger) {
	var authzConfigs []models.AuthzConfig
	err := viper.UnmarshalKey(GetKeyName(constants.AUTHZ_POLICY), &authzConfigs)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to load authz policies")
	}

	// We can validate the authz configurations here if needed

	globals.Global.AuthzConfigs = authzConfigs
}

// Errors are already logged
func GetRoutesFromConfig(logger *zerolog.Logger) ([]models.ProxyRoute, error) {
	if ok := viper.IsSet(GetKeyName(constants.PROXY_ROUTES)); !ok {
		errMsg := "proxy-routes is not set, hence not configuring any routes"
		logger.Debug().Msg(errMsg)
		return nil, errors.New(errMsg)
	}

	var routes []models.ProxyRoute
	err := viper.UnmarshalKey(GetKeyName(constants.PROXY_ROUTES), &routes)
	if err != nil {
		logger.Fatal().Err(err).Msg("invalid proxy-routes configuration")
		return nil, err
	}

	return routes, nil
}
