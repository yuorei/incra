package infrastructure

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/yuorei/incra_api_server/src/domain"
	"github.com/yuorei/incra_api_server/src/domain/repository"
)

type DynamoDBClientRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewClientRepository() repository.ClientRepository {
	return &DynamoDBClientRepository{
		client:    GetDynamoDBClient(),
		tableName: os.Getenv("CLIENT_TABLE_NAME"),
	}
}

func (r *DynamoDBClientRepository) GetClient(clientId string) (domain.Client, error) {
	out, err := r.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"client_id": &types.AttributeValueMemberS{Value: clientId},
		},
	})
	if err != nil {
		return domain.Client{}, err
	}
	if out.Item == nil {
		return domain.Client{}, fmt.Errorf("client not found: %s", clientId)
	}
	var client domain.Client
	if err := attributevalue.UnmarshalMap(out.Item, &client); err != nil {
		return domain.Client{}, err
	}
	return client, nil
}

func (r *DynamoDBClientRepository) GetClientBySlackUserId(slackUserId string) (domain.Client, error) {
	out, err := r.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("slack_user_id-index"),
		KeyConditionExpression: aws.String("slack_user_id = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: slackUserId},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return domain.Client{}, err
	}
	if len(out.Items) == 0 {
		return domain.Client{}, fmt.Errorf("client not found for slack user: %s", slackUserId)
	}
	var client domain.Client
	if err := attributevalue.UnmarshalMap(out.Items[0], &client); err != nil {
		return domain.Client{}, err
	}
	return client, nil
}

func (r *DynamoDBClientRepository) ListClients(registeredBy string) ([]domain.Client, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}
	if registeredBy != "" {
		input.FilterExpression = aws.String("registered_by = :uid")
		input.ExpressionAttributeValues = map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: registeredBy},
		}
	}
	out, err := r.client.Scan(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	var clients []domain.Client
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &clients); err != nil {
		return nil, err
	}
	return clients, nil
}

func (r *DynamoDBClientRepository) CreateClient(client domain.Client) (domain.Client, error) {
	if client.ClientId == "" {
		client.ClientId = uuid.New().String()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	client.CreatedAt = now
	client.UpdatedAt = now
	item, err := attributevalue.MarshalMap(client)
	if err != nil {
		return domain.Client{}, err
	}
	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return domain.Client{}, err
	}
	return client, nil
}

func (r *DynamoDBClientRepository) UpdateClient(client domain.Client) (domain.Client, error) {
	client.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	item, err := attributevalue.MarshalMap(client)
	if err != nil {
		return domain.Client{}, err
	}
	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return domain.Client{}, err
	}
	return client, nil
}

func (r *DynamoDBClientRepository) DeleteClient(clientId string) error {
	_, err := r.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"client_id": &types.AttributeValueMemberS{Value: clientId},
		},
	})
	return err
}
