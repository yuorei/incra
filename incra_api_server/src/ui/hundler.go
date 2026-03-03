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

func (s *ServerImpl) SlackSlashsHandler(c echo.Context) error {
	slackToken := os.Getenv("SLACK_TOKEN")
	api := slack.New(slackToken)

	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return err
	}

	billingUserElement := &slack.SelectBlockElement{
		Type:        "users_select",
		Placeholder: slack.NewTextBlockObject(slack.PlainTextType, "担当者を選択", false, false),
		ActionID:    "billing_user_action",
	}
	billingUserSelect := slack.NewInputBlock(
		"billing_user_block",
		slack.NewTextBlockObject(slack.PlainTextType, "請求先担当者", false, false),
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

	itemDescInput := slack.NewInputBlock(
		"item_desc_block",
		slack.NewTextBlockObject(slack.PlainTextType, "品目名", false, false),
		nil,
		slack.NewPlainTextInputBlockElement(
			slack.NewTextBlockObject(slack.PlainTextType, "品目名を入力", false, false),
			"item_desc_action",
		),
	)

	itemQuantityInput := slack.NewInputBlock(
		"item_quantity_block",
		slack.NewTextBlockObject(slack.PlainTextType, "数量", false, false),
		nil,
		slack.NewPlainTextInputBlockElement(
			slack.NewTextBlockObject(slack.PlainTextType, "数量を入力", false, false),
			"item_quantity_action",
		),
	)

	itemUnitPriceInput := slack.NewInputBlock(
		"item_unit_price_block",
		slack.NewTextBlockObject(slack.PlainTextType, "単価", false, false),
		nil,
		slack.NewPlainTextInputBlockElement(
			slack.NewTextBlockObject(slack.PlainTextType, "単価を入力", false, false),
			"item_unit_price_action",
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

	modalView := slack.ModalViewRequest{
		Type:   slack.VTModal,
		Title:  slack.NewTextBlockObject(slack.PlainTextType, "請求書作成", false, false),
		Submit: slack.NewTextBlockObject(slack.PlainTextType, "作成", false, false),
		Close:  slack.NewTextBlockObject(slack.PlainTextType, "キャンセル", false, false),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				billingUserSelect,
				dueDatePicker,
				bankDetailsInput,
				itemDescInput,
				itemQuantityInput,
				itemUnitPriceInput,
				additionalInfoInput,
			},
		},
		PrivateMetadata: slashCommand.UserID,
	}

	_, err = api.OpenView(slashCommand.TriggerID, modalView)
	if err != nil {
		fmt.Printf("failed to open modal: %v\n", err)
		return c.JSON(http.StatusOK, map[string]string{"text": "モーダルの表示に失敗しました"})
	}

	return c.String(http.StatusOK, "")
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

	// 請求先
	billingUser := values["billing_user_block"]["billing_user_action"].SelectedUser

	// 支払期限
	dueDate := values["due_date_block"]["due_date_action"].SelectedDate

	// 振込先
	bankDetails := values["bank_details_block"]["bank_details_action"].Value

	// 明細
	itemDesc := values["item_desc_block"]["item_desc_action"].Value
	quantityStr := values["item_quantity_block"]["item_quantity_action"].Value
	unitPriceStr := values["item_unit_price_block"]["item_unit_price_action"].Value

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		errResp := slack.NewErrorsViewSubmissionResponse(map[string]string{
			"item_quantity_block": "数量は数値で入力してください",
		})
		return c.JSON(http.StatusOK, errResp)
	}

	unitPrice, err := strconv.Atoi(unitPriceStr)
	if err != nil {
		errResp := slack.NewErrorsViewSubmissionResponse(map[string]string{
			"item_unit_price_block": "単価は数値で入力してください",
		})
		return c.JSON(http.StatusOK, errResp)
	}

	amount := quantity * unitPrice

	// 備考
	additionalInfo := values["additional_info_block"]["additional_info_action"].Value

	// 発行者情報
	issuerSlackUserId := interaction.View.PrivateMetadata
	issuerSlackRealName := interaction.User.Name

	invoice := domain.Invoice{
		BillingSlackUserId:  billingUser,
		BillingClientName:   "",
		DueDate:             dueDate,
		BankDetails:         bankDetails,
		AdditionalInfo:      additionalInfo,
		IssuerSlackUserId:   issuerSlackUserId,
		IssuerSlackRealName: issuerSlackRealName,
		Items: []domain.InvoiceItem{
			{
				Date:        time.Now().Format("2006-01-02"),
				Description: itemDesc,
				Quantity:    quantity,
				UnitPrice:   unitPrice,
				Amount:      amount,
			},
		},
	}

	created, err := s.invoiceUseCase.CreateInvoice(invoice)
	if err != nil {
		fmt.Printf("failed to create invoice: %v\n", err)
		return c.String(http.StatusOK, "")
	}

	// ステータスをsentに遷移
	sent, err := s.invoiceUseCase.TransitionStatus(created.InvoiceId, domain.InvoiceStatusSent, issuerSlackRealName, issuerSlackUserId)
	if err != nil {
		fmt.Printf("warning: failed to transition invoice to sent: %v\n", err)
		sent = created
	}

	// 請求先担当者にDM通知（ボタン付き）
	if billingUser != "" {
		if err := infrastructure.SendInvoiceNotificationWithPayButton(billingUser, sent); err != nil {
			fmt.Printf("warning: failed to send billing user notification DM: %v\n", err)
		}
	}

	clientDisplay := sent.BillingClientName
	if clientDisplay == "" {
		clientDisplay = "未指定"
	}
	message := fmt.Sprintf("請求書を作成・送付しました\n• 請求書ID: %s\n• 取引先: %s\n• 合計金額: ¥%s\n• 支払期限: %s",
		sent.InvoiceId, clientDisplay, formatAmount(sent.TotalAmount), sent.DueDate)

	_, _, err = api.PostMessage(issuerSlackUserId, slack.MsgOptionText(message, false))
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
