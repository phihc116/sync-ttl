package infrastructure

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var (
	Table          = os.Getenv("TABLE_NAME")
	dynamoDbClient *dynamodb.Client
	sqlClient      *gorm.DB
)

func InitDynamoDb(ctx context.Context) error {
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
		return fmt.Errorf("error loading AWS config: %w", err)
	}

	if endpoint := os.Getenv("AWS_ENDPOINT"); endpoint != "" {
		cfg.BaseEndpoint = aws.String(endpoint)
	}

	dynamoDbClient = dynamodb.NewFromConfig(cfg)
	return nil
}

func GetDynamoDbClient() *dynamodb.Client {
	if dynamoDbClient == nil {
		panic("DynamoDB client is not initialized. Call InitDynamoDb() first.")
	}
	return dynamoDbClient
}

func InitSQLServer(dsn string) error {
	var err error
	sqlClient, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect SQL Server: %w", err)
	}
	return nil
}

func GetSQLClient() *gorm.DB {
	if sqlClient == nil {
		panic("SQL Server client is not initialized. Call InitSQLServer() first.")
	}
	return sqlClient
}

func CloseSQL() {
	if sqlClient != nil {
		sqlDB, err := sqlClient.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}

func InitDataStore(ctx context.Context) error {
	sqlDsn := os.Getenv("SQL_DSN")
	if err := InitDynamoDb(ctx); err != nil {
		return err
	}
	if err := InitSQLServer(sqlDsn); err != nil {
		return err
	}
	log.Println("All clients initialized")
	return nil
}
