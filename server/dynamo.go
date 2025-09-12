package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	Table = os.Getenv("TABLE_NAME")
)

type Dynamo struct {
	*dynamodb.Client
}

func NewDynamoDb(ctx context.Context) (*Dynamo, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 50,
		},
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(os.Getenv("AWS_REGION")),
		config.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %w", err)
	}

	if endpoint := os.Getenv("AWS_ENDPOINT"); endpoint != "" {
		cfg.BaseEndpoint = aws.String(endpoint)
	}

	db := dynamodb.NewFromConfig(cfg)
	return &Dynamo{db}, nil
}
