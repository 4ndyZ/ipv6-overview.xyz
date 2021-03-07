package main

import (
	"testing"
)

func TestHTMLAnchorify(t *testing.T) {
	cases := map[string]string{
		"Whitespace Test":    "whitespacetest",
		"Banking":            "banking",
		"Tomorrow.one":       "tomorrowone",
		"Travis-CI":          "travis-ci",
		"Company/Subcompany": "companysubcompany",
		".to":                "to",
		"pläpper":            "plpper",
		"we+you":             "weyou"}

	for totest, expected := range cases {
		result := HTMLAnchorify(totest)

		if expected != result {
			t.Errorf("Expected %q but got %q", expected, result)
		}
	}
}
