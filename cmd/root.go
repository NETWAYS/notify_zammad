package cmd

import (
	"os"

	"github.com/NETWAYS/go-check"
	"github.com/spf13/cobra"
)

var Timeout = 30

var config Config

var rootCmd = &cobra.Command{
	Use:   "notify_zammad",
	Short: "An Icinga notification plugin for Zammad",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		go check.HandleTimeout(Timeout)
	},
	Run: Usage,
}

func Execute(version string) {
	defer check.CatchPanic()

	rootCmd.Version = version
	rootCmd.VersionTemplate()

	if err := rootCmd.Execute(); err != nil {
		check.ExitError(err)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.DisableAutoGenTag = true

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	pfs := rootCmd.PersistentFlags()
	pfs.StringVar(&config.zammadAddress, "address", "",
		"Address of the Zammad instance to be used")
	cobra.MarkFlagRequired(pfs, "address")
	pfs.Uint16Var(&config.port, "port", 443,
		"Port of the Prometheus server")
	pfs.StringVar(&config.basicAuthCredentials.username, "basic-auth-user", "",
		"Username for basic authentication")
	pfs.StringVar(&config.basicAuthCredentials.password, "basic-auth-password", "",
		"Password for basic authentication")
	pfs.StringVar(&config.token, "token", "",
		"Token string for HTTP Token Authentication")
	pfs.StringVar(&config.bearerToken, "oauth2-token", "",
		"Token string for OAuth2 authentication")
	//pfs.BoolVar(&config.doNotUseTLS, "no-tls", false,
	//	"Use plain HTTP to connect instead of HTTPS")
	//pfs.BoolVar(&config.doNotVerifyTlsCertificate, "no-certificate-verification", false,
	//	"Token string for OAuth2 authentication")

	rootCmd.Flags().SortFlags = false
	pfs.SortFlags = false
}

func Usage(cmd *cobra.Command, _ []string) {
	_ = cmd.Usage()

	os.Exit(3)
}
