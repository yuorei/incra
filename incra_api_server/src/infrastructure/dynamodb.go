package infrastructure

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	dynamodbClient *dynamodb.Client
	dynamodbOnce   sync.Once
)

func GetDynamoDBClient() *dynamodb.Client {
	dynamodbOnce.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			panic(err)
		}
		dynamodbClient = dynamodb.NewFromConfig(cfg)
	})
	return dynamodbClient
}
