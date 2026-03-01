package infrastructure

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	slacklib "github.com/slack-go/slack"
	"github.com/yuorei/incra_api_server/src/domain"
)

func SendInvoiceNotificationDM(slackUserId string, invoice domain.Invoice) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	webBaseURL := os.Getenv("WEB_BASE_URL")
	api := slacklib.New(slackToken)

	message := fmt.Sprintf(
		"請求書が届きました\n• 請求書ID: %s\n• 発行者: %s\n• 合計金額: ¥%s\n• 支払期限: %s\n<%s/invoices/%s|請求書を確認する>",
		invoice.InvoiceId, invoice.IssuerSlackRealName, slackFormatAmount(invoice.TotalAmount), invoice.DueDate, webBaseURL, invoice.InvoiceId,
	)

	_, _, _, err := api.SendMessage(slackUserId, slacklib.MsgOptionText(message, false))
	if err != nil {
		return fmt.Errorf("failed to send invoice notification DM: %w", err)
	}
	return nil
}

func slackFormatAmount(amount int) string {
	s := strconv.Itoa(amount)
	n := len(s)
	if n <= 3 {
		return s
	}
	var result strings.Builder
	for i, ch := range s {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteRune(ch)
	}
	return result.String()
}
