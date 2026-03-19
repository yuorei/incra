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

func invoiceStatusLabel(status domain.InvoiceStatus) string {
	switch status {
	case domain.InvoiceStatusDraft:
		return "下書き"
	case domain.InvoiceStatusSent:
		return "送付済（支払い待ち）"
	case domain.InvoiceStatusPaid:
		return "支払い報告済（確認待ち）"
	case domain.InvoiceStatusConfirmed:
		return "確認済み"
	case domain.InvoiceStatusCancelled:
		return "キャンセル"
	default:
		return string(status)
	}
}

func BuildHelpBlocks(cmd string) []slacklib.Block {
	header := slacklib.NewHeaderBlock(
		slacklib.NewTextBlockObject(slacklib.PlainTextType, cmd+" コマンド一覧", false, false),
	)
	commands := []struct{ cmd, desc string }{
		{cmd, "請求書作成モーダルを開く"},
		{cmd + " new", "請求書作成モーダルを開く"},
		{cmd + " help", "このヘルプを表示"},
		{cmd + " list", "自分が発行した請求書（直近5件）"},
		{cmd + " issued", "自分が発行した請求書（直近5件）"},
		{cmd + " received", "自分が支払うべき請求書（直近5件）"},
		{cmd + " pay", "自分が支払うべき請求書（直近5件）"},
		{cmd + " unpaid", "自分が発行した未収金（sent + paid）"},
		{cmd + " summary", "発行済み・受領済みのステータス別集計"},
	}
	blocks := []slacklib.Block{header}
	for _, c := range commands {
		text := fmt.Sprintf("*`%s`*\n%s", c.cmd, c.desc)
		blocks = append(blocks, slacklib.NewSectionBlock(
			slacklib.NewTextBlockObject(slacklib.MarkdownType, text, false, false),
			nil, nil,
		))
	}
	return blocks
}

func BuildIssuedListBlocks(invoices []domain.Invoice, webBaseURL string) []slacklib.Block {
	header := slacklib.NewHeaderBlock(
		slacklib.NewTextBlockObject(slacklib.PlainTextType, "発行済み請求書（直近5件）", false, false),
	)
	blocks := []slacklib.Block{header}
	if len(invoices) == 0 {
		blocks = append(blocks, slacklib.NewSectionBlock(
			slacklib.NewTextBlockObject(slacklib.MarkdownType, "発行済み請求書はありません", false, false),
			nil, nil,
		))
		return blocks
	}
	for _, inv := range invoices {
		text := fmt.Sprintf("*%s*  宛先: %s  ¥%s  期限: %s  %s",
			inv.InvoiceId, inv.BillingClientName, slackFormatAmount(inv.TotalAmount), inv.DueDate, invoiceStatusLabel(inv.Status))
		btn := slacklib.NewButtonBlockElement(
			"view_invoice",
			inv.InvoiceId,
			slacklib.NewTextBlockObject(slacklib.PlainTextType, "詳細", false, false),
		)
		btn.URL = fmt.Sprintf("%s/invoices/%s", webBaseURL, inv.InvoiceId)
		blocks = append(blocks, slacklib.NewSectionBlock(
			slacklib.NewTextBlockObject(slacklib.MarkdownType, text, false, false),
			nil,
			slacklib.NewAccessory(btn),
		))
	}
	return blocks
}

func BuildReceivedListBlocks(invoices []domain.Invoice, webBaseURL string) []slacklib.Block {
	header := slacklib.NewHeaderBlock(
		slacklib.NewTextBlockObject(slacklib.PlainTextType, "支払いが必要な請求書（直近5件）", false, false),
	)
	blocks := []slacklib.Block{header}
	if len(invoices) == 0 {
		blocks = append(blocks, slacklib.NewSectionBlock(
			slacklib.NewTextBlockObject(slacklib.MarkdownType, "支払いが必要な請求書はありません", false, false),
			nil, nil,
		))
		return blocks
	}
	for _, inv := range invoices {
		text := fmt.Sprintf("*%s*  発行者: %s  ¥%s  期限: %s",
			inv.InvoiceId, inv.IssuerSlackRealName, slackFormatAmount(inv.TotalAmount), inv.DueDate)
		btn := slacklib.NewButtonBlockElement(
			"view_invoice",
			inv.InvoiceId,
			slacklib.NewTextBlockObject(slacklib.PlainTextType, "詳細", false, false),
		)
		btn.URL = fmt.Sprintf("%s/invoices/%s", webBaseURL, inv.InvoiceId)
		blocks = append(blocks, slacklib.NewSectionBlock(
			slacklib.NewTextBlockObject(slacklib.MarkdownType, text, false, false),
			nil,
			slacklib.NewAccessory(btn),
		))
	}
	return blocks
}

func BuildUnpaidBlocks(sent []domain.Invoice, paid []domain.Invoice, webBaseURL string) []slacklib.Block {
	header := slacklib.NewHeaderBlock(
		slacklib.NewTextBlockObject(slacklib.PlainTextType, "未収金レポート", false, false),
	)
	blocks := []slacklib.Block{header}
	if len(sent) == 0 && len(paid) == 0 {
		blocks = append(blocks, slacklib.NewSectionBlock(
			slacklib.NewTextBlockObject(slacklib.MarkdownType, "未収金はありません 🎉", false, false),
			nil, nil,
		))
		return blocks
	}
	if len(sent) > 0 {
		total := 0
		for _, inv := range sent {
			total += inv.TotalAmount
		}
		blocks = append(blocks, slacklib.NewSectionBlock(
			slacklib.NewTextBlockObject(slacklib.MarkdownType,
				fmt.Sprintf("*支払い待ち（%d件）* 合計: ¥%s", len(sent), slackFormatAmount(total)),
				false, false),
			nil, nil,
		))
		for _, inv := range sent {
			text := fmt.Sprintf("• %s  宛先: %s  ¥%s  期限: %s",
				inv.InvoiceId, inv.BillingClientName, slackFormatAmount(inv.TotalAmount), inv.DueDate)
			blocks = append(blocks, slacklib.NewSectionBlock(
				slacklib.NewTextBlockObject(slacklib.MarkdownType, text, false, false),
				nil, nil,
			))
		}
	}
	if len(paid) > 0 {
		total := 0
		for _, inv := range paid {
			total += inv.TotalAmount
		}
		blocks = append(blocks, slacklib.NewSectionBlock(
			slacklib.NewTextBlockObject(slacklib.MarkdownType,
				fmt.Sprintf("*確認待ち（%d件）* 合計: ¥%s", len(paid), slackFormatAmount(total)),
				false, false),
			nil, nil,
		))
		for _, inv := range paid {
			text := fmt.Sprintf("• %s  宛先: %s  ¥%s  期限: %s",
				inv.InvoiceId, inv.BillingClientName, slackFormatAmount(inv.TotalAmount), inv.DueDate)
			blocks = append(blocks, slacklib.NewSectionBlock(
				slacklib.NewTextBlockObject(slacklib.MarkdownType, text, false, false),
				nil, nil,
			))
		}
	}
	return blocks
}

func BuildSummaryBlocks(issued []domain.Invoice, received []domain.Invoice, webBaseURL string) []slacklib.Block {
	header := slacklib.NewHeaderBlock(
		slacklib.NewTextBlockObject(slacklib.PlainTextType, "請求書サマリー", false, false),
	)
	blocks := []slacklib.Block{header}

	countByStatus := func(invoices []domain.Invoice) map[domain.InvoiceStatus]struct{ count, total int } {
		m := map[domain.InvoiceStatus]struct{ count, total int }{}
		for _, inv := range invoices {
			e := m[inv.Status]
			e.count++
			e.total += inv.TotalAmount
			m[inv.Status] = e
		}
		return m
	}

	statuses := []domain.InvoiceStatus{
		domain.InvoiceStatusDraft,
		domain.InvoiceStatusSent,
		domain.InvoiceStatusPaid,
		domain.InvoiceStatusConfirmed,
		domain.InvoiceStatusCancelled,
	}

	renderSection := func(title string, invoices []domain.Invoice) {
		blocks = append(blocks, slacklib.NewSectionBlock(
			slacklib.NewTextBlockObject(slacklib.MarkdownType, fmt.Sprintf("*%s*", title), false, false),
			nil, nil,
		))
		m := countByStatus(invoices)
		var fields []*slacklib.TextBlockObject
		for _, s := range statuses {
			e := m[s]
			if e.count == 0 {
				continue
			}
			fields = append(fields,
				slacklib.NewTextBlockObject(slacklib.MarkdownType,
					fmt.Sprintf("*%s*\n%d件 / ¥%s", invoiceStatusLabel(s), e.count, slackFormatAmount(e.total)),
					false, false),
			)
		}
		if len(fields) == 0 {
			blocks = append(blocks, slacklib.NewSectionBlock(
				slacklib.NewTextBlockObject(slacklib.MarkdownType, "該当なし", false, false),
				nil, nil,
			))
		} else {
			blocks = append(blocks, slacklib.NewSectionBlock(nil, fields, nil))
		}
	}

	renderSection("発行済み", issued)
	renderSection("受領済み", received)

	blocks = append(blocks, slacklib.NewContextBlock("",
		slacklib.NewTextBlockObject(slacklib.MarkdownType, "直近100件の集計", false, false),
	))
	return blocks
}

func BuildUnknownCommandBlocks(text string) []slacklib.Block {
	msg := fmt.Sprintf("不明なコマンド: `%s`\n`/incra help` で確認できます", text)
	return []slacklib.Block{
		slacklib.NewSectionBlock(
			slacklib.NewTextBlockObject(slacklib.MarkdownType, msg, false, false),
			nil, nil,
		),
	}
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
