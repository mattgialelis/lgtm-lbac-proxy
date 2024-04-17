package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mattgialelis/lgtm-rbac-proxy/pkg/satokengen"
)

func AuthMiddleware(store *Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the token from the Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing Authorization header")
			}

			// The Authorization header should be in the format "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid Authorization header format")
			}
			token := parts[1]

			decoded, err := satokengen.Decode(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			hash, err := decoded.Hash()
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			// Check if the token is in the store
			user, ok, err := store.Get(hash)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "error checking token")
			}
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			// Load the TenantId and Conditions into the context
			c.Set("KeyData", user)

			return next(c)
		}
	}
}
