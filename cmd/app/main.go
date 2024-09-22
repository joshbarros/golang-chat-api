package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/joshbarros/golang-chat-api/docs"
	"github.com/joshbarros/golang-chat-api/internal/config"
	"github.com/joshbarros/golang-chat-api/internal/delivery/http"
	"github.com/joshbarros/golang-chat-api/internal/repository"
	"github.com/joshbarros/golang-chat-api/internal/usecase"
	"github.com/joshbarros/golang-chat-api/internal/workerpool"
	db_pkg "github.com/joshbarros/golang-chat-api/pkg/db"
	"github.com/joshbarros/golang-chat-api/pkg/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func initTracer() *trace.TracerProvider {
	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create OpenTelemetry exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("golang-chat-api"),
		)),
	)

	otel.SetTracerProvider(tp)
	return tp
}

// ApplyMigrations runs the migrations in the migrations folder
func ApplyMigrations(connStr string) error {
	m, err := migrate.New(
		"file://db/migrations",
		connStr)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}
	log.Println("Migrations applied successfully")
	return nil
}

func connectWithRetry(connStr string, maxRetries int) (*sql.DB, error) {
	var db *sql.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Successfully connected to PostgreSQL")
				return db, nil
			}
		}
		log.Printf("Failed to connect to database, retrying... (%d/%d)", i+1, maxRetries)
		time.Sleep(10 * time.Second)
	}
	return nil, fmt.Errorf("could not connect to database after %d retries: %v", maxRetries, err)
}

// @title Golang Chat API
// @version 1.0
// @description This is a Golang Chat API for real-time chat.
// @contact.name API Support
// @contact.url https://www.josuebarros.com/support
// @contact.email goldenglowitsolutions@gmail.com
// @license.name MIT License
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
func main() {
	// Initialize OpenTelemetry
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Initialize Gin router
	router := gin.Default()

  // Use CORS middleware globally
  router.Use(middleware.SetupCORS())

	// Use RateLimiter middleware globally
	router.Use(middleware.RateLimiter())

	// Load configuration from env
	cfg := config.LoadConfig()

	// Construct the database connection string
	dbConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// Apply migrations
	if err := ApplyMigrations(dbConnStr); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	// Retry connecting to PostgreSQL with a retry mechanism
	db, err := connectWithRetry(dbConnStr, 5)  // Retry 5 times
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Set up repositories
	redisClient := db_pkg.InitRedisClient(cfg.RedisHost, cfg.RedisPort)
	userRepo := repository.NewUserRepository(db)
  roomRepo := repository.NewRoomRepository(db)
  messageRepo := repository.NewMessageRepository(db)

  // Initialize Worker Pool with, e.g., 10 workers
  workerPool := workerpool.NewWorkerPool(10, messageRepo)

	// Set up use cases
	userUsecase := usecase.NewUserUsecase(userRepo)
	chatUsecase := usecase.NewChatUsecase(messageRepo, roomRepo, workerPool)

	// Set up handlers
	userHandler := http.NewUserHandler(userUsecase)
	wsHandler := http.NewWSHandler(chatUsecase, redisClient)

	// Public routes
	router.POST("/register", userHandler.Register)
	router.POST("/login", userHandler.Login)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.JWTAuthMiddleware())
  protected.POST("/rooms", wsHandler.CreateRoom)
  protected.GET("/rooms", wsHandler.GetRooms)
  protected.GET("/rooms/:roomID/messages", wsHandler.GetRoomMessages)
	protected.GET("/ws/:roomID", wsHandler.WebSocketHandler)

	// Prometheus metrics
	router.GET("/metrics", middleware.PrometheusHandler())

  // Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start the server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
