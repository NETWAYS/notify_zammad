package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strings"
	"testing"
)

func TestNotify_ConnectionRefused(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "--zammad-port", "9999", "--notification-type", "Problem")
	out, _ := cmd.CombinedOutput()

	actual := string(out)
	expected := "[UNKNOWN] - could not search for tickets: Get \"http://localhost:9999/api/v1/tickets/search"

	if !strings.Contains(actual, expected) {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}

type NotifyTest struct {
	name     string
	server   *httptest.Server
	args     []string
	expected string
}

func TestNotifyZammad(t *testing.T) {
	tests := []NotifyTest{
		{
			name: "with-missing-flags",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{}`))
			})),
			args:     []string{"run", "../main.go"},
			expected: "[UNKNOWN] - required flag",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.server.Close()

			// We need the random Port extracted
			u, _ := url.Parse(test.server.URL)
			cmd := exec.Command("go", append(test.args, "--zammad-port", u.Port())...)
			out, _ := cmd.CombinedOutput()

			actual := string(out)

			if !strings.Contains(actual, test.expected) {
				t.Error("\nActual: ", actual, "\nExpected: ", test.expected)
			}

		})
	}
}
