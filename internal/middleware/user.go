package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

func UserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		var accessToken string

		accessTokenHeader := c.Request().Header.Get("Authorization")

		accessToken = strings.TrimPrefix(accessTokenHeader, "Bearer ")

		if accessToken == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "access token is required",
			})
		}

		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("SECRET_KEY")), nil
		}, jwt.WithValidMethods([]string{"HS256"}))

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid access token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid access token",
			})
		}

		sub, err := claims.GetSubject()
		if err != nil || sub == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid access token",
			})
		}

		userID, err := uuid.Parse(sub)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid access token",
			})
		}

		c.Set("user_id", userID)
		return next(c)
	}
}
