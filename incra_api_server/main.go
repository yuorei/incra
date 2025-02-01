package main

import (
	"context"
	"os"

	petstore "github.com/yuorei/incra_api_server/api/v1"
	"github.com/yuorei/incra_api_server/src/ui"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var echoLambda *echoadapter.EchoLambda

func init() {
	e := echo.New()

	// ミドルウェアを追加
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS対応を追加
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	// Slackイベントを受け取るエンドポイント
	e.POST("/slack/events", ui.SlackEventsHandler)

	// サーバーの実装インスタンスを作成
	server := &ui.ServerImpl{}

	// generated.go で定義された RegisterHandlers 関数を使用してルートをセットアップ
	petstore.RegisterHandlers(e, server)

	if os.Getenv("LOCAL") == "true" {
		e.Logger.Fatal(e.Start(":8080"))
	}

	echoLambda = echoadapter.New(e)
}

func main() {
	lambda.Start(Handler)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return echoLambda.ProxyWithContext(ctx, req)
}
