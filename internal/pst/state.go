package pst

import (
	"encoding/json"
	"os"
	"sync"
)

type MigrationState struct {
	mu sync.Mutex

	Processed map[string]bool `json:"processed"`
}

func LoadState(path string) (*MigrationState, error) {

	state := &MigrationState{
		Processed: map[string]bool{},
	}

	f, err := os.Open(path)

	if err != nil {

		if os.IsNotExist(err) {
			return state, nil
		}

		return nil, err
	}

	defer f.Close()

	err = json.NewDecoder(f).Decode(state)

	return state, err
}

func (s *MigrationState) Save(
	path string,
) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	tmp := path + ".tmp"

	f, err := os.Create(tmp)

	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	err = enc.Encode(s)

	_ = f.Close()

	if err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func (s *MigrationState) MarkProcessed(
	path string,
) {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Processed[path] = true
}

func (s *MigrationState) IsProcessed(
	path string,
) bool {

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Processed[path]
}
