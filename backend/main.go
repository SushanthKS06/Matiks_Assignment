package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"leaderboard-backend/config"
	"leaderboard-backend/handlers"
	"leaderboard-backend/middleware"
	"leaderboard-backend/services"
	"leaderboard-backend/store"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const persistenceFile = "data/leaderboard.json"

func main() {
	cfg := config.Load()

	ratingIndex := store.NewRatingBucketIndex()
	memoryStore := store.NewMemoryStore(ratingIndex)
	persistence := store.NewPersistence(persistenceFile)

	// Load existing data if available
	if persistence.Exists() {
		fmt.Println("Loading existing data from disk...")
		if err := persistence.Load(memoryStore, ratingIndex); err != nil {
			log.Printf("Warning: failed to load data: %v\n", err)
		} else {
			fmt.Printf("Loaded %d users from disk\n", memoryStore.GetUserCount())
		}
	}

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

	// Initialize middleware
	rateLimiter := middleware.NewRateLimiter(100, 200) // 100 req/sec, burst of 200
	rateLimiter.CleanupOldVisitors(time.Minute * 10)

	logger := middleware.NewLogger()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "ngrok-skip-browser-warning"},
		AllowCredentials: true,
	})

	// Chain middleware: CORS -> RateLimiter -> Logger -> Router
	handler := c.Handler(rateLimiter.Limit(logger.LogRequest(router)))

	// Create server with proper shutdown handling
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Println("\nShutting down server...")

		// Stop simulator
		simulator.Stop()

		// Save data to disk
		fmt.Println("Saving data to disk...")
		if err := persistence.Save(memoryStore); err != nil {
			log.Printf("Warning: failed to save data: %v\n", err)
		} else {
			fmt.Printf("Saved %d users to disk\n", memoryStore.GetUserCount())
		}

		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}

		close(done)
	}()

	fmt.Printf("Leaderboard Server starting on port %s\n", cfg.Port)
	fmt.Printf("Rating range: %d - %d\n", cfg.MinRating, cfg.MaxRating)
	fmt.Printf("Initial users: %d\n", cfg.InitialUsers)
	fmt.Printf("Update interval: %dms\n", cfg.UpdateInterval)
	fmt.Printf("Rate limiting: 100 req/sec, burst 200\n")
	fmt.Printf("Persistence: %s\n", persistenceFile)
	fmt.Println("\nAPI Endpoints:")
	fmt.Println("  GET  /api/leaderboard     - Get paginated leaderboard")
	fmt.Println("  GET  /api/search?q=query  - Search users by username")
	fmt.Println("  POST /api/seed            - Seed initial users")
	fmt.Println("  GET  /api/users/{id}      - Get user by ID")
	fmt.Println("  PATCH /api/users/{id}/rating - Update user rating")
	fmt.Println("  GET  /api/health          - Health check with stats")
	fmt.Println("  POST /api/simulator/start - Start score simulator")
	fmt.Println("  POST /api/simulator/stop  - Stop score simulator")
	fmt.Println("  GET  /api/simulator/status - Get simulator status")
	fmt.Println("\nPress Ctrl+C to save and exit gracefully")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	<-done
	fmt.Println("Server stopped gracefully")
}
