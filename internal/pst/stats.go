package pst

import (
	"fmt"
	"sync"
	"time"
)

type Stats struct {
	mu sync.Mutex

	StartedAt  time.Time
	FinishedAt time.Time

	TotalMessages int
	Processed     int
	Imported      int
	Failed        int
	Skipped       int
}

func NewStats(total int, startedAt time.Time) *Stats {

	return &Stats{
		StartedAt:     startedAt,
		TotalMessages: total,
	}
}

func (s *Stats) IncrementImported() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Imported++
	s.Processed++
}

func (s *Stats) IncrementFailed() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Failed++
	s.Processed++
}

func (s *Stats) IncrementSkipped() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Skipped++
	s.Processed++
}

func (s *Stats) Finish() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.FinishedAt = time.Now()
}

func (s *Stats) Duration() time.Duration {

	if s.FinishedAt.IsZero() {
		return time.Since(
			s.StartedAt,
		)
	}

	return s.FinishedAt.Sub(
		s.StartedAt,
	)
}

func (s *Stats) Progress() float64 {

	if s.TotalMessages == 0 {
		return 0
	}

	return float64(s.Processed) / float64(s.TotalMessages) * 100
}

func (s *Stats) String() string {

	s.mu.Lock()

	processed :=
		s.Processed

	total :=
		s.TotalMessages

	imported :=
		s.Imported

	failed :=
		s.Failed

	skipped :=
		s.Skipped

	started :=
		s.StartedAt

	finished :=
		s.FinishedAt

	s.mu.Unlock()

	var d time.Duration

	if finished.IsZero() {
		d = time.Since(started)
	} else {
		d = finished.Sub(started)
	}

	if d < time.Second {
		d =
			d.Round(
				time.Millisecond,
			)
	} else {
		d =
			d.Round(
				time.Second,
			)
	}

	progress :=
		0.0

	if total > 0 {
		progress =
			float64(processed) /
				float64(total) *
				100
	}

	return fmt.Sprintf(
		`Processed: %d/%d
Imported: %d
Failed: %d
Skipped: %d
Progress: %.1f%%
Duration: %s`,
		processed,
		total,
		imported,
		failed,
		skipped,
		progress,
		d,
	)
}
