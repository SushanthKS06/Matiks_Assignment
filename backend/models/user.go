package models

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

type UserWithRank struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Rank     int    `json:"rank"`
}

type LeaderboardResponse struct {
	Users      []UserWithRank `json:"users"`
	TotalUsers int            `json:"total_users"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	HasMore    bool           `json:"has_more"`
}

type SearchResponse struct {
	Users []UserWithRank `json:"users"`
	Query string         `json:"query"`
	Count int            `json:"count"`
}

type UpdateRatingRequest struct {
	Rating int `json:"rating"`
}

type SeedResponse struct {
	Message    string `json:"message"`
	UsersAdded int    `json:"users_added"`
}

type HealthResponse struct {
	Status     string `json:"status"`
	TotalUsers int    `json:"total_users"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
