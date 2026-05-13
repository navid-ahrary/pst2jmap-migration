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
		"Top-of-Information-Store": true,
		"Inbox":                    true,
		"Sent Items":               true,
		"Sent-Items":               true,
		"Drafts":                   true,
		"Deleted Items":            true,
		"Deleted-Items":            true,
		"Junk Email":               true,
		"Junk-Email":               true,
	}

	return r.file.WalkFolders(func(folder *pst.Folder) error {
		if folder.Name == "Top-of-Information-Store" {
			return nil
		}

		if !allowedFolders[folder.Name] {
			return nil
		}

		fmt.Printf("Folder: %s\n", folder.Name)

		messageIterator, err := folder.GetMessageIterator()

		if eris.Is(err, pst.ErrMessagesNotFound) {
			return nil
		}

		if err != nil {
			return err
		}

		count := 0

		for messageIterator.Next() {
			count++
		}

		fmt.Printf("Messages: %d\n\n", count)

		return messageIterator.Err()
	})
}
