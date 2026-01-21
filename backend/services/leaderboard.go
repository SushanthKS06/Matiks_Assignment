package services

import (
	"leaderboard-backend/models"
	"leaderboard-backend/store"
)

type LeaderboardService struct {
	store       *store.MemoryStore
	ratingIndex *store.RatingBucketIndex
}

func NewLeaderboardService(s *store.MemoryStore, ri *store.RatingBucketIndex) *LeaderboardService {
	return &LeaderboardService{
		store:       s,
		ratingIndex: ri,
	}
}

func (l *LeaderboardService) GetLeaderboard(limit, offset int) *models.LeaderboardResponse {
	users := l.store.GetTopUsers(limit, offset)
	totalUsers := l.store.GetUserCount()

	usersWithRank := make([]models.UserWithRank, 0, len(users))
	for _, user := range users {
		rank := l.ratingIndex.GetRank(user.Rating)
		usersWithRank = append(usersWithRank, models.UserWithRank{
			ID:       user.ID,
			Username: user.Username,
			Rating:   user.Rating,
			Rank:     rank,
		})
	}

	hasMore := offset+limit < totalUsers

	return &models.LeaderboardResponse{
		Users:      usersWithRank,
		TotalUsers: totalUsers,
		Page:       offset/limit + 1,
		PageSize:   limit,
		HasMore:    hasMore,
	}
}

func (l *LeaderboardService) SearchUsers(query string) *models.SearchResponse {
	users := l.store.SearchUsers(query)

	usersWithRank := make([]models.UserWithRank, 0, len(users))
	for _, user := range users {
		rank := l.ratingIndex.GetRank(user.Rating)
		usersWithRank = append(usersWithRank, models.UserWithRank{
			ID:       user.ID,
			Username: user.Username,
			Rating:   user.Rating,
			Rank:     rank,
		})
	}

	return &models.SearchResponse{
		Users: usersWithRank,
		Query: query,
		Count: len(usersWithRank),
	}
}

func (l *LeaderboardService) GetUserWithRank(id string) (*models.UserWithRank, error) {
	user, err := l.store.GetUser(id)
	if err != nil {
		return nil, err
	}

	rank := l.ratingIndex.GetRank(user.Rating)

	return &models.UserWithRank{
		ID:       user.ID,
		Username: user.Username,
		Rating:   user.Rating,
		Rank:     rank,
	}, nil
}
