package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/slack-go/slack"
	"github.com/yuorei/incra_api_server/src/domain"
)

func main() {
	lambda.Start(Handler)
}

func Handler(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	db := dynamodb.NewFromConfig(cfg)
	tableName := os.Getenv("INVOICE_TABLE_NAME")
	slackToken := os.Getenv("SLACK_TOKEN")
	webBaseURL := os.Getenv("WEB_BASE_URL")
	api := slack.New(slackToken)

	// status="sent"の請求書をScan
	out, err := db.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("#s = :status"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: "sent"},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to scan invoices: %w", err)
	}

	var invoices []domain.Invoice
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &invoices); err != nil {
		return fmt.Errorf("failed to unmarshal invoices: %w", err)
	}

	now := time.Now()
	for _, inv := range invoices {
		dueDate, err := time.Parse("2006-01-02", inv.DueDate)
		if err != nil {
			fmt.Printf("warning: invalid due_date format for %s: %v\n", inv.InvoiceId, err)
			continue
		}

		daysUntilDue := int(time.Until(dueDate).Hours() / 24)

		var message string
		if daysUntilDue < 0 {
			message = fmt.Sprintf(
				"*請求書リマインダー* :warning:\n%s (%s) の支払い期限を *%d日超過* しています (%s)\n請求金額: ¥%s\n<%s/invoices/%s|請求書を確認する>",
				inv.InvoiceId, inv.BillingClientName, -daysUntilDue, inv.DueDate, formatAmount(inv.TotalAmount), webBaseURL, inv.InvoiceId,
			)
		} else if daysUntilDue <= 3 {
			message = fmt.Sprintf(
				"*請求書リマインダー*\n%s (%s) の支払い期限まで残り *%d日* です (%s)\n請求金額: ¥%s\n<%s/invoices/%s|請求書を確認する>",
				inv.InvoiceId, inv.BillingClientName, daysUntilDue, inv.DueDate, formatAmount(inv.TotalAmount), webBaseURL, inv.InvoiceId,
			)
		} else {
			continue
		}

		// Slack DM送信
		_, _, _, err = api.SendMessage(inv.IssuerSlackUserId, slack.MsgOptionText(message, false))
		if err != nil {
			fmt.Printf("warning: failed to send DM to %s for %s: %v\n", inv.IssuerSlackUserId, inv.InvoiceId, err)
		} else {
			fmt.Printf("sent reminder for %s to %s (due: %s, days: %d)\n", inv.InvoiceId, inv.IssuerSlackUserId, inv.DueDate, daysUntilDue)
		}
	}

	_ = now
	return nil
}

func formatAmount(amount int) string {
	s := fmt.Sprintf("%d", amount)
	n := len(s)
	if n <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
