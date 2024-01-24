package cmd

import (
	"net/url"
	"testing"
)

func TestConfig(t *testing.T) {
	c := cliConfig.NewClient()

	expected := url.URL{
		Scheme: "http",
		Host:   "localhost:443",
	}

	if c.URL.String() != "http://localhost:443" {
		t.Error("\nActual: ", c.URL.String(), "\nExpected: ", expected.String())
	}
}
