package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type ZammadTicketState int

const (
	New    ZammadTicketState = 0
	Open   ZammadTicketState = 1
	Closed ZammadTicketState = 4
	// WaitingForClosing ZammadTicketState =
)

type ZammadTicket struct {
	Id            int    `json:"id"`
	Title         string `json:"title"`
	Group         string `json:"group"`
	Customer      string `json:"customer"`
	IcingaHost    string `json:"icinga_host"`
	IcingaService string `json:"icinga_service"`
}

type ZammadTicketId uint

type ZammadTicketSearchResult struct {
	Tickets      []ZammadTicketId               `json:"tickets"`
	TicketsCount uint                           `json:"tickets_counts"`
	Assets       ZammadTicketSearchResultAssets `json:"assets"`
}

type ZammadTicketSearchResultAssets struct {
	Tickets map[ZammadTicketId]ZammadTicket `json:"Ticket"`
}

type ZammadApiClient struct {
	Client  http.Client
	URL     url.URL
	Ctx     context.Context
	Headers http.Header
}

func NewClient(url url.URL, rt http.RoundTripper) *ZammadApiClient {

	// Small wrapper
	c := &http.Client{
		Transport: rt,
	}

	return &ZammadApiClient{
		URL:     url,
		Client:  *c,
		Headers: http.Header{},
	}
}

func (client *ZammadApiClient) Get(url url.URL) (*http.Response, error) {

	request, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header = client.Headers

	return client.Client.Do(request)

}

func (client *ZammadApiClient) searchTicketForHostHelper(icingaHost string) (map[ZammadTicketId]ZammadTicket, error) {
	query := "icinga_host:" + icingaHost + " AND (state_id:" + strconv.Itoa(int(New)) + " OR state_id:" + strconv.Itoa(int(Open)) + ")"
	queryUrl := client.URL.JoinPath("/api/v1/tickets/search")

	tmp := queryUrl.Query()
	tmp.Set("query", query)

	queryUrl.RawQuery = tmp.Encode()

	resp, err := client.Get(*queryUrl)
	if err != nil {
		return map[ZammadTicketId]ZammadTicket{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return map[ZammadTicketId]ZammadTicket{}, fmt.Errorf("could not get %s - Error: %d", queryUrl, resp.StatusCode)
	}

	searchResult := ZammadTicketSearchResult{}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&searchResult)

	if err != nil {
		return map[ZammadTicketId]ZammadTicket{}, err
	}

	return searchResult.Assets.Tickets, nil
}

func (client *ZammadApiClient) SearchTicketForHost(icingaHost string) (map[ZammadTicketId]ZammadTicket, error) {
	tmpTickets, err := client.searchTicketForHostHelper(icingaHost)
	if err != nil {
		return map[ZammadTicketId]ZammadTicket{}, err
	}

	result := make(map[ZammadTicketId]ZammadTicket, 0)

	for k, v := range tmpTickets {
		if v.IcingaService == "" {
			result[k] = v
		}
	}

	return result, nil
}

func (client *ZammadApiClient) SearchTicketForService(icingaHost string, icingaService string) (map[ZammadTicketId]ZammadTicket, error) {
	tmpTickets, err := client.searchTicketForHostHelper(icingaHost)
	if err != nil {
		return map[ZammadTicketId]ZammadTicket{}, err
	}

	result := make(map[ZammadTicketId]ZammadTicket, 0)

	for k, v := range tmpTickets {
		if v.IcingaService == icingaService {
			result[k] = v
		}
	}

	return result, nil
}

type ZammadArticle struct {
	TicketId    ZammadTicketId `json:"ticket_id"`
	Subject     string         `json:"subject"`
	Body        string         `json:"body"`
	ContentType string         `json:"content_type"` // "text/html"
	Type        string         `json:"type"`         // "phone"
	Internal    bool           `json:"internal"`     // false
	Sender      string         `json:"sender"`       // "Agent"
	TimeUnit    string         `json:"time_unit"`    // "15"
}

func (client *ZammadApiClient) AddArticleToTicket(article ZammadArticle) error {
	queryUrl := client.URL.JoinPath("/api/v1/ticket_articles")

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(article)
	if err != nil {
		return err
	}

	data := b.Bytes()

	_, err = client.Post(*queryUrl, &data)

	return err
}

func (client *ZammadApiClient) ChangeTicketState(ticketId ZammadTicketId, newState ZammadTicketState) error {
	queryUrl := client.URL.JoinPath("/api/v1/tickets/" + strconv.Itoa(int(ticketId)))

	request, err := http.NewRequest("PUT", queryUrl.String(), nil)
	if err != nil {
		return err
	}

	readCloser := io.NopCloser(bytes.NewReader([]byte("{state: " + strconv.Itoa(int(Closed)) + "}")))

	request.Body = readCloser
	request.Header = client.Headers

	_, err = client.Client.Do(request)

	return err
}

type ZammadNewTicket struct {
	Title         string        `json:"title"`
	Group         string        `json:"group"`
	Customer      string        `json:"customer"`
	Article       ZammadArticle `json:"article"`
	IcingaHost    string        `json:"icinga_host"`
	IcingaService string        `json:"icinga_service"`
}

func (client *ZammadApiClient) CreateTicket(newTicket ZammadNewTicket) error {
	queryUrl := client.URL.JoinPath("/api/v1/tickets/create")

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(newTicket)
	if err != nil {
		return err
	}

	data := b.Bytes()

	_, err = client.Post(*queryUrl, &data)
	if err != nil {
		return err
	}

	return nil
}

func (client *ZammadApiClient) Post(url url.URL, data *[]byte) (*http.Response, error) {

	request, err := http.NewRequest("POST", url.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header = client.Headers

	return client.Client.Do(request)

}
