package pst

import (
	"encoding/json"
	"os"
	"time"
)

type MigrationReport struct {
	StartedAt     time.Time `json:"started_at"`
	FinishedAt    time.Time `json:"finished_at"`
	Duration      string    `json:"duration"`
	TotalMessages int       `json:"total_messages"`
	Processed     int       `json:"processed"`
	Imported      int       `json:"imported"`

	Failed   int       `json:"failed"`
	Failures []Failure `json:"failures"`

	Skipped int            `json:"skipped"`
	Folders map[string]int `json:"folders"`
}

type Failure struct {
	Path      string `json:"path"`
	Subject   string `json:"subject,omitempty"`
	Operation string `json:"operation"`
	Error     string `json:"error"`
}

func WriteReport(path string, stats *Stats, folders map[string]int, failures []Failure) error {

	snapshot := stats.Snapshot()

	report := MigrationReport{
		StartedAt:     snapshot.StartedAt,
		FinishedAt:    snapshot.FinishedAt,
		Duration:      stats.Duration().String(),
		TotalMessages: snapshot.TotalMessages,
		Processed:     snapshot.Processed,
		Imported:      snapshot.Imported,
		Failed:        snapshot.Failed,
		Failures:      failures,
		Skipped:       snapshot.Skipped,
		Folders:       folders,
	}

	f, err := os.Create(path)

	if err != nil {
		return err
	}

	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	return enc.Encode(report)
}
