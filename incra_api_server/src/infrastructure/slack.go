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

func SendInvoiceNotificationWithPayButton(slackUserId string, invoice domain.Invoice) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	webBaseURL := os.Getenv("WEB_BASE_URL")
	api := slacklib.New(slackToken)

	text := fmt.Sprintf(
		"請求書が届きました\n• 請求書ID: %s\n• 発行者: %s\n• 合計金額: ¥%s\n• 支払期限: %s\n<%s/invoices/%s|請求書を確認する>",
		invoice.InvoiceId, invoice.IssuerSlackRealName, slackFormatAmount(invoice.TotalAmount), invoice.DueDate, webBaseURL, invoice.InvoiceId,
	)

	sectionBlock := slacklib.NewSectionBlock(
		slacklib.NewTextBlockObject(slacklib.MarkdownType, text, false, false),
		nil, nil,
	)

	payBtn := slacklib.NewButtonBlockElement(
		"mark_paid",
		invoice.InvoiceId,
		slacklib.NewTextBlockObject(slacklib.PlainTextType, "支払った", false, false),
	)
	payBtn.Style = slacklib.StylePrimary

	actionsBlock := slacklib.NewActionBlock("pay_actions", payBtn)

	_, _, _, err := api.SendMessage(
		slackUserId,
		slacklib.MsgOptionBlocks(sectionBlock, actionsBlock),
		slacklib.MsgOptionText(text, false),
	)
	if err != nil {
		return fmt.Errorf("failed to send invoice notification with pay button: %w", err)
	}
	return nil
}

func SendPaymentConfirmationRequestDM(issuerSlackUserId string, invoice domain.Invoice) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	webBaseURL := os.Getenv("WEB_BASE_URL")
	api := slacklib.New(slackToken)

	text := fmt.Sprintf(
		"支払い報告がありました\n• 請求書ID: %s\n• 取引先: %s\n• 合計金額: ¥%s\n<%s/invoices/%s|請求書を確認する>",
		invoice.InvoiceId, invoice.BillingClientName, slackFormatAmount(invoice.TotalAmount), webBaseURL, invoice.InvoiceId,
	)

	sectionBlock := slacklib.NewSectionBlock(
		slacklib.NewTextBlockObject(slacklib.MarkdownType, text, false, false),
		nil, nil,
	)

	confirmBtn := slacklib.NewButtonBlockElement(
		"confirm_payment",
		invoice.InvoiceId,
		slacklib.NewTextBlockObject(slacklib.PlainTextType, "承認", false, false),
	)
	confirmBtn.Style = slacklib.StylePrimary

	rejectBtn := slacklib.NewButtonBlockElement(
		"reject_payment",
		invoice.InvoiceId,
		slacklib.NewTextBlockObject(slacklib.PlainTextType, "差し戻し", false, false),
	)
	rejectBtn.Style = slacklib.StyleDanger

	actionsBlock := slacklib.NewActionBlock("confirm_actions", confirmBtn, rejectBtn)

	_, _, _, err := api.SendMessage(
		issuerSlackUserId,
		slacklib.MsgOptionBlocks(sectionBlock, actionsBlock),
		slacklib.MsgOptionText(text, false),
	)
	if err != nil {
		return fmt.Errorf("failed to send payment confirmation request DM: %w", err)
	}
	return nil
}

func SendPaymentConfirmedDM(slackUserId string, invoice domain.Invoice) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	webBaseURL := os.Getenv("WEB_BASE_URL")
	api := slacklib.New(slackToken)

	message := fmt.Sprintf(
		"支払いが承認されました\n• 請求書ID: %s\n• 合計金額: ¥%s\n<%s/invoices/%s|請求書を確認する>",
		invoice.InvoiceId, slackFormatAmount(invoice.TotalAmount), webBaseURL, invoice.InvoiceId,
	)

	_, _, _, err := api.SendMessage(slackUserId, slacklib.MsgOptionText(message, false))
	if err != nil {
		return fmt.Errorf("failed to send payment confirmed DM: %w", err)
	}
	return nil
}

func SendPaymentRejectedDM(slackUserId string, invoice domain.Invoice) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	webBaseURL := os.Getenv("WEB_BASE_URL")
	api := slacklib.New(slackToken)

	text := fmt.Sprintf(
		"支払いが差し戻されました\n• 請求書ID: %s\n• 発行者: %s\n• 合計金額: ¥%s\n• 支払期限: %s\n<%s/invoices/%s|請求書を確認する>",
		invoice.InvoiceId, invoice.IssuerSlackRealName, slackFormatAmount(invoice.TotalAmount), invoice.DueDate, webBaseURL, invoice.InvoiceId,
	)

	sectionBlock := slacklib.NewSectionBlock(
		slacklib.NewTextBlockObject(slacklib.MarkdownType, text, false, false),
		nil, nil,
	)

	payBtn := slacklib.NewButtonBlockElement(
		"mark_paid",
		invoice.InvoiceId,
		slacklib.NewTextBlockObject(slacklib.PlainTextType, "支払った", false, false),
	)
	payBtn.Style = slacklib.StylePrimary

	actionsBlock := slacklib.NewActionBlock("pay_actions", payBtn)

	_, _, _, err := api.SendMessage(
		slackUserId,
		slacklib.MsgOptionBlocks(sectionBlock, actionsBlock),
		slacklib.MsgOptionText(text, false),
	)
	if err != nil {
		return fmt.Errorf("failed to send payment rejected DM: %w", err)
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
