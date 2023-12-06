package main

import (
	"os"
)

type Config struct {
	// Zammad  configuration
	zammadAddress   string
	port            uint
	basicAuthCreats struct {
		username string
		password string
	}
	token                     string
	bearerToken               string
	useTls                    bool
	doNotVerifyTlsCertificate bool

	// Icinga 2 notification data
	hostName    string
	serviceName string // optional if host notification

	checkState  string
	checkOutput string

	author  string
	comment string
	date    string
}

func validate_config(config Config) error {

	return nil
}

func main() {

	config := Config{}

	err := validate_config(config)
	if err != nil {
		print(err)
		os.Exit(1)
	}
}
