package pst

import "sync"

type FolderCounts struct {
	mu sync.Mutex

	Counts map[string]int
}

func NewFolderCounts() *FolderCounts {
	return &FolderCounts{
		Counts: make(map[string]int),
	}
}

func (f *FolderCounts) Increment(folder string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.Counts[folder]++
}

func (f *FolderCounts) Snapshot() map[string]int {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := make(map[string]int)

	for k, v := range f.Counts {
		result[k] = v
	}

	return result
}
