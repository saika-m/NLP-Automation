package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type CommandService struct {
	aiService *AIService
}

type CommandRequest struct {
	Description string `json:"description"`
	WorkingDir  string `json:"working_dir,omitempty"`
	DryRun      bool   `json:"dry_run,omitempty"`
}

type CommandResponse struct {
	Commands       []string `json:"commands"`
	Explanation    string   `json:"explanation"`
	SafetyWarnings []string `json:"safety_warnings"`
	Output         string   `json:"output,omitempty"`
	Error          string   `json:"error,omitempty"`
	ExecutionTime  string   `json:"execution_time,omitempty"`
}

func NewCommandService(aiService *AIService) *CommandService {
	return &CommandService{
		aiService: aiService,
	}
}

func (s *CommandService) GenerateAndExecuteCommand(ctx context.Context, req CommandRequest) (*CommandResponse, error) {
	// Generate commands using AI
	aiReq := ScriptGenerationRequest{
		Description: req.Description,
		Platform:    s.getCurrentPlatform(),
		Language:    s.getShellForPlatform(),
	}

	fmt.Printf("Generating commands for: %s\n", req.Description)
	fmt.Printf("Platform: %s, Language: %s\n", aiReq.Platform, aiReq.Language)

	// Update AI service to generate commands instead of scripts
	aiResp, err := s.aiService.GenerateCommands(ctx, aiReq)
	if err != nil {
		fmt.Printf("Error generating commands: %v\n", err)
		return nil, fmt.Errorf("failed to generate commands: %w", err)
	}

	fmt.Printf("AI Response received, parsing commands...\n")
	response := &CommandResponse{
		Commands:       s.parseCommands(aiResp.Script),
		Explanation:    aiResp.Explanation,
		SafetyWarnings: aiResp.SafetyWarnings,
	}

	fmt.Printf("Generated %d commands\n", len(response.Commands))

	// If dry run, just return the commands without executing
	if req.DryRun {
		return response, nil
	}

	// Execute the commands
	fmt.Printf("Executing commands...\n")
	output, execErr := s.executeCommands(ctx, response.Commands, req.WorkingDir)
	response.Output = output
	if execErr != nil {
		response.Error = execErr.Error()
	}

	return response, nil
}

func (s *CommandService) executeCommands(ctx context.Context, commands []string, workingDir string) (string, error) {
	if len(commands) == 0 {
		return "", fmt.Errorf("no commands to execute")
	}

	// Set working directory
	if workingDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		workingDir = homeDir
	}

	// Validate working directory exists
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		return "", fmt.Errorf("working directory does not exist: %s", workingDir)
	}

	var outputs []string
	startTime := time.Now()

	for i, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(command, "#") || strings.HasPrefix(command, "//") {
			continue
		}

		fmt.Printf("Executing command %d: %s\n", i+1, command)

		var output string
		var err error

		// Check if command needs sudo and use appropriate execution method
		if s.needsSudo(command) {
			output, err = s.executeCommandWithSudo(ctx, command, workingDir)
		} else {
			output, err = s.executeCommand(ctx, command, workingDir)
		}

		if err != nil {
			return strings.Join(outputs, "\n"), fmt.Errorf("command %d failed: %w", i+1, err)
		}

		if output != "" {
			outputs = append(outputs, fmt.Sprintf("Command %d output: %s", i+1, output))
		}
	}

	executionTime := time.Since(startTime)
	outputs = append(outputs, fmt.Sprintf("Total execution time: %s", executionTime))

	return strings.Join(outputs, "\n"), nil
}

func (s *CommandService) executeCommand(ctx context.Context, command, workingDir string) (string, error) {
	// Parse command based on platform
	shell, args := s.getShellCommand(command)

	cmd := exec.CommandContext(ctx, shell, args...)
	cmd.Dir = workingDir

	// Set environment variables
	cmd.Env = os.Environ()

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %w", err)
	}

	return string(output), nil
}

func (s *CommandService) executeCommandWithSudo(ctx context.Context, command, workingDir string) (string, error) {
	// Parse command based on platform and add sudo prefix
	shell, args := s.getShellCommandWithSudo(command)

	cmd := exec.CommandContext(ctx, shell, args...)
	cmd.Dir = workingDir

	// Set environment variables
	cmd.Env = os.Environ()

	// Execute command with sudo
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("sudo command execution failed: %w", err)
	}

	return string(output), nil
}

func (s *CommandService) getShellCommand(command string) (string, []string) {
	platform := runtime.GOOS

	switch platform {
	case "windows":
		return "cmd", []string{"/c", command}
	case "darwin", "linux":
		return "/bin/bash", []string{"-c", command}
	default:
		return "/bin/sh", []string{"-c", command}
	}
}

func (s *CommandService) getShellCommandWithSudo(command string) (string, []string) {
	platform := runtime.GOOS

	switch platform {
	case "windows":
		// On Windows, use runas for elevated privileges
		return "cmd", []string{"/c", command}
	case "darwin", "linux":
		// Prefix command with sudo
		sudoCommand := fmt.Sprintf("sudo %s", command)
		return "/bin/bash", []string{"-c", sudoCommand}
	default:
		sudoCommand := fmt.Sprintf("sudo %s", command)
		return "/bin/sh", []string{"-c", sudoCommand}
	}
}

func (s *CommandService) needsSudo(command string) bool {
	// Check if command likely needs sudo privileges
	command = strings.ToLower(strings.TrimSpace(command))

	// Commands that typically need sudo
	sudoCommands := []string{
		"apt", "yum", "dnf", "pacman", "zypper", // package managers
		"systemctl", "service", // system services
		"mount", "umount", // filesystem operations
		"fdisk", "mkfs", "fsck", // disk operations
		"iptables", "ufw", // firewall
		"modprobe", "insmod", "rmmod", // kernel modules
	}

	// Check if command starts with any sudo-requiring command
	for _, sudoCmd := range sudoCommands {
		if strings.HasPrefix(command, sudoCmd+" ") || command == sudoCmd {
			return true
		}
	}

	// Check for operations on system directories
	systemPaths := []string{
		"/etc/", "/var/", "/usr/", "/opt/", "/sys/", "/proc/",
	}

	for _, path := range systemPaths {
		if strings.Contains(command, path) {
			return true
		}
	}

	// Check if already has sudo
	if strings.HasPrefix(command, "sudo ") {
		return false // already has sudo
	}

	return false
}

func (s *CommandService) getCurrentPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	default:
		return "linux"
	}
}

func (s *CommandService) getShellForPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return "cmd"
	case "darwin":
		return "bash"
	case "linux":
		return "bash"
	default:
		return "bash"
	}
}

func (s *CommandService) parseCommands(script string) []string {
	// Handle JSON format that might be returned from AI
	script = strings.TrimSpace(script)

	// Check if the script looks like JSON
	if strings.HasPrefix(script, "{") && strings.HasSuffix(script, "}") {
		// Try to extract script array from JSON using a more robust approach
		if strings.Contains(script, `"script":[`) {
			// Find the script array in the JSON
			start := strings.Index(script, `"script":[`)
			if start != -1 {
				start += len(`"script":[`)

				// Find the matching closing bracket for the array
				bracketCount := 1
				end := start
				for i := start; i < len(script) && bracketCount > 0; i++ {
					if script[i] == '[' {
						bracketCount++
					} else if script[i] == ']' {
						bracketCount--
					}
					if bracketCount == 0 {
						end = i
						break
					}
				}

				if end > start {
					scriptArray := script[start:end]

					// Parse the JSON array more carefully
					var commands []string
					inQuotes := false
					escaped := false
					currentCommand := ""

					for i := 0; i < len(scriptArray); i++ {
						char := scriptArray[i]

						if escaped {
							currentCommand += string(char)
							escaped = false
							continue
						}

						if char == '\\' {
							escaped = true
							continue
						}

						if char == '"' {
							inQuotes = !inQuotes
							continue // Don't include quotes in the command
						}

						if char == ',' && !inQuotes {
							// End of command
							command := strings.TrimSpace(currentCommand)
							if command != "" {
								commands = append(commands, command)
							}
							currentCommand = ""
						} else if inQuotes {
							currentCommand += string(char)
						}
					}

					// Add the last command
					if currentCommand != "" {
						command := strings.TrimSpace(currentCommand)
						if command != "" {
							commands = append(commands, command)
						}
					}

					return commands
				}
			}
		}
	}

	// Original string parsing logic
	lines := strings.Split(script, "\n")
	var commands []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip shebang lines
		if strings.HasPrefix(line, "#!") {
			continue
		}

		// Skip comments but include them for context
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			commands = append(commands, line)
			continue
		}

		// Skip variable assignments that are just for script context
		if strings.Contains(line, "=") && !strings.Contains(line, "mkdir") && !strings.Contains(line, "cp") && !strings.Contains(line, "mv") {
			// This might be a variable assignment, skip if it's just for script context
			continue
		}

		commands = append(commands, line)
	}

	return commands
}

func (s *CommandService) ValidateCommand(ctx context.Context, command string) ([]string, error) {
	// Check if command contains dangerous operations
	dangerous := []string{
		"rm -rf",
		"del /f",
		"format",
		"fdisk",
		"dd if=",
		"shutdown",
		"reboot",
		"sudo rm",
		"sudo dd",
		"chmod 777",
		"chown root",
	}

	var warnings []string
	commandLower := strings.ToLower(command)

	for _, danger := range dangerous {
		if strings.Contains(commandLower, danger) {
			warnings = append(warnings, fmt.Sprintf("Warning: Command contains potentially dangerous operation: %s", danger))
		}
	}

	// Check for network operations
	network := []string{
		"curl",
		"wget",
		"ssh",
		"scp",
		"ftp",
		"telnet",
	}

	for _, net := range network {
		if strings.Contains(commandLower, net) {
			warnings = append(warnings, fmt.Sprintf("Warning: Command performs network operation: %s", net))
		}
	}

	return warnings, nil
}

func (s *CommandService) GetFileSystemInfo(path string) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Check if path exists
	stat, err := os.Stat(path)
	if err != nil {
		return info, err
	}

	info["exists"] = true
	info["is_dir"] = stat.IsDir()
	info["size"] = stat.Size()
	info["mod_time"] = stat.ModTime()
	info["permissions"] = stat.Mode().String()

	// If it's a directory, list contents
	if stat.IsDir() {
		files, err := os.ReadDir(path)
		if err != nil {
			return info, err
		}

		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.Name())
		}
		info["contents"] = fileNames
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err == nil {
		info["absolute_path"] = absPath
	}

	return info, nil
}
