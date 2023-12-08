package cmd

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/NETWAYS/notify_zammad/api"
	"github.com/spf13/cobra"
)

type Config struct {
	zammadGroup    string
	zammadCustomer string

	// Icinga 2 notification data
	hostName    string
	serviceName string // optional if host notification

	checkState  string
	checkOutput string

	notificationType string

	author  string
	comment string
	date    string

	debuglevel int

	// Zammad  configuration
	zammadAddress        string
	basicAuthCredentials struct {
		username string
		password string
	}
	token       string
	bearerToken string
	doNotUseTLS bool
	// doNotVerifyTLSCertificate bool
	port uint16
}

const Copyright = `
Copyright (C) 2022 NETWAYS GmbH <info@netways.de>
`

const License = `
Copyright (C) 2022 NETWAYS GmbH <info@netways.de>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see https://www.gnu.org/licenses/.
`

func (c *Config) ConfigSanityCheck(_ *cobra.Command) error {
	if c.bearerToken == "" && c.token == "" {
		// using basic auth, so both fields must be set
		if c.basicAuthCredentials.username == "" {
			if c.basicAuthCredentials.password == "" {
				return fmt.Errorf("no authentication method provided")
			}

			return fmt.Errorf("no basic authentication username provided")
		}

		if c.basicAuthCredentials.password == "" {
			return fmt.Errorf("no basic authentication password provided")
		}
	}

	return nil
}

func (c *Config) NewClient() (*api.ZammadAPIClient, error) {
	u := url.URL{
		Scheme: "https",
		Host:   c.zammadAddress + ":" + strconv.FormatUint(uint64(c.port), 10),
	}

	if c.doNotUseTLS {
		u.Scheme = "http"
	}

	var rt http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := api.NewClient(u, rt)

	// nolint:gocritic
	if c.bearerToken != "" {
		client.Headers.Add("Authorization", "Bearer "+c.bearerToken)
	} else if c.token != "" {
		client.Headers.Add("Authorization", "Token token="+c.token)
	} else {
		data := []byte(c.basicAuthCredentials.username + ":" + c.basicAuthCredentials.password)
		str := base64.StdEncoding.EncodeToString(data)
		client.Headers.Add("Authorization", "Basic "+str)
	}

	client.Headers.Add("Content-Type", "application/json")

	return client, nil
}
