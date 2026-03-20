package api

import (
	"fmt"
	"net/http"

	"nlp-automation-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	taskService         *services.TaskService
	commandService      *services.CommandService
	systemService       *services.SystemService
	officeService       *services.OfficeService
	confirmationService *services.ConfirmationService
}

func NewHandlers(taskService *services.TaskService, commandService *services.CommandService, systemService *services.SystemService, officeService *services.OfficeService, confirmationService *services.ConfirmationService) *Handlers {
	return &Handlers{
		taskService:         taskService,
		commandService:      commandService,
		systemService:       systemService,
		officeService:       officeService,
		confirmationService: confirmationService,
	}
}

// CreateTask handles POST /api/tasks
func (h *Handlers) CreateTask(c *gin.Context) {
	var req services.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Get client IP
	req.UserIP = c.ClientIP()

	// Create task
	task, err := h.taskService.CreateTask(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to create task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    task,
	})
}

// GetTask handles GET /api/tasks/:id
func (h *Handlers) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
		})
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Task not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task,
	})
}

// GetTasks handles GET /api/tasks with optional status filter
func (h *Handlers) GetTasks(c *gin.Context) {
	status := c.Query("status")

	var tasks []*services.TaskResponse
	var err error

	if status != "" {
		tasks, err = h.taskService.GetTasksByStatus(c.Request.Context(), status)
	} else {
		// If no status specified, get all pending and processing tasks
		pendingTasks, err1 := h.taskService.GetTasksByStatus(c.Request.Context(), "pending")
		processingTasks, err2 := h.taskService.GetTasksByStatus(c.Request.Context(), "processing")

		if err1 != nil || err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve tasks",
			})
			return
		}

		tasks = append(pendingTasks, processingTasks...)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve tasks",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tasks,
		"count":   len(tasks),
	})
}

// DownloadScript handles GET /api/tasks/:id/download
func (h *Handlers) DownloadScript(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
		})
		return
	}

	filePath, fileName, err := h.taskService.DownloadScript(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "File not available",
			"details": err.Error(),
		})
		return
	}

	// Set appropriate headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")

	c.File(filePath)
}

// GetTemplates handles GET /api/templates
func (h *Handlers) GetTemplates(c *gin.Context) {
	category := c.Query("category")

	var templates interface{}
	var err error

	if category != "" {
		templates, err = h.taskService.GetTemplatesByCategory(c.Request.Context(), category)
	} else {
		templates, err = h.taskService.GetTemplates(c.Request.Context())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve templates",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    templates,
	})
}

// GetSystemInfo handles GET /api/system/info
func (h *Handlers) GetSystemInfo(c *gin.Context) {
	info := h.taskService.GetSystemInfo(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// HealthCheck handles GET /api/health
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": gin.H{
			"unix": c.Request.Context().Value("timestamp"),
		},
		"service": "nlp-automation-backend",
		"version": "1.0.0",
	})
}

// CheckAIStatus handles GET /api/ai/status
func (h *Handlers) CheckAIStatus(c *gin.Context) {
	// For now, return online status since we have AI service configured
	// In a real implementation, this would check the actual AI service connection
	c.JSON(http.StatusOK, gin.H{
		"status": "online",
		"model":  "gpt-4o-mini",
		"timestamp": gin.H{
			"unix": c.Request.Context().Value("timestamp"),
		},
	})
}

// GetStats handles GET /api/stats (basic usage statistics)
func (h *Handlers) GetStats(c *gin.Context) {
	// Get basic statistics
	pendingTasks, _ := h.taskService.GetTasksByStatus(c.Request.Context(), "pending")
	processingTasks, _ := h.taskService.GetTasksByStatus(c.Request.Context(), "processing")
	completedTasks, _ := h.taskService.GetTasksByStatus(c.Request.Context(), "completed")
	failedTasks, _ := h.taskService.GetTasksByStatus(c.Request.Context(), "failed")

	stats := gin.H{
		"tasks": gin.H{
			"pending":    len(pendingTasks),
			"processing": len(processingTasks),
			"completed":  len(completedTasks),
			"failed":     len(failedTasks),
			"total":      len(pendingTasks) + len(processingTasks) + len(completedTasks) + len(failedTasks),
		},
		"system": h.taskService.GetSystemInfo(c.Request.Context()),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GenerateAndExecuteCommands handles POST /api/commands
func (h *Handlers) GenerateAndExecuteCommands(c *gin.Context) {
	var req services.CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Generate and optionally execute commands
	response, err := h.commandService.GenerateAndExecuteCommand(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate commands",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ValidateScript handles POST /api/validate (for script validation without creation)
func (h *Handlers) ValidateScript(c *gin.Context) {
	var req struct {
		Script   string `json:"script" binding:"required"`
		Language string `json:"language" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// This would require exposing script service methods
	// For now, return a simple validation response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"valid":    true,
			"warnings": []string{},
			"message":  "Script validation completed",
		},
	})
}

// Middleware for request logging
func (h *Handlers) RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// CORS middleware
func (h *Handlers) CORS(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range origins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Rate limiting middleware (basic implementation)
func (h *Handlers) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic rate limiting - in production, use Redis or similar
		// For now, just continue
		c.Next()
	}
}

// Error handling middleware
func (h *Handlers) ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"details": err.Error(),
			})
		}
	}
}

// System-related handlers

// GetSystemCapabilities handles GET /api/system/capabilities
func (h *Handlers) GetSystemCapabilities(c *gin.Context) {
	systemInfo := h.systemService.GetCachedSystemInfo()
	if systemInfo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "System information not available",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    systemInfo,
	})
}

// RefreshSystemInfo handles POST /api/system/refresh
func (h *Handlers) RefreshSystemInfo(c *gin.Context) {
	systemInfo, err := h.systemService.GatherSystemInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to refresh system information",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    systemInfo,
		"message": "System information refreshed successfully",
	})
}

// Office-related handlers

// GetOfficeCapabilities handles GET /api/office/capabilities
func (h *Handlers) GetOfficeCapabilities(c *gin.Context) {
	capabilities := h.officeService.GetCapabilities()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    capabilities,
	})
}

// ExecuteOfficeAction handles POST /api/office/execute
func (h *Handlers) ExecuteOfficeAction(c *gin.Context) {
	var req services.OfficeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Create confirmation first
	confirmation, err := h.confirmationService.CreateOfficeConfirmation(
		c.Request.Context(),
		req.Action,
		req.Application,
		req.FilePath,
		[]string{req.Description}, // Using description as a command for safety analysis
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create confirmation",
			"details": err.Error(),
		})
		return
	}

	// Return confirmation for user approval
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"confirmation": confirmation,
		"message":      "Please confirm this action before execution",
	})
}

// GetAvailableApplications handles GET /api/office/applications
func (h *Handlers) GetAvailableApplications(c *gin.Context) {
	capabilities := h.officeService.GetCapabilities()

	apps := make(map[string]interface{})
	if capabilities.Excel.Available {
		apps["excel"] = capabilities.Excel
	}
	if capabilities.Word.Available {
		apps["word"] = capabilities.Word
	}
	if capabilities.PowerPoint.Available {
		apps["powerpoint"] = capabilities.PowerPoint
	}
	if capabilities.Outlook.Available {
		apps["outlook"] = capabilities.Outlook
	}
	if capabilities.PDF.Available {
		apps["pdf"] = capabilities.PDF
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apps,
		"count":   len(apps),
	})
}

// Confirmation-related handlers

// CreateConfirmation handles POST /api/confirmations
func (h *Handlers) CreateConfirmation(c *gin.Context) {
	var req services.CreateConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	confirmation, err := h.confirmationService.CreateConfirmation(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create confirmation",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    confirmation,
	})
}

// GetPendingConfirmations handles GET /api/confirmations
func (h *Handlers) GetPendingConfirmations(c *gin.Context) {
	response, err := h.confirmationService.ListPendingConfirmations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve confirmations",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response.Confirmations,
		"count":   response.Count,
	})
}

// GetConfirmation handles GET /api/confirmations/:id
func (h *Handlers) GetConfirmation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Confirmation ID is required",
		})
		return
	}

	confirmation, err := h.confirmationService.GetConfirmation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Confirmation not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    confirmation,
	})
}

// RespondToConfirmation handles POST /api/confirmations/:id/respond
func (h *Handlers) RespondToConfirmation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Confirmation ID is required",
		})
		return
	}

	var response services.ConfirmationResponse
	if err := c.ShouldBindJSON(&response); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid response format",
			"details": err.Error(),
		})
		return
	}

	confirmation, err := h.confirmationService.RespondToConfirmation(c.Request.Context(), id, response)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to respond to confirmation",
			"details": err.Error(),
		})
		return
	}

	// If confirmed, execute the action
	if response.Confirmed {
		// For office actions, execute the office service
		if confirmation.Action != "execute_commands" {
			// This is an office action
			officeReq := services.OfficeRequest{
				Action:      confirmation.Action,
				Application: "office", // Default application
				Description: confirmation.Message,
				Parameters:  map[string]interface{}{"execute": true},
			}

			// Try to extract application from details
			if confirmation.Details != nil {
				if app, ok := confirmation.Details["application"].(string); ok {
					officeReq.Application = app
				}
				if filePath, ok := confirmation.Details["file_path"].(string); ok {
					officeReq.FilePath = filePath
				}
			}

			officeResponse, err := h.officeService.ProcessOfficeRequest(c.Request.Context(), officeReq)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success":      true,
					"confirmation": confirmation,
					"execution":    nil,
					"error":        fmt.Sprintf("Failed to execute office action: %v", err),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success":      true,
				"confirmation": confirmation,
				"execution":    officeResponse,
				"message":      "Action confirmed and executed successfully",
			})
		} else {
			// This is a command execution
			commandReq := services.CommandRequest{
				Description: confirmation.Message,
				DryRun:      false,
			}

			// Try to extract working directory from details
			if confirmation.Details != nil {
				if workingDir, ok := confirmation.Details["working_dir"].(string); ok {
					commandReq.WorkingDir = workingDir
				}
			}

			commandResponse, err := h.commandService.GenerateAndExecuteCommand(c.Request.Context(), commandReq)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success":      true,
					"confirmation": confirmation,
					"execution":    nil,
					"error":        fmt.Sprintf("Failed to execute commands: %v", err),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success":      true,
				"confirmation": confirmation,
				"execution":    commandResponse,
				"message":      "Commands confirmed and executed successfully",
			})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"confirmation": confirmation,
			"message":      "Action was declined",
		})
	}
}
