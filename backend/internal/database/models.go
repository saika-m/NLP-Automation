package database

import (
	"time"

	"gorm.io/gorm"
)

// Task represents a user's automation task request
type Task struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Description string    `json:"description" gorm:"not null"`
	Platform    string    `json:"platform" gorm:"not null"` // windows, linux, macos
	Language    string    `json:"language" gorm:"not null"` // powershell, bash, python, go
	Status      string    `json:"status" gorm:"not null"`   // pending, processing, completed, failed
	UserIP      string    `json:"user_ip"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Generated content
	GeneratedScript string `json:"generated_script,omitempty"`
	Explanation     string `json:"explanation,omitempty"`
	SafetyWarnings  string `json:"safety_warnings,omitempty"`
	FilePath        string `json:"file_path,omitempty"`
	FileName        string `json:"file_name,omitempty"`
	FileSize        int64  `json:"file_size,omitempty"`

	// AI Processing
	AIPrompt   string `json:"ai_prompt,omitempty"`
	AIResponse string `json:"ai_response,omitempty"`

	// Error handling
	ErrorMessage string `json:"error_message,omitempty"`
}

// ScriptTemplate represents predefined script templates
type ScriptTemplate struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Category    string    `json:"category"` // file_management, system_admin, data_processing, etc.
	Platform    string    `json:"platform"`
	Language    string    `json:"language"`
	Template    string    `json:"template" gorm:"type:text"`
	Variables   string    `json:"variables" gorm:"type:text"` // JSON string of variable definitions
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Usage tracking for analytics
type Usage struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	TaskID    string    `json:"task_id"`
	UserIP    string    `json:"user_ip"`
	Platform  string    `json:"platform"`
	Language  string    `json:"language"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`

	Task Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// Safety check results
type SafetyCheck struct {
	ID        string `json:"id" gorm:"primaryKey"`
	TaskID    string `json:"task_id"`
	CheckType string `json:"check_type"` // dangerous_commands, file_operations, network_access, etc.
	Severity  string `json:"severity"`   // low, medium, high, critical
	Message   string `json:"message"`
	Passed    bool   `json:"passed"`

	Task Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// BeforeCreate hook for Task
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	t.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook for Task
func (t *Task) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate hook for ScriptTemplate
func (st *ScriptTemplate) BeforeCreate(tx *gorm.DB) error {
	if st.CreatedAt.IsZero() {
		st.CreatedAt = time.Now()
	}
	st.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook for ScriptTemplate
func (st *ScriptTemplate) BeforeUpdate(tx *gorm.DB) error {
	st.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate hook for Usage
func (u *Usage) BeforeCreate(tx *gorm.DB) error {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	return nil
}
