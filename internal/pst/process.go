package pst

import (
	"context"
	"fmt"
	"time"

	"github.com/navid/pst2jmap-migration/internal/jmap"
)

func ProcessEmail(
	ctx context.Context,
	path string,
	client *jmap.Client,
	mailboxes map[string]string,
	state *MigrationState,
	stats *Stats,
	folderCounts *FolderCounts,
	failures *Failures,
	stateFile string,
) error {

	if state.IsProcessed(path) {
		stats.IncrementSkipped()
		fmt.Printf("SKIPPED (already processed): %s\n", path)
		return nil
	}

	if err := checkCancelled(ctx); err != nil {
		return err
	}

	msg, err := ParseMessage(path)

	if err != nil {
		failures.Add(Failure{
			Path:      path,
			Operation: "parse",
			Error:     err.Error(),
		})

		stats.IncrementFailed()
		fmt.Printf("PARSE FAILED: %v\n", err)
		return nil
	}

	if state.HasMessageID(msg.MessageID) {
		stats.IncrementSkipped()
		fmt.Printf("DUPLICATE MESSAGE-ID: %s\n", msg.MessageID)
		return nil
	}

	mailboxID := jmap.ResolveMailboxID(msg.Folder, mailboxes)

	fmt.Println()
	fmt.Printf("[%d/%d] %s\n", stats.Processed+1, stats.TotalMessages, path)
	fmt.Printf("Folder: %s\n", msg.Folder)
	fmt.Printf("Subject: %s\n", msg.Subject)
	fmt.Printf("Mailbox: %s\n", mailboxID)

	var blobID string

	if err := checkCancelled(ctx); err != nil {
		return err
	}

	err = jmap.Retry(3, 2*time.Second,
		func() error {
			var uploadErr error

			blobID, uploadErr = client.UploadEML(path)

			return uploadErr
		},
	)

	if err != nil {
		failures.Add(Failure{
			Path:      path,
			Subject:   msg.Subject,
			Operation: "upload",
			Error:     err.Error(),
		})
		stats.IncrementFailed()
		fmt.Printf("UPLOAD FAILED: %v\n", err)
		return nil
	}

	fmt.Printf("Uploaded blob: %s\n", blobID)

	if err := checkCancelled(ctx); err != nil {
		return err
	}

	err = jmap.Retry(3, 2*time.Second,
		func() error {
			return client.ImportEmail(blobID, mailboxID)
		},
	)

	if err != nil {
		failures.Add(Failure{
			Path:      path,
			Subject:   msg.Subject,
			Operation: "import",
			Error:     err.Error(),
		})
		stats.IncrementFailed()
		fmt.Printf("IMPORT FAILED: %v\n", err)
		return nil
	}

	state.MarkProcessed(path)
	state.MarkMessageID(msg.MessageID)

	if err := state.Save(stateFile); err != nil {
		fmt.Printf("WARNING: failed to save state: %v\n", err)
	}

	stats.IncrementImported()
	folderCounts.Increment(msg.Folder)

	fmt.Println("SUCCESS")

	return nil
}

func checkCancelled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
