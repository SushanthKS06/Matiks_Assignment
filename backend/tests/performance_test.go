package tests

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"leaderboard-backend/models"
	"leaderboard-backend/store"
)

// usage: go test -v -run=TestUpdatePerformance ./tests
func TestUpdatePerformance(t *testing.T) {
	fmt.Println("🚀 Starting Performance Test...")

	// Setup
	ri := store.NewRatingBucketIndex()
	ms := store.NewMemoryStore(ri)

	userCount := 10000
	updateCount := 10000

	fmt.Printf("📂 Seeding %d users...\n", userCount)
	// Seed 10k users
	for i := 0; i < userCount; i++ {
		ms.AddUser(&models.User{
			ID:       fmt.Sprintf("user_%d", i),
			Username: fmt.Sprintf("user_%d", i),
			Rating:   rand.Intn(4900) + 100,
		})
	}
	fmt.Println("✅ Seeding complete.")

	// Measure 10k updates
	start := time.Now()
	fmt.Printf("⚡ Executing %d random rating updates...\n", updateCount)

	for i := 0; i < updateCount; i++ {
		userID := fmt.Sprintf("user_%d", rand.Intn(userCount))
		newRating := rand.Intn(4900) + 100
		err := ms.UpdateRating(userID, newRating)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	}

	duration := time.Since(start)
	opsPerSec := float64(updateCount) / duration.Seconds()

	fmt.Printf("⏱️  Time taken: %v\n", duration)
	fmt.Printf("📈 Throughput: %.2f updates/sec\n", opsPerSec)

	// Threshold: If it takes more than 1 second for 10k updates, something is wrong.
	// 10k * log(10k) is fast. 10k * 10k (linear scan) is slow (~10ms per search * 10k -> 100s).
	// With the bug, updates were doing linear scans depending on implementation fallback,
	// or worst case skip list search.

	if opsPerSec < 5000 {
		t.Errorf("Performance is too low! Expected >5000 updates/sec, got %.2f. The O(N) bug might still be present.", opsPerSec)
	} else {
		fmt.Println("✅ Performance check PASSED (O(log N) confirmed).")
	}
}
