package main

import (
   "log"
   "net/http"

   "github.com/go-chi/chi/v5"
   "github.com/go-chi/chi/v5/middleware"
   "github.com/zkiss/kb-codex/internal/config"
   "github.com/zkiss/kb-codex/internal/db"
   "github.com/zkiss/kb-codex/internal/handlers"
)

func main() {
   cfg := config.Load()

   dbConn, err := db.ConnectAndMigrate(cfg.DatabaseURL)
   if err != nil {
       log.Fatalf("could not connect to database: %v", err)
   }
   defer dbConn.Close()

   authHandler := handlers.NewAuthHandler(dbConn, []byte(cfg.JWTSecret))

   r := chi.NewRouter()
   r.Use(middleware.Logger)

   r.Post("/api/register", authHandler.Register)
   r.Post("/api/login", authHandler.Login)

   // Serve static files
   fileServer := http.FileServer(http.Dir("static"))
   r.Handle("/*", fileServer)

   log.Printf("Starting server on %s", cfg.Addr)
   if err := http.ListenAndServe(cfg.Addr, r); err != nil {
       log.Fatalf("could not start server: %v", err)
   }
}
