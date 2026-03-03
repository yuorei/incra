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

	for _, inv := range invoices {
		dueDate, err := time.Parse("2006-01-02", inv.DueDate)
		if err != nil {
			fmt.Printf("warning: invalid due_date format for %s: %v\n", inv.InvoiceId, err)
			continue
		}

		daysUntilDue := int(time.Until(dueDate).Hours() / 24)
		if daysUntilDue > 3 {
			continue
		}

		// 期限の表示テキストを組み立て
		var dueText string
		if daysUntilDue < 0 {
			dueText = fmt.Sprintf(":warning: 支払い期限を *%d日超過* しています（%s）", -daysUntilDue, inv.DueDate)
		} else if daysUntilDue == 0 {
			dueText = fmt.Sprintf(":warning: 支払い期限は *本日* です（%s）", inv.DueDate)
		} else {
			dueText = fmt.Sprintf("支払い期限まで残り *%d日* です（%s）", daysUntilDue, inv.DueDate)
		}

		// 請求先の表示名（名前がなければSlackメンション）
		billingName := inv.BillingClientName
		if billingName == "" && inv.BillingSlackUserId != "" {
			billingName = fmt.Sprintf("<@%s>", inv.BillingSlackUserId)
		}

		linkURL := fmt.Sprintf("%s/invoices/%s", webBaseURL, inv.InvoiceId)

		// --- 発行者向けリマインド（Block Kit） ---
		issuerText := fmt.Sprintf(
			"*:bell: 請求書リマインダー*\n\n• 請求書ID: %s\n• 請求先: %s\n• 請求金額: ¥%s\n• %s\n\n<%s|請求書を確認する>",
			inv.InvoiceId, billingName, formatAmount(inv.TotalAmount), dueText, linkURL,
		)

		issuerSection := slack.NewSectionBlock(
			slack.NewTextBlockObject(slack.MarkdownType, issuerText, false, false),
			nil, nil,
		)

		_, _, _, err = api.SendMessage(
			inv.IssuerSlackUserId,
			slack.MsgOptionBlocks(issuerSection),
			slack.MsgOptionText(issuerText, false),
		)
		if err != nil {
			fmt.Printf("warning: failed to send issuer reminder to %s for %s: %v\n", inv.IssuerSlackUserId, inv.InvoiceId, err)
		} else {
			fmt.Printf("sent issuer reminder for %s to %s (due: %s, days: %d)\n", inv.InvoiceId, inv.IssuerSlackUserId, inv.DueDate, daysUntilDue)
		}

		// --- 請求先ユーザー向けリマインド（Block Kit + 支払ったボタン） ---
		if inv.BillingSlackUserId != "" {
			billingText := fmt.Sprintf(
				"*:bell: 請求書リマインダー*\n\n• 請求書ID: %s\n• 発行者: %s\n• 請求金額: ¥%s\n• %s\n\n<%s|請求書を確認する>",
				inv.InvoiceId, inv.IssuerSlackRealName, formatAmount(inv.TotalAmount), dueText, linkURL,
			)

			billingSection := slack.NewSectionBlock(
				slack.NewTextBlockObject(slack.MarkdownType, billingText, false, false),
				nil, nil,
			)

			payBtn := slack.NewButtonBlockElement(
				"mark_paid",
				inv.InvoiceId,
				slack.NewTextBlockObject(slack.PlainTextType, "支払った", false, false),
			)
			payBtn.Style = slack.StylePrimary

			actionsBlock := slack.NewActionBlock("pay_actions", payBtn)

			_, _, _, err = api.SendMessage(
				inv.BillingSlackUserId,
				slack.MsgOptionBlocks(billingSection, actionsBlock),
				slack.MsgOptionText(billingText, false),
			)
			if err != nil {
				fmt.Printf("warning: failed to send billing reminder to %s for %s: %v\n", inv.BillingSlackUserId, inv.InvoiceId, err)
			} else {
				fmt.Printf("sent billing reminder for %s to %s (due: %s, days: %d)\n", inv.InvoiceId, inv.BillingSlackUserId, inv.DueDate, daysUntilDue)
			}
		}
	}
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
