package store

import (
	"leaderboard-backend/models"
	"math/rand"
	"sync"
)

const (
	MaxLevel    = 16   
	Probability = 0.25 // Probability for level promotion
)

// SkipListNode represents a node in the skip list
type SkipListNode struct {
	User    *models.User
	forward []*SkipListNode
}

// SkipList is a probabilistic data structure for O(log N) operations
type SkipList struct {
	mu      sync.RWMutex
	head    *SkipListNode
	level   int
	length  int
	nodeMap map[string]*SkipListNode // userID -> node for O(1) lookup
}

// NewSkipList creates a new skip list
func NewSkipList() *SkipList {
	head := &SkipListNode{
		User:    nil,
		forward: make([]*SkipListNode, MaxLevel),
	}
	return &SkipList{
		head:    head,
		level:   0,
		length:  0,
		nodeMap: make(map[string]*SkipListNode),
	}
}

// randomLevel generates a random level for a new node
func (sl *SkipList) randomLevel() int {
	level := 0
	for level < MaxLevel-1 && rand.Float64() < Probability {
		level++
	}
	return level
}


func compare(a, b *models.User) int {
	if a.Rating > b.Rating {
		return 1 // a comes first (higher rating)
	}
	if a.Rating < b.Rating {
		return -1 // a comes later (lower rating)
	}
	// Same rating, sort by username ascending for stable order
	if a.Username < b.Username {
		return 1 
	}
	if a.Username > b.Username {
		return -1 
	}
	return 0
}

// Insert adds a user to the skip list - O(log N)
func (sl *SkipList) Insert(user *models.User) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Check if already exists
	if _, exists := sl.nodeMap[user.ID]; exists {
		return
	}

	update := make([]*SkipListNode, MaxLevel)
	current := sl.head

	// Find position (descending by rating, ascending by username)
	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && compare(current.forward[i].User, user) > 0 {
			current = current.forward[i]
		}
		update[i] = current
	}

	// Generate random level for new node
	newLevel := sl.randomLevel()

	// Update skip list level if needed
	if newLevel > sl.level {
		for i := sl.level + 1; i <= newLevel; i++ {
			update[i] = sl.head
		}
		sl.level = newLevel
	}

	// Create new node
	newNode := &SkipListNode{
		User:    user,
		forward: make([]*SkipListNode, newLevel+1),
	}

	// Insert node at each level
	for i := 0; i <= newLevel; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	sl.nodeMap[user.ID] = newNode
	sl.length++
}

// Remove deletes a user from the skip list - O(log N)
func (sl *SkipList) Remove(userID string) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	node, exists := sl.nodeMap[userID]
	if !exists {
		return false
	}

	user := node.User
	update := make([]*SkipListNode, MaxLevel)
	current := sl.head

	// Find the node
	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && compare(current.forward[i].User, user) > 0 {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	// Find exact node (might be different node with same rating/username)
	for current != nil && current != node {
		if compare(current.User, user) != 0 {
			break
		}
		for i := sl.level; i >= 0; i-- {
			if update[i].forward[i] == current {
				update[i] = current
			}
		}
		current = current.forward[0]
	}

	if current != node {
		// Node not found in expected position, search manually
		current = sl.head
		for i := sl.level; i >= 0; i-- {
			for current.forward[i] != nil && current.forward[i] != node {
				current = current.forward[i]
			}
			update[i] = current
		}
	}

	// Remove node from each level
	for i := 0; i <= sl.level && update[i].forward[i] == node; i++ {
		update[i].forward[i] = node.forward[i]
	}

	// Update level if needed
	for sl.level > 0 && sl.head.forward[sl.level] == nil {
		sl.level--
	}

	delete(sl.nodeMap, userID)
	sl.length--
	return true
}

// Update removes and re-inserts a user with new rating - O(log N)
func (sl *SkipList) Update(user *models.User) {
	sl.Remove(user.ID)
	sl.Insert(user)
}

// GetTopN returns top N users starting from offset - O(log N + limit)
func (sl *SkipList) GetTopN(limit, offset int) []*models.User {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if offset >= sl.length {
		return []*models.User{}
	}

	// Skip to offset
	current := sl.head.forward[0]
	for i := 0; i < offset && current != nil; i++ {
		current = current.forward[0]
	}

	// Collect limit users
	result := make([]*models.User, 0, limit)
	for i := 0; i < limit && current != nil; i++ {
		userCopy := *current.User
		result = append(result, &userCopy)
		current = current.forward[0]
	}

	return result
}

// Length returns the number of elements in the skip list
func (sl *SkipList) Length() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.length
}

// Contains checks if a user exists in the skip list
func (sl *SkipList) Contains(userID string) bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	_, exists := sl.nodeMap[userID]
	return exists
}

// Clear removes all elements from the skip list
func (sl *SkipList) Clear() {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.head = &SkipListNode{
		User:    nil,
		forward: make([]*SkipListNode, MaxLevel),
	}
	sl.level = 0
	sl.length = 0
	sl.nodeMap = make(map[string]*SkipListNode)
}

// GetAllUserIDs returns all user IDs (for simulator)
func (sl *SkipList) GetAllUserIDs() []string {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	ids := make([]string, 0, sl.length)
	for id := range sl.nodeMap {
		ids = append(ids, id)
	}
	return ids
}
