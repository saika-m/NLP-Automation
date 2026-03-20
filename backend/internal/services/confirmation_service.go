package services

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ConfirmationService struct {
	pendingConfirmations map[string]*ConfirmationRequest
}

type ConfirmationRequest struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Action      string                 `json:"action"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Commands    []string               `json:"commands,omitempty"`
	SafetyLevel string                 `json:"safety_level"` // low, medium, high, critical
	Timestamp   time.Time              `json:"timestamp"`
	Response    *ConfirmationResponse  `json:"response,omitempty"`
	Completed   bool                   `json:"completed"`
}

type ConfirmationResponse struct {
	Confirmed bool                   `json:"confirmed"`
	Reason    string                 `json:"reason,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type CreateConfirmationRequest struct {
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Action      string                 `json:"action"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Commands    []string               `json:"commands,omitempty"`
	SafetyLevel string                 `json:"safety_level,omitempty"`
}

type ConfirmationListResponse struct {
	Confirmations []*ConfirmationRequest `json:"confirmations"`
	Count         int                    `json:"count"`
}

func NewConfirmationService() *ConfirmationService {
	return &ConfirmationService{
		pendingConfirmations: make(map[string]*ConfirmationRequest),
	}
}

func (s *ConfirmationService) CreateConfirmation(ctx context.Context, req CreateConfirmationRequest) (*ConfirmationRequest, error) {
	// Generate unique ID
	id := s.generateConfirmationID()

	// Determine safety level if not provided
	safetyLevel := req.SafetyLevel
	if safetyLevel == "" {
		safetyLevel = s.determineSafetyLevel(req.Commands, req.Action)
	}

	confirmation := &ConfirmationRequest{
		ID:          id,
		Title:       req.Title,
		Message:     req.Message,
		Action:      req.Action,
		Details:     req.Details,
		Commands:    req.Commands,
		SafetyLevel: safetyLevel,
		Timestamp:   time.Now(),
		Completed:   false,
	}

	s.pendingConfirmations[id] = confirmation
	return confirmation, nil
}

func (s *ConfirmationService) RespondToConfirmation(ctx context.Context, id string, response ConfirmationResponse) (*ConfirmationRequest, error) {
	confirmation, exists := s.pendingConfirmations[id]
	if !exists {
		return nil, fmt.Errorf("confirmation with ID %s not found", id)
	}

	if confirmation.Completed {
		return nil, fmt.Errorf("confirmation with ID %s already completed", id)
	}

	response.Timestamp = time.Now()
	confirmation.Response = &response
	confirmation.Completed = true

	return confirmation, nil
}

func (s *ConfirmationService) GetConfirmation(ctx context.Context, id string) (*ConfirmationRequest, error) {
	confirmation, exists := s.pendingConfirmations[id]
	if !exists {
		return nil, fmt.Errorf("confirmation with ID %s not found", id)
	}

	return confirmation, nil
}

func (s *ConfirmationService) ListPendingConfirmations(ctx context.Context) (*ConfirmationListResponse, error) {
	var pending []*ConfirmationRequest

	for _, confirmation := range s.pendingConfirmations {
		if !confirmation.Completed {
			pending = append(pending, confirmation)
		}
	}

	return &ConfirmationListResponse{
		Confirmations: pending,
		Count:         len(pending),
	}, nil
}

func (s *ConfirmationService) CleanupCompletedConfirmations(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	for id, confirmation := range s.pendingConfirmations {
		if confirmation.Completed && confirmation.Timestamp.Before(cutoff) {
			delete(s.pendingConfirmations, id)
		}
	}

	return nil
}

func (s *ConfirmationService) generateConfirmationID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("conf_%d", timestamp)
}

func (s *ConfirmationService) determineSafetyLevel(commands []string, action string) string {
	// Analyze commands and action to determine safety level
	dangerousPatterns := map[string]string{
		"rm -rf":     "critical",
		"del /f":     "critical",
		"format":     "critical",
		"fdisk":      "critical",
		"dd if=":     "critical",
		"shutdown":   "high",
		"reboot":     "high",
		"sudo rm":    "critical",
		"sudo dd":    "critical",
		"chmod 777":  "high",
		"chown root": "high",
	}

	networkPatterns := []string{
		"curl", "wget", "ssh", "scp", "ftp", "telnet",
	}

	systemPatterns := []string{
		"systemctl", "service", "mount", "umount",
	}

	highestLevel := "low"

	// Check each command for dangerous patterns
	for _, command := range commands {
		commandLower := strings.ToLower(command)

		// Check for critical/dangerous patterns
		for pattern, level := range dangerousPatterns {
			if strings.Contains(commandLower, pattern) {
				if level == "critical" {
					return "critical"
				}
				if level == "high" && highestLevel != "critical" {
					highestLevel = "high"
				}
			}
		}

		// Check for network operations
		for _, pattern := range networkPatterns {
			if strings.Contains(commandLower, pattern) {
				if highestLevel == "low" {
					highestLevel = "medium"
				}
			}
		}

		// Check for system operations
		for _, pattern := range systemPatterns {
			if strings.Contains(commandLower, pattern) {
				if highestLevel == "low" {
					highestLevel = "medium"
				}
			}
		}
	}

	// Check action type
	actionLower := strings.ToLower(action)
	if strings.Contains(actionLower, "delete") || strings.Contains(actionLower, "remove") {
		if highestLevel == "low" {
			highestLevel = "medium"
		}
	}

	if strings.Contains(actionLower, "format") || strings.Contains(actionLower, "wipe") {
		return "critical"
	}

	return highestLevel
}

func (s *ConfirmationService) GetSafetyLevelColor(level string) string {
	switch level {
	case "low":
		return "#28a745" // Green
	case "medium":
		return "#ffc107" // Yellow
	case "high":
		return "#fd7e14" // Orange
	case "critical":
		return "#dc3545" // Red
	default:
		return "#6c757d" // Gray
	}
}

func (s *ConfirmationService) GetSafetyLevelDescription(level string) string {
	switch level {
	case "low":
		return "This operation is safe and has minimal risk."
	case "medium":
		return "This operation has moderate risk and should be reviewed."
	case "high":
		return "This operation has high risk and requires careful consideration."
	case "critical":
		return "This operation is potentially dangerous and could cause data loss or system damage."
	default:
		return "Safety level unknown."
	}
}

// Helper method to create office-specific confirmations
func (s *ConfirmationService) CreateOfficeConfirmation(ctx context.Context, action string, application string, filePath string, commands []string) (*ConfirmationRequest, error) {
	title := fmt.Sprintf("Confirm %s Operation", strings.Title(application))

	message := fmt.Sprintf("You are about to perform a %s operation", action)
	if filePath != "" {
		message += fmt.Sprintf(" on file: %s", filePath)
	}

	details := map[string]interface{}{
		"application":   application,
		"action":        action,
		"file_path":     filePath,
		"command_count": len(commands),
	}

	req := CreateConfirmationRequest{
		Title:       title,
		Message:     message,
		Action:      action,
		Details:     details,
		Commands:    commands,
		SafetyLevel: "", // Will be determined automatically
	}

	return s.CreateConfirmation(ctx, req)
}

// Helper method to create system command confirmations
func (s *ConfirmationService) CreateCommandConfirmation(ctx context.Context, description string, commands []string, workingDir string) (*ConfirmationRequest, error) {
	title := "Confirm Command Execution"
	message := fmt.Sprintf("You are about to execute: %s", description)

	details := map[string]interface{}{
		"description":   description,
		"working_dir":   workingDir,
		"command_count": len(commands),
	}

	req := CreateConfirmationRequest{
		Title:       title,
		Message:     message,
		Action:      "execute_commands",
		Details:     details,
		Commands:    commands,
		SafetyLevel: "", // Will be determined automatically
	}

	return s.CreateConfirmation(ctx, req)
}
