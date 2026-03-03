package ui

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		slackUserId := c.Request().Header.Get("X-Slack-User-Id")
		if slackUserId == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "X-Slack-User-Id header is required"})
		}
		return next(c)
	}
}
