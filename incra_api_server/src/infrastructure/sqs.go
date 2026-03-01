package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/yuorei/incra_api_server/src/domain"
)

func SendPDFGenerateMessage(invoice domain.Invoice) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	client := sqs.NewFromConfig(cfg)
	queueUrl := os.Getenv("QUEUE_URL")
	body, err := json.Marshal(invoice)
	if err != nil {
		return err
	}
	result, err := client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueUrl),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return err
	}
	fmt.Println("PDF generate message sent, ID:", *result.MessageId)
	return nil
}
