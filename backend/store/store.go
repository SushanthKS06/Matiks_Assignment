package store

import "leaderboard-backend/models"

type Store interface {
	AddUser(user *models.User) error
	GetUser(id string) (*models.User, error)
	UpdateRating(id string, newRating int) error
	GetAllUsers() []*models.User
	GetUsersByRating(rating int) []*models.User
	GetUserCount() int
	SearchUsers(query string) []*models.User
	GetTopUsers(limit int, offset int) []*models.User
	Clear()
}

type RankingIndex interface {
	GetRank(rating int) int
	IncrementBucket(rating int)
	DecrementBucket(rating int)
	UpdateRating(oldRating, newRating int)
	GetUsersAbove(rating int) int
	GetTotalUsers() int
	Clear()
}
