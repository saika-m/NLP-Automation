package api

import (
	"context"
	"net/http"
	"time"

	"tashi-backend/internal/config"
	"tashi-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config              *config.Config
	router              *gin.Engine
	handlers            *Handlers
	httpServer          *http.Server
	taskService         *services.TaskService
	commandService      *services.CommandService
	systemService       *services.SystemService
	officeService       *services.OfficeService
	confirmationService *services.ConfirmationService
}

func NewServer(cfg *config.Config, taskService *services.TaskService, commandService *services.CommandService, systemService *services.SystemService, officeService *services.OfficeService, confirmationService *services.ConfirmationService) *Server {
	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Create handlers
	handlers := NewHandlers(taskService, commandService, systemService, officeService, confirmationService)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(handlers.RequestLogger())
	router.Use(handlers.CORS(cfg.CORSOrigins))
	router.Use(handlers.RateLimit())
	router.Use(handlers.ErrorHandler())
	router.Use(gin.Recovery())

	// Create server
	server := &Server{
		config:              cfg,
		router:              router,
		handlers:            handlers,
		taskService:         taskService,
		commandService:      commandService,
		systemService:       systemService,
		officeService:       officeService,
		confirmationService: confirmationService,
	}

	// Setup routes
	server.setupRoutes()

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return server
}

func (s *Server) setupRoutes() {
	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", s.handlers.HealthCheck)

		// AI status
		ai := v1.Group("/ai")
		{
			ai.GET("/status", s.handlers.CheckAIStatus)
		}

		// System information
		system := v1.Group("/system")
		{
			system.GET("/info", s.handlers.GetSystemInfo)
			system.GET("/capabilities", s.handlers.GetSystemCapabilities)
			system.POST("/refresh", s.handlers.RefreshSystemInfo)
		}

		// Statistics
		v1.GET("/stats", s.handlers.GetStats)

		// Tasks
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", s.handlers.CreateTask)
			tasks.GET("", s.handlers.GetTasks)
			tasks.GET("/:id", s.handlers.GetTask)
			tasks.GET("/:id/download", s.handlers.DownloadScript)
		}

		// Templates
		v1.GET("/templates", s.handlers.GetTemplates)

		// Script validation
		v1.POST("/validate", s.handlers.ValidateScript)

		// Commands
		v1.POST("/commands", s.handlers.GenerateAndExecuteCommands)

		// Office automation
		office := v1.Group("/office")
		{
			office.GET("/capabilities", s.handlers.GetOfficeCapabilities)
			office.POST("/execute", s.handlers.ExecuteOfficeAction)
			office.GET("/applications", s.handlers.GetAvailableApplications)
		}

		// Confirmations
		confirmations := v1.Group("/confirmations")
		{
			confirmations.POST("", s.handlers.CreateConfirmation)
			confirmations.GET("", s.handlers.GetPendingConfirmations)
			confirmations.GET("/:id", s.handlers.GetConfirmation)
			confirmations.POST("/:id/respond", s.handlers.RespondToConfirmation)
		}
	}

	// Legacy API routes (without version prefix)
	api := s.router.Group("/api")
	{
		// Health check
		api.GET("/health", s.handlers.HealthCheck)

		// AI status
		ai := api.Group("/ai")
		{
			ai.GET("/status", s.handlers.CheckAIStatus)
		}

		// System information
		system := api.Group("/system")
		{
			system.GET("/info", s.handlers.GetSystemInfo)
			system.GET("/capabilities", s.handlers.GetSystemCapabilities)
			system.POST("/refresh", s.handlers.RefreshSystemInfo)
		}

		// Statistics
		api.GET("/stats", s.handlers.GetStats)

		// Tasks
		tasks := api.Group("/tasks")
		{
			tasks.POST("", s.handlers.CreateTask)
			tasks.GET("", s.handlers.GetTasks)
			tasks.GET("/:id", s.handlers.GetTask)
			tasks.GET("/:id/download", s.handlers.DownloadScript)
		}

		// Templates
		api.GET("/templates", s.handlers.GetTemplates)

		// Script validation
		api.POST("/validate", s.handlers.ValidateScript)

		// Commands
		api.POST("/commands", s.handlers.GenerateAndExecuteCommands)

		// Office automation
		office := api.Group("/office")
		{
			office.GET("/capabilities", s.handlers.GetOfficeCapabilities)
			office.POST("/execute", s.handlers.ExecuteOfficeAction)
			office.GET("/applications", s.handlers.GetAvailableApplications)
		}

		// Confirmations
		confirmations := api.Group("/confirmations")
		{
			confirmations.POST("", s.handlers.CreateConfirmation)
			confirmations.GET("", s.handlers.GetPendingConfirmations)
			confirmations.GET("/:id", s.handlers.GetConfirmation)
			confirmations.POST("/:id/respond", s.handlers.RespondToConfirmation)
		}
	}

	// Root routes
	s.router.GET("/", s.rootHandler)
	s.router.GET("/health", s.handlers.HealthCheck)

	// Serve static files if needed
	s.router.Static("/static", "./static")

	// Handle 404
	s.router.NoRoute(s.notFoundHandler)
}

func (s *Server) rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to Tashi Backend API",
		"version": "1.0.0",
		"docs":    "Visit /api/system/info for system information",
		"endpoints": gin.H{
			"health":    "/api/health",
			"tasks":     "/api/tasks",
			"templates": "/api/templates",
			"system":    "/api/system/info",
			"stats":     "/api/stats",
		},
	})
}

func (s *Server) notFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error":   "Endpoint not found",
		"message": "The requested endpoint does not exist",
		"path":    c.Request.URL.Path,
	})
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
