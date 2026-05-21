package pst

import (
	"net/mail"
	"os"

	"github.com/navid/pst2jmap-migration/internal/model"
)

func ParseEML(
	path string,
) (*model.Message, error) {

	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	msg, err := mail.ReadMessage(f)

	if err != nil {
		return nil, err
	}

	return &model.Message{
		Subject: msg.Header.Get("Subject"),
		From:    msg.Header.Get("From"),
	}, nil
}
