package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"leaderboard-backend/models"
	"leaderboard-backend/services"
	"leaderboard-backend/store"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	userService        *services.UserService
	leaderboardService *services.LeaderboardService
	simulator          *services.ScoreSimulator
	initialUsers       int
	ratingIndex        *store.RatingBucketIndex
	memoryStore        *store.MemoryStore
}

func NewUserHandler(
	userService *services.UserService,
	leaderboardService *services.LeaderboardService,
	simulator *services.ScoreSimulator,
	initialUsers int,
	ratingIndex *store.RatingBucketIndex,
	memoryStore *store.MemoryStore,
) *UserHandler {
	return &UserHandler{
		userService:        userService,
		leaderboardService: leaderboardService,
		simulator:          simulator,
		initialUsers:       initialUsers,
		ratingIndex:        ratingIndex,
		memoryStore:        memoryStore,
	}
}

func (h *UserHandler) SeedUsers(w http.ResponseWriter, r *http.Request) {
	countStr := r.URL.Query().Get("count")
	count := h.initialUsers

	if countStr != "" {
		if parsed, err := strconv.Atoi(countStr); err == nil && parsed > 0 && parsed <= 100000 {
			count = parsed
		}
	}

	h.userService.Clear()

	added, err := h.userService.SeedUsers(count)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "seed_failed",
			Message: err.Error(),
		})
		return
	}

	h.simulator.Start()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SeedResponse{
		Message:    "Successfully seeded users",
		UsersAdded: added,
	})
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	userWithRank, err := h.leaderboardService.GetUserWithRank(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userWithRank)
}

func (h *UserHandler) UpdateRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid JSON body",
		})
		return
	}

	if err := h.userService.UpdateRating(id, req.Rating); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	userWithRank, err := h.leaderboardService.GetUserWithRank(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userWithRank)
}

// Health returns comprehensive health check with system stats
func (h *UserHandler) Health(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ratingStats := h.ratingIndex.GetStats()
	storeStats := h.memoryStore.GetStats()
	simulatorStats := h.simulator.GetStats()

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"users": map[string]interface{}{
			"total": h.userService.GetUserCount(),
		},
		"rating_index": ratingStats,
		"memory_store": storeStats,
		"simulator":    simulatorStats,
		"memory": map[string]interface{}{
			"alloc_mb":       m.Alloc / 1024 / 1024,
			"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
			"sys_mb":         m.Sys / 1024 / 1024,
			"num_gc":         m.NumGC,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) StartSimulator(w http.ResponseWriter, r *http.Request) {
	h.simulator.Start()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Simulator started",
		"running": h.simulator.IsRunning(),
	})
}

func (h *UserHandler) StopSimulator(w http.ResponseWriter, r *http.Request) {
	h.simulator.Stop()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Simulator stopped",
		"running": h.simulator.IsRunning(),
	})
}

func (h *UserHandler) SimulatorStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.simulator.GetStats())
}
