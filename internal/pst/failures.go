package pst

import "sync"

type Failures struct {
	mu sync.Mutex

	Items []Failure
}

func NewFailures() *Failures {
	return &Failures{}
}

func (f *Failures) Add(item Failure) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.Items = append(f.Items, item)
}

func (f *Failures) Snapshot() []Failure {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := make([]Failure, len(f.Items))

	copy(result, f.Items)

	return result
}
