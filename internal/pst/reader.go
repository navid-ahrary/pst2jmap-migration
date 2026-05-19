package pstreader

import (
	"fmt"
	"os"
	"sort"

	charsets "github.com/emersion/go-message/charset"

	"github.com/navid/pst2jmap-migration/internal/model"

	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
	"golang.org/x/text/encoding"
)

type Reader struct {
	file   *pst.File
	reader *os.File
}

func Open(path string) (*Reader, error) {

	// Register charset support
	pst.ExtendCharsets(func(name string, enc encoding.Encoding) {
		charsets.RegisterEncoding(name, enc)
	})

	reader, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	pstFile, err := pst.New(reader)

	if err != nil {
		return nil, err
	}

	return &Reader{
		file:   pstFile,
		reader: reader,
	}, nil
}

func (r *Reader) Close() error {

	r.file.Cleanup()

	return r.reader.Close()
}

func (r *Reader) WalkMailboxFolders() error {

	allowedFolders := map[string]bool{
		"Inbox":         true,
		"Sent Items":    true,
		"Sent-Items":    true,
		"Drafts":        true,
		"Deleted Items": true,
		"Deleted-Items": true,
		"Junk Email":    true,
		"Junk-Email":    true,
	}

	return r.file.WalkFolders(func(folder *pst.Folder) error {

		if !allowedFolders[folder.Name] {
			return nil
		}

		fmt.Printf("\nFolder: %s\n", folder.Name)

		messageIterator, err := folder.GetMessageIterator()

		if eris.Is(err, pst.ErrMessagesNotFound) {
			return nil
		}

		if err != nil {
			return err
		}

		var messages []*model.Message

		for messageIterator.Next() {

			message := messageIterator.Value()

			// Convert PST message -> internal model
			msg, err := ExtractMessage(folder, message)

			if err != nil {

				fmt.Printf(
					"Failed to extract message: %v\n",
					err,
				)

				continue
			}

			if msg == nil {
				continue
			}

			// ONLY STORE
			messages = append(messages, msg)
		}

		// Sort by date descending
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].Date.After(messages[j].Date)
		})

		// Print ONLY ONCE
		for i, msg := range messages {

			fmt.Printf(
				"[%d] %s | %s | %s\n",
				i+1,
				msg.Date.Format("2006-01-02 15:04"),
				msg.FromEmail,
				msg.Subject,
			)
		}

		fmt.Printf(
			"\nProcessed messages: %d\n",
			len(messages),
		)

		return messageIterator.Err()
	})
}
