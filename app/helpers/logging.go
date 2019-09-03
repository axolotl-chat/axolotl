package helpers

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func SetupLogging() error {
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
	return nil
}
