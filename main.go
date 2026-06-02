package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"os/signal"
	"syscall"

	"github.com/navid/pst2jmap-migration/internal/jmap"
	"github.com/navid/pst2jmap-migration/internal/pst"
)

const (
	JMAP_URL = "https://postmaster.collab24.net/jmap"
)

var version = "dev"

func main() {

	startedAt := time.Now()

	var (
		pstFile  string
		username string
		password string
	)

	flag.StringVar(&pstFile, "pst", "", "Src. Path to PST file")
	flag.StringVar(&username, "user", "", "Dest. Username")
	flag.StringVar(&password, "password", "", "Dest. Password")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	validate(pstFile, username, password)

	migrationDir := pst.MigrationDir(pstFile)

	err := os.MkdirAll(migrationDir, 0755)

	if err != nil {
		exit("failed to create migration directory", err)
	}

	stateFile := filepath.Join(migrationDir, "state.json")

	reportFile := filepath.Join(migrationDir, "report.json")

	fmt.Println("Migration workspace:", migrationDir)

	state, err := pst.LoadState(stateFile)

	if err != nil {
		exit("failed to load migration state", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals

		fmt.Println()
		fmt.Println("Interrupt received, stopping after current email...")

		cancel()
	}()

	fmt.Println("Extracting PST...")

	reader, err := pst.NewReader(ctx, pstFile)

	if err != nil {
		exit("failed to extract PST", err)
	}

	defer reader.Close()

	client := jmap.NewClient(JMAP_URL, username, password)

	fmt.Println("Connecting to JMAP...")

	err = client.Connect()

	if err != nil {
		exit("failed to connect jmap", err)
	}

	fmt.Println("Authenticated as:", client.Session.Username)

	fmt.Println()

	mailboxes, err := client.GetMailboxIDs()

	if err != nil {
		exit("failed to get mailboxes", err)
	}

	fmt.Println("Detected mailboxes:")

	for role, id := range mailboxes {
		fmt.Printf("  %-10s %s\n", role, id)
	}

	fmt.Println()

	total, err := pst.CountMessages(reader.OutputDir)

	if err != nil {
		exit("failed to count messages", err)
	}

	stats := pst.NewStats(total, startedAt)

	fmt.Printf("Found %d emails\n\n", total)

	fmt.Println("Starting migration...")

	folderCounts := map[string]int{}

	err = reader.Walk(
		func(path string) error {
			select {
			case <-ctx.Done():

				fmt.Println()
				fmt.Println("Migration interrupted.")

				return filepath.SkipAll

			default:
			}

			if state.IsProcessed(path) {
				stats.IncrementSkipped()

				fmt.Printf("SKIPPED: %s\n", path)

				return nil
			}

			msg, err := pst.ParseMessage(path)

			if err != nil {
				stats.IncrementFailed()

				fmt.Printf("PARSE FAILED: %v\n", err)

				return nil
			}

			mailboxID := jmap.ResolveMailboxID(msg.Folder, mailboxes)

			if state.HasMessageID(msg.MessageID) {
				stats.IncrementSkipped()

				fmt.Printf("DUPLICATE MESSAGE-ID: %s\n", msg.MessageID)

				return nil
			}

			fmt.Println()
			fmt.Printf("[%d/%d] %s\n", stats.Processed+1, stats.TotalMessages, path)

			fmt.Printf("Folder: %s\n", msg.Folder)
			fmt.Printf("Subject: %s\n", msg.Subject)
			fmt.Printf("Mailbox: %s\n", mailboxID)

			var blobID string

			err = jmap.Retry(3, 2*time.Second, func() error {
				var uploadErr error
				blobID, uploadErr = client.UploadEML(path)
				return uploadErr
			})

			if err != nil {
				stats.IncrementFailed()
				fmt.Printf("UPLOAD FAILED: %v\n", err)
				return nil
			}

			fmt.Printf("Uploaded blob: %s\n", blobID)

			err = jmap.Retry(3, 2*time.Second, func() error {
				return client.ImportEmail(blobID, mailboxID)
			})

			if err != nil {
				stats.IncrementFailed()
				fmt.Printf("IMPORT FAILED: %v\n", err)
				return nil
			}

			state.MarkMessageID(msg.MessageID)
			state.MarkProcessed(path)

			err = state.Save(stateFile)

			if err != nil {
				fmt.Printf("WARNING: failed to save state: %v\n", err)
			}

			stats.IncrementImported()

			folderCounts[msg.Folder]++

			fmt.Println("SUCCESS")

			return nil
		},
	)

	if err != nil {
		exit("migration failed", err)
	}

	stats.Finish()

	if ctx.Err() != nil {
		fmt.Println()
		fmt.Println("Migration stopped by user.")
	}

	err = pst.WriteReport(reportFile, stats, folderCounts)

	if err != nil {
		fmt.Printf("WARNING: failed to write report: %v\n", err)
	}

	fmt.Println()
	fmt.Println("Folder summary:")

	for folder, count := range folderCounts {

		fmt.Printf("  %-15s %d\n", folder, count)
	}

	fmt.Println()
	fmt.Println(stats)
}

func validate(pstFile string, user string, pass string,
) {

	required := map[string]string{
		"--pst":      pstFile,
		"--user":     user,
		"--password": pass,
	}

	for k, v := range required {
		if v == "" {
			fmt.Fprintf(os.Stderr, "ERROR: missing required option %s\n", k)
			os.Exit(1)
		}
	}
}

func exit(msg string, err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", msg, err)
	os.Exit(1)
}
