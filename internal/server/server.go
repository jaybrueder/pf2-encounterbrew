package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/joho/godotenv/autoload"

	"pf2.encounterbrew.com/internal/database"
)

type Server struct {
	port int
	db   database.Service
}

func NewServer() (*http.Server, error) {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080 // Default port
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL environment variable not set")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		return nil, fmt.Errorf("MIGRATIONS_PATH environment variable not set")
	}

	// Ensure the path is in the correct format
	migrationsURL := fmt.Sprintf("file://%s", migrationsPath)

	// Database Initialization
	dbService := database.New()
	if dbService == nil { // Or check for an error if database.New() returns one
		return nil, fmt.Errorf("failed to initialize database service")
	}

	log.Println("Running database migrations...")
	err := runMigrations(dbURL, migrationsURL)
	if err != nil {
		return nil, fmt.Errorf("could not run database migrations: %w", err)
	}
	log.Println("Database migrations finished successfully.")

	// Server setup
	newServer := &Server{
		port: port,
		db:   dbService,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Server starting on port %d\n", newServer.port)

	return server, nil
}

// Helper function to run migrations
func runMigrations(databaseURL string, migrationsURL string) error {
	// databaseURL needs to be in the format expected by the migrate driver
	// e.g., "postgres://user:pass@host:port/db?sslmode=disable"
	// migrationsURL needs to be in the format "file://path/to/dir"

	m, err := migrate.New(migrationsURL, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Apply all available "up" migrations
	if err := m.Up(); err != nil {
		// ErrNoChange means migrations were already up-to-date
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No new migrations to apply.")
			return nil // Not an error in this context
		}
		// For other errors, return them
		return fmt.Errorf("migration failed: %w", err)
	}

	// Check for dirty migrations after running Up (optional but good practice)
	version, dirty, err := m.Version()
	if err != nil {
		// This check might fail if ErrNoChange occurred, handle appropriately if needed
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Printf("Warning: Could not get migration version after run: %v", err)
		}
	} else if dirty {
		// This indicates an issue during the last migration attempt (before this run or during)
		return fmt.Errorf("migration check failed: database is in dirty state at version %d", version)
	}

	log.Printf("Database migrated to version %d", version)
	return nil
}
