package migration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/navid/pst2jmap-migration/internal/jmap"
	"github.com/navid/pst2jmap-migration/internal/model"
	"github.com/navid/pst2jmap-migration/internal/pst"
)

func Run(
	ctx context.Context,
	pstFile string,
	username string,
	password string,
	jmapURL string,
	workers int,
	mailboxWorkers int,
) error {

	startedAt := time.Now()

	migrationDir := pst.MigrationDir(pstFile)

	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		return fmt.Errorf(
			"create migration directory: %w",
			err,
		)
	}

	stateFile := filepath.Join(
		migrationDir,
		"state.json",
	)

	reportFile := filepath.Join(
		migrationDir,
		"report.json",
	)

	fmt.Println("Migration workspace:", migrationDir)

	state, err := pst.LoadState(stateFile)

	if err != nil {
		return fmt.Errorf(
			"load migration state: %w",
			err,
		)
	}

	fmt.Println("Extracting PST...")

	reader, err := pst.NewReader(
		ctx,
		pstFile,
	)

	if err != nil {
		return fmt.Errorf(
			"extract PST: %w",
			err,
		)
	}

	defer reader.Close()

	client := jmap.NewClient(
		jmapURL,
		username,
		password,
	)

	fmt.Println("Connecting to JMAP...")

	if err := client.Connect(); err != nil {
		return fmt.Errorf(
			"connect JMAP: %w",
			err,
		)
	}

	fmt.Println(
		"Authenticated as:",
		client.Session.Username,
	)

	fmt.Println()

	mailboxes, err := client.GetMailboxIDs()

	if err != nil {
		return fmt.Errorf(
			"get mailboxes: %w",
			err,
		)
	}

	fmt.Println("Detected mailboxes:")

	for role, id := range mailboxes {
		fmt.Printf(
			"  %-10s %s\n",
			role,
			id,
		)
	}

	fmt.Println()

	total, err := pst.CountMessages(
		reader.OutputDir,
	)

	if err != nil {
		return fmt.Errorf(
			"count messages: %w",
			err,
		)
	}

	stats := pst.NewStats(
		total,
		startedAt,
	)

	failures := pst.NewFailures()

	folderCounts := pst.NewFolderCounts()

	fmt.Printf(
		"Found %d emails\n\n",
		total,
	)

	fmt.Println("Starting migration...")

	jobs := make(
		chan model.Job,
		100,
	)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {

		wg.Add(1)

		go pst.StartWorker(
			ctx,
			&wg,
			jobs,
			func(ctx context.Context, path string) error {
				return pst.ProcessEmail(
					ctx,
					path,
					client,
					mailboxes,
					state,
					stats,
					folderCounts,
					failures,
					stateFile,
				)
			},
		)
	}

	err = reader.Walk(
		func(path string) error {

			select {

			case <-ctx.Done():
				return filepath.SkipAll

			default:
			}

			jobs <- model.Job{
				PSTFile: path,
			}

			return nil
		},
	)

	close(jobs)

	wg.Wait()

	if err != nil &&
		err != filepath.SkipAll {
		return fmt.Errorf(
			"walk extracted emails: %w",
			err,
		)
	}

	stats.Finish()

	if ctx.Err() != nil {

		fmt.Println()
		fmt.Println(
			"Migration stopped by user.",
		)
	}

	if err := pst.WriteReport(
		reportFile,
		stats,
		folderCounts.Snapshot(),
		failures.Snapshot(),
	); err != nil {

		fmt.Printf(
			"WARNING: failed to write report: %v\n",
			err,
		)
	}

	fmt.Println()
	fmt.Println("Folder summary:")

	for folder, count := range folderCounts.Snapshot() {

		fmt.Printf(
			"  %-15s %d\n",
			folder,
			count,
		)
	}

	fmt.Println()
	fmt.Println(stats)

	return nil
}
