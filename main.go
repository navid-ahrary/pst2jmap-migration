package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"os/signal"
	"syscall"

	"github.com/navid/pst2jmap-migration/internal/jmap"
	"github.com/navid/pst2jmap-migration/internal/model"
	"github.com/navid/pst2jmap-migration/internal/pst"
)

const JMAP_URL = "https://postmaster.collab24.net/jmap"

var version = "dev"

func main() {
	startedAt := time.Now()

	var (
		pstFile  string
		username string
		password string
		workers  int
	)

	flag.StringVar(&pstFile, "pst", "", "Src. Path to PST file")
	flag.StringVar(&username, "user", "", "Dest. Username")
	flag.StringVar(&password, "password", "", "Dest. Password")
	flag.IntVar(&workers, "workers", 20, "Number of transfer workers")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	validate(pstFile, username, password, workers)

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
	failures := pst.NewFailures()

	fmt.Printf("Found %d emails\n\n", total)
	fmt.Println("Starting migration...")

	folderCounts := pst.NewFolderCounts()

	jobs := make(chan model.Job, 100)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go pst.StartWorker(&wg, jobs, func(path string) error {
			return pst.ProcessEmail(path, client, mailboxes, state, stats, folderCounts, failures, stateFile)
		})
	}

	err = reader.Walk(func(path string) error {
		select {

		case <-ctx.Done():
			return filepath.SkipAll

		default:
		}

		jobs <- model.Job{
			Path: path,
		}

		return nil
	})

	close(jobs)
	wg.Wait()

	if err != nil {
		exit("migration failed", err)
	}

	stats.Finish()

	if ctx.Err() != nil {
		fmt.Println()
		fmt.Println("Migration stopped by user.")
	}

	err = pst.WriteReport(reportFile, stats, folderCounts.Snapshot(), failures.Snapshot())

	if err != nil {
		fmt.Printf("WARNING: failed to write report: %v\n", err)
	}

	fmt.Println()
	fmt.Println("Folder summary:")

	for folder, count := range folderCounts.Snapshot() {
		fmt.Printf("  %-15s %d\n", folder, count)
	}

	fmt.Println()
	fmt.Println(stats)
}

func validate(pstFile string, user string, pass string, workers int) {
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

	if workers < 1 {
		fmt.Fprintln(
			os.Stderr,
			"ERROR: --workers must be greater than 0",
		)
		os.Exit(1)
	}
}

func exit(msg string, err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", msg, err)
	os.Exit(1)
}
