package db

import (
    "database/sql"
    "log"

    _ "github.com/lib/pq"
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

// ConnectAndMigrate opens a database connection and applies migrations.
func ConnectAndMigrate(databaseURL string) (*sql.DB, error) {
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    m, err := migrate.New("file://migrations", databaseURL)
    if err != nil {
        return nil, err
    }
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return nil, err
    }
    log.Println("database migrated")
    return db, nil
}
