package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	petstore "github.com/yuorei/incra_api_server/api/v1"
	"github.com/yuorei/incra_api_server/src/domain"
	"github.com/yuorei/incra_api_server/src/infrastructure"
	"github.com/yuorei/incra_api_server/src/usecase"
)

type ServerImpl struct {
	invoiceUseCase usecase.InvoiceUseCase
}

func NewServerImpl() *ServerImpl {
	return &ServerImpl{
		invoiceUseCase: usecase.NewInvoiceUseCase(),
	}
}

func (s *ServerImpl) GetHealth(ctx echo.Context) error {
	health := "OK"
	return ctx.JSON(http.StatusOK, petstore.Health{Message: &health})
}

func (s *ServerImpl) CreateInvoice(ctx echo.Context) error {
	var req petstore.CreateInvoiceRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	items := make([]domain.InvoiceItem, len(req.Items))
	for i, item := range req.Items {
		memo := ""
		if item.Memo != nil {
			memo = *item.Memo
		}
		items[i] = domain.InvoiceItem{
			Date:        item.Date,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Amount:      item.Amount,
			Memo:        memo,
		}
	}

	additionalInfo := ""
	if req.AdditionalInfo != nil {
		additionalInfo = *req.AdditionalInfo
	}

	billingClientId := ""
	if req.BillingClientId != nil {
		billingClientId = *req.BillingClientId
	}
	billingClientName := ""
	if req.BillingClientName != nil {
		billingClientName = *req.BillingClientName
	}
	if req.BillingSlackUserId == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "billing_slack_user_id is required"})
	}

	issuerSlackUserId := ctx.Request().Header.Get("X-Slack-User-Id")
	if req.BillingSlackUserId == issuerSlackUserId {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "自分自身に請求することはできません"})
	}

	invoice := domain.Invoice{
		BillingClientId:     billingClientId,
		BillingClientName:   billingClientName,
		BillingSlackUserId:  req.BillingSlackUserId,
		DueDate:             req.DueDate,
		BankDetails:         req.BankDetails,
		AdditionalInfo:      additionalInfo,
		Items:               items,
		IssuerSlackUserId:   issuerSlackUserId,
		IssuerSlackRealName: ctx.Request().Header.Get("X-Slack-User-Name"),
	}

	created, err := s.invoiceUseCase.CreateInvoice(invoice)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, domainInvoiceToAPI(created))
}

func (s *ServerImpl) ListInvoices(ctx echo.Context, params petstore.ListInvoicesParams) error {
	status := ""
	if params.Status != nil {
		status = string(*params.Status)
	}
	limit := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	lastKey := ""
	if params.LastKey != nil {
		lastKey = *params.LastKey
	}
	slackUserId := ctx.Request().Header.Get("X-Slack-User-Id")

	var invoices []domain.Invoice
	var nextKey string
	var err error

	if params.Type != nil && *params.Type == "received" {
		invoices, nextKey, err = s.invoiceUseCase.ListReceivedInvoices(slackUserId, status, limit, lastKey)
	} else {
		invoices, nextKey, err = s.invoiceUseCase.ListInvoices(slackUserId, status, limit, lastKey)
	}
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	apiInvoices := make([]petstore.Invoice, len(invoices))
	for i, inv := range invoices {
		apiInvoices[i] = domainInvoiceToAPI(inv)
	}
	resp := petstore.InvoiceListResponse{
		Invoices: &apiInvoices,
		NextKey:  &nextKey,
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *ServerImpl) GetInvoice(ctx echo.Context, invoiceId string) error {
	invoice, err := s.invoiceUseCase.GetInvoice(invoiceId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, domainInvoiceToAPI(invoice))
}

func (s *ServerImpl) UpdateInvoice(ctx echo.Context, invoiceId string) error {
	var req petstore.UpdateInvoiceRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	invoice := domain.Invoice{InvoiceId: invoiceId}
	if req.BillingClientId != nil {
		invoice.BillingClientId = *req.BillingClientId
	}
	if req.BillingClientName != nil {
		invoice.BillingClientName = *req.BillingClientName
	}
	if req.BillingSlackUserId != nil {
		if *req.BillingSlackUserId == ctx.Request().Header.Get("X-Slack-User-Id") {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "自分自身に請求することはできません"})
		}
		invoice.BillingSlackUserId = *req.BillingSlackUserId
	}
	if req.DueDate != nil {
		invoice.DueDate = *req.DueDate
	}
	if req.BankDetails != nil {
		invoice.BankDetails = *req.BankDetails
	}
	if req.AdditionalInfo != nil {
		invoice.AdditionalInfo = *req.AdditionalInfo
	}
	if req.Items != nil {
		items := make([]domain.InvoiceItem, len(*req.Items))
		for i, item := range *req.Items {
			memo := ""
			if item.Memo != nil {
				memo = *item.Memo
			}
			items[i] = domain.InvoiceItem{
				Date:        item.Date,
				Description: item.Description,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				Amount:      item.Amount,
				Memo:        memo,
			}
		}
		invoice.Items = items
	}

	updated, err := s.invoiceUseCase.UpdateInvoice(invoice)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, domainInvoiceToAPI(updated))
}

func (s *ServerImpl) UpdateInvoiceStatus(ctx echo.Context, invoiceId string) error {
	var req petstore.UpdateInvoiceStatusRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	changedBy := ctx.Request().Header.Get("X-Slack-User-Name")
	changedByUserId := ctx.Request().Header.Get("X-Slack-User-Id")
	if changedBy == "" {
		changedBy = changedByUserId
	}
	updated, err := s.invoiceUseCase.TransitionStatus(invoiceId, domain.InvoiceStatus(req.Status), changedBy, changedByUserId)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, domainInvoiceToAPI(updated))
}

func (s *ServerImpl) DeleteInvoice(ctx echo.Context, invoiceId string) error {
	if err := s.invoiceUseCase.DeleteInvoice(invoiceId); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (s *ServerImpl) SlackEventsHandler(c echo.Context) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	api := slack.New(slackToken)

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		return err
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var res *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &res); err != nil {
			return err
		}

		return c.String(http.StatusOK, res.Challenge)

	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		switch event := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			message := strings.Split(event.Text, " ")
			if len(message) < 2 {
				return err
			}
			command := message[1]
			switch command {
			case "ping":
				_, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("pong", false))
				if err != nil {
					return err
				}
			}
		}

	}

	return c.NoContent(http.StatusOK)
}

// SlackModalMetadata is the JSON-encoded private metadata for the invoice creation modal.
type SlackModalMetadata struct {
	IssuerID  string `json:"issuer_id"`
	ItemCount int    `json:"item_count"`
}

// buildItemInputBlocks creates the three input blocks (品目名/数量/単価) for an item at the given index.
func buildItemInputBlocks(index int) []slack.Block {
	prefix := fmt.Sprintf("item_%d", index)
	label := fmt.Sprintf("品目 %d", index+1)

	descInput := slack.NewInputBlock(
		prefix+"_desc_block",
		slack.NewTextBlockObject(slack.PlainTextType, label+" - 品目名", false, false),
		nil,
		slack.NewPlainTextInputBlockElement(
			slack.NewTextBlockObject(slack.PlainTextType, "品目名を入力", false, false),
			prefix+"_desc_action",
		),
	)

	quantityInput := slack.NewInputBlock(
		prefix+"_quantity_block",
		slack.NewTextBlockObject(slack.PlainTextType, label+" - 数量", false, false),
		nil,
		slack.NewPlainTextInputBlockElement(
			slack.NewTextBlockObject(slack.PlainTextType, "数量を入力", false, false),
			prefix+"_quantity_action",
		),
	)

	unitPriceInput := slack.NewInputBlock(
		prefix+"_unit_price_block",
		slack.NewTextBlockObject(slack.PlainTextType, label+" - 単価", false, false),
		nil,
		slack.NewPlainTextInputBlockElement(
			slack.NewTextBlockObject(slack.PlainTextType, "単価を入力", false, false),
			prefix+"_unit_price_action",
		),
	)

	return []slack.Block{descInput, quantityInput, unitPriceInput}
}

const invoiceDateFormat = "2006-01-02"

// buildInvoiceModalView builds the full ModalViewRequest for the invoice creation modal.
func buildInvoiceModalView(meta SlackModalMetadata) (slack.ModalViewRequest, error) {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return slack.ModalViewRequest{}, fmt.Errorf("failed to marshal modal metadata: %w", err)
	}

	billingUserElement := &slack.MultiSelectBlockElement{
		Type:        slack.MultiOptTypeUser,
		Placeholder: slack.NewTextBlockObject(slack.PlainTextType, "担当者を選択（複数可）", false, false),
		ActionID:    "billing_user_action",
	}
	billingUserSelect := slack.NewInputBlock(
		"billing_user_block",
		slack.NewTextBlockObject(slack.PlainTextType, "請求先", false, false),
		nil,
		billingUserElement,
	)
	billingUserSelect.Optional = false

	dueDatePicker := slack.NewInputBlock(
		"due_date_block",
		slack.NewTextBlockObject(slack.PlainTextType, "支払期限", false, false),
		nil,
		slack.NewDatePickerBlockElement("due_date_action"),
	)

	bankDetailsInput := slack.NewInputBlock(
		"bank_details_block",
		slack.NewTextBlockObject(slack.PlainTextType, "振込先", false, false),
		nil,
		slack.NewPlainTextInputBlockElement(
			slack.NewTextBlockObject(slack.PlainTextType, "振込先を入力", false, false),
			"bank_details_action",
		),
	)

	additionalInfoInput := slack.NewInputBlock(
		"additional_info_block",
		slack.NewTextBlockObject(slack.PlainTextType, "備考", false, false),
		nil,
		slack.NewPlainTextInputBlockElement(
			slack.NewTextBlockObject(slack.PlainTextType, "備考を入力（任意）", false, false),
			"additional_info_action",
		),
	)
	additionalInfoInput.Optional = true

	blocks := []slack.Block{billingUserSelect, dueDatePicker, bankDetailsInput}
	for i := 0; i < meta.ItemCount; i++ {
		blocks = append(blocks, buildItemInputBlocks(i)...)
	}

	addItemButton := slack.NewButtonBlockElement("add_item_action", "add_item",
		slack.NewTextBlockObject(slack.PlainTextType, "＋ アイテムを追加", false, false))
	blocks = append(blocks, slack.NewActionBlock("add_item_actions_block", addItemButton))
	blocks = append(blocks, additionalInfoInput)

	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           slack.NewTextBlockObject(slack.PlainTextType, "請求書作成", false, false),
		Submit:          slack.NewTextBlockObject(slack.PlainTextType, "作成", false, false),
		Close:           slack.NewTextBlockObject(slack.PlainTextType, "キャンセル", false, false),
		Blocks:          slack.Blocks{BlockSet: blocks},
		PrivateMetadata: string(metaJSON),
	}, nil
}

func (s *ServerImpl) SlackSlashsHandler(c echo.Context) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	api := slack.New(slackToken)

	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return err
	}

	subcommand := strings.TrimSpace(strings.ToLower(slashCommand.Text))
	switch subcommand {
	case "", "new":
		return s.handleSlashNew(c, api, slashCommand)
	case "help":
		return s.handleSlashHelp(c, slashCommand.Command)
	case "list", "issued":
		return s.handleSlashIssued(c, slashCommand.UserID)
	case "received", "pay":
		return s.handleSlashReceived(c, slashCommand.UserID)
	case "unpaid":
		return s.handleSlashUnpaid(c, slashCommand.UserID)
	case "summary":
		return s.handleSlashSummary(c, slashCommand.UserID)
	default:
		return s.handleSlashUnknown(c, slashCommand.Text)
	}
}

func (s *ServerImpl) handleSlashNew(c echo.Context, api *slack.Client, slashCommand slack.SlashCommand) error {
	meta := SlackModalMetadata{IssuerID: slashCommand.UserID, ItemCount: 1}
	modalView, err := buildInvoiceModalView(meta)
	if err != nil {
		fmt.Printf("failed to build modal: %v\n", err)
		return c.JSON(http.StatusOK, map[string]string{"text": "モーダルの構築に失敗しました"})
	}
	_, err = api.OpenView(slashCommand.TriggerID, modalView)
	if err != nil {
		fmt.Printf("failed to open modal: %v\n", err)
		return c.JSON(http.StatusOK, map[string]string{"text": "モーダルの表示に失敗しました"})
	}
	return c.String(http.StatusOK, "")
}

func (s *ServerImpl) handleSlashHelp(c echo.Context, cmd string) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"response_type": "ephemeral",
		"blocks":        infrastructure.BuildHelpBlocks(cmd),
	})
}

func (s *ServerImpl) handleSlashIssued(c echo.Context, userID string) error {
	webBaseURL := os.Getenv("WEB_BASE_URL")
	invoices, _, err := s.invoiceUseCase.ListInvoices(userID, "", 5, "")
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"text": "請求書の取得に失敗しました"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"response_type": "ephemeral",
		"blocks":        infrastructure.BuildIssuedListBlocks(invoices, webBaseURL),
	})
}

func (s *ServerImpl) handleSlashReceived(c echo.Context, userID string) error {
	webBaseURL := os.Getenv("WEB_BASE_URL")
	invoices, _, err := s.invoiceUseCase.ListReceivedInvoices(userID, "sent", 5, "")
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"text": "請求書の取得に失敗しました"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"response_type": "ephemeral",
		"blocks":        infrastructure.BuildReceivedListBlocks(invoices, webBaseURL),
	})
}

func (s *ServerImpl) handleSlashUnpaid(c echo.Context, userID string) error {
	webBaseURL := os.Getenv("WEB_BASE_URL")
	sent, _, err := s.invoiceUseCase.ListInvoices(userID, "sent", 100, "")
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"text": "請求書の取得に失敗しました"})
	}
	paid, _, err := s.invoiceUseCase.ListInvoices(userID, "paid", 100, "")
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"text": "請求書の取得に失敗しました"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"response_type": "ephemeral",
		"blocks":        infrastructure.BuildUnpaidBlocks(sent, paid, webBaseURL),
	})
}

func (s *ServerImpl) handleSlashSummary(c echo.Context, userID string) error {
	webBaseURL := os.Getenv("WEB_BASE_URL")
	issued, _, err := s.invoiceUseCase.ListInvoices(userID, "", 100, "")
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"text": "請求書の取得に失敗しました"})
	}
	received, _, err := s.invoiceUseCase.ListReceivedInvoices(userID, "", 100, "")
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"text": "請求書の取得に失敗しました"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"response_type": "ephemeral",
		"blocks":        infrastructure.BuildSummaryBlocks(issued, received, webBaseURL),
	})
}

func (s *ServerImpl) handleSlashUnknown(c echo.Context, text string) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"response_type": "ephemeral",
		"blocks":        infrastructure.BuildUnknownCommandBlocks(text),
	})
}

func (s *ServerImpl) SlackInteractionHandler(c echo.Context) error {
	payload := c.FormValue("payload")
	if payload == "" {
		return c.String(http.StatusBadRequest, "missing payload")
	}

	var interaction slack.InteractionCallback
	if err := json.Unmarshal([]byte(payload), &interaction); err != nil {
		return c.String(http.StatusBadRequest, "invalid payload")
	}

	switch interaction.Type {
	case slack.InteractionTypeViewSubmission:
		return s.handleViewSubmission(c, interaction)
	case slack.InteractionTypeBlockActions:
		return s.handleBlockActions(c, interaction)
	default:
		return c.String(http.StatusOK, "")
	}
}

func (s *ServerImpl) handleViewSubmission(c echo.Context, interaction slack.InteractionCallback) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	api := slack.New(slackToken)

	values := interaction.View.State.Values

	// PrivateMetadata を JSON パース（旧形式: 単なるユーザーID文字列にも対応）
	var meta SlackModalMetadata
	if err := json.Unmarshal([]byte(interaction.View.PrivateMetadata), &meta); err != nil {
		meta = SlackModalMetadata{IssuerID: interaction.View.PrivateMetadata, ItemCount: 1}
	}
	if meta.ItemCount <= 0 {
		meta.ItemCount = 1
	}

	// 請求先（複数選択）
	billingUsers := values["billing_user_block"]["billing_user_action"].SelectedUsers

	// 支払期限
	dueDate := values["due_date_block"]["due_date_action"].SelectedDate

	// 振込先
	bankDetails := values["bank_details_block"]["bank_details_action"].Value

	// 明細（複数アイテム）
	validationErrors := map[string]string{}
	items := make([]domain.InvoiceItem, 0, meta.ItemCount)
	today := time.Now().Format(invoiceDateFormat)
	for i := 0; i < meta.ItemCount; i++ {
		descKey := fmt.Sprintf("item_%d_desc_block", i)
		quantityKey := fmt.Sprintf("item_%d_quantity_block", i)
		unitPriceKey := fmt.Sprintf("item_%d_unit_price_block", i)

		desc := values[descKey][fmt.Sprintf("item_%d_desc_action", i)].Value
		quantityStr := values[quantityKey][fmt.Sprintf("item_%d_quantity_action", i)].Value
		unitPriceStr := values[unitPriceKey][fmt.Sprintf("item_%d_unit_price_action", i)].Value

		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			// 全アイテムのエラーを収集してから一度に返すため continue する
			validationErrors[quantityKey] = "数量は数値で入力してください"
			continue
		}
		unitPrice, err := strconv.Atoi(unitPriceStr)
		if err != nil {
			// 全アイテムのエラーを収集してから一度に返すため continue する
			validationErrors[unitPriceKey] = "単価は数値で入力してください"
			continue
		}
		items = append(items, domain.InvoiceItem{
			Date:        today,
			Description: desc,
			Quantity:    quantity,
			UnitPrice:   unitPrice,
			Amount:      quantity * unitPrice,
		})
	}

	if len(validationErrors) > 0 {
		errResp := slack.NewErrorsViewSubmissionResponse(validationErrors)
		return c.JSON(http.StatusOK, errResp)
	}

	// 備考
	additionalInfo := values["additional_info_block"]["additional_info_action"].Value

	// 発行者情報
	issuerSlackUserId := meta.IssuerID
	issuerSlackRealName := interaction.User.Name

	webBaseURL := os.Getenv("WEB_BASE_URL")

	// 請求先ごとに請求書を作成・送付
	var createdInvoiceIds []string
	var failedUsers []string
	for _, billingUser := range billingUsers {
		if billingUser == issuerSlackUserId {
			continue
		}

		invoice := domain.Invoice{
			BillingSlackUserId:  billingUser,
			BillingClientName:   "",
			DueDate:             dueDate,
			BankDetails:         bankDetails,
			AdditionalInfo:      additionalInfo,
			IssuerSlackUserId:   issuerSlackUserId,
			IssuerSlackRealName: issuerSlackRealName,
			Items:               items,
		}

		created, err := s.invoiceUseCase.CreateInvoice(invoice)
		if err != nil {
			fmt.Printf("failed to create invoice for %s: %v\n", billingUser, err)
			failedUsers = append(failedUsers, fmt.Sprintf("<@%s>", billingUser))
			continue
		}

		// ステータスをsentに遷移
		sent, err := s.invoiceUseCase.TransitionStatus(created.InvoiceId, domain.InvoiceStatusSent, issuerSlackRealName, issuerSlackUserId)
		if err != nil {
			fmt.Printf("warning: failed to transition invoice to sent: %v\n", err)
			sent = created
		}

		// 請求先担当者にDM通知（ボタン付き）
		if err := infrastructure.SendInvoiceNotificationWithPayButton(billingUser, sent); err != nil {
			fmt.Printf("warning: failed to send billing user notification DM: %v\n", err)
		}

		createdInvoiceIds = append(createdInvoiceIds, fmt.Sprintf("<%s/invoices/%s|%s>（<@%s>宛）", webBaseURL, sent.InvoiceId, sent.InvoiceId, billingUser))
	}

	var messageParts []string
	if len(createdInvoiceIds) > 0 {
		messageParts = append(messageParts, fmt.Sprintf("請求書を%d件作成・送付しました\n%s", len(createdInvoiceIds), strings.Join(createdInvoiceIds, "\n")))
	}
	if len(failedUsers) > 0 {
		messageParts = append(messageParts, fmt.Sprintf("以下の請求先への請求書作成に失敗しました: %s", strings.Join(failedUsers, ", ")))
	}

	if len(messageParts) == 0 {
		return c.String(http.StatusOK, "")
	}

	// 発行者に作成結果をDM通知
	_, _, err := api.PostMessage(issuerSlackUserId, slack.MsgOptionText(strings.Join(messageParts, "\n\n"), false))
	if err != nil {
		fmt.Printf("failed to send message: %v\n", err)
	}

	return c.String(http.StatusOK, "")
}

func (s *ServerImpl) handleBlockActions(c echo.Context, interaction slack.InteractionCallback) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	api := slack.New(slackToken)

	if len(interaction.ActionCallback.BlockActions) == 0 {
		return c.String(http.StatusOK, "")
	}

	action := interaction.ActionCallback.BlockActions[0]
	invoiceId := action.Value
	changedByUserId := interaction.User.ID
	changedBy := interaction.User.Name

	var targetStatus domain.InvoiceStatus
	var successMessage string

	switch action.ActionID {
	case "add_item_action":
		var meta SlackModalMetadata
		if err := json.Unmarshal([]byte(interaction.View.PrivateMetadata), &meta); err != nil {
			// 旧形式（ユーザーIDのみの文字列）にも対応
			meta = SlackModalMetadata{IssuerID: interaction.View.PrivateMetadata, ItemCount: 1}
		}
		meta.ItemCount++
		updatedModal, err := buildInvoiceModalView(meta)
		if err != nil {
			fmt.Printf("failed to build updated modal: %v\n", err)
			return c.String(http.StatusOK, "")
		}
		_, err = api.UpdateView(updatedModal, "", interaction.View.Hash, interaction.View.ID)
		if err != nil {
			fmt.Printf("failed to update modal: %v\n", err)
		}
		return c.String(http.StatusOK, "")
	case "mark_paid":
		targetStatus = domain.InvoiceStatusPaid
		successMessage = fmt.Sprintf("請求書 %s の支払いを報告しました", invoiceId)
	case "confirm_payment":
		targetStatus = domain.InvoiceStatusConfirmed
		successMessage = fmt.Sprintf("請求書 %s の支払いを承認しました", invoiceId)
	case "reject_payment":
		targetStatus = domain.InvoiceStatusSent
		successMessage = fmt.Sprintf("請求書 %s の支払いを差し戻しました", invoiceId)
	default:
		return c.String(http.StatusOK, "")
	}

	_, err := s.invoiceUseCase.TransitionStatus(invoiceId, targetStatus, changedBy, changedByUserId)
	if err != nil {
		// エフェメラルメッセージでエラー通知
		_, err2 := api.PostEphemeral(
			interaction.Channel.ID,
			changedByUserId,
			slack.MsgOptionText(err.Error(), false),
		)
		if err2 != nil {
			fmt.Printf("warning: failed to send ephemeral error message: %v\n", err2)
		}
		return c.String(http.StatusOK, "")
	}

	// 成功時: 元メッセージを更新してボタンを消去
	_, _, _, err = api.UpdateMessage(
		interaction.Channel.ID,
		interaction.Message.Timestamp,
		slack.MsgOptionText(successMessage, false),
	)
	if err != nil {
		fmt.Printf("warning: failed to update message: %v\n", err)
	}

	return c.String(http.StatusOK, "")
}

type SlackUser struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	RealName     string `json:"real_name"`
	DisplayName  string `json:"display_name"`
	ProfileImage string `json:"profile_image"`
	TeamID       string `json:"team_id"`
}

func (s *ServerImpl) SlackUsersHandler(c echo.Context) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	api := slack.New(slackToken)

	users, err := api.GetUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get users"})
	}

	var result []SlackUser
	for _, u := range users {
		if u.IsBot || u.Deleted {
			continue
		}
		result = append(result, SlackUser{
			ID:           u.ID,
			Name:         u.Name,
			RealName:     u.RealName,
			DisplayName:  u.Profile.DisplayName,
			ProfileImage: u.Profile.Image48,
			TeamID:       u.TeamID,
		})
	}

	return c.JSON(http.StatusOK, result)
}

func domainInvoiceToAPI(inv domain.Invoice) petstore.Invoice {
	status := petstore.InvoiceStatus(inv.Status)
	totalAmount := inv.TotalAmount
	paidAmount := inv.PaidAmount

	apiItems := make([]petstore.InvoiceItem, len(inv.Items))
	for i, item := range inv.Items {
		apiItem := petstore.InvoiceItem{
			Date:        item.Date,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Amount:      item.Amount,
		}
		if item.Memo != "" {
			memo := item.Memo
			apiItem.Memo = &memo
		}
		apiItems[i] = apiItem
	}

	apiHistory := make([]petstore.HistoryEntry, len(inv.History))
	for i, h := range inv.History {
		changedAt := h.ChangedAt
		oldStatus := h.OldStatus
		newStatus := h.NewStatus
		changedBy := h.ChangedBy
		apiHistory[i] = petstore.HistoryEntry{
			ChangedAt: &changedAt,
			OldStatus: &oldStatus,
			NewStatus: &newStatus,
			ChangedBy: &changedBy,
		}
	}

	result := petstore.Invoice{
		InvoiceId:           &inv.InvoiceId,
		Status:              &status,
		BillingClientId:     &inv.BillingClientId,
		BillingClientName:   &inv.BillingClientName,
		BillingSlackUserId:  &inv.BillingSlackUserId,
		TotalAmount:         &totalAmount,
		DueDate:             &inv.DueDate,
		BankDetails:         &inv.BankDetails,
		CreatedAt:           &inv.CreatedAt,
		UpdatedAt:           &inv.UpdatedAt,
		IssuerSlackUserId:   &inv.IssuerSlackUserId,
		IssuerSlackRealName: &inv.IssuerSlackRealName,
		Items:               &apiItems,
		History:             &apiHistory,
	}
	if inv.AdditionalInfo != "" {
		result.AdditionalInfo = &inv.AdditionalInfo
	}
	if inv.PdfUrl != "" {
		result.PdfUrl = &inv.PdfUrl
	}
	if inv.PaidAmount != 0 {
		result.PaidAmount = &paidAmount
	}
	if inv.PaidDate != "" {
		result.PaidDate = &inv.PaidDate
	}
	if inv.PaidMethod != "" {
		result.PaidMethod = &inv.PaidMethod
	}
	return result
}

func formatAmount(amount int) string {
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
