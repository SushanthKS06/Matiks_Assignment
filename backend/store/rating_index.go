package store

import (
	"sync"
	"sync/atomic"
)

const (
	MinRating   = 100
	MaxRating   = 5000
	RatingRange = MaxRating - MinRating + 1 // 4901 buckets
)

type RatingBucketIndex struct {
	mu         sync.RWMutex
	buckets    [RatingRange]int32 // Count of users at each rating
	cumulative [RatingRange]int32 // Precomputed: users with rating > current
	totalUsers int32
}

func NewRatingBucketIndex() *RatingBucketIndex {
	return &RatingBucketIndex{}
}

func ratingToIndex(rating int) int {
	if rating < MinRating {
		return 0
	}
	if rating > MaxRating {
		return RatingRange - 1
	}
	return rating - MinRating
}

// recalculateCumulative performs full O(4901) recalculation
// Used only when adding/removing users (not for rating updates)
func (r *RatingBucketIndex) recalculateCumulative() {
	var cumSum int32 = 0
	for i := RatingRange - 1; i >= 0; i-- {
		r.cumulative[i] = cumSum
		cumSum += r.buckets[i]
	}
}

// GetRank returns the competition rank for a given rating
// O(1) lookup using precomputed cumulative array
func (r *RatingBucketIndex) GetRank(rating int) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	idx := ratingToIndex(rating)
	return int(r.cumulative[idx]) + 1
}

// IncrementBucket adds a user at the given rating
// O(4901) - only called when adding new users
func (r *RatingBucketIndex) IncrementBucket(rating int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := ratingToIndex(rating)
	r.buckets[idx]++
	atomic.AddInt32(&r.totalUsers, 1)
	r.recalculateCumulative()
}

// DecrementBucket removes a user at the given rating
// O(4901) - only called when removing users
func (r *RatingBucketIndex) DecrementBucket(rating int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := ratingToIndex(rating)
	if r.buckets[idx] > 0 {
		r.buckets[idx]--
		atomic.AddInt32(&r.totalUsers, -1)
	}
	r.recalculateCumulative()
}

// UpdateRating moves a user from oldRating to newRating
// O(|newRating - oldRating|) using incremental update instead of O(4901)
func (r *RatingBucketIndex) UpdateRating(oldRating, newRating int) {
	if oldRating == newRating {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldIdx := ratingToIndex(oldRating)
	newIdx := ratingToIndex(newRating)

	if r.buckets[oldIdx] > 0 {
		r.buckets[oldIdx]--
	}
	r.buckets[newIdx]++

	// Incremental cumulative update - O(|newIdx - oldIdx|) instead of O(4901)
	// cumulative[i] = count of users with rating > (i + MinRating)
	if oldIdx < newIdx {
		// User moved to higher rating
		// Ratings in range [oldIdx, newIdx) now have one MORE user above them
		for i := oldIdx; i < newIdx; i++ {
			r.cumulative[i]++
		}
	} else {
		// User moved to lower rating
		// Ratings in range [newIdx, oldIdx) now have one LESS user above them
		for i := newIdx; i < oldIdx; i++ {
			r.cumulative[i]--
		}
	}
}

// GetUsersAbove returns count of users with rating strictly higher than given
func (r *RatingBucketIndex) GetUsersAbove(rating int) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	idx := ratingToIndex(rating)
	return int(r.cumulative[idx])
}

// GetTotalUsers returns total number of users in the index
func (r *RatingBucketIndex) GetTotalUsers() int {
	return int(atomic.LoadInt32(&r.totalUsers))
}

// Clear removes all users from the index
func (r *RatingBucketIndex) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.buckets {
		r.buckets[i] = 0
		r.cumulative[i] = 0
	}
	atomic.StoreInt32(&r.totalUsers, 0)
}

// GetBucketCount returns number of users at a specific rating
func (r *RatingBucketIndex) GetBucketCount(rating int) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	idx := ratingToIndex(rating)
	return int(r.buckets[idx])
}

// GetRatingsDescending returns all ratings that have at least one user, sorted descending
func (r *RatingBucketIndex) GetRatingsDescending() []int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ratings := make([]int, 0)
	for i := RatingRange - 1; i >= 0; i-- {
		if r.buckets[i] > 0 {
			ratings = append(ratings, i+MinRating)
		}
	}
	return ratings
}

// GetStats returns statistics about the rating index
func (r *RatingBucketIndex) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nonEmptyBuckets := 0
	maxBucketSize := int32(0)
	for _, count := range r.buckets {
		if count > 0 {
			nonEmptyBuckets++
			if count > maxBucketSize {
				maxBucketSize = count
			}
		}
	}

	return map[string]interface{}{
		"total_users":       r.totalUsers,
		"total_buckets":     RatingRange,
		"non_empty_buckets": nonEmptyBuckets,
		"max_bucket_size":   maxBucketSize,
		"min_rating":        MinRating,
		"max_rating":        MaxRating,
	}
}
