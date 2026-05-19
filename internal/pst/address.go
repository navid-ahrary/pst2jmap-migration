package pstreader

import (
	"net/mail"
	"strings"

	"github.com/mooijtech/go-pst/v6/pkg/properties"
)

func extractSenderEmail(
	msg *properties.Message,
) string {

	// Preferred:
	if msg.SenderAddressType != nil &&
		*msg.SenderAddressType == "SMTP" &&
		msg.SenderEmailAddress != nil &&
		*msg.SenderEmailAddress != "" {

		return *msg.SenderEmailAddress
	}

	// Fallback:
	if msg.SentRepresentingAddressType != nil &&
		*msg.SentRepresentingAddressType == "SMTP" &&
		msg.SentRepresentingEmailAddress != nil &&
		*msg.SentRepresentingEmailAddress != "" {

		return *msg.SentRepresentingEmailAddress
	}

	// Final fallback:
	headers := msg.GetTransportMessageHeaders()

	if headers == "" {
		return ""
	}

	parsed, err := mail.ReadMessage(
		strings.NewReader(headers),
	)

	if err != nil {
		return ""
	}

	from := parsed.Header.Get("From")

	addr, err := mail.ParseAddress(from)

	if err != nil {
		return from
	}

	return addr.Address
}
