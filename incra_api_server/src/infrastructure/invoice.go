package infrastructure

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/yuorei/incra_api_server/src/domain"
	"github.com/yuorei/incra_api_server/src/domain/repository"
)

type InvoiceRepository struct {
}

func NewInvoiceRepository() repository.InvoiceRepository {
	return &InvoiceRepository{}
}

func (r *InvoiceRepository) GetInvoice(invoiceId string) (domain.InvoiceResponse, error) {
	return domain.InvoiceResponse{}, nil
}

func (r *InvoiceRepository) CreateInvoice(invoice domain.Invoice) (domain.InvoiceResponse, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return domain.InvoiceResponse{}, err
	}

	client := sqs.NewFromConfig(cfg)

	// SQSキューのURL（Terraformの出力などから取得）
	queueUrl := os.Getenv("QUEUE_URL")

	// メッセージ送信
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueUrl),
		MessageBody: aws.String("Hello from Go Lambda!"),
	}

	result, err := client.SendMessage(context.TODO(), input)
	if err != nil {
		return domain.InvoiceResponse{}, err
	} else {
		fmt.Println("Message sent, ID:", *result.MessageId)
	}

	return domain.InvoiceResponse{}, nil
}
