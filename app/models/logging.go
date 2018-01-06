package models

import (
	"log"
	"os"
)

func SetupLogging() error {
	log.SetOutput(os.Stdout)
	log.Printf("Starting TextSecure")
	return nil
}
