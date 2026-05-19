package pstreader

import (
	"fmt"
	"time"

	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/mooijtech/go-pst/v6/pkg/properties"

	"github.com/navid/pst2jmap-migration/internal/model"
)

func ExtractMessage(
	folder *pst.Folder,
	message *pst.Message,
) (*model.Message, error) {

	msg, ok := message.Properties.(*properties.Message)

	// fmt.Println(msg)
	// if IgnoredMessageClasses[*msg.ContentType] {
	// 	return nil, nil
	// }

	if !ok {
		return nil, fmt.Errorf("unsupported message type")
	}

	var date time.Time

	if msg.ClientSubmitTime != nil {
		date = time.Unix(0, *msg.ClientSubmitTime)
	}

	out := &model.Message{
		Folder: folder.Name,

		Subject: msg.GetSubject(),

		FromName: msg.GetSenderName(),

		FromEmail: extractSenderEmail(msg),

		MessageID: *msg.InternetMessageId,

		Headers: msg.GetTransportMessageHeaders(),

		Body: msg.GetBody(),

		Date: date,
	}

	return out, nil
}
