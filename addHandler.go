package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mattgialelis/lgtm-rbac-proxy/pkg/satokengen"
	"github.com/sirupsen/logrus"
)

// Example Type
//
// {
// "name": "exampleName",
// "tenantIds": ["tenant1", "tenant2"],

// 	"allowedLabels": {
// 		"MustInclude": "{projectId=\"mg-infra-monitoring\", container=\"querier\"}",
// 		"MustExclude": "{projectId=\"mg-bingomysteries-rc-rs-stg\", cluster=\"k8s-1\"}"
// 	}

// }
type TokenRequest struct {
	Name          string            `json:"name"`
	TenantIds     []string          `json:"tenantIds"`
	AllowedLabels map[string]string `json:"allowedLabels"`
}

func createHandler(c echo.Context, store *Store) error {
	var keyData KeyData

	req := new(TokenRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Error: Invalid request body")
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, "Error: name is required")
	}

	if len(req.TenantIds) == 0 {
		return c.JSON(http.StatusBadRequest, "Error: At least one tenantId is required")
	}

	if len(req.AllowedLabels) == 0 {
		return c.JSON(http.StatusBadRequest, "Error: At least one allowedLabel is required")
	}

	keygenRes, err := satokengen.New("rbac")
	if err != nil {
		logrus.Info(err.Error())
	}

	keyData.Name = req.Name
	keyData.TenantIds = req.TenantIds
	keyData.AllowedLabels = make(map[string]string)
	keyData.Token = keygenRes.HashedKey

	for name, value := range req.AllowedLabels {
		_, err := labelParser(value)
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Error: Invalid label")
		}

		keyData.AllowedLabels[name] = value
	}

	err = store.Add(keygenRes.HashedKey, keyData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error: Failed to store key")
	}

	response := TokenResponse{
		Token: keygenRes.ClientSecret,
	}

	logrus.WithFields(logrus.Fields{
		"key": keyData.Name,
	}).Info("Key created")

	return c.JSON(http.StatusOK, response)
}
