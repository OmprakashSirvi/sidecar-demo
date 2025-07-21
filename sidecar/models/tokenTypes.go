package models

type TokenTypes struct {
	Name    string `mapstructure:"name"`
	JwksUrl string `mapstructure:"jwks-url"`
}
