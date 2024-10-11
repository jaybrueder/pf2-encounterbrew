package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rbcervilla/redisstore/v9"
	"github.com/redis/go-redis/v9"

	"pf2.encounterbrew.com/internal/database"
)

type Server struct {
	port         int
	db           database.Service
	sessionStore sessions.Store
}

func NewServer() (*http.Server, error) {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Create a context
	ctx := context.Background()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"), // e.g., "localhost:6379"
	})

	// Create Redis store
	store, err := redisstore.NewRedisStore(ctx, redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis store: %v", err)
	}

	// Set session cookie options
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // TODO: Set to true if using HTTPS
	})

	newServer := &Server{
		port:         port,
		db:           database.New(),
		sessionStore: store,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server, nil
}
