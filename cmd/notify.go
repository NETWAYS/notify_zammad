package cmd

import (
	"errors"

	"github.com/NETWAYS/go-check"
	icingadsl "github.com/NETWAYS/go-icingadsl/types"
	icingadslTypes "github.com/NETWAYS/go-icingadsl/types"
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

func sendNotification(cmd *cobra.Command, args []string) {

	notificationType, err := icingadslTypes.ParseNotificationType(config.notificationType)
	if err != nil {
		check.ExitError(err)

	}

	client, err := config.NewClient()
	if err != nil {
		check.ExitError(err)
	}

	var tickets map[api.ZammadTicketId]api.ZammadTicket

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
	var ticketId api.ZammadTicketId

	if len(tickets) == 0 {
		// No ticket yet
		ticketExists = false
	} else {
		// got a ticket
		ticketExists = true

		// just pick the first element
		for k := range tickets {
			ticketId = k
			break
		}
	}

	switch notificationType {
	case icingadslTypes.Custom:
		// Custom notification, add article to existing ticket
		if ticketExists {
			newArticle := api.ZammadArticle{
				TicketId:    ticketId,
				Subject:     "Custom notification",
				Body:        "Custom notification was triggered",
				ContentType: "text/html",
				Type:        "web",
				Internal:    true,
				Sender:      "Agent",
				TimeUnit:    "0",
			}

			client.AddArticleToTicket(newArticle)
		} else {
			check.ExitRaw(check.OK, "Got a custom notification, but no ticket. Not sending anything, exiting silently")
		}
	case icingadslTypes.Acknowledgement:
		// Acknowledgement for a problem, so search problem ticket and add an article
		// Possibly set the ticket to handled or something
		if !ticketExists {
			check.ExitRaw(check.Warning, "Should send an Acknowledgement, but didn't find the problem")
		}
		newArticle := api.ZammadArticle{
			TicketId:    ticketId,
			Subject:     "Acknowledgement",
			Body:        "Problem was acknowledged",
			ContentType: "text/html",
			Type:        "web",
			Internal:    true,
			Sender:      "Agent",
			TimeUnit:    "0",
		}

		client.AddArticleToTicket(newArticle)
		client.ChangeTicketState(ticketId, 3)

	case icingadslTypes.Problem:
		/*
		 * Problem occured, search for existing ticket
		 * if yes -> add an article and change the ticket title according to new state
		 * if no -> create a new ticket
		 */

		if ticketExists {
			newArticle := api.ZammadArticle{
				TicketId:    ticketId,
				Subject:     "Problem",
				Body:        config.checkState + " " + config.checkOutput,
				ContentType: "text/html",
				Type:        "web",
				Internal:    true,
				Sender:      "Agent",
				TimeUnit:    "0",
			}
			client.AddArticleToTicket(newArticle)
		} else {
			newArticle := api.ZammadArticle{
				TicketId:    ticketId,
				Subject:     "Problem",
				Body:        config.checkState + " " + config.checkOutput,
				ContentType: "text/html",
				Type:        "web",
				Internal:    true,
				Sender:      "Agent",
				TimeUnit:    "0",
			}

			titleText, err := icingadsl.FormatNotificationType(notificationType)
			if err != nil {
				check.ExitError(err)
			}

			titleText += ": "

			if config.serviceName != "" {
				// service problem
				titleText += "Service " + config.serviceName + " on " + config.hostName + " is " + config.checkState
			} else {
				titleText += "Host " + config.hostName + " is " + config.checkState
			}

			newTicket := api.ZammadNewTicket{
				Title:         config.checkState,
				Group:         config.zammadGroup,
				Customer:      config.zammadCustomer,
				Article:       newArticle,
				IcingaHost:    config.hostName,
				IcingaService: config.serviceName,
			}

			client.CreateTicket(newTicket)
		}
	case icingadslTypes.Recovery:
		/*
		 * Recovery, search for existing ticket and resolve (close) it. If none exits, do nothing
		 */

		if !ticketExists {
			// No ticket for that, do nothing
		} else {
			// Post new article and the close ticket
			newArticle := api.ZammadArticle{
				TicketId:    ticketId,
				Subject:     "Recovery",
				Body:        config.checkState + " " + config.checkOutput,
				ContentType: "text/html",
				Type:        "web",
				Internal:    true,
				Sender:      "Agent",
				TimeUnit:    "0",
			}

			client.AddArticleToTicket(newArticle)
			client.ChangeTicketState(ticketId, api.Closed)
		}
	default:
		check.ExitError(errors.New("Unsupported notification type"))
	}

	check.ExitRaw(check.OK, "")
}

func init() {
	rootCmd.AddCommand(notifyCmd)

	fs := notifyCmd.Flags()

	fs.StringVar(&config.hostName, "host-name", "",
		"host name of the Icinga 2 Host object to notify for")
	cobra.MarkFlagRequired(fs, "host-name")

	fs.StringVar(&config.serviceName, "service-name", "",
		"service name of the Icinga 2 Service Object to notify for (optional for Host Notifications)")

	fs.StringVar(&config.checkState, "check-state", "",
		"State of the Object (Up/Down for hosts, OK/Warning/Critical/Unknown for services)")
	fs.StringVar(&config.checkOutput, "check-output", "",
		"Output of the last executed check")

	fs.StringVar(&config.notificationType, "notification-type", "",
		"The type of the notication (Problem/Recover/Acknowledgement/...)")
	cobra.MarkFlagRequired(fs, "notification-type")

	fs.StringVar(&config.author, "notification-author", "",
		"The name of an author if the event was manually triggered")
	fs.StringVar(&config.comment, "notification-comment", "",
		"A comment in case of manually triggered events")
	fs.StringVar(&config.date, "notification-date", "",
		"Date when an event occured")

	fs.StringVar(&config.zammadGroup, "zammad-group", "",
		"The Zammad group to put the ticket in")
	cobra.MarkFlagRequired(fs, "zammad-group")

	fs.StringVar(&config.zammadCustomer, "zammad-customer", "",
		"Customer field in ticket")
	cobra.MarkFlagRequired(fs, "customer")

	fs.SortFlags = false
}
