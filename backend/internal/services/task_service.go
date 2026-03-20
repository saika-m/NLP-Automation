package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"tashi-backend/internal/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskService struct {
	db              *gorm.DB
	repo            *database.Repository
	scriptService   *ScriptService
	aiService       *AIService
	compilerService *CompilerService
}

type CreateTaskRequest struct {
	Description string `json:"description" binding:"required"`
	Platform    string `json:"platform" binding:"required"`
	Language    string `json:"language" binding:"required"`
	UserIP      string `json:"user_ip,omitempty"`
}

type TaskResponse struct {
	ID              string                 `json:"id"`
	Description     string                 `json:"description"`
	Platform        string                 `json:"platform"`
	Language        string                 `json:"language"`
	Status          string                 `json:"status"`
	GeneratedScript string                 `json:"generated_script,omitempty"`
	Explanation     string                 `json:"explanation,omitempty"`
	SafetyWarnings  []string               `json:"safety_warnings,omitempty"`
	FilePath        string                 `json:"file_path,omitempty"`
	FileName        string                 `json:"file_name,omitempty"`
	FileSize        int64                  `json:"file_size,omitempty"`
	Preview         map[string]interface{} `json:"preview,omitempty"`
	CompilationInfo *CompilationResult     `json:"compilation_info,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
}

func NewTaskService(db *gorm.DB, scriptService *ScriptService, aiService *AIService, compilerService *CompilerService) *TaskService {
	return &TaskService{
		db:              db,
		repo:            database.NewRepository(db),
		scriptService:   scriptService,
		aiService:       aiService,
		compilerService: compilerService,
	}
}

// CreateTask creates a new automation task
func (ts *TaskService) CreateTask(ctx context.Context, req CreateTaskRequest) (*TaskResponse, error) {
	// Generate unique task ID
	taskID := uuid.New().String()

	// Validate request
	if err := ts.validateCreateTaskRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create task record
	task := &database.Task{
		ID:          taskID,
		Description: req.Description,
		Platform:    strings.ToLower(req.Platform),
		Language:    strings.ToLower(req.Language),
		Status:      "pending",
		UserIP:      req.UserIP,
	}

	if err := ts.repo.CreateTask(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Start async processing
	go ts.processTaskAsync(context.Background(), taskID)

	return ts.taskToResponse(task), nil
}

// GetTask retrieves a task by ID
func (ts *TaskService) GetTask(ctx context.Context, taskID string) (*TaskResponse, error) {
	task, err := ts.repo.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	response := ts.taskToResponse(task)

	// Add safety checks if available
	if safetyChecks, err := ts.repo.GetSafetyChecksByTask(taskID); err == nil {
		var warnings []string
		for _, check := range safetyChecks {
			if !check.Passed {
				warnings = append(warnings, fmt.Sprintf("[%s] %s", strings.ToUpper(check.Severity), check.Message))
			}
		}
		response.SafetyWarnings = warnings
	}

	return response, nil
}

// GetTasksByStatus retrieves tasks by status
func (ts *TaskService) GetTasksByStatus(ctx context.Context, status string) ([]*TaskResponse, error) {
	tasks, err := ts.repo.GetTasksByStatus(status)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	var responses []*TaskResponse
	for _, task := range tasks {
		responses = append(responses, ts.taskToResponse(&task))
	}

	return responses, nil
}

// DownloadScript provides the script file for download
func (ts *TaskService) DownloadScript(ctx context.Context, taskID string) (string, string, error) {
	task, err := ts.repo.GetTask(taskID)
	if err != nil {
		return "", "", fmt.Errorf("task not found: %w", err)
	}

	if task.Status != "completed" {
		return "", "", fmt.Errorf("task not completed yet")
	}

	if task.FilePath == "" {
		return "", "", fmt.Errorf("no file available for download")
	}

	// Check if file exists
	if _, err := os.Stat(task.FilePath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("file not found on disk")
	}

	return task.FilePath, task.FileName, nil
}

// processTaskAsync processes the task asynchronously
func (ts *TaskService) processTaskAsync(ctx context.Context, taskID string) {
	// Update status to processing
	if err := ts.updateTaskStatus(taskID, "processing"); err != nil {
		log.Printf("Failed to update task status: %v", err)
		return
	}

	// Get task details
	task, err := ts.repo.GetTask(taskID)
	if err != nil {
		log.Printf("Failed to get task: %v", err)
		ts.updateTaskError(taskID, "Failed to retrieve task details")
		return
	}

	// Generate script using AI
	aiReq := ScriptGenerationRequest{
		Description: task.Description,
		Platform:    task.Platform,
		Language:    task.Language,
	}

	aiResp, err := ts.aiService.GenerateScript(ctx, aiReq)
	if err != nil {
		log.Printf("Failed to generate script: %v", err)
		ts.updateTaskError(taskID, fmt.Sprintf("AI generation failed: %v", err))
		return
	}

	// Validate script content
	scriptWarnings := ts.scriptService.ValidateScriptContent(aiResp.Script, task.Language)

	// Combine AI warnings with script validation warnings
	allWarnings := append(aiResp.SafetyWarnings, scriptWarnings...)

	// Save script to file
	filePath, fileName, err := ts.scriptService.SaveScript(aiResp.Script, task.Language, taskID)
	if err != nil {
		log.Printf("Failed to save script: %v", err)
		ts.updateTaskError(taskID, fmt.Sprintf("Failed to save script: %v", err))
		return
	}

	// Get file size
	fileInfo, _ := os.Stat(filePath)
	var fileSize int64
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	// Attempt compilation if needed
	var compilationResult *CompilationResult
	if ts.compilerService.needsCompilation(task.Language) {
		compReq := CompilationRequest{
			SourcePath: filePath,
			Language:   task.Language,
			Platform:   task.Platform,
			TaskID:     taskID,
		}

		compilationResult, err = ts.compilerService.CompileScript(ctx, compReq)
		if err != nil {
			log.Printf("Compilation failed: %v", err)
			// Don't fail the entire task, just log the compilation error
			compilationResult = &CompilationResult{
				Success:      false,
				ErrorMessage: err.Error(),
			}
		}

		// If compilation succeeded, update file path to compiled executable
		if compilationResult.Success && compilationResult.OutputPath != "" {
			filePath = compilationResult.OutputPath
			fileName = compilationResult.OutputName

			// Update file size for compiled executable
			if fileInfo, err := os.Stat(filePath); err == nil {
				fileSize = fileInfo.Size()
			}
		}
	}

	// Update task with results
	task.GeneratedScript = aiResp.Script
	task.Explanation = aiResp.Explanation
	task.SafetyWarnings = strings.Join(allWarnings, "\n")
	task.FilePath = filePath
	task.FileName = fileName
	task.FileSize = fileSize
	task.Status = "completed"

	if err := ts.repo.UpdateTask(task); err != nil {
		log.Printf("Failed to update task: %v", err)
		return
	}

	// Create usage record
	usage := &database.Usage{
		ID:       uuid.New().String(),
		TaskID:   taskID,
		UserIP:   task.UserIP,
		Platform: task.Platform,
		Language: task.Language,
		Success:  true,
	}
	ts.repo.CreateUsage(usage)

	// Store safety checks
	ts.storeSafetyChecks(taskID, allWarnings)

	log.Printf("Task %s completed successfully", taskID)
}

// validateCreateTaskRequest validates the create task request
func (ts *TaskService) validateCreateTaskRequest(req CreateTaskRequest) error {
	// Validate platform
	validPlatforms := []string{"windows", "linux", "macos", "darwin", "cross"}
	if !ts.contains(validPlatforms, strings.ToLower(req.Platform)) {
		return fmt.Errorf("invalid platform: %s", req.Platform)
	}

	// Validate language
	validLanguages := []string{"powershell", "bash", "shell", "python", "go", "golang", "c", "cpp", "c++", "rust", "batch"}
	if !ts.contains(validLanguages, strings.ToLower(req.Language)) {
		return fmt.Errorf("invalid language: %s", req.Language)
	}

	// Validate description
	if len(strings.TrimSpace(req.Description)) < 10 {
		return fmt.Errorf("description must be at least 10 characters long")
	}

	if len(req.Description) > 1000 {
		return fmt.Errorf("description must be less than 1000 characters")
	}

	return nil
}

// updateTaskStatus updates the task status
func (ts *TaskService) updateTaskStatus(taskID, status string) error {
	task, err := ts.repo.GetTask(taskID)
	if err != nil {
		return err
	}

	task.Status = status
	return ts.repo.UpdateTask(task)
}

// updateTaskError updates the task with error information
func (ts *TaskService) updateTaskError(taskID, errorMsg string) {
	task, err := ts.repo.GetTask(taskID)
	if err != nil {
		log.Printf("Failed to get task for error update: %v", err)
		return
	}

	task.Status = "failed"
	task.ErrorMessage = errorMsg

	if err := ts.repo.UpdateTask(task); err != nil {
		log.Printf("Failed to update task with error: %v", err)
	}
}

// storeSafetyChecks stores safety check results
func (ts *TaskService) storeSafetyChecks(taskID string, warnings []string) {
	for _, warning := range warnings {
		severity := "medium"
		if strings.Contains(strings.ToLower(warning), "critical") {
			severity = "critical"
		} else if strings.Contains(strings.ToLower(warning), "high") {
			severity = "high"
		} else if strings.Contains(strings.ToLower(warning), "low") {
			severity = "low"
		}

		check := &database.SafetyCheck{
			ID:        uuid.New().String(),
			TaskID:    taskID,
			CheckType: "script_analysis",
			Severity:  severity,
			Message:   warning,
			Passed:    false,
		}

		ts.repo.CreateSafetyCheck(check)
	}
}

// taskToResponse converts database task to response format
func (ts *TaskService) taskToResponse(task *database.Task) *TaskResponse {
	response := &TaskResponse{
		ID:              task.ID,
		Description:     task.Description,
		Platform:        task.Platform,
		Language:        task.Language,
		Status:          task.Status,
		GeneratedScript: task.GeneratedScript,
		Explanation:     task.Explanation,
		FilePath:        task.FilePath,
		FileName:        task.FileName,
		FileSize:        task.FileSize,
		CreatedAt:       task.CreatedAt,
		UpdatedAt:       task.UpdatedAt,
		ErrorMessage:    task.ErrorMessage,
	}

	// Parse safety warnings
	if task.SafetyWarnings != "" {
		response.SafetyWarnings = strings.Split(task.SafetyWarnings, "\n")
	}

	// Add script preview if script is available
	if task.GeneratedScript != "" {
		response.Preview = ts.scriptService.GetScriptPreview(task.GeneratedScript, task.Language)
	}

	return response
}

// contains checks if a slice contains a string
func (ts *TaskService) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetTemplates returns available script templates
func (ts *TaskService) GetTemplates(ctx context.Context) ([]database.ScriptTemplate, error) {
	return ts.repo.GetTemplates()
}

// GetTemplatesByCategory returns templates filtered by category
func (ts *TaskService) GetTemplatesByCategory(ctx context.Context, category string) ([]database.ScriptTemplate, error) {
	return ts.repo.GetTemplatesByCategory(category)
}

// GetSystemInfo returns system information and capabilities
func (ts *TaskService) GetSystemInfo(ctx context.Context) map[string]interface{} {
	compilerEnv := ts.compilerService.ValidateCompilerEnvironment()
	supportedLanguages := ts.compilerService.GetSupportedLanguages()

	return map[string]interface{}{
		"supported_platforms":  []string{"windows", "linux", "macos", "cross"},
		"supported_languages":  []string{"powershell", "bash", "python", "go", "c", "cpp", "rust", "batch"},
		"compiled_languages":   supportedLanguages,
		"compiler_environment": compilerEnv,
		"version":              "1.0.0",
		"features": map[string]bool{
			"ai_generation":      true,
			"safety_checks":      true,
			"script_compilation": len(supportedLanguages) > 0,
			"cross_platform":     true,
			"templates":          true,
		},
	}
}
