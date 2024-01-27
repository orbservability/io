package server

import (
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

func shutdownTimeout() time.Duration {
	timeoutStr := os.Getenv("SERVER_SHUTDOWN_TIMEOUT")
	if timeoutStr != "" {
		timeoutSec, err := strconv.Atoi(timeoutStr)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid SERVER_SHUTDOWN_TIMEOUT value; using default")
		} else {
			return time.Duration(timeoutSec) * time.Second
		}
	}
	return 30 * time.Second
}
