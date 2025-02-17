package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	petstore "github.com/yuorei/incra_api_server/api/v1"
	"github.com/yuorei/incra_api_server/src/domain"
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

func (s *ServerImpl) GetInvoiceInvoiceId(ctx echo.Context, invoiceRequest petstore.InvoiceRequest) error {
	_, err := s.invoiceUseCase.GetInvoice(*invoiceRequest.InvoiceId)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, petstore.InvoiceResponse{
		InvoiceId: invoiceRequest.InvoiceId,
	})
}

func (s *ServerImpl) PostInvoice(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, petstore.InvoiceResponse{
		InvoiceId: new(string),
	})
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
	now := time.Now()
	fmt.Println(slashCommand)
	invoice := domain.Invoice{
		AdditionalInfo:       &slashCommand.Text,
		BillingSlackRealName: &slashCommand.UserName,
		BillingSlackUserId:   &slashCommand.UserID,
		InvoiceId:            &slashCommand.Command,
		IssuerSlackRealName:  &slashCommand.UserName,
		IssuerSlackUserId:    &slashCommand.UserID,
		PaidAmount:           new(int),
		PaidDate:             &now,
	}
	res, err := s.invoiceUseCase.CreateInvoice(invoice)
	if err != nil {
		return err
	}
	fmt.Println(res)

	_, _, err = api.PostMessage(slashCommand.ChannelID, slack.MsgOptionText(slashCommand.Text, false))
	if err != nil {
		return err
	}

	return nil
}
