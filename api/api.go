package api

type Ticket struct {
	Title         string `json:"title"`
	Group         string `json:"group"`
	Customer      string `json:"customer"`
	IcingaHost    string `json:"icinga_host"`
	IcingaService string `json:"icinga_service"`
}
