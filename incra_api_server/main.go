package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var echoLambda *echoadapter.EchoLambda

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

type Play struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func init() {
	e := echo.New()

	// ミドルウェアの設定
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ルートの設定
	e.GET("/health", healthCheck)

	// ユーザー関連のエンドポイント
	e.GET("/user", getUsers)
	e.POST("/user", createUser)
	e.GET("/user/:id", getUser)
	e.PUT("/user/:id", updateUser)
	e.DELETE("/user/:id", deleteUser)

	// 書籍関連のエンドポイント
	e.GET("/book", getBooks)
	e.POST("/book", createBook)
	e.GET("/book/:id", getBook)
	e.PUT("/book/:id", updateBook)
	e.DELETE("/book/:id", deleteBook)

	// プレイ関連のエンドポイント
	e.GET("/play/:id", getPlay)

	echoLambda = echoadapter.New(e)
}

func main() {
	lambda.Start(Handler)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return echoLambda.ProxyWithContext(ctx, req)
}

// ヘルスチェック
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// ユーザー関連のハンドラー
func getUsers(c echo.Context) error {
	// TODO: データベースからユーザー一覧を取得する実装
	users := []User{
		{ID: "1", Name: "User1"},
		{ID: "2", Name: "User2"},
	}
	return c.JSON(http.StatusOK, users)
}

func createUser(c echo.Context) error {
	user := new(User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	// TODO: データベースにユーザーを保存する実装
	return c.JSON(http.StatusCreated, user)
}

func getUser(c echo.Context) error {
	id := c.Param("id")
	// TODO: データベースからユーザーを取得する実装
	user := User{ID: id, Name: "TestUser"}
	return c.JSON(http.StatusOK, user)
}

func updateUser(c echo.Context) error {
	id := c.Param("id")
	user := new(User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	user.ID = id
	// TODO: データベースのユーザー情報を更新する実装
	return c.JSON(http.StatusOK, user)
}

func deleteUser(c echo.Context) error {
	id := c.Param("id")
	fmt.Println(id)
	// TODO: データベースからユーザーを削除する実装
	return c.JSON(http.StatusNoContent, nil)
}

// 書籍関連のハンドラー
func getBooks(c echo.Context) error {
	// TODO: データベースから書籍一覧を取得する実装
	books := []Book{
		{ID: "1", Title: "Book1", Author: "Author1"},
		{ID: "2", Title: "Book2", Author: "Author2"},
	}
	return c.JSON(http.StatusOK, books)
}

func createBook(c echo.Context) error {
	book := new(Book)
	if err := c.Bind(book); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	// TODO: データベースに書籍を保存する実装
	return c.JSON(http.StatusCreated, book)
}

func getBook(c echo.Context) error {
	id := c.Param("id")
	// TODO: データベースから書籍を取得する実装
	book := Book{ID: id, Title: "TestBook", Author: "TestAuthor"}
	return c.JSON(http.StatusOK, book)
}

func updateBook(c echo.Context) error {
	id := c.Param("id")
	book := new(Book)
	if err := c.Bind(book); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	book.ID = id
	// TODO: データベースの書籍情報を更新する実装
	return c.JSON(http.StatusOK, book)
}

func deleteBook(c echo.Context) error {
	id := c.Param("id")
	fmt.Println(id)
	// TODO: データベースから書籍を削除する実装
	return c.JSON(http.StatusNoContent, nil)
}

// プレイ関連のハンドラー
func getPlay(c echo.Context) error {
	id := c.Param("id")
	// TODO: データベースからプレイ情報を取得する実装
	play := Play{ID: id, Title: "TestPlay", Description: "Test Description"}
	return c.JSON(http.StatusOK, play)
}
