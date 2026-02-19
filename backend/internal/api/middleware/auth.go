package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func KeycloakMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(401, "Missing Authorization Header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.NewHTTPError(401, "Invalid Authorization Header")
		}

		// TODO: Validate token against Keycloak JWKS
		// token := parts[1]

		return next(c)
	}
}
