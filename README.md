# Scalable Leaderboard System

A real-time, tie-aware, scalable leaderboard system built with Golang backend and Web frontend.

## 🚀 Features

- **10,000+ Users**: Handles large-scale user data with capacity for millions
- **O(1) Rank Computation**: Uses rating bucket index for instant rank lookups
- **O(limit) Leaderboard**: Pre-sorted user list for efficient pagination
- **Competition Ranking**: Proper tie handling (same rating = same rank)
- **Real-Time Updates**: Background score simulator with batch updates (10 users/tick)
- **Instant Search**: Fast username search with live global rank (max 100 results)
- **Rate Limiting**: Token bucket rate limiter (100 req/sec, burst 200)
- **Request Logging**: Structured request logging with timing
- **Comprehensive Health**: Detailed stats endpoint with memory usage

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend (Web/Expo)                      │
│  ┌─────────────────┐  ┌─────────────────┐                   │
│  │ Leaderboard     │  │ Search View     │                   │
│  │ (Virtualized)   │  │ (Debounced)     │                   │
│  └────────┬────────┘  └────────┬────────┘                   │
└───────────┼────────────────────┼────────────────────────────┘
            │                    │
            ▼                    ▼
┌─────────────────────────────────────────────────────────────┐
│                    Middleware Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ CORS        │  │ Rate Limit  │  │ Logging     │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└───────────────────────────┬─────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend (Golang)                         │
│  ┌─────────────────┐  ┌─────────────────┐                   │
│  │ Leaderboard API │  │ Search API      │                   │
│  └────────┬────────┘  └────────┬────────┘                   │
│           │                    │                            │
│           ▼                    ▼                            │
│  ┌─────────────────────────────────────┐                    │
│  │         Rating Bucket Index         │ ◄── O(1) Ranking   │
│  │         (4901 buckets)              │                    │
│  └─────────────────────────────────────┘                    │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐  ┌─────────────────┐                   │
│  │ Memory Store    │  │ Score Simulator │                   │
│  │ (Sorted List)   │  │ (Batch Updates) │                   │
│  └─────────────────┘  └─────────────────┘                   │
└─────────────────────────────────────────────────────────────┘
```

## 📊 Ranking Algorithm

**Competition Ranking Formula**:
```
rank = 1 + count of users with strictly higher rating
```

**Example**:
| Rating | Rank |
|--------|------|
| 5000   | 1    |
| 4900   | 2    |
| 4900   | 2    |
| 4800   | 4    |

## 🛠️ Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+ (for frontend/web)

### 1. Start Backend

```bash
cd backend
go mod tidy
go run main.go
```

Server starts at `http://localhost:8080`

### 2. Seed Data

Open browser and click "Seed 10k Users" button, or:

```bash
curl -X POST http://localhost:8080/api/seed
```

### 3. Start Web Frontend

```bash
cd web
npx serve -l 3000
```

Open `http://localhost:3000` in your browser.

### 4. Start Expo Frontend (Optional)

```bash
cd frontend
npm install
npx expo start
```

## 📚 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/leaderboard?limit=50&offset=0` | Get paginated leaderboard |
| GET | `/api/search?q=rahul` | Search users by username |
| GET | `/api/users/{id}` | Get user with rank |
| POST | `/api/seed?count=10000` | Seed initial users |
| PATCH | `/api/users/{id}/rating` | Update user rating |
| GET | `/api/health` | Health check with detailed stats |
| POST | `/api/simulator/start` | Start score simulator |
| POST | `/api/simulator/stop` | Stop score simulator |
| GET | `/api/simulator/status` | Get simulator status |

## 🧪 Testing

```bash
cd backend
go test ./tests/... -v
```

### Test Coverage:
- Basic ranking (single users)
- Tied ranking (multiple users, same rating)
- Boundary ratings (100, 5000)
- Large scale (10,000+ users)
- Competition ranking verification
- Concurrent read/write operations
- Edge cases (thousands with same rating)
- Search with special characters
- Stress testing for GetTopUsers

## 📈 Performance

| Operation | Complexity | Notes |
|-----------|------------|-------|
| Get user rank | O(1) | Precomputed cumulative array |
| Get top N users | O(N) | Pre-sorted list slice |
| Search users | O(M log M) | Limited to 100 results |
| Update rating | O(Δ) | Incremental cumulative update |
| Add user | O(log N) | Binary search insertion |

## 🛡️ Production Features

- **Rate Limiting**: 100 requests/second per IP, burst of 200
- **Request Logging**: Structured logs with timing
- **Health Monitoring**: Memory usage, rating index stats, simulator stats
- **Request Timeouts**: 10-second timeout on frontend API calls
- **Environment Variables**: Configurable API URL via `EXPO_PUBLIC_API_URL`
- **Input Validation**: Search query sanitization
- **Result Limits**: Max 100 search results to prevent memory issues

## 📁 Project Structure

```
Matiks_Assignment/
├── backend/           # Golang backend
│   ├── main.go
│   ├── config/
│   ├── middleware/    # Rate limiting & logging
│   ├── models/
│   ├── store/
│   │   ├── rating_index.go   # O(1) ranking engine
│   │   └── memory_store.go   # Sorted user list
│   ├── services/
│   ├── handlers/
│   └── tests/
│       ├── ranking_test.go
│       ├── concurrency_test.go
│       └── edge_cases_test.go
├── web/               # Web frontend
│   └── index.html
├── frontend/          # React Native/Expo
└── README.md
```

## 🔧 Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Backend server port |
| `INITIAL_USERS` | 10000 | Default seed count |
| `UPDATE_INTERVAL` | 100 | Simulator tick (ms) |
| `EXPO_PUBLIC_API_URL` | localhost:8080/api | Frontend API URL |

## 📄 License

MIT
