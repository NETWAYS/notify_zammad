package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/AlessandroSechi/zammad-go"
	"github.com/NETWAYS/go-check"
	"github.com/spf13/cobra"
)

type Config struct {
	// Zammad  configuration
	zammadAddress        string
	port                 uint16
	basicAuthCredentials struct {
		username string
		password string
	}
	token                     string
	bearerToken               string
	doNotUseTLS               bool
	doNotVerifyTlsCertificate bool

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

func (c *Config) ConfigSanityCheck(cmd *cobra.Command) error {

	//fs := cmd.Flags()

	if c.bearerToken == "" && c.token == "" {
		// using basic auth, so both fields must be set
		if c.basicAuthCredentials.username == "" {
			if c.basicAuthCredentials.password == "" {
				return fmt.Errorf("No authentication method provided")
			}

			return fmt.Errorf("No basic authentication username provided")
		}

		if c.basicAuthCredentials.password == "" {
			return fmt.Errorf("No basic authentication password provided")
		}

	}

	return nil
}

func (c *Config) NewClient() *zammad.Client {
	u := url.URL{
		Scheme: "https",
		Host:   c.zammadAddress + ":" + strconv.FormatUint(uint64(c.port), 10),
	}

	if c.doNotUseTLS {
		u.Scheme = "http"
	}

	client, err := zammad.NewClient(&zammad.Client{
		Username: c.basicAuthCredentials.username,
		Password: c.basicAuthCredentials.password,
		Token:    c.token,
		OAuth:    c.bearerToken,
		Url:      u.String(),
	})

	if err != nil {
		check.ExitError(err)
	}

	return client
}

func (c *Config) timeoutContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
