package store

import (
	"encoding/json"
	"fmt"
	"io"
	"leaderboard-backend/models"
	"os"
	"path/filepath"
	"sync"
)

// Persistence handles saving and loading data
type Persistence struct {
	mu       sync.Mutex
	filePath string
}

// PersistenceData is the structure saved to disk
type PersistenceData struct {
	Users   []*models.User `json:"users"`
	Version int            `json:"version"`
}

// NewPersistence creates a new persistence handler
func NewPersistence(filePath string) *Persistence {
	return &Persistence{
		filePath: filePath,
	}
}

// Save writes all users to disk atomically
func (p *Persistence) Save(store *MemoryStore) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get all users
	users := store.GetAllUsers()

	data := PersistenceData{
		Users:   users,
		Version: 1,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(p.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to temp file first (atomic write)
	tempPath := p.filePath + ".tmp"
	if err := os.WriteFile(tempPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Rename temp file to actual file (atomic on most filesystems)
	if err := os.Rename(tempPath, p.filePath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// Load reads users from disk and populates the store
func (p *Persistence) Load(store *MemoryStore, ratingIndex *RatingBucketIndex) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(p.filePath); os.IsNotExist(err) {
		return nil // No data to load, not an error
	}

	// Open file
	file, err := os.Open(p.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read all content
	jsonData, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal JSON
	var data PersistenceData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Clear existing data
	store.Clear()

	// Load users
	for _, user := range data.Users {
		if err := store.AddUser(user); err != nil {
			// Log but don't fail - continue loading other users
			fmt.Printf("Warning: failed to load user %s: %v\n", user.ID, err)
		}
	}

	return nil
}

// Exists checks if persistence file exists
func (p *Persistence) Exists() bool {
	_, err := os.Stat(p.filePath)
	return err == nil
}

// GetPath returns the persistence file path
func (p *Persistence) GetPath() string {
	return p.filePath
}

// Delete removes the persistence file
func (p *Persistence) Delete() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return os.Remove(p.filePath)
}
