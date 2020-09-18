package config

import (
	"github.com/kelseyhightower/envconfig"
	"log"
)

// EnvConfig godoc
type EnvConfig struct {
	AggregationIntervalMinutes int `envconfig:"AGGREGATION_INTERVAL_MINUTES" default:30`
}

var env EnvConfig

// GetConfig godoc
func GetConfig() EnvConfig {

	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}
	return env
}
