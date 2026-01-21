package tests

import (
	"sync"
	"testing"

	"leaderboard-backend/models"
	"leaderboard-backend/store"
)

func TestConcurrentReads(t *testing.T) {
	idx := store.NewRatingBucketIndex()

	for i := 0; i < 1000; i++ {
		idx.IncrementBucket(100 + (i % 4901))
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(rating int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = idx.GetRank(rating)
			}
		}(100 + (i % 4901))
	}
	wg.Wait()
}

func TestConcurrentReadWrite(t *testing.T) {
	idx := store.NewRatingBucketIndex()

	for i := 0; i < 1000; i++ {
		idx.IncrementBucket(100 + (i % 4901))
	}

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(rating int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = idx.GetRank(rating)
			}
		}(100 + (i % 4901))
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(oldRating, newRating int) {
			defer wg.Done()
			idx.UpdateRating(oldRating, newRating)
		}(100+(i%4901), 100+((i+100)%4901))
	}

	wg.Wait()

	if idx.GetTotalUsers() != 1000 {
		t.Errorf("Total users should remain 1000, got %d", idx.GetTotalUsers())
	}
}

func TestConcurrentMemoryStore(t *testing.T) {
	idx := store.NewRatingBucketIndex()
	ms := store.NewMemoryStore(idx)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			user := &models.User{
				ID:       "user-" + string(rune('a'+i%26)) + "-" + string(rune('0'+i%10)),
				Username: "user" + string(rune('0'+i%10)),
				Rating:   100 + (i * 49),
			}
			_ = ms.AddUser(user)
		}(i)
	}

	wg.Wait()
}
