# notify-zammad

A notification plugin for (mostly) Icinga which manages problems as Zammad tickets.

This plugin opens/updates/closes Zammad tickets via the Zammad API. The user/token for this plugin needs at least the `ticket.agent` permission.

Problem notifications will open a new ticket if none exists. The ticket will be in state `new`. If a ticket exists the plugin will add an article to the this ticket.

Acknowledgement notifications will add an article to an existing ticket.
This will also set the ticket state to `open`.
If no ticket exists nothing will happen.

Downtime/Flapping notifications will add an article to an existing ticket.
If no ticket exists nothing will happen.

Recovery notifications will close an existing ticket.
This will set the ticket state to `closed`.
If no ticket exists nothing will happen.

To track tickets the plugin uses two custom field attributes:

- icinga_host
- icinga_service

The plugin is currently designed to update the last created ticket with matching icinga_host and icinga_service.

**Why not use Zammad's built-in Icinga integration?** The built-in integration uses mails received by Zammad to open/close tickets. We had the requirement to solve the same feature without the use of mail.

## Usage

```bash
An Icinga notification plugin for Zammad

Usage:
  notify_zammad [flags]

Flags:
  -H, --zammad-hostname string        Address of the Zammad instance (NOTIFY_ZAMMAD_HOSTNAME) (default "localhost")
  -p, --zammad-port int               Port of the Zammad instance (default 443)
  -s, --secure                        Use a HTTPS connection
  -T, --token string                  Token for server authentication (NOTIFY_ZAMMAD_TOKEN)
  -u, --user string                   Specify the user name and password for server authentication <user:password> (NOTIFY_ZAMMAD_BASICAUTH)
      --ca-file string                Specify the CA File for TLS authentication (NOTIFY_ZAMMAD_CA_FILE)
      --cert-file string              Specify the Certificate File for TLS authentication (NOTIFY_ZAMMAD_CERT_FILE)
      --key-file string               Specify the Key File for TLS authentication (NOTIFY_ZAMMAD_KEY_FILE)
  -i, --insecure                      Skip the verification of the server\'s TLS certificate
  -t, --timeout int                   Timeout in seconds for the CheckPlugin (default 30)
      --host-name string              Host name of the Icinga 2 Host object
      --service-name string           Service name of the Icinga 2 Service Object (optional for Host Notifications)
      --check-state string            State of the Object (Up/Down for hosts, OK/Warning/Critical/Unknown for services)
      --check-output string           Output of the last executed check
      --notification-type string      Type of the notification (Problem/Recovery/Acknowledgement)
      --notification-author string    Name of an author for manual events
      --notification-comment string   Comment for manual events
      --notification-date string      Date when the event occurred
      --zammad-group string           Custom Zammad Field for the group
      --zammad-customer string        Custom Zammad Field for the customer
  -h, --help                          help for notify_zammad
  -v, --version                       version for notify_zammad
```

The plugin respects the environment variables `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY`.

Various flags can be set with environment variables, refer to the help to see which flags.

### Examples

Open a new Ticket at `https//zammad.example:8080`:

```bash
notify_zammad \
--zammad-hostname zammad.example \
--zammad-port 8080 \
--secure \
--token NoTaReAlToken_CXXoPxX \
--notification-type Problem \
--host-name myPreciousHost01 \
--service-name hostalive \
--check-state Down \
--check-output "CRITICAL - host unreachable" \
--zammad-group Users \
--zammad-customer "jon.snow@zammad"
```

Acknowledge an existing Ticket at `https//zammad.example:8080`:

```bash
notify_zammad \
--zammad-hostname zammad.example \
--zammad-port 8080 \
--secure \
--token NoTaReAlToken_CXXoPxX \
--notification-type Acknowledgement \
--host-name myPreciousHost01 \
--service-name hostalive \
--check-state Down \
--check-output "CRITICAL - host unreachable" \
--zammad-group Users \
--zammad-customer "jon.snow@zammad"
```

Close an existing Ticket at `https//zammad.example:8080`:

```bash
notify_zammad \
--zammad-hostname zammad.example \
--zammad-port 8080 \
--secure \
--token NoTaReAlToken_CXXoPxX \
--notification-type Recovery \
--host-name myPreciousHost01 \
--service-name hostalive \
--check-state Up \
--check-output "PING OK - Packet loss = 0%" \
--zammad-group Users \
--zammad-customer "jon.snow@zammad"
```

## License

Copyright (c) 2024 [NETWAYS GmbH](mailto:info@netways.de)

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public
License as published by the Free Software Foundation, either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not,
see [gnu.org/licenses](https://www.gnu.org/licenses/).
