package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strings"
	"testing"
)

func TestCreateArticleBody(t *testing.T) {
	actual := createArticleBody("foo")
	expected := "<h3>foo</h3>"

	if !strings.Contains(actual, expected) {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}

	expected = "<p>Check State: </p>"
	if !strings.Contains(actual, expected) {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}

func TestNotify_ConnectionRefused(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "--zammad-port", "9999", "--notification-type", "Problem", "--host-name", "foo", "--check-state", "foo", "--check-output", "foo", "--zammad-group", "foo", "--zammad-customer", "foo")
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
		{
			name: "with-wrong-auth",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token := r.Header.Get("Authorization")
				if token == "secret" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))
					return
				}
				w.WriteHeader(http.StatusUnauthorized)
			})),
			args:     []string{"run", "../main.go", "--token", "foo", "--notification-type", "Problem", "--host-name", "Host01", "--service-name", "hostalive", "--check-state", "Down", "--check-output", "CRITICAL - host unreachable", "--zammad-group", "Users", "--zammad-customer", "jon.snow@zammad"},
			expected: "[UNKNOWN] - authentication failed for http://localhost",
		},
		{
			name: "with-wrong-type",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{}`))
			})),
			args:     []string{"run", "../main.go", "--token", "foo", "--notification-type", "NoSuchType", "--host-name", "Host01", "--service-name", "hostalive", "--check-state", "Down", "--check-output", "CRITICAL - host unreachable", "--zammad-group", "Users", "--zammad-customer", "jon.snow@zammad"},
			expected: "[UNKNOWN] - unsupported notification type",
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
