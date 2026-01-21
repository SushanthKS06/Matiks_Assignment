package services

import (
	"leaderboard-backend/store"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type ScoreSimulator struct {
	store       *store.MemoryStore
	ratingIndex *store.RatingBucketIndex
	minRating   int
	maxRating   int
	interval    time.Duration
	running     int32 // atomic for lock-free check
	mu          sync.Mutex
	stopChan    chan struct{}
	updateCount int64
	batchSize   int

	// Cached user IDs to avoid allocations every tick
	cachedIDs    []string
	cacheVersion int64
}

func NewScoreSimulator(s *store.MemoryStore, ri *store.RatingBucketIndex, minRating, maxRating int, intervalMs int) *ScoreSimulator {
	return &ScoreSimulator{
		store:       s,
		ratingIndex: ri,
		minRating:   minRating,
		maxRating:   maxRating,
		interval:    time.Duration(intervalMs) * time.Millisecond,
		stopChan:    make(chan struct{}),
		batchSize:   10, // Update 10 users per tick for more realistic simulation
		cachedIDs:   make([]string, 0),
	}
}

func (s *ScoreSimulator) Start() {
	s.mu.Lock()
	if atomic.LoadInt32(&s.running) == 1 {
		s.mu.Unlock()
		return
	}
	atomic.StoreInt32(&s.running, 1)
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	go s.run()
}

func (s *ScoreSimulator) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if atomic.LoadInt32(&s.running) == 0 {
		return
	}
	atomic.StoreInt32(&s.running, 0)
	close(s.stopChan)
}

func (s *ScoreSimulator) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

func (s *ScoreSimulator) GetUpdateCount() int64 {
	return atomic.LoadInt64(&s.updateCount)
}

func (s *ScoreSimulator) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Refresh cache every 10 seconds
	cacheTicker := time.NewTicker(10 * time.Second)
	defer cacheTicker.Stop()

	// Initial cache
	s.refreshCache()

	for {
		select {
		case <-s.stopChan:
			return
		case <-cacheTicker.C:
			s.refreshCache()
		case <-ticker.C:
			s.updateRandomUsers()
		}
	}
}

// refreshCache updates the cached user IDs
func (s *ScoreSimulator) refreshCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cachedIDs = s.store.GetAllUserIDs()
	s.cacheVersion++
}

// updateRandomUsers updates multiple random users per tick
// Optimized: uses cached IDs, prepares data before locking
func (s *ScoreSimulator) updateRandomUsers() {
	s.mu.Lock()
	ids := s.cachedIDs
	s.mu.Unlock()

	if len(ids) == 0 {
		return
	}

	// Prepare random selections without holding any locks
	batchCount := s.batchSize
	if batchCount > len(ids) {
		batchCount = len(ids)
	}

	selectedIndices := make([]int, batchCount)
	for i := 0; i < batchCount; i++ {
		selectedIndices[i] = rand.Intn(len(ids))
	}

	// Process updates one at a time with minimal lock time
	for _, idx := range selectedIndices {
		randomID := ids[idx]

		user, err := s.store.GetUser(randomID)
		if err != nil {
			continue
		}

		delta := rand.Intn(201) - 100
		newRating := user.Rating + delta

		if newRating < s.minRating {
			newRating = s.minRating
		}
		if newRating > s.maxRating {
			newRating = s.maxRating
		}

		s.store.UpdateRating(randomID, newRating)
		atomic.AddInt64(&s.updateCount, 1)
	}
}

// GetStats returns simulator statistics
func (s *ScoreSimulator) GetStats() map[string]interface{} {
	s.mu.Lock()
	cacheSize := len(s.cachedIDs)
	cacheVer := s.cacheVersion
	s.mu.Unlock()

	return map[string]interface{}{
		"running":       s.IsRunning(),
		"update_count":  atomic.LoadInt64(&s.updateCount),
		"batch_size":    s.batchSize,
		"interval_ms":   s.interval.Milliseconds(),
		"cache_size":    cacheSize,
		"cache_version": cacheVer,
	}
}
