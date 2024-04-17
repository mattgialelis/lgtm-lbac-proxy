package main

import (
	"context"
	"crypto/subtle"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type TokenResponse struct {
	Token string
}

func main() {
	adminPassword := os.Getenv("ADMIN_USER_PASSWORD")
	if adminPassword == "" {
		logrus.Fatal("ADMIN_USER_PASSWORD is required")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		logrus.Fatal("DB_PATH is required")
	}

	config, err := LoadConfig()
	if err != nil {
		logrus.Error(err)
	}

	if config.logType == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	store, err := NewStore(dbPath)
	if err != nil {
		logrus.Fatal(err)
	}

	e := echo.New()

	basicAuthGroup := e.Group("")

	basicAuthGroup.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// Be careful to use constant time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(username), []byte(config.AdminUser.Username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(adminPassword)) == 1 {
			return true, nil
		}
		return false, nil
	}))

	basicAuthGroup.POST("/create", func(c echo.Context) error {
		return createHandler(c, store)
	})

	basicAuthGroup.GET("/tokens", func(c echo.Context) error {
		return getTokensHandler(c, store)
	})

	// Create a group of routes that require authentication
	authGroupToken := e.Group("")
	authGroupToken.Use(AuthMiddleware(store))
	// AUTHENTICATED ROUTES
	authGroupToken.Any("/loki/*", func(c echo.Context) error {
		return lokiReverseProxyHandler(c, config)
	}, labelMiddleware)

	// Start the server in a goroutine so that it doesn't block the signal listening
	go func() {
		if err := e.Start(":8080"); err != nil {
			logrus.Info("Shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1) // Add buffer size of 1
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := e.Shutdown(ctx); err != nil {
		logrus.Fatal(err)
	}

	// Close the store
	if err := store.Close(); err != nil {
		logrus.Fatal(err)
	}

}

func getTokensHandler(c echo.Context, store *Store) error {
	tokens, err := store.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get tokens: " + err.Error()})
	}
	return c.JSON(http.StatusOK, tokens)
}

func lokiReverseProxyHandler(c echo.Context, config *Config) error {
	user := c.Get("KeyData").(KeyData)

	req := c.Request()
	res := c.Response()

	tennatHeader := req.Header.Get("X-Scope-OrgID")
	if tennatHeader == "" {
		logrus.WithFields(logrus.Fields{
			"user": user.Name,
		}).Error("Missing X-Scope-OrgID header")

		return c.String(http.StatusBadRequest, "Missing X-Scope-OrgID header")
	}

	matchingTennat := TenantIdCheck(tennatHeader, user.TenantIds)
	if !matchingTennat {
		logrus.WithFields(logrus.Fields{
			"user": user.Name,
		}).Errorf("TenantId does not match, access denied. Requested TenantId: %s", tennatHeader)

		return c.String(http.StatusForbidden, "TenantId does not match user access")
	}

	target, _ := url.Parse(config.LokiURL)

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Set ErrorHandler to return a custom error message
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"user":  user.Name,
		}).Error("An error occurred")

		c.String(http.StatusBadGateway, "An error occurred while proxying the request.")
	}

	// Update the headers to allow for SSL redirection
	req.URL.Host = target.Host
	req.URL.Scheme = target.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))

	// Forward X-Scope-OrgID header
	if orgID := tennatHeader; orgID != "" {
		req.Header.Set("X-Scope-OrgID", orgID)
	}

	req.Host = target.Host

	// Add Basic Auth
	if config.LokiBasicAuth.Enabled {
		req.SetBasicAuth(config.LokiBasicAuth.Username, config.LokiBasicAuth.Password)
	}

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
	return nil
}
