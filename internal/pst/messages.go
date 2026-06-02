package pst

import (
	"io"
	"mime"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	"github.com/navid/pst2jmap-migration/internal/model"
	"golang.org/x/net/html/charset"
)

func ParseMessage(path string) (*model.Message, error) {
	f, err :=
		os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	msg, err := mail.ReadMessage(f)

	if err != nil {
		return nil, err
	}

	subject := decodeHeader(msg.Header.Get("Subject"))

	from := msg.Header.Get("From")

	messageID := msg.Header.Get("Message-ID")

	return &model.Message{
		Subject:   subject,
		From:      from,
		MessageID: messageID,
		Folder:    mailboxFromPath(path),
	}, nil
}

func decodeHeader(value string) string {
	if value == "" {
		return ""
	}

	decoder := mime.WordDecoder{
		CharsetReader: func(charsetName string, input io.Reader) (io.Reader, error) {
			return charset.NewReaderLabel(charsetName, input)
		},
	}

	decoded, err := decoder.DecodeHeader(value)

	if err != nil {
		return strings.TrimSpace(value)
	}

	return decoded
}

func mailboxFromPath(path string) string {
	dir := filepath.Dir(path)

	for {
		name := filepath.Base(dir)

		switch name {

		case "Inbox":
			return "Inbox"

		case "Sent-Items":
			return "Sent-Items"

		case "Sent Items":
			return "Sent Items"

		case "Deleted-Items":
			return "Deleted-Items"

		case "Deleted Items":
			return "Deleted Items"

		case "Drafts":
			return "Drafts"

		case "Junk-Email":
			return "Junk-Email"

		case "Junk Email":
			return "Junk Email"
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			break
		}

		dir = parent
	}

	return "Unknown"
}
