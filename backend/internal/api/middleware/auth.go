package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Auth(enabled bool, jwksURL string) echo.MiddlewareFunc {
	if !enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Inject mock user claims when auth is disabled
				claims := jwt.MapClaims{"sub": "guest-user", "name": "Guest"}
				c.Set("user", claims)
				ctx := context.WithValue(c.Request().Context(), "user", claims)
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			}
		}
	}

	// Initialize keyfunc
	k, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		panic(fmt.Sprintf("Failed to create keyfunc: %v", err))
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization Header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization Header")
			}

			tokenStr := parts[1]
			token, err := jwt.Parse(tokenStr, k.Keyfunc)

			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Invalid Token: %v", err))
			}

			// Add claims to context
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("user", claims)
				// Inject into request context for services
				ctx := context.WithValue(c.Request().Context(), "user", claims)
				c.SetRequest(c.Request().WithContext(ctx))
			}

			return next(c)
		}
	}
}
