package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"nlp-automation-backend/internal/config"
)

type CompilerService struct {
	config *config.Config
}

type CompilationRequest struct {
	SourcePath string
	Language   string
	Platform   string
	TaskID     string
}

type CompilationResult struct {
	Success      bool   `json:"success"`
	OutputPath   string `json:"output_path,omitempty"`
	OutputName   string `json:"output_name,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	CompileTime  string `json:"compile_time,omitempty"`
}

func NewCompilerService(cfg *config.Config) *CompilerService {
	return &CompilerService{
		config: cfg,
	}
}

// CompileScript compiles source code to executable if applicable
func (c *CompilerService) CompileScript(ctx context.Context, req CompilationRequest) (*CompilationResult, error) {
	startTime := time.Now()

	// Check if compilation is needed for this language
	if !c.needsCompilation(req.Language) {
		return &CompilationResult{
			Success:     true,
			OutputPath:  req.SourcePath,
			OutputName:  filepath.Base(req.SourcePath),
			CompileTime: time.Since(startTime).String(),
		}, nil
	}

	// Create output filename
	outputName := c.getExecutableName(req.TaskID, req.Platform)
	outputPath := filepath.Join(c.config.OutputDir, outputName)

	// Compile based on language
	var err error
	switch strings.ToLower(req.Language) {
	case "go", "golang":
		err = c.compileGo(ctx, req.SourcePath, outputPath, req.Platform)
	case "c":
		err = c.compileC(ctx, req.SourcePath, outputPath, req.Platform)
	case "cpp", "c++":
		err = c.compileCpp(ctx, req.SourcePath, outputPath, req.Platform)
	case "rust":
		err = c.compileRust(ctx, req.SourcePath, outputPath, req.Platform)
	default:
		return nil, fmt.Errorf("compilation not supported for language: %s", req.Language)
	}

	result := &CompilationResult{
		CompileTime: time.Since(startTime).String(),
	}

	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result, nil
	}

	result.Success = true
	result.OutputPath = outputPath
	result.OutputName = outputName

	return result, nil
}

// needsCompilation checks if the language requires compilation
func (c *CompilerService) needsCompilation(language string) bool {
	compiledLanguages := []string{"go", "golang", "c", "cpp", "c++", "rust", "java"}

	for _, lang := range compiledLanguages {
		if strings.ToLower(language) == lang {
			return true
		}
	}

	return false
}

// getExecutableName generates executable filename based on platform
func (c *CompilerService) getExecutableName(taskID, platform string) string {
	timestamp := time.Now().Format("20060102_150405")
	baseName := fmt.Sprintf("nlp-automation_executable_%s_%s", taskID[:8], timestamp)

	switch strings.ToLower(platform) {
	case "windows":
		return baseName + ".exe"
	default:
		return baseName
	}
}

// compileGo compiles Go source code
func (c *CompilerService) compileGo(ctx context.Context, sourcePath, outputPath, platform string) error {
	// Set environment variables for cross-compilation
	env := os.Environ()

	switch strings.ToLower(platform) {
	case "windows":
		env = append(env, "GOOS=windows", "GOARCH=amd64")
	case "linux":
		env = append(env, "GOOS=linux", "GOARCH=amd64")
	case "macos", "darwin":
		env = append(env, "GOOS=darwin", "GOARCH=amd64")
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(c.config.CompilationTimeout)*time.Second)
	defer cancel()

	// Run go build command
	cmd := exec.CommandContext(timeoutCtx, "go", "build", "-o", outputPath, sourcePath)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go compilation failed: %s\nOutput: %s", err.Error(), string(output))
	}

	return nil
}

// compileC compiles C source code
func (c *CompilerService) compileC(ctx context.Context, sourcePath, outputPath, platform string) error {
	compiler := "gcc"
	if runtime.GOOS == "darwin" {
		compiler = "clang"
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(c.config.CompilationTimeout)*time.Second)
	defer cancel()

	args := []string{"-o", outputPath, sourcePath}

	// Add platform-specific flags
	switch strings.ToLower(platform) {
	case "windows":
		if runtime.GOOS != "windows" {
			compiler = "x86_64-w64-mingw32-gcc"
		}
	}

	cmd := exec.CommandContext(timeoutCtx, compiler, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("C compilation failed: %s\nOutput: %s", err.Error(), string(output))
	}

	return nil
}

// compileCpp compiles C++ source code
func (c *CompilerService) compileCpp(ctx context.Context, sourcePath, outputPath, platform string) error {
	compiler := "g++"
	if runtime.GOOS == "darwin" {
		compiler = "clang++"
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(c.config.CompilationTimeout)*time.Second)
	defer cancel()

	args := []string{"-o", outputPath, sourcePath, "-std=c++17"}

	// Add platform-specific flags
	switch strings.ToLower(platform) {
	case "windows":
		if runtime.GOOS != "windows" {
			compiler = "x86_64-w64-mingw32-g++"
		}
	}

	cmd := exec.CommandContext(timeoutCtx, compiler, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("C++ compilation failed: %s\nOutput: %s", err.Error(), string(output))
	}

	return nil
}

// compileRust compiles Rust source code
func (c *CompilerService) compileRust(ctx context.Context, sourcePath, outputPath, platform string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(c.config.CompilationTimeout)*time.Second)
	defer cancel()

	// Create a temporary Cargo project
	tempDir := filepath.Join(c.config.TempDir, fmt.Sprintf("rust_project_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create Cargo.toml
	cargoToml := `[package]
name = "nlp-automation_script"
version = "0.1.0"
edition = "2021"

[[bin]]
name = "main"
path = "src/main.rs"
`

	if err := os.WriteFile(filepath.Join(tempDir, "Cargo.toml"), []byte(cargoToml), 0644); err != nil {
		return fmt.Errorf("failed to create Cargo.toml: %w", err)
	}

	// Create src directory and copy source
	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create src directory: %w", err)
	}

	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(filepath.Join(srcDir, "main.rs"), sourceContent, 0644); err != nil {
		return fmt.Errorf("failed to write main.rs: %w", err)
	}

	// Set target based on platform
	var target string
	switch strings.ToLower(platform) {
	case "windows":
		target = "x86_64-pc-windows-gnu"
	case "linux":
		target = "x86_64-unknown-linux-gnu"
	case "macos", "darwin":
		target = "x86_64-apple-darwin"
	}

	// Build with cargo
	args := []string{"build", "--release"}
	if target != "" {
		args = append(args, "--target", target)
	}

	cmd := exec.CommandContext(timeoutCtx, "cargo", args...)
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Rust compilation failed: %s\nOutput: %s", err.Error(), string(output))
	}

	// Copy the built executable
	var builtPath string
	if target != "" {
		builtPath = filepath.Join(tempDir, "target", target, "release", "main")
	} else {
		builtPath = filepath.Join(tempDir, "target", "release", "main")
	}

	if strings.ToLower(platform) == "windows" {
		builtPath += ".exe"
	}

	sourceFile, err := os.Open(builtPath)
	if err != nil {
		return fmt.Errorf("failed to open built executable: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy executable: %w", err)
	}

	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(outputPath, 0755); err != nil {
			return fmt.Errorf("failed to make executable: %w", err)
		}
	}

	return nil
}

// ValidateCompilerEnvironment checks if required compilers are available
func (c *CompilerService) ValidateCompilerEnvironment() map[string]bool {
	compilers := map[string]string{
		"go":    "go",
		"gcc":   "gcc",
		"g++":   "g++",
		"cargo": "cargo",
		"rustc": "rustc",
	}

	if runtime.GOOS == "darwin" {
		compilers["clang"] = "clang"
		compilers["clang++"] = "clang++"
	}

	availability := make(map[string]bool)

	for name, command := range compilers {
		_, err := exec.LookPath(command)
		availability[name] = err == nil
	}

	return availability
}

// GetSupportedLanguages returns list of languages that can be compiled
func (c *CompilerService) GetSupportedLanguages() []string {
	env := c.ValidateCompilerEnvironment()
	var supported []string

	if env["go"] {
		supported = append(supported, "go", "golang")
	}
	if env["gcc"] {
		supported = append(supported, "c")
	}
	if env["g++"] || env["clang++"] {
		supported = append(supported, "cpp", "c++")
	}
	if env["cargo"] && env["rustc"] {
		supported = append(supported, "rust")
	}

	return supported
}
