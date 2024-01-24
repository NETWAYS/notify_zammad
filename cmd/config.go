package cmd

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	checkhttpconfig "github.com/NETWAYS/go-check-network/http/config"
	"github.com/NETWAYS/notify_zammad/internal/client"
)

type Config struct {
	BasicAuth string `env:"NOTIFY_ZAMMAD_BASICAUTH"`
	Token     string `env:"NOTIFY_ZAMMAD_TOKEN"`
	CAFile    string `env:"NOTIFY_ZAMMAD_CA_FILE"`
	CertFile  string `env:"NOTIFY_ZAMMAD_CERT_FILE"`
	KeyFile   string `env:"NOTIFY_ZAMMAD_KEY_FILE"`
	Hostname  string `env:"NOTIFY_ZAMMAD_HOSTNAME"`

	ZammadGroup            string
	ZammadCustomer         string
	IcingaHostname         string
	IcingaServiceName      string
	IcingaCheckState       string
	IcingaCheckOutput      string
	IcingaNotificationType string
	IcingaAuthor           string
	IcingaComment          string
	IcingaDate             string

	Port int

	Insecure bool
	Secure   bool
}

var cliConfig Config

const Copyright = `
Copyright (C) 2024 NETWAYS GmbH <info@netways.de>
`

const License = `
Copyright (C) 2024 NETWAYS GmbH <info@netways.de>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see https://www.gnu.org/licenses/.
`

func (c *Config) NewClient() *client.Client {
	u := url.URL{
		Scheme: "http",
		Host:   c.Hostname + ":" + strconv.Itoa(c.Port),
	}

	if c.Secure {
		u.Scheme = "https"
	}

	// Create TLS configuration for default RoundTripper
	tlsConfig, err := checkhttpconfig.NewTLSConfig(&checkhttpconfig.TLSConfig{
		InsecureSkipVerify: c.Insecure,
		CAFile:             c.CAFile,
		KeyFile:            c.KeyFile,
		CertFile:           c.CertFile,
	})

	if err != nil {
		fmt.Println("error creating TLS configuration", err)
		os.Exit(1)
	}

	var rt http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
	}

	// Using a Bearer Token for authentication
	if c.Token != "" {
		rt = checkhttpconfig.NewAuthorizationCredentialsRoundTripper("Token", c.Token, rt)
	}

	// Using a BasicAuth for authentication
	if c.BasicAuth != "" {
		s := strings.Split(c.BasicAuth, ":")
		if len(s) != 2 {
			fmt.Println("specify the user name and password for server authentication <user:password>", err)
			os.Exit(1)
		}

		var u = s[0]

		var p = s[1]

		rt = checkhttpconfig.NewBasicAuthRoundTripper(u, p, rt)
	}

	return client.NewClient(u, rt)
}
