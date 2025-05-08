package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	zammad "github.com/NETWAYS/notify_zammad/internal/api"
)

func TestUpdateTicketState(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)

		b, _ := io.ReadAll(r.Body)
		actual := string(b)

		if !strings.Contains(actual, "closed") {
			t.Errorf("Expected state closed got: %s", string(b))
		}

		w.Write([]byte(`{}`))
	}))

	defer ts.Close()

	rt := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	u, _ := url.Parse(ts.URL)

	c := NewClient(*u, rt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticket := zammad.Ticket{}

	err := c.UpdateTicketState(ctx, ticket, zammad.ClosedTicketState)

	if err != nil {
		t.Errorf("Did not except error: %v", err)
	}
}

func TestCreateTicket(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)

		b, _ := io.ReadAll(r.Body)
		actual := string(b)

		if !strings.Contains(actual, "MyNewTicket") {
			t.Errorf("Expected new ticket got: %s", string(b))
		}

		w.Write([]byte(`{}`))
	}))

	defer ts.Close()

	rt := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	u, _ := url.Parse(ts.URL)

	c := NewClient(*u, rt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticket := zammad.NewTicket{
		Title: "MyNewTicket",
	}

	err := c.CreateTicket(ctx, ticket)

	if err != nil {
		t.Errorf("Did not except error: %v", err)
	}
}

func TestSearchTickets(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
  {
    "id": 13,
    "group_id": 1,
    "priority_id": 2,
    "state_id": 1,
    "organization_id": null,
    "number": "65012",
    "title": "[Problem] State: Down for Host: MyHost",
    "owner_id": 1,
    "customer_id": 3,
    "note": null,
    "first_response_at": null,
    "first_response_escalation_at": null,
    "first_response_in_min": null,
    "first_response_diff_in_min": null,
    "close_at": null,
    "close_escalation_at": null,
    "close_in_min": null,
    "close_diff_in_min": null,
    "update_escalation_at": null,
    "update_in_min": null,
    "update_diff_in_min": null,
    "last_close_at": null,
    "last_contact_at": null,
    "last_contact_agent_at": null,
    "last_contact_customer_at": null,
    "last_owner_update_at": null,
    "create_article_type_id": 11,
    "create_article_sender_id": 1,
    "article_count": 1,
    "escalation_at": null,
    "pending_time": null,
    "type": null,
    "time_unit": null,
    "preferences": {},
    "updated_by_id": 3,
    "created_by_id": 3,
    "created_at": "2025-05-05T09:38:25.350Z",
    "updated_at": "2025-05-05T09:38:25.418Z",
    "checklist_id": null,
    "icinga_host": "MyHost",
    "icinga_service": "",
    "referencing_checklist_ids": [],
    "article_ids": [
      21
    ],
    "ticket_time_accounting_ids": []
  }
]`))
	}))

	defer ts.Close()

	rt := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	u, _ := url.Parse(ts.URL)

	c := NewClient(*u, rt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickets, err := c.SearchTickets(ctx, "MyHost", "")

	if err != nil {
		t.Errorf("Did not expect error: %v", err)
	}

	if len(tickets) < 1 {
		t.Errorf("Expected test server to return tickets got: %v", tickets)
	}

	if (tickets[0].IcingaHost) != "MyHost" {
		t.Errorf("Expected ticket to contain host got: %v", tickets[0])
	}
}

func TestSearchTicketsWithNoService(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
  {
    "id": 15,
    "group_id": 1,
    "priority_id": 2,
    "state_id": 1,
    "organization_id": null,
    "number": "65014",
    "title": "[Problem] State: Down for Host: MyHost Service: NoSuchService",
    "owner_id": 1,
    "customer_id": 3,
    "note": null,
    "first_response_at": null,
    "first_response_escalation_at": null,
    "first_response_in_min": null,
    "first_response_diff_in_min": null,
    "close_at": null,
    "close_escalation_at": null,
    "close_in_min": null,
    "close_diff_in_min": null,
    "update_escalation_at": null,
    "update_in_min": null,
    "update_diff_in_min": null,
    "last_close_at": null,
    "last_contact_at": null,
    "last_contact_agent_at": null,
    "last_contact_customer_at": null,
    "last_owner_update_at": null,
    "create_article_type_id": 11,
    "create_article_sender_id": 1,
    "article_count": 3,
    "escalation_at": null,
    "pending_time": null,
    "type": null,
    "time_unit": null,
    "preferences": {},
    "updated_by_id": 3,
    "created_by_id": 3,
    "created_at": "2025-05-05T12:52:36.650Z",
    "updated_at": "2025-05-05T13:08:29.288Z",
    "checklist_id": null,
    "icinga_host": "MyHost",
    "icinga_service": "NoSuchService",
    "referencing_checklist_ids": [],
    "article_ids": [
      33,
      34,
      35
    ],
    "ticket_time_accounting_ids": []
  }
]`))
	}))

	defer ts.Close()

	rt := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	u, _ := url.Parse(ts.URL)

	c := NewClient(*u, rt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickets, err := c.SearchTickets(ctx, "MyHost", "MyService")

	if err != nil {
		t.Errorf("Did not except error: %v", err)
	}

	if len(tickets) != 0 {
		t.Errorf("Expected to return no tickets got: %v", tickets)
	}
}

func TestSearchTicketsWithService(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
  {
    "id": 16,
    "group_id": 1,
    "priority_id": 2,
    "state_id": 1,
    "organization_id": null,
    "number": "65015",
    "title": "[Problem] State: Down for Host: MyHost Service: MyService",
    "owner_id": 1,
    "customer_id": 3,
    "note": null,
    "first_response_at": null,
    "first_response_escalation_at": null,
    "first_response_in_min": null,
    "first_response_diff_in_min": null,
    "close_at": null,
    "close_escalation_at": null,
    "close_in_min": null,
    "close_diff_in_min": null,
    "update_escalation_at": null,
    "update_in_min": null,
    "update_diff_in_min": null,
    "last_close_at": null,
    "last_contact_at": null,
    "last_contact_agent_at": null,
    "last_contact_customer_at": null,
    "last_owner_update_at": null,
    "create_article_type_id": 11,
    "create_article_sender_id": 1,
    "article_count": 1,
    "escalation_at": null,
    "pending_time": null,
    "type": null,
    "time_unit": null,
    "preferences": {},
    "updated_by_id": 3,
    "created_by_id": 3,
    "created_at": "2025-05-05T13:46:37.651Z",
    "updated_at": "2025-05-05T13:46:37.733Z",
    "checklist_id": null,
    "icinga_host": "MyHost",
    "icinga_service": "MyService",
    "referencing_checklist_ids": [],
    "article_ids": [
      36
    ],
    "ticket_time_accounting_ids": []
  }
]`))
	}))

	defer ts.Close()

	rt := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	u, _ := url.Parse(ts.URL)

	c := NewClient(*u, rt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickets, err := c.SearchTickets(ctx, "MyHost", "MyService")

	if err != nil {
		t.Errorf("Did not except error: %v", err)
	}

	if (tickets[0].IcingaService) != "MyService" {
		t.Errorf("Expected ticket to contain service got: %v", tickets[0])
	}
}

func TestAddArticleToTicket(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)

		b, _ := io.ReadAll(r.Body)
		actual := string(b)

		if !strings.Contains(actual, "1337") {
			t.Errorf("Expected new ticket got: %s", string(b))
		}

		w.Write([]byte(`{}`))
	}))

	defer ts.Close()

	rt := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	u, _ := url.Parse(ts.URL)

	c := NewClient(*u, rt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := zammad.Article{
		TicketID: 1337,
		Subject:  "Acknowledgement",
	}

	err := c.AddArticleToTicket(ctx, a)

	if err != nil {
		t.Errorf("Did not except error: %v", err)
	}
}
