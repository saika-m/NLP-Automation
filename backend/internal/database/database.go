package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initialize sets up the database connection and runs migrations
func Initialize(dbPath string) (*gorm.DB, error) {
	// Configure GORM logger
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, err
	}

	// Run auto migrations
	if err := runMigrations(db); err != nil {
		return nil, err
	}

	// Seed initial data
	if err := seedData(db); err != nil {
		log.Printf("Warning: Failed to seed initial data: %v", err)
	}

	return db, nil
}

// runMigrations runs all database migrations
func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&Task{},
		&ScriptTemplate{},
		&Usage{},
		&SafetyCheck{},
	)
}

// seedData inserts initial data into the database
func seedData(db *gorm.DB) error {
	// Check if templates already exist
	var count int64
	db.Model(&ScriptTemplate{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	// Seed script templates
	templates := []ScriptTemplate{
		{
			ID:          "file-rename-date",
			Name:        "Rename Files by Date",
			Description: "Rename files in a directory by adding date prefix",
			Category:    "file_management",
			Platform:    "windows",
			Language:    "powershell",
			Template: `# Rename files by date
$sourceDir = "{{.SourceDirectory}}"
$dateFormat = "{{.DateFormat}}"

Get-ChildItem -Path $sourceDir -File | ForEach-Object {
    $newName = (Get-Date -Format $dateFormat) + "_" + $_.Name
    Rename-Item -Path $_.FullName -NewName $newName
    Write-Host "Renamed: $($_.Name) -> $newName"
}`,
			Variables: `{"SourceDirectory": {"type": "string", "description": "Directory containing files to rename"}, "DateFormat": {"type": "string", "description": "Date format (e.g., yyyy-MM-dd)", "default": "yyyy-MM-dd"}}`,
		},
		{
			ID:          "file-organize-type",
			Name:        "Organize Files by Type",
			Description: "Move files to folders based on their extensions",
			Category:    "file_management",
			Platform:    "linux",
			Language:    "bash",
			Template: `#!/bin/bash
# Organize files by type
SOURCE_DIR="{{.SourceDirectory}}"
TARGET_DIR="{{.TargetDirectory}}"

cd "$SOURCE_DIR" || exit 1

for file in *.*; do
    if [ -f "$file" ]; then
        extension="${file##*.}"
        mkdir -p "$TARGET_DIR/$extension"
        mv "$file" "$TARGET_DIR/$extension/"
        echo "Moved: $file -> $TARGET_DIR/$extension/"
    fi
done`,
			Variables: `{"SourceDirectory": {"type": "string", "description": "Source directory path"}, "TargetDirectory": {"type": "string", "description": "Target directory for organized files"}}`,
		},
		{
			ID:          "csv-data-process",
			Name:        "Process CSV Data",
			Description: "Read, filter, and process CSV data",
			Category:    "data_processing",
			Platform:    "cross",
			Language:    "python",
			Template: `#!/usr/bin/env python3
import pandas as pd
import sys

def process_csv():
    input_file = "{{.InputFile}}"
    output_file = "{{.OutputFile}}"
    filter_column = "{{.FilterColumn}}"
    filter_value = "{{.FilterValue}}"
    
    try:
        # Read CSV file
        df = pd.read_csv(input_file)
        print(f"Loaded {len(df)} rows from {input_file}")
        
        # Apply filter if specified
        if filter_column and filter_value:
            df = df[df[filter_column] == filter_value]
            print(f"Filtered to {len(df)} rows where {filter_column} = {filter_value}")
        
        # Save processed data
        df.to_csv(output_file, index=False)
        print(f"Saved processed data to {output_file}")
        
    except Exception as e:
        print(f"Error processing CSV: {e}")
        sys.exit(1)

if __name__ == "__main__":
    process_csv()`,
			Variables: `{"InputFile": {"type": "string", "description": "Input CSV file path"}, "OutputFile": {"type": "string", "description": "Output CSV file path"}, "FilterColumn": {"type": "string", "description": "Column to filter by (optional)"}, "FilterValue": {"type": "string", "description": "Value to filter for (optional)"}}`,
		},
	}

	for _, template := range templates {
		if err := db.Create(&template).Error; err != nil {
			return err
		}
	}

	log.Println("Database seeded with initial script templates")
	return nil
}

// Repository interface for database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Task operations
func (r *Repository) CreateTask(task *Task) error {
	return r.db.Create(task).Error
}

func (r *Repository) GetTask(id string) (*Task, error) {
	var task Task
	err := r.db.First(&task, "id = ?", id).Error
	return &task, err
}

func (r *Repository) UpdateTask(task *Task) error {
	return r.db.Save(task).Error
}

func (r *Repository) GetTasksByStatus(status string) ([]Task, error) {
	var tasks []Task
	err := r.db.Where("status = ?", status).Find(&tasks).Error
	return tasks, err
}

// Template operations
func (r *Repository) GetTemplates() ([]ScriptTemplate, error) {
	var templates []ScriptTemplate
	err := r.db.Find(&templates).Error
	return templates, err
}

func (r *Repository) GetTemplatesByCategory(category string) ([]ScriptTemplate, error) {
	var templates []ScriptTemplate
	err := r.db.Where("category = ?", category).Find(&templates).Error
	return templates, err
}

// Usage tracking
func (r *Repository) CreateUsage(usage *Usage) error {
	return r.db.Create(usage).Error
}

// Safety check operations
func (r *Repository) CreateSafetyCheck(check *SafetyCheck) error {
	return r.db.Create(check).Error
}

func (r *Repository) GetSafetyChecksByTask(taskID string) ([]SafetyCheck, error) {
	var checks []SafetyCheck
	err := r.db.Where("task_id = ?", taskID).Find(&checks).Error
	return checks, err
}
