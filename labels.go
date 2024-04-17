package main

import (
	"fmt"
	"net/url"

	"github.com/grafana/loki/v3/pkg/logql/syntax"
	"github.com/prometheus/prometheus/model/labels"
)

func TenantIdCheck(tenantId string, tenantIds []string) bool {
	for _, id := range tenantIds {
		if tenantId == id {
			return true
		}
	}
	return false
}

func CompareLabels(ToMatch map[string][]labels.Matcher, ToCompare []labels.Matcher) error {

	mustInclude, includeOk := ToMatch["MustInclude"]
	mustExclude, excludeOk := ToMatch["MustExclude"]

	if includeOk {
		for _, matcher := range mustInclude {
			found := false
			for _, compare := range ToCompare {
				if matcher.Name == compare.Name && matcher.Value == compare.Value {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("missing required label: %s=%s", matcher.Name, matcher.Value)
			}
		}
	}

	if excludeOk {
		for _, matcher := range mustExclude {
			for _, compare := range ToCompare {
				if matcher.Name == compare.Name && matcher.Value == compare.Value {
					return fmt.Errorf("found excluded label: %s=%s", matcher.Name, matcher.Value)
				}
			}
		}
	}

	return nil
}

func parseQueryUrl(queryUrl string) ([]labels.Matcher, error) {
	u, err := url.Parse(queryUrl)
	if err != nil {
		return nil, err
	}

	query, err := url.QueryUnescape(u.Query().Get("query"))
	if err != nil {
		return nil, err
	}

	return labelParser(query)
}

func labelParser(query string) ([]labels.Matcher, error) {
	var allLabels []labels.Matcher

	expr, err := syntax.ParseExpr(query)
	if err != nil {
		return nil, err
	}

	expr.Walk(func(e syntax.Expr) {
		switch e := e.(type) {
		case *syntax.MatchersExpr:
			// Do something with the match expression.
			labels, _ := syntax.ParseMatchers(e.String(), false)

			for _, label := range labels {
				allLabels = append(allLabels, *label)
			}

		default:
			// Do something with the other expressions.
		}
	})

	return allLabels, nil
}
