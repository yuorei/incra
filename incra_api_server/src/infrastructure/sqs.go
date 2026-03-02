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

type PDFGenerateMessage struct {
	domain.Invoice
	BillingClientSlackUserId string `json:"billing_client_slack_user_id"`
}

func SendPDFGenerateMessage(invoice domain.Invoice, billingClientSlackUserId string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	client := sqs.NewFromConfig(cfg)
	queueUrl := os.Getenv("QUEUE_URL")
	msg := PDFGenerateMessage{
		Invoice:                  invoice,
		BillingClientSlackUserId: billingClientSlackUserId,
	}
	body, err := json.Marshal(msg)
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
