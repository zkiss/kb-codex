package app

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/zkiss/kb-codex/internal/db"
	"github.com/zkiss/kb-codex/internal/handlers"
)

// Dependencies holds the runtime dependencies required to build the HTTP server.
type Dependencies struct {
	// DatabaseURL is the postgres connection string.
	DatabaseURL string
	// JWTSecret is used for signing auth tokens.
	JWTSecret []byte
	// AIClient provides embedding and completion APIs.
	AIClient handlers.AIClient
}

// New initializes the database, applies migrations and returns the DB connection
// and router ready to be served.
func New(deps Dependencies) (*sql.DB, http.Handler, error) {
	conn, err := db.ConnectAndMigrate(deps.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	authHandler := handlers.NewAuthHandler(conn, deps.JWTSecret)
	kbHandler := handlers.NewKBHandler(conn, deps.AIClient)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Post("/api/register", authHandler.Register)
	r.Post("/api/login", authHandler.Login)
	r.Get("/api/kbs", kbHandler.ListKB)
	r.Post("/api/kbs", kbHandler.CreateKB)
	r.Get("/api/kbs/{kbID}/files", kbHandler.ListFiles)
	r.Post("/api/kbs/{kbID}/files", kbHandler.UploadFile)
	r.Post("/api/kbs/{kbID}/ask", kbHandler.AskQuestion)

	r.Get("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	return conn, r, nil
}
