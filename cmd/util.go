package cmd

import (
	"log"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func checkDebug() {
	if debugFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Info("debug!")
	}
}

func newCfg() aws.Config {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(regionFlag))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	return cfg
}
