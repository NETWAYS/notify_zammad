package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-icingadsl"
	"github.com/spf13/cobra"

	zammad "github.com/NETWAYS/notify_zammad/internal/api"
	"github.com/NETWAYS/notify_zammad/internal/client"
)

// Timeout is the default timout for the plugin
var Timeout = 30

var rootCmd = &cobra.Command{
	Use: "notify_zammad",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		go check.HandleTimeout(Timeout)
	},
	Short: "An Icinga notification plugin for Zammad",
	Long:  "An Icinga notification plugin for Zammad",
	Run:   sendNotification,
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
	// Configuration for the connection
	pfs.StringVarP(&cliConfig.Hostname, "zammad-hostname", "H", "localhost",
		"Address of the Zammad instance (NOTIFY_ZAMMAD_HOSTNAME)")
	pfs.IntVarP(&cliConfig.Port, "zammad-port", "p", 443,
		"Port of the Zammad instance")
	pfs.BoolVarP(&cliConfig.Secure, "secure", "s", false,
		"Use a HTTPS connection")
	pfs.StringVarP(&cliConfig.Token, "token", "T", "",
		"Token for server authentication (NOTIFY_ZAMMAD_TOKEN)")
	pfs.StringVarP(&cliConfig.BasicAuth, "user", "u", "",
		"Specify the user name and password for server authentication <user:password> (NOTIFY_ZAMMAD_BASICAUTH)")
	pfs.StringVarP(&cliConfig.CAFile, "ca-file", "", "",
		"Specify the CA File for TLS authentication (NOTIFY_ZAMMAD_CA_FILE)")
	pfs.StringVarP(&cliConfig.CertFile, "cert-file", "", "",
		"Specify the Certificate File for TLS authentication (NOTIFY_ZAMMAD_CERT_FILE)")
	pfs.StringVarP(&cliConfig.KeyFile, "key-file", "", "",
		"Specify the Key File for TLS authentication (NOTIFY_ZAMMAD_KEY_FILE)")
	pfs.BoolVarP(&cliConfig.Insecure, "insecure", "i", false,
		"Skip the verification of the server's TLS certificate")
	pfs.IntVarP(&Timeout, "timeout", "t", Timeout,
		"Timeout in seconds for the plugin")

	rootCmd.MarkFlagsMutuallyExclusive("user", "token")

	// Configuration for the notification
	pfs.StringVar(&cliConfig.IcingaHostname, "host-name", "",
		"Host name of the Icinga 2 Host object")
	pfs.StringVar(&cliConfig.IcingaServiceName, "service-name", "",
		"Service name of the Icinga 2 Service Object (optional for Host Notifications)")
	pfs.StringVar(&cliConfig.IcingaCheckState, "check-state", "",
		"State of the Object (Up/Down for hosts, OK/Warning/Critical/Unknown for services)")
	pfs.StringVar(&cliConfig.IcingaCheckOutput, "check-output", "",
		"Output of the last executed check")
	pfs.StringVar(&cliConfig.IcingaNotificationType, "notification-type", "",
		"Type of the notification (Problem/Recovery/Acknowledgement)")
	pfs.StringVar(&cliConfig.IcingaAuthor, "notification-author", "",
		"Name of an author for manual events")
	pfs.StringVar(&cliConfig.IcingaComment, "notification-comment", "",
		"Comment for manual events")
	pfs.StringVar(&cliConfig.IcingaDate, "notification-date", "",
		"Date when the event occurred")
	pfs.StringVar(&cliConfig.ZammadGroup, "zammad-group", "",
		"Custom Zammad Field for the group")
	pfs.StringVar(&cliConfig.ZammadCustomer, "zammad-customer", "",
		"Custom Zammad Field for the customer")

	_ = cobra.MarkFlagRequired(pfs, "notification-type")
	_ = cobra.MarkFlagRequired(pfs, "host-name")
	_ = cobra.MarkFlagRequired(pfs, "check-state")
	_ = cobra.MarkFlagRequired(pfs, "check-output")
	_ = cobra.MarkFlagRequired(pfs, "zammad-group")
	_ = cobra.MarkFlagRequired(pfs, "zammad-customer")

	rootCmd.Flags().SortFlags = false
	pfs.SortFlags = false
}

// sendNotification is the cobra.Command that is executed
func sendNotification(_ *cobra.Command, _ []string) {
	notificationType, err := icingadsl.ParseNotificationType(cliConfig.IcingaNotificationType)

	if err != nil {
		check.ExitError(err)
	}

	// Creating an client and connecting to the API
	c := cliConfig.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(Timeout)*time.Second)
	defer cancel()

	// Search for existing Tickets
	tickets, err := c.SearchTickets(ctx, cliConfig.IcingaHostname, cliConfig.IcingaServiceName)

	if err != nil {
		check.ExitError(err)
	}

	var ticket zammad.Ticket

	var notificationErr error

	if len(tickets) > 0 {
		// Using the first ticket found for the notification,
		// the SearchTickets methods returns the tickets by created_at.
		// If no ticket is found the zammad.Ticket type will be empty,
		// which can be used to detect if a new ticket needs to be created.
		ticket = tickets[0]
	}

	switch notificationType {
	case icingadsl.Custom:
		// If ticket exists, adds message to existing ticket
		notificationErr = handleCustomNotification(ctx, c, ticket)
	case icingadsl.Acknowledgement:
		// If ticket exists, adds message to existing ticket
		notificationErr = handleAcknowledgeNotification(ctx, c, ticket)
	case icingadsl.Problem:
		// Opens a new ticket if none exists
		// If one exists, adds message to existing ticket
		notificationErr = handleProblemNotification(ctx, c, ticket)
	case icingadsl.Recovery:
		// Closes a ticket if one exists
		// If ticket is open, adds message to existing ticket
		// If ticket is closed, reopens the ticket with the message
		notificationErr = handleRecoveryNotification(ctx, c, ticket)
	case icingadsl.DowntimeStart:
		// Currently no implemented
	case icingadsl.DowntimeEnd:
		// Currently no implemented
	case icingadsl.DowntimeRemoved:
		// Currently no implemented
	case icingadsl.FlappingStart:
		// Currently no implemented
	case icingadsl.FlappingEnd:
		// Currently no implemented
	default:
		check.ExitError(fmt.Errorf("unsupported notification type"))
	}

	if notificationErr != nil {
		check.ExitError(notificationErr)
	}

	check.ExitRaw(check.OK, "")
}

// handleProblemNotification opens a new ticket if none exists,
// If one exists, adds message to existing ticket.
func handleProblemNotification(ctx context.Context, c *client.Client, ticket zammad.Ticket) error {
	var err error

	a := zammad.Article{
		TicketID:    ticket.ID,
		Subject:     "Problem",
		Body:        fmt.Sprintf("%s %s", cliConfig.IcingaCheckState, cliConfig.IcingaCheckOutput),
		ContentType: "text/html",
		Type:        "web",
		Internal:    true,
		Sender:      "Agent",
	}

	// If a Zammad Ticket exists, add the article to this ticket.
	if ticket.ID != 0 {
		err = c.AddArticleToTicket(ctx, a)
		return err
	}

	// Open a new Ticket with the given data
	var title strings.Builder

	title.WriteString("[Problem] ")

	title.WriteString(fmt.Sprintf("State: %s for", cliConfig.IcingaCheckState))
	title.WriteString(fmt.Sprintf(" Host: %s", cliConfig.IcingaHostname))

	if cliConfig.IcingaServiceName != "" {
		title.WriteString(fmt.Sprintf(" Service: %s", cliConfig.IcingaServiceName))
	}

	ticket.Title = title.String()
	ticket.Group = cliConfig.ZammadGroup
	ticket.Customer = cliConfig.ZammadCustomer
	ticket.IcingaHost = cliConfig.IcingaHostname
	ticket.IcingaService = cliConfig.IcingaServiceName
	ticket.Article = a

	err = c.CreateTicket(ctx, ticket)

	return err
}

// handleAcknowledgeNotification adds a new article to an existing ticket
// If the ticket is in state new, it will be set to state open
// If no ticket exists an error is returned
func handleAcknowledgeNotification(ctx context.Context, c *client.Client, ticket zammad.Ticket) error {
	// If no Zammad Ticket exists, we cannot add an article and thus return an error
	// and notify the user
	if ticket.ID == 0 {
		return fmt.Errorf("no open or new ticket found to add acknowledgement")
	}

	a := zammad.Article{
		TicketID:    ticket.ID,
		Subject:     "Acknowledgement",
		Body:        fmt.Sprintf("Acknowledgement for: %s %s", cliConfig.IcingaCheckState, cliConfig.IcingaCheckOutput),
		ContentType: "text/html",
		Type:        "web",
		Internal:    true,
		Sender:      "Agent",
	}

	err := c.AddArticleToTicket(ctx, a)

	if err != nil {
		return err
	}

	// Update the ticket state to open
	err = c.UpdateTicketState(ctx, ticket, zammad.OpenTicketState)

	return err
}

// handleRecoveryNotification closes an existing ticket
// If the existing ticket is open, adds an article to ticket and sets the state to closed
// If ticket is closed, reopens the ticket with an article
func handleRecoveryNotification(ctx context.Context, c *client.Client, ticket zammad.Ticket) error {
	if ticket.ID == 0 {
		return fmt.Errorf("no open or new ticket found to add recovery")
	}

	a := zammad.Article{
		TicketID:    ticket.ID,
		Subject:     "Recovery",
		Body:        fmt.Sprintf("Recovery for: %s %s", cliConfig.IcingaCheckState, cliConfig.IcingaCheckOutput),
		ContentType: "text/html",
		Type:        "web",
		Internal:    true,
		Sender:      "Agent",
	}

	err := c.AddArticleToTicket(ctx, a)

	if err != nil {
		return err
	}

	// Update the ticket state to open
	err = c.UpdateTicketState(ctx, ticket, zammad.ClosedTicketState)

	return err
}

// handleCustomNotification adds an article to an existing ticket
// If no ticket exists nothing happens and the function returns
func handleCustomNotification(ctx context.Context, c *client.Client, ticket zammad.Ticket) error {
	if ticket.ID == 0 {
		return nil
	}

	a := zammad.Article{
		TicketID:    ticket.ID,
		Subject:     "Recovery",
		Body:        fmt.Sprintf("Recovery for: %s %s", cliConfig.IcingaCheckState, cliConfig.IcingaCheckOutput),
		ContentType: "text/html",
		Type:        "web",
		Internal:    true,
		Sender:      "Agent",
	}

	err := c.AddArticleToTicket(ctx, a)

	return err
}
