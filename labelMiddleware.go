package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/sirupsen/logrus"
)

func labelMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var MatchLabels map[string][]labels.Matcher
		MatchLabels = make(map[string][]labels.Matcher)
		// Access the values set in the context by authMiddleware
		KeyData := c.Get("KeyData").(KeyData)

		for name, value := range KeyData.AllowedLabels {
			labels, err := labelParser(value)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"key":   KeyData.Name,
				}).Errorf("Error parsing labels")
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error parsing labels: %v", err))

			}
			MatchLabels[name] = labels
		}

		url := c.Request().URL

		// Pass the URL to the extractUrlsLabels function
		urlsLabels, err := parseQueryUrl(url.String())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"key":   KeyData.Name,
			}).Errorf("Error extracting URLs and Labels")
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error extracting URLs and Labels: %v", err))
		}

		err = CompareLabels(MatchLabels, urlsLabels)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"key":   KeyData.Name,
			}).Errorf("Error comparing labels")
			return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("error comparing labels: %v", err))
		}

		logrus.Info("Labels match %+v", urlsLabels)
		return next(c)
	}
}
