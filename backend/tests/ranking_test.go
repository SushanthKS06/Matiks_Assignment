package tests

import (
	"testing"

	"leaderboard-backend/store"
)

func TestRatingBucketIndex_BasicRanking(t *testing.T) {
	idx := store.NewRatingBucketIndex()

	idx.IncrementBucket(5000)
	idx.IncrementBucket(4900)
	idx.IncrementBucket(4900)
	idx.IncrementBucket(4800)

	tests := []struct {
		rating       int
		expectedRank int
	}{
		{5000, 1},
		{4900, 2},
		{4800, 4},
	}

	for _, tt := range tests {
		rank := idx.GetRank(tt.rating)
		if rank != tt.expectedRank {
			t.Errorf("Rating %d: expected rank %d, got %d", tt.rating, tt.expectedRank, rank)
		}
	}
}

func TestRatingBucketIndex_TiedRanking(t *testing.T) {
	idx := store.NewRatingBucketIndex()

	for i := 0; i < 5; i++ {
		idx.IncrementBucket(4500)
	}
	for i := 0; i < 3; i++ {
		idx.IncrementBucket(4000)
	}
	for i := 0; i < 2; i++ {
		idx.IncrementBucket(3500)
	}

	if rank := idx.GetRank(4500); rank != 1 {
		t.Errorf("Rating 4500: expected rank 1, got %d", rank)
	}

	if rank := idx.GetRank(4000); rank != 6 {
		t.Errorf("Rating 4000: expected rank 6, got %d", rank)
	}

	if rank := idx.GetRank(3500); rank != 9 {
		t.Errorf("Rating 3500: expected rank 9, got %d", rank)
	}
}

func TestRatingBucketIndex_BoundaryRatings(t *testing.T) {
	idx := store.NewRatingBucketIndex()

	idx.IncrementBucket(100)
	idx.IncrementBucket(5000)
	idx.IncrementBucket(2500)

	if rank := idx.GetRank(5000); rank != 1 {
		t.Errorf("Max rating 5000: expected rank 1, got %d", rank)
	}

	if rank := idx.GetRank(2500); rank != 2 {
		t.Errorf("Mid rating 2500: expected rank 2, got %d", rank)
	}

	if rank := idx.GetRank(100); rank != 3 {
		t.Errorf("Min rating 100: expected rank 3, got %d", rank)
	}
}

func TestRatingBucketIndex_UpdateRating(t *testing.T) {
	idx := store.NewRatingBucketIndex()

	idx.IncrementBucket(4000)
	idx.IncrementBucket(3000)

	if rank := idx.GetRank(4000); rank != 1 {
		t.Errorf("Before update - Rating 4000: expected rank 1, got %d", rank)
	}

	idx.UpdateRating(4000, 2000)

	if rank := idx.GetRank(3000); rank != 1 {
		t.Errorf("After update - Rating 3000: expected rank 1, got %d", rank)
	}

	if rank := idx.GetRank(2000); rank != 2 {
		t.Errorf("After update - Rating 2000: expected rank 2, got %d", rank)
	}
}

func TestRatingBucketIndex_LargeScale(t *testing.T) {
	idx := store.NewRatingBucketIndex()

	for i := 0; i < 10000; i++ {
		rating := 100 + (i % 4901)
		idx.IncrementBucket(rating)
	}

	if rank := idx.GetRank(5000); rank != 1 {
		t.Errorf("Max rating should be rank 1, got %d", rank)
	}

	if total := idx.GetTotalUsers(); total != 10000 {
		t.Errorf("Expected 10000 total users, got %d", total)
	}
}

func TestRatingBucketIndex_CompetitionRankingExample(t *testing.T) {
	idx := store.NewRatingBucketIndex()

	idx.IncrementBucket(4600)
	idx.IncrementBucket(3900)
	idx.IncrementBucket(3900)
	idx.IncrementBucket(1234)

	tests := []struct {
		rating       int
		expectedRank int
	}{
		{4600, 1},
		{3900, 2},
		{1234, 4},
	}

	for _, tt := range tests {
		rank := idx.GetRank(tt.rating)
		if rank != tt.expectedRank {
			t.Errorf("Rating %d: expected rank %d, got %d", tt.rating, tt.expectedRank, rank)
		}
	}
}
