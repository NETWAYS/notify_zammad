package cmd

import (
	"errors"
	"fmt"

	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-icingadsl"
	"github.com/NETWAYS/notify_zammad/api"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:     "notify",
	Short:   "Send a notification for an Icinga 2 event to Zammad",
	Long:    `TODO`,
	Example: `TODO`,
	Run:     sendNotification,
}

// nolint:funlen
func sendNotification(cmd *cobra.Command, args []string) {
	notificationType, err := icingadsl.ParseNotificationType(config.notificationType)
	if err != nil {
		check.ExitError(err)
	}

	if config.debuglevel > 0 {
		ntString, err := icingadsl.FormatNotificationType(notificationType)
		if err != nil {
			check.ExitError(err)
		}

		fmt.Printf("Got notification type: %s (%d)\n", ntString, notificationType)
	}

	client, err := config.NewClient()
	if err != nil {
		check.ExitError(err)
	}

	var tickets map[api.ZammadTicketID]api.ZammadTicket

	if config.serviceName == "" {
		// Searching for a host ticket
		tickets, err = client.SearchTicketForHost(config.hostName)
		if err != nil {
			check.ExitError(err)
		}
	} else {
		// Searching for a service ticket
		tickets, err = client.SearchTicketForService(config.hostName, config.serviceName)
		if err != nil {
			check.ExitError(err)
		}
	}

	var ticketExists bool

	var ticketID api.ZammadTicketID

	if len(tickets) == 0 {
		// No ticket yet
		ticketExists = false
	} else {
		// got a ticket
		ticketExists = true

		// just pick the first element
		for k := range tickets {
			ticketID = k
			break
		}
	}

	if config.debuglevel > 0 {
		if ticketExists {
			fmt.Printf("Ticket already exists with ID: %d\n", ticketID)
		} else {
			fmt.Println("No existing ticket for this host/service combination found")
		}
	}

	switch notificationType {
	case icingadsl.Custom:
		err = customNotificationHelper(client, ticketExists, ticketID)
		if err != nil {
			check.ExitError(err)
		}
	case icingadsl.Acknowledgement:
		err = acknowledgeNotificationHelper(client, ticketExists, ticketID)
		if err != nil {
			check.ExitError(err)
		}
	case icingadsl.Problem:
		err = problemNotificationHelper(client, ticketExists, ticketID)
		if err != nil {
			check.ExitError(err)
		}
	case icingadsl.Recovery:
		err = recoveryNotificationHelper(client, ticketExists, ticketID)
		if err != nil {
			check.ExitError(err)
		}
	case icingadsl.DowntimeStart:
	case icingadsl.DowntimeEnd:
	case icingadsl.DowntimeRemoved:
	case icingadsl.FlappingStart:
	case icingadsl.FlappingEnd:
	default:
		check.ExitError(errors.New("unsupported notification type"))
	}

	check.ExitRaw(check.OK, "")
}

func init() {
	rootCmd.AddCommand(notifyCmd)

	fs := notifyCmd.Flags()

	fs.StringVar(&config.hostName, "host-name", "",
		"host name of the Icinga 2 Host object to notify for")

	_ = cobra.MarkFlagRequired(fs, "host-name")

	fs.StringVar(&config.serviceName, "service-name", "",
		"service name of the Icinga 2 Service Object to notify for (optional for Host Notifications)")

	fs.StringVar(&config.checkState, "check-state", "",
		"State of the Object (Up/Down for hosts, OK/Warning/Critical/Unknown for services)")

	_ = cobra.MarkFlagRequired(fs, "check-state")

	fs.StringVar(&config.checkOutput, "check-output", "",
		"Output of the last executed check")

	_ = cobra.MarkFlagRequired(fs, "check-output")

	fs.StringVar(&config.notificationType, "notification-type", "",
		"The type of the notication (Problem/Recover/Acknowledgement/...)")

	_ = cobra.MarkFlagRequired(fs, "notification-type")

	fs.StringVar(&config.author, "notification-author", "",
		"The name of an author if the event was manually triggered")
	fs.StringVar(&config.comment, "notification-comment", "",
		"A comment in case of manually triggered events")
	fs.StringVar(&config.date, "notification-date", "",
		"Date when an event occurred")

	fs.StringVar(&config.zammadGroup, "zammad-group", "",
		"The Zammad group to put the ticket in")

	_ = cobra.MarkFlagRequired(fs, "zammad-group")

	fs.StringVar(&config.zammadCustomer, "zammad-customer", "",
		"Customer field in ticket")

	_ = cobra.MarkFlagRequired(fs, "customer")

	fs.SortFlags = false
}

func customNotificationHelper(client *api.ZammadAPIClient, ticketExists bool, ticketID api.ZammadTicketID) error {
	// Custom notification, add article to existing ticket
	if !ticketExists {
		return nil
	}

	newArticle := api.ZammadArticle{
		TicketID:    ticketID,
		Subject:     "Custom notification",
		Body:        "Custom notification was triggered",
		ContentType: "text/html",
		Type:        "web",
		Internal:    true,
		Sender:      "Agent",
		TimeUnit:    "0",
	}

	err := client.AddArticleToTicket(newArticle)

	return err
}

func acknowledgeNotificationHelper(client *api.ZammadAPIClient, ticketExists bool, ticketID api.ZammadTicketID) error {
	// Acknowledgement for a problem, so search problem ticket and add an article
	// Possibly set the ticket to handled or something
	if !ticketExists {
		return fmt.Errorf("should send an Acknowledgement, but didn't find the problem")
	}

	newArticle := api.ZammadArticle{
		TicketID:    ticketID,
		Subject:     "Acknowledgement",
		Body:        "Problem was acknowledged",
		ContentType: "text/html",
		Type:        "web",
		Internal:    true,
		Sender:      "Agent",
		TimeUnit:    "0",
	}

	err := client.AddArticleToTicket(newArticle)
	if err != nil {
		return err
	}

	err = client.ChangeTicketState(ticketID, 3)

	return err
}

// nolint:funlen
func problemNotificationHelper(client *api.ZammadAPIClient, ticketExists bool, ticketID api.ZammadTicketID) error {
	/*
	 * Problem occurred, search for existing ticket
	 * if yes -> add an article and change the ticket title according to new state
	 * if no -> create a new ticket
	 */
	if ticketExists {
		newArticle := api.ZammadArticle{
			TicketID:    ticketID,
			Subject:     "Problem",
			Body:        config.checkState + " " + config.checkOutput,
			ContentType: "text/html",
			Type:        "web",
			Internal:    true,
			Sender:      "Agent",
			TimeUnit:    "0",
		}

		err := client.AddArticleToTicket(newArticle)

		return err
	}

	if config.debuglevel > 0 {
		fmt.Println("Creating new problem ticket")
	}

	newArticle := api.ZammadArticle{
		TicketID:    ticketID,
		Subject:     "Problem",
		Body:        config.checkState + " " + config.checkOutput,
		ContentType: "text/html",
		Type:        "web",
		Internal:    true,
		Sender:      "Agent",
		TimeUnit:    "0",
	}

	titleText, err := icingadsl.FormatNotificationType(icingadsl.Problem)
	if err != nil {
		return err
	}

	titleText += ": "

	if config.serviceName != "" {
		// service problem
		titleText += "Service " + config.serviceName + " on " + config.hostName + " is " + config.checkState
	} else {
		titleText += "Host " + config.hostName + " is " + config.checkState
	}

	newTicket := api.ZammadNewTicket{
		Title:         titleText,
		Group:         config.zammadGroup,
		Customer:      config.zammadCustomer,
		Article:       newArticle,
		IcingaHost:    config.hostName,
		IcingaService: config.serviceName,
	}

	if config.debuglevel > 1 {
		fmt.Printf("New problem ticket: %#v\n", newTicket)
	}

	err = client.CreateTicket(newTicket)

	return err
}

func recoveryNotificationHelper(client *api.ZammadAPIClient, ticketExists bool, ticketID api.ZammadTicketID) error {
	/*
	 * Recovery, search for existing ticket and resolve (close) it. If none exits, do nothing
	 */
	if !ticketExists {
		// No ticket for that, do nothing
		return nil
	}

	// Post new article and the close ticket
	newArticle := api.ZammadArticle{
		TicketID:    ticketID,
		Subject:     "Recovery",
		Body:        config.checkState + " " + config.checkOutput,
		ContentType: "text/html",
		Type:        "web",
		Internal:    true,
		Sender:      "Agent",
		TimeUnit:    "0",
	}

	err := client.AddArticleToTicket(newArticle)
	if err != nil {
		return err
	}

	err = client.ChangeTicketState(ticketID, api.Closed)

	return err
}
