package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestStore(t *testing.T) (*Store, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := NewWithPath(dbPath)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	return s, func() {
		s.Close()
		os.RemoveAll(dir)
	}
}

func TestNewWithPath(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "subdir", "test.db")
	s, err := NewWithPath(dbPath)
	if err != nil {
		t.Fatalf("NewWithPath failed: %v", err)
	}
	defer s.Close()

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file should be created")
	}
}

func TestSearchHistory(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// Initially empty
	entries, err := s.GetSearchHistory(10)
	if err != nil {
		t.Fatalf("GetSearchHistory failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}

	// Save a few entries
	for i := 0; i < 5; i++ {
		err := s.SaveSearch(SearchEntry{
			URL:       "https://example.com/" + string(rune('a'+i)),
			Title:     "Search " + string(rune('A'+i)),
			Timestamp: time.Now(),
			Type:      "search",
		})
		if err != nil {
			t.Fatalf("SaveSearch failed: %v", err)
		}
		time.Sleep(time.Millisecond) // ensure different timestamps
	}

	// Get all
	entries, err = s.GetSearchHistory(10)
	if err != nil {
		t.Fatalf("GetSearchHistory failed: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}

	// Should be newest first
	if entries[0].Title != "Search E" {
		t.Errorf("newest entry should be last saved, got %q", entries[0].Title)
	}

	// Limit works
	entries, err = s.GetSearchHistory(2)
	if err != nil {
		t.Fatalf("GetSearchHistory with limit failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries with limit, got %d", len(entries))
	}
}

func TestClearHistory(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	s.SaveSearch(SearchEntry{URL: "test", Title: "test", Timestamp: time.Now()})

	err := s.ClearHistory()
	if err != nil {
		t.Fatalf("ClearHistory failed: %v", err)
	}

	entries, _ := s.GetSearchHistory(10)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(entries))
	}
}

func TestDownloadCRUD(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	record := DownloadRecord{
		ID:         "test-id-1",
		VideoID:    "abc123",
		Title:      "Test Video",
		URL:        "https://youtube.com/watch?v=abc123",
		FormatID:   "137",
		OutputPath: "/tmp/test.mp4",
		State:      "downloading",
		FileSize:   1000000,
		CreatedAt:  time.Now(),
	}

	// Save
	err := s.SaveDownload(record)
	if err != nil {
		t.Fatalf("SaveDownload failed: %v", err)
	}

	// Get
	got, err := s.GetDownload("test-id-1")
	if err != nil {
		t.Fatalf("GetDownload failed: %v", err)
	}
	if got.Title != "Test Video" {
		t.Errorf("Title = %q, want %q", got.Title, "Test Video")
	}
	if got.State != "downloading" {
		t.Errorf("State = %q, want %q", got.State, "downloading")
	}

	// Get non-existent
	_, err = s.GetDownload("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent download")
	}
}

func TestUpdateDownloadState(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	record := DownloadRecord{
		ID:        "update-test",
		Title:     "Update Test",
		State:     "downloading",
		CreatedAt: time.Now(),
	}
	s.SaveDownload(record)

	// Update to completed
	err := s.UpdateDownloadState("update-test", "completed")
	if err != nil {
		t.Fatalf("UpdateDownloadState failed: %v", err)
	}

	got, _ := s.GetDownload("update-test")
	if got.State != "completed" {
		t.Errorf("State = %q, want %q", got.State, "completed")
	}
	if got.CompletedAt.IsZero() {
		t.Error("CompletedAt should be set when state is completed")
	}

	// Update non-existent
	err = s.UpdateDownloadState("nonexistent", "completed")
	if err == nil {
		t.Error("expected error for non-existent download")
	}
}

func TestGetAllDownloads(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// Save multiple records with different creation times
	for i := 0; i < 3; i++ {
		s.SaveDownload(DownloadRecord{
			ID:        string(rune('a' + i)),
			Title:     "Video " + string(rune('A'+i)),
			State:     "completed",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		})
	}

	records, err := s.GetAllDownloads()
	if err != nil {
		t.Fatalf("GetAllDownloads failed: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}

	// Should be newest first
	if records[0].ID != "c" {
		t.Errorf("newest should be first, got %q", records[0].ID)
	}
}

func TestGetIncomplete(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	s.SaveDownload(DownloadRecord{ID: "1", State: "completed", CreatedAt: time.Now()})
	s.SaveDownload(DownloadRecord{ID: "2", State: "downloading", CreatedAt: time.Now()})
	s.SaveDownload(DownloadRecord{ID: "3", State: "cancelled", CreatedAt: time.Now()})
	s.SaveDownload(DownloadRecord{ID: "4", State: "failed", CreatedAt: time.Now()})
	s.SaveDownload(DownloadRecord{ID: "5", State: "paused", CreatedAt: time.Now()})

	incomplete, err := s.GetIncomplete()
	if err != nil {
		t.Fatalf("GetIncomplete failed: %v", err)
	}
	// Should include downloading, failed, paused (not completed, not cancelled)
	if len(incomplete) != 3 {
		t.Errorf("expected 3 incomplete, got %d", len(incomplete))
	}
}

func TestDeleteDownload(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	s.SaveDownload(DownloadRecord{ID: "del-test", Title: "Delete Me", State: "completed", CreatedAt: time.Now()})

	err := s.DeleteDownload("del-test")
	if err != nil {
		t.Fatalf("DeleteDownload failed: %v", err)
	}

	_, err = s.GetDownload("del-test")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestSettings(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// Get defaults
	settings, err := s.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	defaults := DefaultSettings()
	if settings.Theme != defaults.Theme {
		t.Errorf("default Theme = %q, want %q", settings.Theme, defaults.Theme)
	}
	if settings.MaxConcurrent != defaults.MaxConcurrent {
		t.Errorf("default MaxConcurrent = %d, want %d", settings.MaxConcurrent, defaults.MaxConcurrent)
	}

	// Save custom settings
	custom := Settings{
		DownloadDir:   "/custom/dir",
		MaxConcurrent: 5,
		DefaultFormat: "bestvideo+bestaudio",
		Theme:         "latte",
		EmbedSubs:     true,
		EmbedMetadata: true,
		EmbedChapters: false,
		MpvPath:       "/usr/local/bin/mpv",
	}
	err = s.SaveSettings(custom)
	if err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	// Retrieve
	got, err := s.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings after save failed: %v", err)
	}
	if got.Theme != "latte" {
		t.Errorf("Theme = %q, want %q", got.Theme, "latte")
	}
	if got.MaxConcurrent != 5 {
		t.Errorf("MaxConcurrent = %d, want 5", got.MaxConcurrent)
	}
	if !got.EmbedSubs {
		t.Error("EmbedSubs should be true")
	}
}

func TestSearchEntryMarshalUnmarshal(t *testing.T) {
	entry := SearchEntry{
		URL:       "https://example.com",
		Title:     "Test",
		Timestamp: time.Now().Truncate(time.Millisecond),
		Type:      "search",
	}

	data, err := entry.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got SearchEntry
	err = got.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.URL != entry.URL {
		t.Errorf("URL = %q, want %q", got.URL, entry.URL)
	}
	if got.Title != entry.Title {
		t.Errorf("Title = %q, want %q", got.Title, entry.Title)
	}
}

func TestDownloadRecordMarshalUnmarshal(t *testing.T) {
	record := DownloadRecord{
		ID:        "test-id",
		VideoID:   "abc",
		Title:     "Test Video",
		State:     "completed",
		CreatedAt: time.Now().Truncate(time.Millisecond),
	}

	data, err := record.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got DownloadRecord
	err = got.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.ID != record.ID {
		t.Errorf("ID = %q, want %q", got.ID, record.ID)
	}
	if got.State != record.State {
		t.Errorf("State = %q, want %q", got.State, record.State)
	}
}

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()
	if s.DownloadDir == "" {
		t.Error("DownloadDir should not be empty")
	}
	if s.MaxConcurrent <= 0 {
		t.Error("MaxConcurrent should be positive")
	}
	if s.DefaultFormat == "" {
		t.Error("DefaultFormat should not be empty")
	}
	if s.Theme == "" {
		t.Error("Theme should not be empty")
	}
}
