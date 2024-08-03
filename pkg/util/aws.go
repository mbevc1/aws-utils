package util

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Testing struct approach
type Aws struct {
	Cfg    aws.Config
	Region string
}

func New(region string) (*Aws, error) {
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)

		return &Aws{}, err
	}

	return &Aws{
		Cfg:    cfg,
		Region: region,
	}, nil
}
