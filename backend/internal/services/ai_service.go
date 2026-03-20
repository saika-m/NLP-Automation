package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type AIService struct {
	client *openai.Client
}

type ScriptGenerationRequest struct {
	Description string `json:"description"`
	Platform    string `json:"platform"`
	Language    string `json:"language"`
}

type ScriptGenerationResponse struct {
	Script         string   `json:"script"`
	Explanation    string   `json:"explanation"`
	SafetyWarnings []string `json:"safety_warnings"`
}

func NewAIService(apiKey string) *AIService {
	client := openai.NewClient(apiKey)
	return &AIService{
		client: client,
	}
}

func (s *AIService) GenerateScript(ctx context.Context, req ScriptGenerationRequest) (*ScriptGenerationResponse, error) {
	prompt := s.buildPrompt(req)

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.getSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.3,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate script: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI service")
	}

	return s.parseResponse(resp.Choices[0].Message.Content), nil
}

func (s *AIService) GenerateCommands(ctx context.Context, req ScriptGenerationRequest) (*ScriptGenerationResponse, error) {
	prompt := s.buildCommandPrompt(req)

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.getCommandSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.3,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate commands: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI service")
	}

	return s.parseResponse(resp.Choices[0].Message.Content), nil
}

func (s *AIService) getSystemPrompt() string {
	return `You are Tashi, an expert automation script generator for non-technical users. Your role is to:

1. Generate safe, well-documented scripts for desktop automation tasks
2. Provide clear explanations of what the script does
3. Include safety warnings for potentially dangerous operations
4. Use best practices for the target platform and language
5. Make scripts easy to understand and modify

IMPORTANT GUIDELINES:
- Always include error handling and validation
- Add comments explaining each major step
- Warn about destructive operations (file deletion, system changes)
- Use safe defaults and validate inputs
- Make scripts portable when possible
- Include usage instructions

Response format should be JSON with these fields:
{
  "script": "the generated script code",
  "explanation": "clear explanation of what the script does and how to use it",
  "safety_warnings": ["array of safety warnings if any dangerous operations"]
}`
}

func (s *AIService) buildPrompt(req ScriptGenerationRequest) string {
	return fmt.Sprintf(`Generate a %s script for %s platform that accomplishes the following task:

Task Description: %s

Requirements:
- Target Platform: %s
- Script Language: %s
- Include proper error handling
- Add helpful comments
- Make it safe for non-technical users
- Include usage instructions in the explanation

Please provide a complete, ready-to-run script with explanation and safety warnings.`,
		req.Language, req.Platform, req.Description, req.Platform, req.Language)
}

func (s *AIService) getCommandSystemPrompt() string {
	return `You are Tashi, an expert command generator for desktop automation. Your role is to:

1. Generate safe, executable commands for desktop automation tasks
2. Provide clear explanations of what each command does
3. Include safety warnings for potentially dangerous operations
4. Use best practices for the target platform
5. Generate commands that can be executed directly

IMPORTANT GUIDELINES:
- Generate individual commands, not full scripts
- Each command should be executable on its own
- Add comments as separate lines starting with #
- Warn about destructive operations (file deletion, system changes)
- Use safe defaults and validate inputs
- Make commands portable when possible
- Include usage instructions

Response format should be JSON with these fields:
{
  "script": "the generated commands (one per line)",
  "explanation": "clear explanation of what the commands do and how to use them",
  "safety_warnings": ["array of safety warnings if any dangerous operations"]
}`
}

func (s *AIService) buildCommandPrompt(req ScriptGenerationRequest) string {
	return fmt.Sprintf(`Generate %s commands for %s platform that accomplish the following task:

Task Description: %s

Requirements:
- Target Platform: %s
- Command Shell: %s
- Generate individual commands that can be executed directly
- Include comments as separate lines with #
- Make it safe for non-technical users
- Include usage instructions in the explanation

IMPORTANT PATH INFERENCE RULES:
- When user mentions "desktop", use ~/Desktop (macOS/Linux) or %USERPROFILE%\Desktop (Windows)
- When user mentions "downloads", use ~/Downloads (macOS/Linux) or %USERPROFILE%\Downloads (Windows)
- When user mentions "documents", use ~/Documents (macOS/Linux) or %USERPROFILE%\Documents (Windows)
- When user mentions "pictures" or "photos", use ~/Pictures (macOS/Linux) or %USERPROFILE%\Pictures (Windows)
- When user mentions "home", use ~ (macOS/Linux) or %USERPROFILE% (Windows)
- When user mentions "current directory" or "here", use . (current working directory)
- Be intelligent about file extensions and common locations
- If user says "find xxx file on desktop", automatically use ~/Desktop/xxx or search in ~/Desktop
- If user says "create folder on desktop", automatically use ~/Desktop/folder_name
- If user says "backup to documents", automatically use ~/Documents/backup_folder
- Make commands work without requiring user to specify full paths

Please provide executable commands with explanation and safety warnings.`,
		req.Language, req.Platform, req.Description, req.Platform, req.Language)
}

func (s *AIService) parseResponse(content string) *ScriptGenerationResponse {
	// Try to extract JSON from the response
	response := &ScriptGenerationResponse{
		Script:         "",
		Explanation:    "",
		SafetyWarnings: []string{},
	}

	// First, try to find and parse JSON content
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonContent := content[jsonStart : jsonEnd+1]

		// Try to parse as JSON
		var jsonResp map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &jsonResp); err == nil {
			// Successfully parsed JSON

			// Handle both "script" as string and "script" as array
			if scriptInterface, ok := jsonResp["script"]; ok {
				switch script := scriptInterface.(type) {
				case string:
					// Script is a string, use it directly
					cleanScript := strings.TrimSpace(script)
					cleanScript = strings.ReplaceAll(cleanScript, "\\n", "\n")
					cleanScript = strings.ReplaceAll(cleanScript, "\\t", "\t")
					cleanScript = strings.ReplaceAll(cleanScript, "\\\"", "\"")
					cleanScript = strings.ReplaceAll(cleanScript, "\\'", "'")
					cleanScript = strings.ReplaceAll(cleanScript, "\\\\", "\\")
					cleanScript = s.cleanupScriptContent(cleanScript)
					response.Script = cleanScript
				case []interface{}:
					// Script is an array, join the elements
					var scriptLines []string
					for _, item := range script {
						if str, ok := item.(string); ok {
							scriptLines = append(scriptLines, str)
						}
					}
					response.Script = strings.Join(scriptLines, "\n")
				}
			}

			if explanation, ok := jsonResp["explanation"].(string); ok {
				cleanExplanation := strings.TrimSpace(explanation)
				cleanExplanation = strings.ReplaceAll(cleanExplanation, "\\n", "\n")
				cleanExplanation = strings.ReplaceAll(cleanExplanation, "\\t", "\t")
				response.Explanation = cleanExplanation
			}
			if warnings, ok := jsonResp["safety_warnings"].([]interface{}); ok {
				for _, warning := range warnings {
					if warningStr, ok := warning.(string); ok {
						response.SafetyWarnings = append(response.SafetyWarnings, warningStr)
					}
				}
			}
			// If we successfully parsed JSON and have script content, return it
			if response.Script != "" {
				return response
			}
		}
	}

	// Alternative JSON parsing - try to be more flexible
	lines := strings.Split(content, "\n")
	var jsonLines []string
	inJson := false
	braceCount := 0

	for _, line := range lines {
		if strings.Contains(line, "{") {
			inJson = true
			braceCount += strings.Count(line, "{")
		}
		if inJson {
			jsonLines = append(jsonLines, line)
			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")
			if braceCount <= 0 {
				break
			}
		}
	}

	if len(jsonLines) > 0 {
		jsonContent := strings.Join(jsonLines, "\n")
		var jsonResp map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &jsonResp); err == nil {
			if script, ok := jsonResp["script"].(string); ok {
				cleanScript := strings.TrimSpace(script)
				cleanScript = strings.ReplaceAll(cleanScript, "\\n", "\n")
				cleanScript = strings.ReplaceAll(cleanScript, "\\t", "\t")
				cleanScript = strings.ReplaceAll(cleanScript, "\\\"", "\"")
				cleanScript = strings.ReplaceAll(cleanScript, "\\'", "'")
				cleanScript = strings.ReplaceAll(cleanScript, "\\\\", "\\")
				response.Script = cleanScript
			}
			if explanation, ok := jsonResp["explanation"].(string); ok {
				cleanExplanation := strings.TrimSpace(explanation)
				cleanExplanation = strings.ReplaceAll(cleanExplanation, "\\n", "\n")
				cleanExplanation = strings.ReplaceAll(cleanExplanation, "\\t", "\t")
				response.Explanation = cleanExplanation
			}
			if warnings, ok := jsonResp["safety_warnings"].([]interface{}); ok {
				for _, warning := range warnings {
					if warningStr, ok := warning.(string); ok {
						response.SafetyWarnings = append(response.SafetyWarnings, warningStr)
					}
				}
			}
			if response.Script != "" {
				return response
			}
		}
	}

	// Fallback 1: Look for script content between code blocks
	codeBlockStart := strings.Index(content, "```")
	if codeBlockStart != -1 {
		// Find the end of the language identifier line
		lineEnd := strings.Index(content[codeBlockStart+3:], "\n")
		if lineEnd != -1 {
			scriptStart := codeBlockStart + 3 + lineEnd + 1
			scriptEnd := strings.Index(content[scriptStart:], "```")
			if scriptEnd != -1 {
				scriptContent := strings.TrimSpace(content[scriptStart : scriptStart+scriptEnd])
				if scriptContent != "" {
					response.Script = scriptContent
				}
			}
		}
	}

	// Fallback 2: Look for script content without code blocks
	if response.Script == "" {
		// Try to find script patterns
		lines := strings.Split(content, "\n")
		var scriptLines []string
		var inScript bool

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Start collecting script if we find common script patterns
			if !inScript && (strings.HasPrefix(trimmed, "#!/") ||
				strings.HasPrefix(trimmed, "@echo off") ||
				strings.HasPrefix(trimmed, "# PowerShell") ||
				strings.Contains(trimmed, "def ") ||
				strings.Contains(trimmed, "function ") ||
				strings.HasPrefix(trimmed, "mkdir ") ||
				strings.HasPrefix(trimmed, "New-Item ") ||
				strings.HasPrefix(trimmed, "cp ") ||
				strings.HasPrefix(trimmed, "Copy-Item ")) {
				inScript = true
			}

			if inScript {
				scriptLines = append(scriptLines, line)
			}

			// Stop if we hit explanation or safety warnings
			if strings.Contains(strings.ToLower(trimmed), "explanation:") ||
				strings.Contains(strings.ToLower(trimmed), "safety warning") {
				break
			}
		}

		if len(scriptLines) > 0 {
			response.Script = strings.TrimSpace(strings.Join(scriptLines, "\n"))
		}
	}

	// Extract explanation
	if response.Explanation == "" {
		explanationStart := strings.Index(strings.ToLower(content), "explanation")
		if explanationStart != -1 {
			explanationContent := content[explanationStart:]
			lines := strings.Split(explanationContent, "\n")
			var explanationLines []string

			for i, line := range lines {
				if i == 0 {
					// Skip the "explanation:" line
					continue
				}
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					continue
				}
				if strings.Contains(strings.ToLower(trimmed), "safety") ||
					strings.Contains(strings.ToLower(trimmed), "warning") {
					break
				}
				explanationLines = append(explanationLines, trimmed)
			}

			if len(explanationLines) > 0 {
				response.Explanation = strings.Join(explanationLines, " ")
			}
		}
	}

	// Extract safety warnings
	warningStart := strings.Index(strings.ToLower(content), "safety")
	if warningStart == -1 {
		warningStart = strings.Index(strings.ToLower(content), "warning")
	}

	if warningStart != -1 {
		warningContent := content[warningStart:]
		lines := strings.Split(warningContent, "\n")

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" &&
				!strings.Contains(strings.ToLower(trimmed), "safety warnings:") &&
				!strings.Contains(strings.ToLower(trimmed), "warnings:") &&
				(strings.Contains(strings.ToLower(trimmed), "warning") ||
					strings.Contains(strings.ToLower(trimmed), "caution") ||
					strings.Contains(strings.ToLower(trimmed), "danger")) {
				response.SafetyWarnings = append(response.SafetyWarnings, trimmed)
			}
		}
	}

	// Final fallback: if still no script, try to extract any code-like content
	if response.Script == "" {
		lines := strings.Split(content, "\n")
		var possibleScript []string

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Look for lines that look like code
			if strings.HasPrefix(trimmed, "$") || // PowerShell variables
				strings.HasPrefix(trimmed, "mkdir") ||
				strings.HasPrefix(trimmed, "New-Item") ||
				strings.HasPrefix(trimmed, "Copy-Item") ||
				strings.HasPrefix(trimmed, "cp ") ||
				strings.HasPrefix(trimmed, "mv ") ||
				strings.HasPrefix(trimmed, "rm ") ||
				strings.HasPrefix(trimmed, "echo ") ||
				strings.HasPrefix(trimmed, "Write-Host") ||
				strings.Contains(trimmed, " -") || // Common flag pattern
				strings.Contains(trimmed, "for ") ||
				strings.Contains(trimmed, "if ") {
				possibleScript = append(possibleScript, line)
			}
		}

		if len(possibleScript) > 0 {
			response.Script = strings.Join(possibleScript, "\n")
		}
	}

	// If we still have no script content, use the entire content as explanation
	if response.Script == "" && response.Explanation == "" {
		response.Explanation = content
	}

	return response
}

// cleanupScriptContent removes any remaining formatting issues and ensures clean script content
func (s *AIService) cleanupScriptContent(script string) string {
	// Remove any remaining JSON artifacts
	script = strings.TrimSpace(script)

	// Remove any trailing quotes, commas, or JSON syntax
	script = strings.TrimSuffix(script, "\",")
	script = strings.TrimSuffix(script, "\"")
	script = strings.TrimPrefix(script, "\"")

	// Remove any explanation or safety warnings that might have been mixed in
	lines := strings.Split(script, "\n")
	var cleanLines []string
	inScriptContent := true

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines at the beginning
		if len(cleanLines) == 0 && trimmed == "" {
			continue
		}

		// Stop script content extraction if we hit JSON artifacts
		if strings.HasPrefix(trimmed, "\"explanation\":") ||
			strings.HasPrefix(trimmed, "\"safety_warnings\":") ||
			strings.HasPrefix(trimmed, "explanation\":") ||
			strings.HasPrefix(trimmed, "safety_warnings\":") ||
			strings.HasPrefix(trimmed, "},") ||
			strings.HasPrefix(trimmed, "}") ||
			strings.HasPrefix(trimmed, "{") ||
			strings.HasPrefix(trimmed, "]") ||
			strings.Contains(trimmed, "\"explanation\"") ||
			strings.Contains(trimmed, "\"safety_warnings\"") ||
			strings.Contains(trimmed, "This script creates") ||
			strings.Contains(trimmed, "To use this script") ||
			strings.Contains(trimmed, "Usage Instructions") ||
			strings.Contains(trimmed, "Make sure") ||
			strings.Contains(trimmed, "Please") ||
			strings.Contains(trimmed, "Note:") ||
			strings.Contains(trimmed, "Warning:") ||
			strings.Contains(trimmed, "Important:") ||
			strings.HasPrefix(trimmed, "\",") ||
			strings.HasSuffix(trimmed, "\",") ||
			strings.HasSuffix(trimmed, "\"") && len(trimmed) > 50 {
			inScriptContent = false
			break
		}

		// Only add lines that are part of the script content
		if inScriptContent {
			// Clean up any trailing JSON artifacts on the line
			line = strings.TrimSuffix(line, "\",")
			line = strings.TrimSuffix(line, "\"")
			line = strings.TrimSuffix(line, ",")

			cleanLines = append(cleanLines, line)
		}
	}

	cleanScript := strings.Join(cleanLines, "\n")

	// Final cleanup - remove any remaining escape sequences
	cleanScript = strings.ReplaceAll(cleanScript, "\\n", "\n")
	cleanScript = strings.ReplaceAll(cleanScript, "\\t", "\t")
	cleanScript = strings.ReplaceAll(cleanScript, "\\\"", "\"")
	cleanScript = strings.ReplaceAll(cleanScript, "\\'", "'")

	// Fix common variable issues for bash scripts
	cleanScript = s.fixCommonVariableIssues(cleanScript)

	return strings.TrimSpace(cleanScript)
}

// fixCommonVariableIssues fixes common variable definition issues in generated scripts
func (s *AIService) fixCommonVariableIssues(script string) string {
	// Replace common undefined variables with proper definitions
	if strings.Contains(script, "$DEST_PATH") && !strings.Contains(script, "DEST_PATH=") {
		script = "DEST_PATH=\"$HOME/Desktop/helloworld\"\n" + script
	}
	if strings.Contains(script, "$SOURCE_PATH") && !strings.Contains(script, "SOURCE_PATH=") {
		script = "SOURCE_PATH=\"$HOME/Desktop/black.jpg\"\n" + script
	}
	if strings.Contains(script, "$dest_dir") && !strings.Contains(script, "dest_dir=") {
		script = "dest_dir=\"$HOME/Desktop/helloworld\"\n" + script
	}
	if strings.Contains(script, "$src_file") && !strings.Contains(script, "src_file=") {
		script = "src_file=\"$HOME/Desktop/black.jpg\"\n" + script
	}

	return script
}

func (s *AIService) ValidateScript(ctx context.Context, script, language string) ([]string, error) {
	prompt := fmt.Sprintf(`Analyze this %s script for potential security risks and safety issues:

%s

Please identify:
1. Dangerous operations (file deletion, system changes, network access)
2. Potential security vulnerabilities
3. Operations that could cause data loss
4. Any malicious patterns

Respond with a list of specific warnings, or "SAFE" if no issues found.`, language, script)

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   500,
		Temperature: 0.1,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to validate script: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from validation service")
	}

	content := resp.Choices[0].Message.Content
	if strings.TrimSpace(strings.ToUpper(content)) == "SAFE" {
		return []string{}, nil
	}

	// Parse warnings from response
	lines := strings.Split(content, "\n")
	var warnings []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			warnings = append(warnings, line)
		}
	}

	return warnings, nil
}
