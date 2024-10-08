package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// Insert inserts data into a table.
	Insert(table string, columns []string, values ...interface{}) (sql.Result, error)

	// Query executes a query that returns rows, typically a SELECT.
    Query(query string, args ...interface{}) (*sql.Rows, error)

    // QueryRow executes a query that is expected to return at most one row.
    QueryRow(query string, args ...interface{}) *sql.Row

    // Exec executes a query without returning any rows.
    Exec(query string, args ...interface{}) (sql.Result, error)

    // Begin starts a transaction.
    Begin() (*sql.Tx, error)
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err)) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}

func (s *service) Insert(table string, columns []string, values ...interface{}) (sql.Result, error) {
    // Build the query string
    query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
        table,
        strings.Join(columns, ", "),
        strings.Join(strings.Split(strings.Repeat("?", len(columns)), ""), ", "))

    // Replace ? with $1, $2, etc. for PostgreSQL
    query = convertToPostgresPlaceholders(query)

    // Prepare the statement
    stmt, err := s.db.Prepare(query)
    if err != nil {
        return nil, fmt.Errorf("error preparing statement: %w", err)
    }
    defer stmt.Close()

    // Execute the statement
    return stmt.Exec(values...)
}

// Helper function to convert ? placeholders to $1, $2, etc.
func convertToPostgresPlaceholders(query string) string {
    for i := 1; strings.Contains(query, "?"); i++ {
        query = strings.Replace(query, "?", fmt.Sprintf("$%d", i), 1)
    }
    return query
}

func (s *service) Query(query string, args ...interface{}) (*sql.Rows, error) {
    return s.db.Query(query, args...)
}

func (s *service) QueryRow(query string, args ...interface{}) *sql.Row {
    return s.db.QueryRow(query, args...)
}

func (s *service) Exec(query string, args ...interface{}) (sql.Result, error) {
    return s.db.Exec(query, args...)
}

func (s *service) Begin() (*sql.Tx, error) {
	return s.db.Begin()
}
