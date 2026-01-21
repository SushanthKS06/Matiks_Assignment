package tests

import (
	"sync"
	"testing"

	"leaderboard-backend/models"
	"leaderboard-backend/store"
)

func TestThousandUsersWithSameRating(t *testing.T) {
	idx := store.NewRatingBucketIndex()
	ms := store.NewMemoryStore(idx)

	// Add 1000 users with rating 3000
	for i := 0; i < 1000; i++ {
		user := &models.User{
			ID:       "user-same-" + string(rune('a'+i%26)) + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)),
			Username: "player" + string(rune('0'+i%10)),
			Rating:   3000,
		}
		_ = ms.AddUser(user)
	}

	// All users with same rating should have rank 1
	rank := idx.GetRank(3000)
	if rank != 1 {
		t.Errorf("Expected rank 1 for rating 3000, got %d", rank)
	}

	// Add one user with higher rating
	topUser := &models.User{
		ID:       "top-user",
		Username: "champion",
		Rating:   5000,
	}
	_ = ms.AddUser(topUser)

	// Now all 3000-rated users should have rank 2
	rank = idx.GetRank(3000)
	if rank != 2 {
		t.Errorf("Expected rank 2 for rating 3000 after adding higher user, got %d", rank)
	}

	// Top user should be rank 1
	rank = idx.GetRank(5000)
	if rank != 1 {
		t.Errorf("Expected rank 1 for rating 5000, got %d", rank)
	}

	// Total users should be 1001
	if idx.GetTotalUsers() != 1001 {
		t.Errorf("Expected 1001 total users, got %d", idx.GetTotalUsers())
	}
}

func TestExactBoundaryRatings(t *testing.T) {
	idx := store.NewRatingBucketIndex()

	// Test minimum rating boundary
	idx.IncrementBucket(100)
	if rank := idx.GetRank(100); rank != 1 {
		t.Errorf("Min rating 100: expected rank 1, got %d", rank)
	}

	// Test maximum rating boundary
	idx.IncrementBucket(5000)
	if rank := idx.GetRank(5000); rank != 1 {
		t.Errorf("Max rating 5000: expected rank 1, got %d", rank)
	}

	// Now min rating should be rank 2
	if rank := idx.GetRank(100); rank != 2 {
		t.Errorf("After adding 5000, rating 100: expected rank 2, got %d", rank)
	}

	// Test below minimum (should clamp to 100)
	idx.IncrementBucket(50) // Should be treated as 100
	if rank := idx.GetRank(100); rank != 2 {
		t.Errorf("After adding sub-min rating, rating 100: expected rank 2, got %d", rank)
	}

	// Test above maximum (should clamp to 5000)
	idx.IncrementBucket(6000) // Should be treated as 5000
	if rank := idx.GetRank(5000); rank != 1 {
		t.Errorf("After adding over-max rating, rating 5000: expected rank 1, got %d", rank)
	}
}

func TestIncrementalCumulativeUpdate(t *testing.T) {
	idx := store.NewRatingBucketIndex()

	// Setup: Add users at various ratings
	idx.IncrementBucket(1000)
	idx.IncrementBucket(2000)
	idx.IncrementBucket(3000)
	idx.IncrementBucket(4000)
	idx.IncrementBucket(5000)

	// Initial ranks
	if rank := idx.GetRank(5000); rank != 1 {
		t.Errorf("Initial: 5000 should be rank 1, got %d", rank)
	}
	if rank := idx.GetRank(4000); rank != 2 {
		t.Errorf("Initial: 4000 should be rank 2, got %d", rank)
	}
	if rank := idx.GetRank(3000); rank != 3 {
		t.Errorf("Initial: 3000 should be rank 3, got %d", rank)
	}

	// Move user from 3000 to 4500 (incremental update test)
	idx.UpdateRating(3000, 4500)

	// New ranks after move
	if rank := idx.GetRank(5000); rank != 1 {
		t.Errorf("After move: 5000 should be rank 1, got %d", rank)
	}
	if rank := idx.GetRank(4500); rank != 2 {
		t.Errorf("After move: 4500 should be rank 2, got %d", rank)
	}
	if rank := idx.GetRank(4000); rank != 3 {
		t.Errorf("After move: 4000 should be rank 3, got %d", rank)
	}
	// 3000 bucket is now empty, but rank should still work
	if rank := idx.GetRank(3000); rank != 4 {
		t.Errorf("After move: 3000 should be rank 4, got %d", rank)
	}
}

func TestSearchWithSpecialCharacters(t *testing.T) {
	idx := store.NewRatingBucketIndex()
	ms := store.NewMemoryStore(idx)

	// Add users with various names
	users := []struct {
		id       string
		username string
		rating   int
	}{
		{"1", "john_doe", 3000},
		{"2", "jane-doe", 2500},
		{"3", "bob123", 2000},
		{"4", "alice", 1500},
	}

	for _, u := range users {
		user := &models.User{ID: u.id, Username: u.username, Rating: u.rating}
		_ = ms.AddUser(user)
	}

	// Search with empty string should return empty
	results := ms.SearchUsers("")
	if len(results) != 0 {
		t.Errorf("Empty search should return 0 results, got %d", len(results))
	}

	// Search with spaces should be trimmed
	results = ms.SearchUsers("   ")
	if len(results) != 0 {
		t.Errorf("Whitespace search should return 0 results, got %d", len(results))
	}

	// Normal search should work
	results = ms.SearchUsers("john")
	if len(results) != 1 {
		t.Errorf("Search 'john' should return 1 result, got %d", len(results))
	}
}

func TestStressGetTopUsers(t *testing.T) {
	idx := store.NewRatingBucketIndex()
	ms := store.NewMemoryStore(idx)

	// Add 10000 users with varying ratings
	for i := 0; i < 10000; i++ {
		rating := 100 + (i % 4901)
		user := &models.User{
			ID:       "user-" + string(rune('a'+i%26)) + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)) + string(rune('0'+(i/100)%10)),
			Username: "player" + string(rune('0'+i%10)),
			Rating:   rating,
		}
		_ = ms.AddUser(user)
	}

	// GetTopUsers should be fast (O(limit) not O(N log N))
	topUsers := ms.GetTopUsers(50, 0)
	if len(topUsers) != 50 {
		t.Errorf("Expected 50 top users, got %d", len(topUsers))
	}

	// Verify sorted order
	for i := 1; i < len(topUsers); i++ {
		if topUsers[i].Rating > topUsers[i-1].Rating {
			t.Errorf("Users not sorted correctly: rating %d > %d at positions %d, %d",
				topUsers[i].Rating, topUsers[i-1].Rating, i-1, i)
		}
	}

	// Test pagination
	page2 := ms.GetTopUsers(50, 50)
	if len(page2) != 50 {
		t.Errorf("Expected 50 users on page 2, got %d", len(page2))
	}

	// First user of page 2 should have rating <= last user of page 1
	if page2[0].Rating > topUsers[49].Rating {
		t.Errorf("Page 2 first user (%d) has higher rating than page 1 last (%d)",
			page2[0].Rating, topUsers[49].Rating)
	}
}

func TestConcurrentRatingUpdates(t *testing.T) {
	idx := store.NewRatingBucketIndex()
	ms := store.NewMemoryStore(idx)

	// Add initial users
	for i := 0; i < 100; i++ {
		user := &models.User{
			ID:       "user-" + string(rune('a'+i%26)) + string(rune('0'+i%10)),
			Username: "user" + string(rune('0'+i%10)),
			Rating:   2500, // All start at 2500
		}
		_ = ms.AddUser(user)
	}

	// Concurrent updates
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				userID := "user-" + string(rune('a'+i%26)) + string(rune('0'+i%10))
				newRating := 100 + (i*j)%4901
				_ = ms.UpdateRating(userID, newRating)
			}
		}(i)
	}

	// Concurrent reads during updates
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = ms.GetTopUsers(50, 0)
			}
		}()
	}

	wg.Wait()

	// System should still be consistent
	totalUsers := ms.GetUserCount()
	if totalUsers != 100 {
		t.Errorf("Expected 100 users after concurrent updates, got %d", totalUsers)
	}
}
