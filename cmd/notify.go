package cmd

import (
	"errors"

	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/icingadsl"
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

	client := config.NewClient()

	notificationType, err := icingadsl.ParseNotificationType(config.notificationType)
	if err != nil {
		check.ExitError(err)
	}

	switch notificationType {
	case icingadsl.Custom:
	case icingadsl.Acknowledgment:
	case icingadsl.Problem:
	case icingadsl.Recovery:
	default:
		check.ExitError(errors.New("Unsupported notification type"))
	}

	/*
		tickets, err := client.TicketList() //Get Users

		if err != nil {
			check.ExitError(err)
		}

		for idx, ticket := range *tickets {
			fmt.Printf("%v\n", idx)
			b, err := json.MarshalIndent(ticket, "", "  ")
			if err != nil {
				fmt.Println("error:", err)
			}
			fmt.Print(string(b))
		}
	*/

	tmpMap := make(map[string]interface{})
	tmpMap["icinga_host"] = config.hostName
	tmpMap["group"] = config.zammadGroup
	tmpMap["customer"] = config.zammadCustomer

	// Determine notification type
	if config.serviceName != "" {
		// got a service name, so this is related to a service
		tmpMap["title"] = "Service: " + config.serviceName + " on " + config.hostName
		tmpMap["icinga_service"] = config.serviceName
	} else {
		// got no service name, so this is related to a host
		tmpMap["title"] = "Host: " + config.hostName
	}

	/*
			article := `   "article": {
		      "subject": "My subject",
		      "body": "I am a message!",
		      "type": "note",
		      "internal": false
		   }`

			tmpMap["article"] = article
	*/

	// ticket "open"

	_, err := client.TicketCreate(&tmpMap)
	if err != nil {
		check.ExitError(err)
	}
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
