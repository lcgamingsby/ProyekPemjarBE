package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"backend/config"
	dbpkg "backend/database"
	"backend/handlers"
	"backend/middleware"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg := config.Load()

	db, err := dbpkg.Connect(cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	if err != nil {
		log.Fatalf("db connect err: %v", err)
	}
	defer db.Close()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// Public routes
	api := r.Group("/api")
	authH := handlers.NewAuthHandler(db, cfg.JWTSecret, cfg.JWTExpiresH)
	api.POST("/register", authH.Register)
	api.POST("/login", authH.Login)

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	sessionsH := handlers.NewSessionsHandler(db)
	protected.GET("/sessions", sessionsH.ListSessions)
	protected.POST("/sessions", sessionsH.CreateSession)
	protected.GET("/profile", func(c *gin.Context) {
		claims, _ := c.Get(middleware.JwtClaimsKey)
		c.JSON(200, gin.H{"profile": claims})
	})

	// Public join (by code)
	api.GET("/sessions/join/:code", sessionsH.GetSessionByCode)

	// WebSocket endpoint (placeholder) — expects token in query or auth header
	r.GET("/ws", middleware.JWTAuth(cfg.JWTSecret), func(c *gin.Context) {
		// For now simple response show connected
		c.String(200, "WebSocket endpoint placeholder - token valid")
	})

	log.Printf("starting server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
