package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"
)

// --- Search History ---

// SaveSearch persists a search entry keyed by reverse timestamp
func (s *Store) SaveSearch(entry SearchEntry) error {
	if s.readOnly {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSearchHistory)
		// Key: reverse timestamp for natural ordering (newest first)
		key := fmt.Sprintf("%020d", time.Now().UnixNano())
		data, err := entry.Marshal()
		if err != nil {
			return err
		}
		return b.Put([]byte(key), data)
	})
}

// GetSearchHistory returns the most recent `limit` search entries
func (s *Store) GetSearchHistory(limit int) ([]SearchEntry, error) {
	var entries []SearchEntry
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSearchHistory)
		c := b.Cursor()
		count := 0
		// Iterate in reverse (newest first)
		for k, v := c.Last(); k != nil && count < limit; k, v = c.Prev() {
			var e SearchEntry
			if err := e.Unmarshal(v); err != nil {
				continue
			}
			entries = append(entries, e)
			count++
		}
		return nil
	})
	return entries, err
}

// DeleteSearch removes a search entry by its URL
func (s *Store) DeleteSearch(url string) error {
	if s.readOnly {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSearchHistory)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var e SearchEntry
			if err := e.Unmarshal(v); err != nil {
				continue
			}
			if e.URL == url {
				if err := b.Delete(k); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// ClearHistory removes all search history
func (s *Store) ClearHistory() error {
	if s.readOnly {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(bucketSearchHistory); err != nil {
			return err
		}
		_, err := tx.CreateBucket(bucketSearchHistory)
		return err
	})
}

// --- Downloads ---

// SaveDownload persists a download record keyed by UUID
func (s *Store) SaveDownload(record DownloadRecord) error {
	if s.readOnly {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDownloads)
		data, err := record.Marshal()
		if err != nil {
			return err
		}
		return b.Put([]byte(record.ID), data)
	})
}

// GetDownload retrieves a download record by ID
func (s *Store) GetDownload(id string) (*DownloadRecord, error) {
	var record DownloadRecord
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDownloads)
		data := b.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("download %s not found", id)
		}
		return record.Unmarshal(data)
	})
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetAllDownloads returns all download records sorted by creation time (newest first)
func (s *Store) GetAllDownloads() ([]DownloadRecord, error) {
	var records []DownloadRecord
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDownloads)
		return b.ForEach(func(k, v []byte) error {
			var r DownloadRecord
			if err := r.Unmarshal(v); err != nil {
				return nil // skip corrupt records
			}
			records = append(records, r)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records, nil
}

// GetIncomplete returns downloads that aren't completed or cancelled
func (s *Store) GetIncomplete() ([]DownloadRecord, error) {
	all, err := s.GetAllDownloads()
	if err != nil {
		return nil, err
	}
	var incomplete []DownloadRecord
	for _, r := range all {
		if r.State != "completed" && r.State != "cancelled" {
			incomplete = append(incomplete, r)
		}
	}
	return incomplete, nil
}

// UpdateDownloadState updates just the state field of a download
func (s *Store) UpdateDownloadState(id, state string) error {
	if s.readOnly {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDownloads)
		data := b.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("download %s not found", id)
		}
		var r DownloadRecord
		if err := r.Unmarshal(data); err != nil {
			return err
		}
		r.State = state
		if state == "completed" {
			r.CompletedAt = time.Now()
		}
		updated, err := r.Marshal()
		if err != nil {
			return err
		}
		return b.Put([]byte(id), updated)
	})
}

// DeleteDownload removes a download record
func (s *Store) DeleteDownload(id string) error {
	if s.readOnly {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDownloads)
		return b.Delete([]byte(id))
	})
}

// --- Settings ---

const settingsKey = "app_settings"

// GetSettings retrieves app settings, returning defaults if not set
func (s *Store) GetSettings() (Settings, error) {
	settings := DefaultSettings()
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSettings)
		data := b.Get([]byte(settingsKey))
		if data == nil {
			return nil // return defaults
		}
		return json.Unmarshal(data, &settings)
	})
	return settings, err
}

// SaveSettings persists app settings
func (s *Store) SaveSettings(settings Settings) error {
	if s.readOnly {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSettings)
		data, err := json.Marshal(settings)
		if err != nil {
			return err
		}
		return b.Put([]byte(settingsKey), data)
	})
}
