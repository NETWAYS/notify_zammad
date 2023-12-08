package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ZammadTicketState int

const (
	New    ZammadTicketState = 1
	Open   ZammadTicketState = 2
	Closed ZammadTicketState = 4
	// WaitingForClosing ZammadTicketState =
	Undefined = 255
)

func FormatZammadTicketState(s ZammadTicketState) string {
	switch s {
	case New:
		return "new"
	case Open:
		return "open"
	case Closed:
		return "closed"
	default:
		return ""
	}
}

func ParseZammadTicketState(s string) ZammadTicketState {
	switch strings.ToLower(s) {
	case "new":
		return New
	case "open":
		return Open
	case "closed":
		return Closed
	default:
		return Undefined
	}
}

type ZammadTicket struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Group         string `json:"group"`
	Customer      string `json:"customer"`
	IcingaHost    string `json:"icinga_host"`
	IcingaService string `json:"icinga_service"`
}

type ZammadTicketID uint

type ZammadTicketSearchResult struct {
	Tickets      []ZammadTicketID               `json:"tickets"`
	TicketsCount uint                           `json:"tickets_counts"`
	Assets       ZammadTicketSearchResultAssets `json:"assets"`
}

type ZammadTicketSearchResultAssets struct {
	Tickets map[ZammadTicketID]ZammadTicket `json:"Ticket"`
}

type ZammadAPIClient struct {
	Client  http.Client
	URL     url.URL
	Headers http.Header
}

func NewClient(url url.URL, rt http.RoundTripper) *ZammadAPIClient {
	// Small wrapper
	c := &http.Client{
		Transport: rt,
	}

	return &ZammadAPIClient{
		URL:     url,
		Client:  *c,
		Headers: http.Header{},
	}
}

// nolint:interfacer
func (client *ZammadAPIClient) Get(url url.URL) (*http.Response, error) {
	// nolint:noctx
	request, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header = client.Headers

	return client.Client.Do(request)
}

func (client *ZammadAPIClient) searchTicketForHostHelper(icingaHost string) (map[ZammadTicketID]ZammadTicket, error) {
	query := "icinga_host:" + icingaHost + " AND (state_id:" + strconv.Itoa(int(New)) + " OR state_id:" + strconv.Itoa(int(Open)) + ")"
	queryURL := client.URL.JoinPath("/api/v1/tickets/search")

	tmp := queryURL.Query()
	tmp.Set("query", query)

	queryURL.RawQuery = tmp.Encode()

	resp, err := client.Get(*queryURL)
	if err != nil {
		return map[ZammadTicketID]ZammadTicket{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return map[ZammadTicketID]ZammadTicket{}, fmt.Errorf("could not get %s - Error: %d", queryURL, resp.StatusCode)
	}

	searchResult := ZammadTicketSearchResult{}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&searchResult)

	if err != nil {
		return map[ZammadTicketID]ZammadTicket{}, err
	}

	return searchResult.Assets.Tickets, nil
}

func (client *ZammadAPIClient) SearchTicketForHost(icingaHost string) (map[ZammadTicketID]ZammadTicket, error) {
	tmpTickets, err := client.searchTicketForHostHelper(icingaHost)
	if err != nil {
		return map[ZammadTicketID]ZammadTicket{}, err
	}

	result := make(map[ZammadTicketID]ZammadTicket, 0)

	for k, v := range tmpTickets {
		if v.IcingaService == "" {
			result[k] = v
		}
	}

	return result, nil
}

func (client *ZammadAPIClient) SearchTicketForService(icingaHost string, icingaService string) (map[ZammadTicketID]ZammadTicket, error) {
	tmpTickets, err := client.searchTicketForHostHelper(icingaHost)
	if err != nil {
		return map[ZammadTicketID]ZammadTicket{}, err
	}

	result := make(map[ZammadTicketID]ZammadTicket, 0)

	for k, v := range tmpTickets {
		if v.IcingaService == icingaService {
			result[k] = v
		}
	}

	return result, nil
}

type ZammadArticle struct {
	TicketID    ZammadTicketID `json:"ticket_id,omitempty"`
	Subject     string         `json:"subject"`
	Body        string         `json:"body"`
	ContentType string         `json:"content_type"` // "text/html"
	Type        string         `json:"type"`         // "phone"
	Internal    bool           `json:"internal"`     // false
	Sender      string         `json:"sender"`       // "Agent"
	TimeUnit    string         `json:"time_unit"`    // "15"
}

func (client *ZammadAPIClient) AddArticleToTicket(article ZammadArticle) error {
	queryURL := client.URL.JoinPath("/api/v1/ticket_articles")

	b := new(bytes.Buffer)

	err := json.NewEncoder(b).Encode(article)
	if err != nil {
		return err
	}

	data := b.Bytes()

	resp, err := client.Post(*queryURL, &data)
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("could not get %s - Error: %d", queryURL, resp.StatusCode)
	}

	resp.Body.Close()

	return err
}

func (client *ZammadAPIClient) ChangeTicketState(ticketID ZammadTicketID, newState ZammadTicketState) error {
	queryURL := client.URL.JoinPath("/api/v1/tickets/" + strconv.Itoa(int(ticketID)))

	data := []byte("{\"state\": \"" + FormatZammadTicketState(newState) + "\"}")
	bodyReader := bytes.NewReader(data)

	// nolint:noctx
	request, err := http.NewRequest(http.MethodPut, queryURL.String(), bodyReader)
	if err != nil {
		return err
	}

	request.Header = client.Headers

	resp, err := client.Client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not get %s - Error: %d", queryURL, resp.StatusCode)
	}

	return nil
}

type ZammadNewTicket struct {
	Title         string        `json:"title"`
	Group         string        `json:"group"`
	Customer      string        `json:"customer"`
	Article       ZammadArticle `json:"article"`
	IcingaHost    string        `json:"icinga_host"`
	IcingaService string        `json:"icinga_service"`
}

func (client *ZammadAPIClient) CreateTicket(newTicket ZammadNewTicket) error {
	queryURL := client.URL.JoinPath("/api/v1/tickets")

	b := new(bytes.Buffer)

	err := json.NewEncoder(b).Encode(newTicket)
	if err != nil {
		return err
	}

	data := b.Bytes()

	resp, err := client.Post(*queryURL, &data)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("could not get %s - Error: %d", queryURL, resp.StatusCode)
	}

	defer resp.Body.Close()

	return nil
}

// nolint:interfacer
func (client *ZammadAPIClient) Post(url url.URL, data *[]byte) (*http.Response, error) {
	bodyReader := bytes.NewReader(*data)
	// nolint:noctx
	request, err := http.NewRequest(http.MethodPost, url.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	request.Header = client.Headers

	return client.Client.Do(request)
}
