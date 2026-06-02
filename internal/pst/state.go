package pst

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

type MigrationState struct {
	mu sync.Mutex

	Processed  map[string]bool `json:"processed"`
	MessageIDs map[string]bool `json:"message_ids"`
}

func LoadState(path string) (*MigrationState, error) {

	state := &MigrationState{
		Processed:  map[string]bool{},
		MessageIDs: map[string]bool{},
	}

	f, err := os.Open(path)

	if err != nil {

		if os.IsNotExist(err) {
			return state, nil
		}

		return nil, err
	}

	defer f.Close()

	info, err := f.Stat()

	if err != nil {
		return nil, err
	}

	if info.Size() == 0 {
		return state, nil
	}

	err = json.NewDecoder(f).Decode(state)

	if err != nil {

		if err == io.EOF {
			return state, nil
		}

		return nil, err
	}

	return state, nil

}

func (s *MigrationState) HasMessageID(id string) bool {

	if id == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.MessageIDs[id]
}

func (s *MigrationState) MarkMessageID(id string) {

	if id == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.MessageIDs[id] = true
}

func (s *MigrationState) Save(path string) error {

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

func (s *MigrationState) MarkProcessed(path string) {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Processed[path] = true
}

func (s *MigrationState) IsProcessed(path string) bool {

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Processed[path]
}
