package main

import (
	"context"
	"net/http"

	petstore "github.com/yuorei/incra_api_server/api/v1"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var echoLambda *echoadapter.EchoLambda

// ServerImpl は generated.go で定義されたインターフェイスを実装する構造体です
type ServerImpl struct{}

// GetHello GetGreeting は /hello エンドポイントのハンドラー関数です
func (s *ServerImpl) GetHello(ctx echo.Context) error {
	hello := "Hello, World!"
	return ctx.JSON(http.StatusOK, petstore.Hello{Message: &hello})
}

// GetGoodbye は /goodbye エンドポイントのハンドラー関数です
func (s *ServerImpl) GetGoodbye(ctx echo.Context) error {
	goodbye := "Goodbye, World!"
	return ctx.JSON(http.StatusOK, petstore.Goodbye{Message: &goodbye})
}

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

	// サーバーの実装インスタンスを作成
	server := &ServerImpl{}

	// generated.go で定義された RegisterHandlers 関数を使用してルートをセットアップ
	petstore.RegisterHandlers(e, server)

	echoLambda = echoadapter.New(e)
}

func main() {
	lambda.Start(Handler)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return echoLambda.ProxyWithContext(ctx, req)
}
