package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"leaderboard-backend/config"
	"leaderboard-backend/handlers"
	"leaderboard-backend/models"
	"leaderboard-backend/services"
	"leaderboard-backend/store"

	"github.com/gorilla/mux"
)

// setupTestServer creates a test server with all handlers
func setupTestServer() (*mux.Router, *store.MemoryStore, *store.RatingBucketIndex, *services.ScoreSimulator) {
	cfg := config.Load()

	ratingIndex := store.NewRatingBucketIndex()
	memoryStore := store.NewMemoryStore(ratingIndex)

	userService := services.NewUserService(memoryStore, ratingIndex, cfg.MinRating, cfg.MaxRating)
	leaderboardService := services.NewLeaderboardService(memoryStore, ratingIndex)
	simulator := services.NewScoreSimulator(memoryStore, ratingIndex, cfg.MinRating, cfg.MaxRating, cfg.UpdateInterval)

	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService)
	userHandler := handlers.NewUserHandler(userService, leaderboardService, simulator, cfg.InitialUsers, ratingIndex, memoryStore)

	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	api.HandleFunc("/leaderboard", leaderboardHandler.GetLeaderboard).Methods("GET")
	api.HandleFunc("/search", leaderboardHandler.SearchUsers).Methods("GET")
	api.HandleFunc("/seed", userHandler.SeedUsers).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	api.HandleFunc("/users/{id}/rating", userHandler.UpdateRating).Methods("PATCH")
	api.HandleFunc("/health", userHandler.Health).Methods("GET")
	api.HandleFunc("/simulator/start", userHandler.StartSimulator).Methods("POST")
	api.HandleFunc("/simulator/stop", userHandler.StopSimulator).Methods("POST")
	api.HandleFunc("/simulator/status", userHandler.SimulatorStatus).Methods("GET")

	return router, memoryStore, ratingIndex, simulator
}

func TestAPI_Health(t *testing.T) {
	router, _, _, _ := setupTestServer()

	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health endpoint returned wrong status: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}

func TestAPI_Seed(t *testing.T) {
	router, _, _, simulator := setupTestServer()
	defer simulator.Stop()

	req, err := http.NewRequest("POST", "/api/seed?count=100", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Seed endpoint returned wrong status: got %v want %v", status, http.StatusOK)
	}

	var response models.SeedResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.UsersAdded != 100 {
		t.Errorf("Expected 100 users added, got %d", response.UsersAdded)
	}
}

func TestAPI_Leaderboard(t *testing.T) {
	router, memoryStore, ratingIndex, _ := setupTestServer()

	// Add test users
	for i := 0; i < 50; i++ {
		user := &models.User{
			ID:       "user-" + string(rune('a'+i%26)) + string(rune('0'+i%10)),
			Username: "testuser" + string(rune('0'+i%10)),
			Rating:   4000 - i*50,
		}
		memoryStore.AddUser(user)
	}

	req, err := http.NewRequest("GET", "/api/leaderboard?limit=10&offset=0", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Leaderboard endpoint returned wrong status: got %v want %v", status, http.StatusOK)
	}

	var response models.LeaderboardResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Users) != 10 {
		t.Errorf("Expected 10 users, got %d", len(response.Users))
	}

	if response.TotalUsers != 50 {
		t.Errorf("Expected 50 total users, got %d", response.TotalUsers)
	}

	// Verify rank correctness
	for i, user := range response.Users {
		expectedRank := ratingIndex.GetRank(user.Rating)
		if user.Rank != expectedRank {
			t.Errorf("User %d: expected rank %d, got %d", i, expectedRank, user.Rank)
		}
	}

	// Verify sorted order
	for i := 1; i < len(response.Users); i++ {
		if response.Users[i].Rating > response.Users[i-1].Rating {
			t.Errorf("Users not sorted: rating %d > %d at positions %d, %d",
				response.Users[i].Rating, response.Users[i-1].Rating, i, i-1)
		}
	}
}

func TestAPI_Search(t *testing.T) {
	router, memoryStore, _, _ := setupTestServer()

	// Add test users
	testUsers := []models.User{
		{ID: "id1", Username: "rahul_kumar", Rating: 4500},
		{ID: "id2", Username: "rahul_sharma", Rating: 4200},
		{ID: "id3", Username: "priya_singh", Rating: 4000},
		{ID: "id4", Username: "rahul_gupta", Rating: 3800},
	}

	for _, u := range testUsers {
		user := u
		memoryStore.AddUser(&user)
	}

	req, err := http.NewRequest("GET", "/api/search?q=rahul", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Search endpoint returned wrong status: got %v want %v", status, http.StatusOK)
	}

	var response models.SearchResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Count != 3 {
		t.Errorf("Expected 3 results for 'rahul', got %d", response.Count)
	}

	// Verify all results contain "rahul"
	for _, user := range response.Users {
		if user.Username[:5] != "rahul" {
			t.Errorf("Search result doesn't match query: %s", user.Username)
		}
	}

	// Verify sorted by rating descending
	for i := 1; i < len(response.Users); i++ {
		if response.Users[i].Rating > response.Users[i-1].Rating {
			t.Errorf("Results not sorted by rating")
		}
	}
}

func TestAPI_SearchEmpty(t *testing.T) {
	router, _, _, _ := setupTestServer()

	req, err := http.NewRequest("GET", "/api/search?q=", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Empty search returned wrong status: got %v want %v", status, http.StatusOK)
	}

	var response models.SearchResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Count != 0 {
		t.Errorf("Expected 0 results for empty search, got %d", response.Count)
	}
}

func TestAPI_UpdateRating(t *testing.T) {
	router, memoryStore, ratingIndex, _ := setupTestServer()

	// Add test user
	user := &models.User{
		ID:       "update-test-user",
		Username: "testuser",
		Rating:   2500,
	}
	memoryStore.AddUser(user)

	// Update rating
	body := bytes.NewBuffer([]byte(`{"rating": 3500}`))
	req, err := http.NewRequest("PATCH", "/api/users/update-test-user/rating", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("UpdateRating returned wrong status: got %v want %v, body: %s", status, http.StatusOK, rr.Body.String())
	}

	var response models.UserWithRank
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Rating != 3500 {
		t.Errorf("Expected rating 3500, got %d", response.Rating)
	}

	// Verify rank is correct
	expectedRank := ratingIndex.GetRank(3500)
	if response.Rank != expectedRank {
		t.Errorf("Expected rank %d, got %d", expectedRank, response.Rank)
	}
}

func TestAPI_GetUser(t *testing.T) {
	router, memoryStore, _, _ := setupTestServer()

	// Add test user
	user := &models.User{
		ID:       "get-test-user",
		Username: "testgetuser",
		Rating:   3000,
	}
	memoryStore.AddUser(user)

	req, err := http.NewRequest("GET", "/api/users/get-test-user", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetUser returned wrong status: got %v want %v", status, http.StatusOK)
	}

	var response models.UserWithRank
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Username != "testgetuser" {
		t.Errorf("Expected username 'testgetuser', got %s", response.Username)
	}

	if response.Rating != 3000 {
		t.Errorf("Expected rating 3000, got %d", response.Rating)
	}
}

func TestAPI_GetUserNotFound(t *testing.T) {
	router, _, _, _ := setupTestServer()

	req, err := http.NewRequest("GET", "/api/users/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("GetUser for nonexistent returned wrong status: got %v want %v", status, http.StatusNotFound)
	}
}

func TestAPI_SimulatorStartStop(t *testing.T) {
	router, memoryStore, _, simulator := setupTestServer()
	defer simulator.Stop()

	// Add some users first
	for i := 0; i < 10; i++ {
		user := &models.User{
			ID:       "sim-user-" + string(rune('a'+i)),
			Username: "simuser" + string(rune('0'+i)),
			Rating:   2500,
		}
		memoryStore.AddUser(user)
	}

	// Start simulator
	req, err := http.NewRequest("POST", "/api/simulator/start", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Start simulator returned wrong status: got %v want %v", status, http.StatusOK)
	}

	// Check status
	req, _ = http.NewRequest("GET", "/api/simulator/status", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var status map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&status)

	if status["running"] != true {
		t.Errorf("Simulator should be running")
	}

	// Stop simulator
	req, _ = http.NewRequest("POST", "/api/simulator/stop", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Stop simulator returned wrong status: got %v want %v", rr.Code, http.StatusOK)
	}
}

func TestAPI_LeaderboardPagination(t *testing.T) {
	router, memoryStore, _, _ := setupTestServer()

	// Add 100 users
	for i := 0; i < 100; i++ {
		user := &models.User{
			ID:       "page-user-" + string(rune('a'+i%26)) + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)),
			Username: "pageuser" + string(rune('0'+i%10)),
			Rating:   5000 - i*10,
		}
		memoryStore.AddUser(user)
	}

	// Get first page
	req, _ := http.NewRequest("GET", "/api/leaderboard?limit=20&offset=0", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var page1 models.LeaderboardResponse
	json.NewDecoder(rr.Body).Decode(&page1)

	if len(page1.Users) != 20 {
		t.Errorf("Page 1: expected 20 users, got %d", len(page1.Users))
	}

	if !page1.HasMore {
		t.Error("Page 1 should have more")
	}

	// Get second page
	req, _ = http.NewRequest("GET", "/api/leaderboard?limit=20&offset=20", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var page2 models.LeaderboardResponse
	json.NewDecoder(rr.Body).Decode(&page2)

	if len(page2.Users) != 20 {
		t.Errorf("Page 2: expected 20 users, got %d", len(page2.Users))
	}

	// Verify no overlap between pages
	page1IDs := make(map[string]bool)
	for _, u := range page1.Users {
		page1IDs[u.ID] = true
	}

	for _, u := range page2.Users {
		if page1IDs[u.ID] {
			t.Errorf("User %s appears in both pages", u.ID)
		}
	}

	// Verify page continuity (last of page1 >= first of page2)
	if page1.Users[19].Rating < page2.Users[0].Rating {
		t.Error("Page 2 first user has higher rating than page 1 last user")
	}
}

func TestAPI_RankTies(t *testing.T) {
	router, memoryStore, _, _ := setupTestServer()

	// Add users with tied ratings
	for i := 0; i < 5; i++ {
		user := &models.User{
			ID:       "tie-user-" + string(rune('a'+i)),
			Username: "tieuser" + string(rune('a'+i)),
			Rating:   4000, // All same rating
		}
		memoryStore.AddUser(user)
	}

	req, _ := http.NewRequest("GET", "/api/leaderboard?limit=10&offset=0", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var response models.LeaderboardResponse
	json.NewDecoder(rr.Body).Decode(&response)

	// All users should have rank 1 (competition ranking)
	for _, user := range response.Users {
		if user.Rank != 1 {
			t.Errorf("User %s with rating %d should have rank 1, got %d",
				user.Username, user.Rating, user.Rank)
		}
	}
}
