package client

// TODO Currently contains some duplicated code that could be
// DRY refactored later.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	zammad "github.com/NETWAYS/notify_zammad/internal/api"
)

type Client struct {
	Client  http.Client
	URL     url.URL
	Headers http.Header
}

func NewClient(url url.URL, rt http.RoundTripper) *Client {
	// Small wrapper for the http.Client that we feed with a custom RoundTripper
	c := &http.Client{
		Transport: rt,
	}

	return &Client{
		URL:    url,
		Client: *c,
	}
}

// SearchTickets searches tickets for the given hostname and service.
// If only the hostname is provided all tickets with this hostname are returned,
// if a service is provided only tickets with matching service and hostname are returned.
func (c *Client) SearchTickets(ctx context.Context, hostname, service string) ([]zammad.Ticket, error) {
	query := fmt.Sprintf("icinga_host: %s AND (state.name: new OR state.name: open)", hostname)

	u := c.URL.JoinPath("/api/v1/tickets/search")

	// Add ?search URL parameter with the given query
	search := u.Query()
	search.Set("query", query)
	// The Zammad API returns the tickets sorted by updated_at by default,
	// we use the more stable created_at field.
	// This will return the newest ticket first
	search.Set("sort_by", "created_at")
	search.Set("order_by", "desc")
	u.RawQuery = search.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)

	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("could not search for tickets: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed for %s", c.URL.String())
	}

	var result zammad.TicketSearchResult

	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("unable to parse search results: %w", err)
	}

	// We only care about the tickets, thus we create a slice to easier work with them
	tickets := make([]zammad.Ticket, 0, len(result.Assets.Tickets))

	for _, ticket := range result.Assets.Tickets {
		// If no service is provided we add the ticket and are done
		if service == "" {
			tickets = append(tickets, ticket)
			continue
		}

		// If a service is provided and it is matching the ticket's service
		if service != "" && ticket.IcingaService == service {
			tickets = append(tickets, ticket)
		}
	}

	return tickets, nil
}

// AddArticleToTicket adds an article to an existing ticket
func (c *Client) AddArticleToTicket(ctx context.Context, article zammad.Article) error {
	url := c.URL.JoinPath("/api/v1/ticket_articles")

	data, err := json.Marshal(article)

	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	resp, err := c.Client.Do(req)

	// Retrieve response body since to have details on potential errors
	b, _ := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("could not add article: %w - Error: %s", err, string(b))
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("could not add article: %s - Error: %s", url.String(), string(b))
	}

	defer resp.Body.Close()

	return err
}

// CreateTicket create a new ticket in Zammad
func (c *Client) CreateTicket(ctx context.Context, ticket zammad.Ticket) error {
	url := c.URL.JoinPath("/api/v1/tickets")

	data, err := json.Marshal(ticket)

	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	resp, err := c.Client.Do(req)

	// Retrieve response body since to have details on potential errors
	b, _ := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("could not create ticket: %w - Error: %s", err, string(b))
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("could not create ticket: %s - Error: %s", url.String(), string(b))
	}

	defer resp.Body.Close()

	return err
}

// UpdateTicketState updates the ticket to the given state
func (c *Client) UpdateTicketState(ctx context.Context, ticket zammad.Ticket, state zammad.TicketState) error {
	url := c.URL.JoinPath("/api/v1/tickets", strconv.Itoa(ticket.ID))

	// Set the state field with the given state to be sent to the API
	data := []byte(fmt.Sprintf("{\"state\": \"%s\"}", state))

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url.String(), bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	resp, err := c.Client.Do(req)

	// Retrieve response body since to have details on potential errors
	b, _ := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("could not update ticket: %w - Error: %s", err, string(b))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not update ticket: %s - Error: %s", url.String(), string(b))
	}

	defer resp.Body.Close()

	return err
}
