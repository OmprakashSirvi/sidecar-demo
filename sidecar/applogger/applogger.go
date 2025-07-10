package applogger

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogging() {
	log.Logger = zerolog.New(os.Stdout).With().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func GetLogger() zerolog.Logger {
	return zerolog.New(os.Stdout).With().Logger()
}

// Gets logger from gin context
func GetCtxLogger(ctx *gin.Context) zerolog.Logger {
	var logger zerolog.Logger
	ctxLogger, exists := ctx.Get("logger")
	if !exists {
		return GetLogger()
	}

	logger, ok := ctxLogger.(zerolog.Logger)
	if !ok {
		return GetLogger()
	}

	return logger
}
