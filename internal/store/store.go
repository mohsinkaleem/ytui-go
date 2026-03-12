package store

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	bolt "go.etcd.io/bbolt"
)

var (
	bucketSearchHistory = []byte("search_history")
	bucketDownloads     = []byte("downloads")
	bucketSettings      = []byte("settings")
)

// allBuckets lists every bucket the store needs.
var allBuckets = [][]byte{bucketSearchHistory, bucketDownloads, bucketSettings}

// initBuckets creates all required buckets in the database.
func initBuckets(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		for _, b := range allBuckets {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
}

// Store wraps bbolt for persistence
type Store struct {
	db       *bolt.DB
	path     string
	readOnly bool // true when opened with a shared lock (another instance holds the write lock)
}

// New opens or creates the bbolt database.
// If the database is locked by another instance, it falls back to read-only mode.
func New() (*Store, error) {
	dbDir := filepath.Join(xdg.DataHome, "ytui-go")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dbDir, "ytui-go.db")

	// Try read-write first
	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err == nil {
		if err := initBuckets(db); err != nil {
			db.Close()
			return nil, fmt.Errorf("init buckets: %w", err)
		}
		return &Store{db: db, path: dbPath, readOnly: false}, nil
	}

	// Read-write failed (likely locked by another instance) – try read-only
	db, roErr := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 1 * time.Second, ReadOnly: true})
	if roErr != nil {
		return nil, fmt.Errorf("open db (read-write: %v; read-only: %w)", err, roErr)
	}

	return &Store{db: db, path: dbPath, readOnly: true}, nil
}

// NewWithPath opens a bbolt database at a specific path (useful for testing)
func NewWithPath(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := initBuckets(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("init buckets: %w", err)
	}

	return &Store{db: db, path: dbPath, readOnly: false}, nil
}

// Close closes the database
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Path returns the database file path
func (s *Store) Path() string {
	return s.path
}

// IsReadOnly returns true when the store was opened in read-only mode
// (e.g. another instance holds the write lock).
func (s *Store) IsReadOnly() bool {
	return s.readOnly
}
