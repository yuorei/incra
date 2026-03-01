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
	"github.com/yuorei/incra_api_server/src/domain"
	"github.com/yuorei/incra_api_server/src/domain/repository"
)

type DynamoDBInvoiceRepository struct {
	client           *dynamodb.Client
	tableName        string
	counterTableName string
}

func NewInvoiceRepository() repository.InvoiceRepository {
	return &DynamoDBInvoiceRepository{
		client:           GetDynamoDBClient(),
		tableName:        os.Getenv("INVOICE_TABLE_NAME"),
		counterTableName: os.Getenv("COUNTER_TABLE_NAME"),
	}
}

func (r *DynamoDBInvoiceRepository) GetInvoice(invoiceId string) (domain.Invoice, error) {
	out, err := r.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"invoice_id": &types.AttributeValueMemberS{Value: invoiceId},
		},
	})
	if err != nil {
		return domain.Invoice{}, err
	}
	if out.Item == nil {
		return domain.Invoice{}, fmt.Errorf("invoice not found: %s", invoiceId)
	}
	var invoice domain.Invoice
	if err := attributevalue.UnmarshalMap(out.Item, &invoice); err != nil {
		return domain.Invoice{}, err
	}
	return invoice, nil
}

func (r *DynamoDBInvoiceRepository) ListInvoices(issuerSlackUserId string, status string, limit int, lastKey string) ([]domain.Invoice, string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("issuer_slack_user_id-created_at-index"),
		KeyConditionExpression: aws.String("issuer_slack_user_id = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: issuerSlackUserId},
		},
		ScanIndexForward: aws.Bool(false),
	}
	if limit > 0 {
		input.Limit = aws.Int32(int32(limit))
	}
	if status != "" {
		input.FilterExpression = aws.String("#s = :status")
		input.ExpressionAttributeNames = map[string]string{"#s": "status"}
		input.ExpressionAttributeValues[":status"] = &types.AttributeValueMemberS{Value: status}
	}
	out, err := r.client.Query(context.TODO(), input)
	if err != nil {
		return nil, "", err
	}
	var invoices []domain.Invoice
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &invoices); err != nil {
		return nil, "", err
	}
	return invoices, "", nil
}

func (r *DynamoDBInvoiceRepository) CreateInvoice(invoice domain.Invoice) (domain.Invoice, error) {
	item, err := attributevalue.MarshalMap(invoice)
	if err != nil {
		return domain.Invoice{}, err
	}
	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return domain.Invoice{}, err
	}
	return invoice, nil
}

func (r *DynamoDBInvoiceRepository) UpdateInvoice(invoice domain.Invoice) (domain.Invoice, error) {
	item, err := attributevalue.MarshalMap(invoice)
	if err != nil {
		return domain.Invoice{}, err
	}
	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return domain.Invoice{}, err
	}
	return invoice, nil
}

func (r *DynamoDBInvoiceRepository) UpdateInvoiceStatus(invoiceId string, status domain.InvoiceStatus, changedBy string) (domain.Invoice, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	current, err := r.GetInvoice(invoiceId)
	if err != nil {
		return domain.Invoice{}, err
	}
	entry := domain.HistoryEntry{
		ChangedAt: now,
		OldStatus: string(current.Status),
		NewStatus: string(status),
		ChangedBy: changedBy,
	}
	entryAttr, err := attributevalue.MarshalMap(entry)
	if err != nil {
		return domain.Invoice{}, err
	}
	_, err = r.client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"invoice_id": &types.AttributeValueMemberS{Value: invoiceId},
		},
		UpdateExpression: aws.String("SET #s = :status, updated_at = :now, history = list_append(if_not_exists(history, :empty_list), :entry)"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":     &types.AttributeValueMemberS{Value: string(status)},
			":now":        &types.AttributeValueMemberS{Value: now},
			":entry":      &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberM{Value: entryAttr}}},
			":empty_list": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
		},
	})
	if err != nil {
		return domain.Invoice{}, err
	}
	return r.GetInvoice(invoiceId)
}

func (r *DynamoDBInvoiceRepository) DeleteInvoice(invoiceId string) error {
	_, err := r.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"invoice_id": &types.AttributeValueMemberS{Value: invoiceId},
		},
	})
	return err
}

func (r *DynamoDBInvoiceRepository) NextInvoiceNumber(year int) (string, error) {
	out, err := r.client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(r.counterTableName),
		Key: map[string]types.AttributeValue{
			"counter_name": &types.AttributeValueMemberS{Value: "invoice_number"},
		},
		UpdateExpression: aws.String("ADD #val :one SET #year = :year"),
		ExpressionAttributeNames: map[string]string{
			"#val":  "value",
			"#year": "year",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":one":  &types.AttributeValueMemberN{Value: "1"},
			":year": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", year)},
		},
		ReturnValues: types.ReturnValueUpdatedNew,
	})
	if err != nil {
		return "", err
	}
	var result struct {
		Value int `dynamodbav:"value"`
	}
	if err := attributevalue.UnmarshalMap(out.Attributes, &result); err != nil {
		return "", err
	}
	return fmt.Sprintf("INV-%d-%04d", year, result.Value), nil
}
