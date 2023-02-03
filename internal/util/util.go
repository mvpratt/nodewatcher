package util

import (
	"log"
	"os"
)

// RequireEnvVar returns the variable specified, or exits if environment variable not defined
func RequireEnvVar(varName string) string {
	env := os.Getenv(varName)
	if env == "" {
		log.Fatalf("\nERROR: %s environment variable must be set.", varName)
	}
	return env
}
