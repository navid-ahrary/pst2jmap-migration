package pstreader

import (
	"fmt"
	"os"

	charsets "github.com/emersion/go-message/charset"

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

	var processed int

	err := r.file.WalkFolders(func(folder *pst.Folder) error {

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

		var folderProcessed int

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

			processed++
			folderProcessed++

			// TODO:
			// Future migration step:
			//
			// err = migrateMessage(msg)
			// if err != nil {
			//     ...
			// }

			_ = msg

			// Print progress every 100 messages
			if processed%100 == 0 {

				fmt.Printf(
					"\rProcessed: %d",
					processed,
				)
			}
		}

		fmt.Printf(
			"\nFolder completed: %d messages\n",
			folderProcessed,
		)

		return messageIterator.Err()
	})

	if err != nil {
		return err
	}

	fmt.Printf(
		"\n\nCompleted. Total processed: %d\n",
		processed,
	)

	return nil
}
