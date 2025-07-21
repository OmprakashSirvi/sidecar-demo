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

	setDefaults()
	loadAuthzConfigs(&logger)
	loadValidTokenTypes(&logger)
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

func setDefaults() {
	// Setting the default environment variables
	viper.SetDefault(constants.MY_ENV, "local")

	// Set default values for your configuration keys.
	viper.SetDefault(GetKeyName(constants.MAX_CONNECTION_LIMIT), 50)
	viper.SetDefault(GetKeyName(constants.REQUEST_TIMEOUT), 10)
	viper.SetDefault(GetKeyName(constants.MAX_REQUESTS_PER_SECOND), constants.DefaultMaxRequestPerSecond)
	viper.SetDefault(GetKeyName(constants.BURST_THRESHOLD), constants.DefaultBurstThreshold)
	viper.SetDefault(GetKeyName(constants.USER_RATE_LIMIT_WINDOW), constants.DefaultUserRateLimitWindow)

	// Set the config values into the application's global configuration
	globals.Global.MaxConnectionLimit = viper.GetInt(GetKeyName(constants.MAX_CONNECTION_LIMIT))
	globals.Global.RequestTimeout = viper.GetInt(GetKeyName(constants.REQUEST_TIMEOUT))
	globals.Global.MaxRequestsPerSecond = viper.GetFloat64(GetKeyName(constants.MAX_REQUESTS_PER_SECOND))
	globals.Global.BurstThreshold = viper.GetInt(GetKeyName(constants.BURST_THRESHOLD))
	globals.Global.RateLimitWindow = viper.GetInt(GetKeyName(constants.USER_RATE_LIMIT_WINDOW))
	globals.Global.ProxyBackend = viper.GetString(GetKeyName(constants.PROXY_BACKEND))
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

func loadValidTokenTypes(logger *zerolog.Logger) {
	var validTokenTypes []models.TokenTypes
	err := viper.UnmarshalKey(GetKeyName(constants.TOKEN_TYPES), &validTokenTypes)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to load valid token types")
	}

	globals.Global.ValidTokenTypes = validTokenTypes
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
