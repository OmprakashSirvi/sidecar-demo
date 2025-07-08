package config

import (
	"fmt"
	"path/filepath"
	"sidecar/constants"

	"github.com/spf13/viper"
)

// TODO: Specify the pathname and config file name dynamically
func InitConfig() {
	viper.SetConfigName("proxy")
	viper.SetConfigType("yaml")
	// TODO: Should get the config directory dynamically while initializing the service
	// Keep this as default if config directory is not provided
	abs, _ := filepath.Abs("/conf")
	viper.AddConfigPath(abs)
	err := viper.ReadInConfig()
	if err != nil {
		errStr := fmt.Sprintf("not able to read proxy.yaml, error: %v", err)
		panic(errStr)
	}
	viper.AutomaticEnv()

	viper.SetDefault(constants.MY_ENV, "local")
}

// Get the key name with the current env
func GetKeyNameForEnv(key string) string {
	env := viper.GetString(constants.MY_ENV)
	keyName := fmt.Sprintf("%v.%v", env, key)
	if viper.IsSet(keyName) {
		// This means there is an override available for this key in current env configuration
		return keyName
	}
	// No override is available for this key, use the existing one..
	return key
}

