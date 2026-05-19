package pstreader

import (
	pst "github.com/mooijtech/go-pst/v6/pkg"
	"github.com/rotisserie/eris"
)

func (r *Reader) CountMessages() (
	[]FolderStats,
	int,
	error,
) {

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

	var (
		total int
		stats []FolderStats
	)

	err := r.file.WalkFolders(func(folder *pst.Folder) error {

		if !allowedFolders[folder.Name] {
			return nil
		}

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

		stats = append(stats, FolderStats{
			Name:  folder.Name,
			Count: count,
		})

		total += count

		return messageIterator.Err()
	})

	return stats, total, err
}
