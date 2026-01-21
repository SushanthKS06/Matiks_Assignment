package store

import (
	"fmt"
	"leaderboard-backend/models"
	"sort"
	"strings"
	"sync"
)

const (
	MaxPrefixLength = 4 // Limit prefix indexing to first 4 characters for memory efficiency
)

type MemoryStore struct {
	mu          sync.RWMutex
	users       map[string]*models.User // id -> user
	usersByName map[string][]string     // username prefix -> user ids (for search)
	ratingIndex *RatingBucketIndex
	skipList    *SkipList // O(log N) sorted user list
}

func NewMemoryStore(ratingIndex *RatingBucketIndex) *MemoryStore {
	return &MemoryStore{
		users:       make(map[string]*models.User),
		usersByName: make(map[string][]string),
		ratingIndex: ratingIndex,
		skipList:    NewSkipList(),
	}
}

func (m *MemoryStore) AddUser(user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.ID]; exists {
		return fmt.Errorf("user with ID %s already exists", user.ID)
	}

	m.users[user.ID] = user
	m.indexUsername(user.ID, user.Username)
	m.ratingIndex.IncrementBucket(user.Rating)

	// Insert into skip list - O(log N)
	m.skipList.Insert(user)

	return nil
}

func (m *MemoryStore) indexUsername(userID, username string) {
	lowerName := strings.ToLower(username)
	maxLen := len(lowerName)
	if maxLen > MaxPrefixLength {
		maxLen = MaxPrefixLength // Limit prefix length for memory efficiency
	}
	for i := 1; i <= maxLen; i++ {
		prefix := lowerName[:i]
		m.usersByName[prefix] = append(m.usersByName[prefix], userID)
	}
	// Also index full name for exact matches
	if len(lowerName) > MaxPrefixLength {
		m.usersByName[lowerName] = append(m.usersByName[lowerName], userID)
	}
}

func (m *MemoryStore) removeUsernameIndex(userID, username string) {
	lowerName := strings.ToLower(username)
	maxLen := len(lowerName)
	if maxLen > MaxPrefixLength {
		maxLen = MaxPrefixLength
	}
	for i := 1; i <= maxLen; i++ {
		prefix := lowerName[:i]
		ids := m.usersByName[prefix]
		for j, id := range ids {
			if id == userID {
				m.usersByName[prefix] = append(ids[:j], ids[j+1:]...)
				break
			}
		}
	}
	// Remove full name index
	if len(lowerName) > MaxPrefixLength {
		ids := m.usersByName[lowerName]
		for j, id := range ids {
			if id == userID {
				m.usersByName[lowerName] = append(ids[:j], ids[j+1:]...)
				break
			}
		}
	}
}

func (m *MemoryStore) GetUser(id string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID %s not found", id)
	}

	userCopy := *user
	return &userCopy, nil
}

func (m *MemoryStore) UpdateRating(id string, newRating int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		return fmt.Errorf("user with ID %s not found", id)
	}

	oldRating := user.Rating
	if oldRating != newRating {
		
		m.skipList.Remove(id)

		user.Rating = newRating
		m.ratingIndex.UpdateRating(oldRating, newRating)

		m.skipList.Insert(user)
	}

	return nil
}

func (m *MemoryStore) GetAllUsers() []*models.User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]*models.User, 0, len(m.users))
	for _, user := range m.users {
		userCopy := *user
		users = append(users, &userCopy)
	}
	return users
}

func (m *MemoryStore) GetUsersByRating(rating int) []*models.User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]*models.User, 0)
	for _, user := range m.users {
		if user.Rating == rating {
			userCopy := *user
			users = append(users, &userCopy)
		}
	}
	return users
}

func (m *MemoryStore) GetUserCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.users)
}

func (m *MemoryStore) SearchUsers(query string) []*models.User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if query == "" {
		return []*models.User{}
	}

	lowerQuery := strings.ToLower(strings.TrimSpace(query))
	if lowerQuery == "" {
		return []*models.User{}
	}

	lookupKey := lowerQuery
	if len(lookupKey) > MaxPrefixLength {
		lookupKey = lowerQuery[:MaxPrefixLength]
	}

	userIDs := m.usersByName[lookupKey]
	seen := make(map[string]bool)
	users := make([]*models.User, 0)

	for _, id := range userIDs {
		if seen[id] {
			continue
		}
		seen[id] = true

		if user, exists := m.users[id]; exists {
			if strings.Contains(strings.ToLower(user.Username), lowerQuery) {
				userCopy := *user
				users = append(users, &userCopy)
			}
		}
	}

	// Sort by rating descending
	sort.Slice(users, func(i, j int) bool {
		return users[i].Rating > users[j].Rating
	})

	// Limit results to prevent memory issues
	const maxSearchResults = 100
	if len(users) > maxSearchResults {
		users = users[:maxSearchResults]
	}

	return users
}

// GetTopUsers returns top N users by rating - O(log N + limit) using skip list
func (m *MemoryStore) GetTopUsers(limit int, offset int) []*models.User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Delegate to skip list - O(log N + limit)
	return m.skipList.GetTopN(limit, offset)
}

func (m *MemoryStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users = make(map[string]*models.User)
	m.usersByName = make(map[string][]string)
	m.skipList.Clear()
	m.ratingIndex.Clear()
}

func (m *MemoryStore) GetRandomUserID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for id := range m.users {
		return id
	}
	return ""
}

func (m *MemoryStore) GetAllUserIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.users))
	for id := range m.users {
		ids = append(ids, id)
	}
	return ids
}

// GetStats returns statistics about the memory store
func (m *MemoryStore) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_users":            len(m.users),
		"skip_list_size":         m.skipList.Length(),
		"username_index_entries": len(m.usersByName),
	}
}
