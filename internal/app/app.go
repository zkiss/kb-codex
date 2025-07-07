package app

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/zkiss/kb-codex/internal/config"
	"github.com/zkiss/kb-codex/internal/db"
	"github.com/zkiss/kb-codex/internal/handlers"
)

// App encapsulates the application state and logic.
type App struct {
	cfg    *config.Config
	db     *sql.DB
	router http.Handler
}

// New initializes the database, applies migrations and returns the App instance ready to be served.
func New(cfg *config.Config, aiClient handlers.AIClient) (*App, error) {
	conn, err := db.ConnectAndMigrate(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	authHandler := handlers.NewAuthHandler(conn, cfg.JWTSecret)
	kbHandler := handlers.NewKBHandler(conn, aiClient)

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

	return &App{cfg: cfg, db: conn, router: r}, nil
}

func (a *App) Close() error {
	return a.db.Close()
}

// Listen starts the HTTP server using the provided listener function.
func (a *App) Listen(listener func(port uint16, router http.Handler) error) error {
	return listener(a.cfg.Port, a.router)
}
