package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type kycRecord struct {
	ID             string       `json:"id"`
	ProcessID      string       `json:"processId,omitempty"`
	Status         string       `json:"status"`
	Result         string       `json:"result,omitempty"`
	Identity       identityData `json:"identity,omitempty"`
	IdentityEdited bool         `json:"identityEdited,omitempty"`
	Error          string       `json:"error,omitempty"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type identityData struct {
	CURP      string `json:"curp"`
	BirthDate string `json:"birthDate"`
	FullName  string `json:"fullName"`
}

type recordStore struct {
	mu       sync.Mutex
	filePath string
	records  map[string]kycRecord
}

func newRecordStore(filePath string) (*recordStore, error) {
	store := &recordStore{
		filePath: filePath,
		records:  make(map[string]kycRecord),
	}
	data, err := os.ReadFile(filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err == nil {
		var records []kycRecord
		if err := json.Unmarshal(data, &records); err != nil {
			return nil, err
		}
		for _, record := range records {
			store.records[record.ID] = record
		}
	}
	return store, nil
}

func (s *recordStore) save(record kycRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := make(map[string]kycRecord, len(s.records)+1)
	for id, current := range s.records {
		next[id] = current
	}
	next[record.ID] = record
	records := make([]kycRecord, 0, len(next))
	for _, current := range next {
		records = append(records, current)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.filePath), 0700); err != nil {
		return err
	}
	tempPath := s.filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(tempPath, s.filePath); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	s.records = next
	return nil
}

func (s *recordStore) get(id string) (kycRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.records[id]
	return record, ok
}

func (s *recordStore) list() []kycRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	records := make([]kycRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records
}
