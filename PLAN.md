# Web Portfolio API — Implementation Plan

## Overview

A Go HTTP API for the portfolio backend. The first endpoint tracks page views.

**Stack:**
- **Router:** `chi` — a small library that maps URLs to handler functions. Go's built-in router works but is awkward; chi makes routes and middleware readable.
- **CORS:** `github.com/go-chi/cors` — when your Next.js site (domain A) fetches from this API (domain B), browsers block it by default. CORS headers tell the browser "requests from my portfolio domain are allowed." Without this, your frontend fetch calls will fail with a CORS error.
- **Storage:** SQLite via `modernc.org/sqlite` — a single file on disk, no separate database server needed. Survives restarts, handles a personal portfolio's traffic easily.
- **Auth:** Secret string in the URL query param, stored in `.env`

---

## Auth: Secret in the URL

Instead of a custom header, the increment endpoint checks for a `?secret=` query parameter.

```
POST /views?secret=yoursecretstring   ← increments the count
GET  /views                           ← returns the count (public)
```

The secret lives in `.env` (gitignored) and is pulled in via `os.Getenv`. You include it in the URL when calling from your Next.js server — the browser never sees the Go API URL at all (see Next.js section below).

---

## API Endpoints

| Method | Path                         | Auth   | Description                         |
| ------ | ---------------------------- | ------ | ----------------------------------- |
| GET    | /views                       | None   | Returns current view count as JSON  |
| POST   | /views?secret=yoursecret     | Secret | Increments count, returns new count |

**Responses:**

```json
// GET /views  or  POST /views?secret=correct
{ "count": 42 }

// POST /views  (missing or wrong secret)
// HTTP 401
{ "error": "unauthorized" }
```

---

## File Structure

```
web-portfolio-api/
├── go.mod
├── go.sum
├── main.go                    # Entry point: config, router, server start
├── internal/
│   ├── db/
│   │   └── db.go              # SQLite connection, schema, GetCount, Increment
│   └── handler/
│       └── views.go           # HTTP handlers: GetViews, IncrementViews
├── .env.example               # Documents required env vars (commit this)
└── .env                       # Actual secrets (gitignore this)
```

No separate middleware file — the secret check is a few lines directly in the handler.

---

## Step 1 — Install Dependencies

```bash
go get github.com/go-chi/chi/v5
go get github.com/go-chi/cors
go get modernc.org/sqlite
```

---

## Step 2 — `internal/db/db.go`

```go
package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS page_views (
			id    INTEGER PRIMARY KEY CHECK (id = 1),
			count INTEGER NOT NULL DEFAULT 0
		);
		INSERT OR IGNORE INTO page_views (id, count) VALUES (1, 0);
	`)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) GetCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(ctx, `SELECT count FROM page_views WHERE id = 1`).Scan(&count)
	return count, err
}

func (s *Store) Increment(ctx context.Context) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE page_views SET count = count + 1 WHERE id = 1`)
	if err != nil {
		return 0, err
	}

	var count int64
	err = tx.QueryRowContext(ctx, `SELECT count FROM page_views WHERE id = 1`).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, tx.Commit()
}
```

---

## Step 3 — `internal/handler/views.go`

The secret check happens here — reads `?secret=` from the URL and compares it to the value loaded from `.env`.

```go
package handler

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"web-portfolio-api/internal/db"
)

type ViewsHandler struct {
	store  *db.Store
	secret string // loaded from API_SECRET env var in main.go
}

func NewViewsHandler(store *db.Store, secret string) *ViewsHandler {
	return &ViewsHandler{store: store, secret: secret}
}

type viewsResponse struct {
	Count int64 `json:"count"`
}

func (h *ViewsHandler) GetViews(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.GetCount(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(viewsResponse{Count: count})
}

func (h *ViewsHandler) IncrementViews(w http.ResponseWriter, r *http.Request) {
	provided := r.URL.Query().Get("secret")
	// subtle.ConstantTimeCompare prevents timing attacks (comparing char by char would
	// leak info about how many characters match — this compares all at once)
	if subtle.ConstantTimeCompare([]byte(provided), []byte(h.secret)) != 1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	count, err := h.store.Increment(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(viewsResponse{Count: count})
}
```

---

## Step 4 — `main.go`

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"web-portfolio-api/internal/db"
	"web-portfolio-api/internal/handler"
)

func main() {
	secret := os.Getenv("API_SECRET")
	if secret == "" {
		log.Fatal("API_SECRET environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "views.db"
	}

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*" // allow all origins in local dev; set your domain in prod
	}

	store, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	viewsHandler := handler.NewViewsHandler(store, secret)

	r := chi.NewRouter()

	// CORS: tells browsers "requests from my portfolio domain are allowed"
	// Without this, your Next.js fetch calls will be blocked by the browser
	r.Use(cors.New(cors.Options{
		AllowedOrigins: []string{allowedOrigin},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         300,
	}).Handler)

	r.Get("/views", viewsHandler.GetViews)
	r.Post("/views", viewsHandler.IncrementViews)

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
```

---

## Step 5 — `.env.example`

```
API_SECRET=your-secret-string-here
PORT=8080
DB_PATH=views.db
ALLOWED_ORIGIN=https://yourportfolio.com
```

`.gitignore` — add these two lines:

```
.env
views.db
```

---

## Next.js Integration

Call the increment from a **server-side** Next.js API Route so the secret never appears in the browser:

```ts
// app/api/track-view/route.ts
export async function POST() {
  const secret = process.env.PORTFOLIO_API_SECRET;
  const apiUrl = process.env.PORTFOLIO_API_URL; // e.g. http://localhost:8080

  await fetch(`${apiUrl}/views?secret=${secret}`, {
    method: 'POST',
    cache: 'no-store',
  });

  return new Response(null, { status: 204 });
}
```

```ts
// Call this from a layout or page (Server Component)
// app/layout.tsx
async function trackView() {
  await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/api/track-view`, {
    method: 'POST',
    cache: 'no-store',
  });
}
```

```ts
// Display the count — just fetch GET /views directly (it's public)
const res = await fetch(`${process.env.PORTFOLIO_API_URL}/views`, { cache: 'no-store' });
const { count } = await res.json();
```

---

## Testing Locally

```bash
# 1. Install deps
go get github.com/go-chi/chi/v5 github.com/go-chi/cors modernc.org/sqlite

# 2. Set env and run
export API_SECRET=test123
go run ./main.go

# 3. Read the count (public)
curl http://localhost:8080/views
# → {"count":0}

# 4. Increment without secret — should fail
curl -X POST http://localhost:8080/views
# → 401 {"error":"unauthorized"}

# 5. Increment with correct secret
curl -X POST "http://localhost:8080/views?secret=test123"
# → {"count":1}

# 6. Stop server, restart, GET /views → {"count":1}  (confirms SQLite persisted it)
```

---

## Deployment Notes

- SQLite needs a **persistent disk**. Good options: Fly.io (volumes), Railway (volume mounts), or a plain VPS. Avoid platforms with ephemeral filesystems — the `views.db` file will be wiped on restart.
- Set `ALLOWED_ORIGIN` to your exact portfolio domain in production (e.g. `https://yourname.com`).
- `API_SECRET` should be set as a secret in your deployment environment, not in the repo.
