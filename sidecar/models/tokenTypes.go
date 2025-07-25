package models

type TokenType struct {
	Name    string `mapstructure:"name"`
	JwksUrl string `mapstructure:"jwks-url"`
}
