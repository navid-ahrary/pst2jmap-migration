package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	csvreader "github.com/navid/pst2jmap-migration/internal/csv"
	"github.com/navid/pst2jmap-migration/internal/migration"
	"github.com/navid/pst2jmap-migration/internal/model"
)

const DefaultJMAPURL = "https://postmaster.collab24.net/jmap"

var version = "dev"

func main() {
	var (
		pstFile        string
		username       string
		password       string
		csvFile        string
		jmapURL        string
		workers        int
		mailboxWorkers int
	)

	flag.StringVar(&pstFile, "pst", "", "Source PST file")
	flag.StringVar(&username, "user", "", "Destination username")
	flag.StringVar(&password, "password", "", "Destination password")
	flag.StringVar(&csvFile, "csv", "", "CSV job file")
	flag.StringVar(&jmapURL, "jmap", DefaultJMAPURL, "JMAP endpoint")
	flag.IntVar(&mailboxWorkers, "mailbox-workers", 3, "Number of parallel mailbox migrations")
	flag.IntVar(&workers, "workers", 20, "Number of transfer workers")

	showVersion := flag.Bool("version", false, "Show version")

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	// CSV mode
	if csvFile != "" {
		if err := runCSV(ctx, csvFile, jmapURL, workers, mailboxWorkers); err != nil {
			exit("csv migration failed", err)
		}
		return
	}

	// Single mailbox mode
	validateSingle(pstFile, username, password, workers)
	if err := migration.Run(ctx, pstFile, username, password, jmapURL, workers, 1); err != nil {
		exit("migration failed", err)
	}
}

func runCSV(
	ctx context.Context,
	csvPath string,
	jmapURL string,
	workers int,
	mailboxWorkers int,
) error {

	jobs, err := csvreader.ReadJobs(csvPath)
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}

	fmt.Printf("Loaded %d migration jobs\n", len(jobs))

	jobChan := make(chan model.Job)

	var (
		wg sync.WaitGroup

		successCount int
		failedCount  int
		skippedCount int

		results []model.MigrationResult

		mu sync.Mutex
	)

	for i := 0; i < mailboxWorkers; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for job := range jobChan {

				select {

				case <-ctx.Done():
					return

				default:
				}

				fmt.Printf("\n[Mailbox Worker %d] Row %d - %s\n", workerID, job.Row, job.Email)

				if err := validateJob(job); err != nil {
					mu.Lock()

					skippedCount++

					results = append(
						results,
						model.MigrationResult{
							Row:     job.Row,
							Email:   job.Email,
							PSTFile: job.PSTFile,
							Status:  "SKIPPED",
							Error:   err.Error(),
						},
					)

					mu.Unlock()

					fmt.Printf("SKIPPED row %d (%s): %v\n", job.Row, job.Email, err)
					continue
				}

				err := migration.Run(
					ctx,
					job.PSTFile,
					job.Email,
					job.Password,
					jmapURL,
					workers,
					mailboxWorkers,
				)

				mu.Lock()

				if err != nil {
					failedCount++

					results = append(
						results,
						model.MigrationResult{
							Row:     job.Row,
							Email:   job.Email,
							PSTFile: job.PSTFile,
							Status:  "FAILED",
							Error:   err.Error(),
						},
					)

					mu.Unlock()
					fmt.Printf("FAILED row %d (%s): %v\n", job.Row, job.Email, err)

					continue
				}

				successCount++

				results = append(
					results,
					model.MigrationResult{
						Row:     job.Row,
						Email:   job.Email,
						PSTFile: job.PSTFile,
						Status:  "SUCCESS",
					},
				)

				mu.Unlock()
				fmt.Printf("SUCCESS row %d (%s)\n", job.Row, job.Email)
			}

		}(i + 1)
	}

	go func() {
		defer close(jobChan)

		for _, job := range jobs {
			select {
			case <-ctx.Done():
				return

			case jobChan <- job:
			}
		}
	}()

	wg.Wait()

	resultFile := csvPath + ".result.csv"

	if err := csvreader.WriteResults(resultFile, results); err != nil {
		fmt.Printf("WARNING: failed to write result file: %v\n", err)
	}

	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("Migration Summary")
	fmt.Println("==================================================")

	fmt.Printf("Total Jobs : %d\n", len(jobs))
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed    : %d\n", failedCount)
	fmt.Printf("Skipped   : %d\n", skippedCount)

	fmt.Printf("\nResult report written to: %s\n", resultFile)

	return nil
}

func setupSignalHandler(cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals

		fmt.Println()
		fmt.Println("Interrupt received, stopping...")

		cancel()
	}()
}

func validateSingle(
	pstFile string,
	user string,
	pass string,
	workers int,
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

	if workers < 1 {
		fmt.Fprintln(os.Stderr, "ERROR: --workers must be greater than 0")
		os.Exit(1)
	}
}

func validateJob(job model.Job) error {
	if job.Email == "" {
		return fmt.Errorf("missing email")
	}

	if job.Password == "" {
		return fmt.Errorf("missing password")
	}

	if job.PSTFile == "" {
		return fmt.Errorf("missing pst file")
	}

	if _, err := os.Stat(job.PSTFile); err != nil {
		return fmt.Errorf("pst file not found: %s", job.PSTFile)
	}

	return nil
}

func exit(msg string, err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", msg, err)
	os.Exit(1)
}
