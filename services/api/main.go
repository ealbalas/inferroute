package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/inferroute/inferroute/internal/cache"
	"github.com/inferroute/inferroute/internal/database"
	"github.com/inferroute/inferroute/internal/observability"
	"github.com/inferroute/inferroute/services/api/handlers"
	"github.com/inferroute/inferroute/services/api/middleware"
)

func main() {
	log := observability.NewLogger()

	// ── Config ────────────────────────────────────────────────────────────────
	port       := getEnv("PORT", "8080")
	dbURL      := getEnv("DATABASE_URL", "postgres://inferroute:inferroute@localhost:5432/inferroute?sslmode=disable")
	redisURL   := getEnv("REDIS_URL", "redis://localhost:6379")
	routerURL  := getEnv("ROUTER_URL", "http://localhost:8081")
	jwtSecret  := getEnv("JWT_SECRET", "change-me-in-production")

	// ── Dependencies ──────────────────────────────────────────────────────────
	ctx := context.Background()

	db, err := database.Connect(ctx, dbURL)
	if err != nil {
		log.Error("connect to postgres", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient, err := cache.New(redisURL)
	if err != nil {
		log.Error("connect to redis", "err", err)
		os.Exit(1)
	}

	// ── Router ────────────────────────────────────────────────────────────────
	gin.SetMode(getEnv("GIN_MODE", "release"))
	r := gin.New()
	r.Use(
		middleware.RequestLogger(log),
		gin.Recovery(),
		corsMiddleware(),
	)

	r.GET("/health", handlers.Health)

	v1 := r.Group("/v1")
	{
		// Public — auth endpoints
		v1.POST("/auth/register", handlers.Register(db, jwtSecret))
		v1.POST("/auth/login", handlers.Login(db, jwtSecret))

		// Protected by API key — inference endpoints
		inference := v1.Group("")
		inference.Use(middleware.APIKeyAuth(db))
		inference.Use(middleware.RateLimit(redisClient, 10, time.Second))
		{
			inference.POST("/generate", handlers.Generate(db, redisClient, routerURL))
			inference.GET("/models", handlers.Models)
		}

		// Protected by JWT — dashboard endpoints
		dash := v1.Group("")
		dash.Use(middleware.JWTAuth(jwtSecret))
		{
			dash.GET("/usage",    handlers.Usage(db))
			dash.GET("/requests", handlers.Requests(db))
			dash.GET("/workers",  handlers.Workers(routerURL))
			dash.GET("/keys",     handlers.ListAPIKeys(db))
			dash.POST("/keys",    handlers.CreateAPIKey(db))
			dash.DELETE("/keys/:id", handlers.DeleteAPIKey(db))
		}
	}

	// ── Server ────────────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("api service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("shutdown complete")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func init() {
	_ = slog.Default() // ensure logger is initialized
}
