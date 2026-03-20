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

type OfficeService struct {
	systemService *SystemService
	aiService     *AIService
}

type OfficeRequest struct {
	Action      string                 `json:"action"`
	Application string                 `json:"application"` // excel, word, powerpoint, etc.
	FilePath    string                 `json:"file_path,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Description string                 `json:"description"`
}

type OfficeResponse struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	FilePath   string   `json:"file_path,omitempty"`
	Output     string   `json:"output,omitempty"`
	Error      string   `json:"error,omitempty"`
	Commands   []string `json:"commands,omitempty"`
	ScriptPath string   `json:"script_path,omitempty"`
}

type OfficeCapabilities struct {
	Excel      ExcelCapabilities      `json:"excel"`
	Word       WordCapabilities       `json:"word"`
	PowerPoint PowerPointCapabilities `json:"powerpoint"`
	Outlook    OutlookCapabilities    `json:"outlook"`
	PDF        PDFCapabilities        `json:"pdf"`
	System     SystemCapabilities     `json:"system"`
}

type ExcelCapabilities struct {
	Available   bool     `json:"available"`
	Version     string   `json:"version"`
	Operations  []string `json:"operations"`
	ScriptTypes []string `json:"script_types"`
}

type WordCapabilities struct {
	Available   bool     `json:"available"`
	Version     string   `json:"version"`
	Operations  []string `json:"operations"`
	ScriptTypes []string `json:"script_types"`
}

type PowerPointCapabilities struct {
	Available   bool     `json:"available"`
	Version     string   `json:"version"`
	Operations  []string `json:"operations"`
	ScriptTypes []string `json:"script_types"`
}

type OutlookCapabilities struct {
	Available   bool     `json:"available"`
	Version     string   `json:"version"`
	Operations  []string `json:"operations"`
	ScriptTypes []string `json:"script_types"`
}

type PDFCapabilities struct {
	Available  bool     `json:"available"`
	Tools      []string `json:"tools"`
	Operations []string `json:"operations"`
}

type SystemCapabilities struct {
	Platform        string   `json:"platform"`
	ScriptLanguages []string `json:"script_languages"`
	Automation      []string `json:"automation"`
}

func NewOfficeService(systemService *SystemService, aiService *AIService) *OfficeService {
	return &OfficeService{
		systemService: systemService,
		aiService:     aiService,
	}
}

func (s *OfficeService) GetCapabilities() *OfficeCapabilities {
	caps := &OfficeCapabilities{}

	// Detect Excel capabilities
	caps.Excel = s.detectExcelCapabilities()

	// Detect Word capabilities
	caps.Word = s.detectWordCapabilities()

	// Detect PowerPoint capabilities
	caps.PowerPoint = s.detectPowerPointCapabilities()

	// Detect Outlook capabilities
	caps.Outlook = s.detectOutlookCapabilities()

	// Detect PDF capabilities
	caps.PDF = s.detectPDFCapabilities()

	// System capabilities
	caps.System = s.detectSystemCapabilities()

	return caps
}

func (s *OfficeService) detectExcelCapabilities() ExcelCapabilities {
	caps := ExcelCapabilities{
		Operations: []string{
			"create_workbook", "open_workbook", "save_workbook",
			"add_worksheet", "delete_worksheet", "rename_worksheet",
			"read_cell", "write_cell", "read_range", "write_range",
			"insert_row", "delete_row", "insert_column", "delete_column",
			"create_chart", "format_cells", "apply_formula",
			"sort_data", "filter_data", "pivot_table", "conditional_formatting",
			"merge_cells", "unmerge_cells", "auto_fit", "freeze_panes",
		},
	}

	switch runtime.GOOS {
	case "darwin":
		// Check for Microsoft Excel on macOS
		if s.isApplicationAvailable("/Applications/Microsoft Excel.app") {
			caps.Available = true
			caps.Version = s.getExcelVersionMac()
			caps.ScriptTypes = []string{"applescript", "javascript", "python"}
		}
	case "windows":
		// Check for Microsoft Excel on Windows
		if s.isCommandAvailable("excel") {
			caps.Available = true
			caps.Version = s.getExcelVersionWindows()
			caps.ScriptTypes = []string{"vba", "powershell", "python", "com"}
		}
	case "linux":
		// Check for LibreOffice Calc
		if s.isCommandAvailable("libreoffice") || s.isCommandAvailable("soffice") {
			caps.Available = true
			caps.Version = s.getLibreOfficeVersion()
			caps.ScriptTypes = []string{"python", "basic", "bash"}
		}
	}

	return caps
}

func (s *OfficeService) detectWordCapabilities() WordCapabilities {
	caps := WordCapabilities{
		Operations: []string{
			"create_document", "open_document", "save_document",
			"insert_text", "replace_text", "format_text", "insert_image",
			"create_table", "insert_header", "insert_footer", "page_break",
			"apply_style", "create_toc", "mail_merge", "track_changes",
			"insert_hyperlink", "spell_check", "word_count",
		},
	}

	switch runtime.GOOS {
	case "darwin":
		if s.isApplicationAvailable("/Applications/Microsoft Word.app") {
			caps.Available = true
			caps.Version = s.getWordVersionMac()
			caps.ScriptTypes = []string{"applescript", "javascript", "python"}
		}
	case "windows":
		if s.isCommandAvailable("winword") {
			caps.Available = true
			caps.Version = s.getWordVersionWindows()
			caps.ScriptTypes = []string{"vba", "powershell", "python", "com"}
		}
	case "linux":
		if s.isCommandAvailable("libreoffice") || s.isCommandAvailable("soffice") {
			caps.Available = true
			caps.Version = s.getLibreOfficeVersion()
			caps.ScriptTypes = []string{"python", "basic", "bash"}
		}
	}

	return caps
}

func (s *OfficeService) detectPowerPointCapabilities() PowerPointCapabilities {
	caps := PowerPointCapabilities{
		Operations: []string{
			"create_presentation", "open_presentation", "save_presentation",
			"add_slide", "delete_slide", "duplicate_slide", "reorder_slides",
			"insert_text", "insert_image", "insert_chart", "insert_video",
			"apply_template", "apply_theme", "slide_transition", "animation",
			"export_pdf", "export_images", "speaker_notes",
		},
	}

	switch runtime.GOOS {
	case "darwin":
		if s.isApplicationAvailable("/Applications/Microsoft PowerPoint.app") {
			caps.Available = true
			caps.Version = s.getPowerPointVersionMac()
			caps.ScriptTypes = []string{"applescript", "javascript", "python"}
		}
	case "windows":
		if s.isCommandAvailable("powerpnt") {
			caps.Available = true
			caps.Version = s.getPowerPointVersionWindows()
			caps.ScriptTypes = []string{"vba", "powershell", "python", "com"}
		}
	case "linux":
		if s.isCommandAvailable("libreoffice") || s.isCommandAvailable("soffice") {
			caps.Available = true
			caps.Version = s.getLibreOfficeVersion()
			caps.ScriptTypes = []string{"python", "basic", "bash"}
		}
	}

	return caps
}

func (s *OfficeService) detectOutlookCapabilities() OutlookCapabilities {
	caps := OutlookCapabilities{
		Operations: []string{
			"send_email", "read_emails", "create_calendar_event", "schedule_meeting",
			"create_contact", "create_task", "send_meeting_request", "reply_email",
			"forward_email", "search_emails", "manage_folders", "auto_reply",
		},
	}

	switch runtime.GOOS {
	case "darwin":
		if s.isApplicationAvailable("/Applications/Microsoft Outlook.app") {
			caps.Available = true
			caps.Version = s.getOutlookVersionMac()
			caps.ScriptTypes = []string{"applescript", "javascript"}
		}
	case "windows":
		if s.isCommandAvailable("outlook") {
			caps.Available = true
			caps.Version = s.getOutlookVersionWindows()
			caps.ScriptTypes = []string{"vba", "powershell", "com"}
		}
	}

	return caps
}

func (s *OfficeService) detectPDFCapabilities() PDFCapabilities {
	caps := PDFCapabilities{
		Operations: []string{
			"merge_pdfs", "split_pdf", "extract_text", "extract_images",
			"compress_pdf", "convert_to_pdf", "add_watermark", "encrypt_pdf",
			"decrypt_pdf", "extract_pages", "rotate_pages", "add_annotations",
		},
	}

	tools := []string{}

	// Check for various PDF tools
	if s.isCommandAvailable("pdftk") {
		tools = append(tools, "pdftk")
	}
	if s.isCommandAvailable("gs") || s.isCommandAvailable("ghostscript") {
		tools = append(tools, "ghostscript")
	}
	if s.isCommandAvailable("qpdf") {
		tools = append(tools, "qpdf")
	}
	if s.isCommandAvailable("pdfunite") {
		tools = append(tools, "poppler-utils")
	}

	caps.Tools = tools
	caps.Available = len(tools) > 0

	return caps
}

func (s *OfficeService) detectSystemCapabilities() SystemCapabilities {
	caps := SystemCapabilities{
		Platform: runtime.GOOS,
	}

	// Detect available script languages
	languages := []string{}
	if s.isCommandAvailable("python") || s.isCommandAvailable("python3") {
		languages = append(languages, "python")
	}
	if s.isCommandAvailable("node") {
		languages = append(languages, "javascript")
	}
	if s.isCommandAvailable("powershell") || s.isCommandAvailable("pwsh") {
		languages = append(languages, "powershell")
	}
	if runtime.GOOS != "windows" {
		languages = append(languages, "bash")
	}
	if runtime.GOOS == "darwin" {
		languages = append(languages, "applescript")
	}

	caps.ScriptLanguages = languages

	// Detect automation capabilities
	automation := []string{}
	switch runtime.GOOS {
	case "darwin":
		automation = append(automation, "applescript", "automator", "shortcuts")
	case "windows":
		automation = append(automation, "task_scheduler", "com_automation", "powershell_remoting")
	case "linux":
		automation = append(automation, "cron", "systemd", "dbus")
	}

	caps.Automation = automation

	return caps
}

func (s *OfficeService) ProcessOfficeRequest(ctx context.Context, req OfficeRequest) (*OfficeResponse, error) {
	// Use AI to generate appropriate automation script
	aiReq := ScriptGenerationRequest{
		Description: req.Description,
		Platform:    runtime.GOOS,
		Language:    s.getBestScriptLanguage(req.Application),
	}

	// Enhance the description with office-specific context
	enhancedDesc := s.enhanceDescriptionForOffice(req)
	aiReq.Description = enhancedDesc

	// Generate script using AI
	aiResp, err := s.aiService.GenerateScript(ctx, aiReq)
	if err != nil {
		return &OfficeResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to generate office automation script: %v", err),
		}, err
	}

	// Save the script to a temporary file
	scriptPath, err := s.saveScript(aiResp.Script, aiReq.Language)
	if err != nil {
		return &OfficeResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to save script: %v", err),
		}, err
	}

	response := &OfficeResponse{
		Success:    true,
		Message:    aiResp.Explanation,
		ScriptPath: scriptPath,
		Commands:   s.parseCommands(aiResp.Script),
	}

	// Execute the script if requested
	if req.Parameters != nil {
		if execute, ok := req.Parameters["execute"].(bool); ok && execute {
			output, execErr := s.executeScript(ctx, scriptPath, aiReq.Language)
			response.Output = output
			if execErr != nil {
				response.Error = execErr.Error()
				response.Success = false
			}
		}
	}

	return response, nil
}

func (s *OfficeService) enhanceDescriptionForOffice(req OfficeRequest) string {
	enhanced := req.Description

	// Add context about the application
	if req.Application != "" {
		enhanced = fmt.Sprintf("[%s] %s", strings.ToUpper(req.Application), enhanced)
	}

	// Add file path context if provided
	if req.FilePath != "" {
		enhanced += fmt.Sprintf(" (File: %s)", req.FilePath)
	}

	// Add platform-specific hints
	switch runtime.GOOS {
	case "darwin":
		enhanced += " [Use AppleScript for Office automation on macOS]"
	case "windows":
		enhanced += " [Use COM automation or PowerShell for Office on Windows]"
	case "linux":
		enhanced += " [Use LibreOffice or command-line tools for document processing on Linux]"
	}

	// Add specific hints based on application
	switch strings.ToLower(req.Application) {
	case "excel":
		enhanced += " [Focus on spreadsheet operations like cell manipulation, formulas, charts, and data analysis]"
	case "word":
		enhanced += " [Focus on document operations like text formatting, tables, images, and templates]"
	case "powerpoint":
		enhanced += " [Focus on presentation operations like slides, animations, and media insertion]"
	case "outlook":
		enhanced += " [Focus on email and calendar operations]"
	case "pdf":
		enhanced += " [Focus on PDF manipulation like merging, splitting, and text extraction]"
	}

	return enhanced
}

func (s *OfficeService) getBestScriptLanguage(application string) string {
	switch runtime.GOOS {
	case "darwin":
		// On macOS, AppleScript is best for Office automation
		if application == "excel" || application == "word" || application == "powerpoint" || application == "outlook" {
			return "applescript"
		}
		return "python"
	case "windows":
		// On Windows, PowerShell or VBA/COM is best for Office
		if application == "excel" || application == "word" || application == "powerpoint" || application == "outlook" {
			return "powershell"
		}
		return "powershell"
	case "linux":
		// On Linux, Python is usually best
		return "python"
	default:
		return "python"
	}
}

func (s *OfficeService) saveScript(script, language string) (string, error) {
	// Create scripts directory if it doesn't exist
	scriptsDir := filepath.Join(os.TempDir(), "tashi_office_scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create scripts directory: %w", err)
	}

	// Determine file extension
	var ext string
	switch language {
	case "python":
		ext = ".py"
	case "applescript":
		ext = ".scpt"
	case "powershell":
		ext = ".ps1"
	case "javascript":
		ext = ".js"
	case "bash":
		ext = ".sh"
	default:
		ext = ".txt"
	}

	// Create unique filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("office_automation_%s%s", timestamp, ext)
	scriptPath := filepath.Join(scriptsDir, filename)

	// Write script to file
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return "", fmt.Errorf("failed to write script file: %w", err)
	}

	return scriptPath, nil
}

func (s *OfficeService) executeScript(ctx context.Context, scriptPath, language string) (string, error) {
	var cmd *exec.Cmd

	switch language {
	case "python":
		if s.isCommandAvailable("python3") {
			cmd = exec.CommandContext(ctx, "python3", scriptPath)
		} else {
			cmd = exec.CommandContext(ctx, "python", scriptPath)
		}
	case "applescript":
		cmd = exec.CommandContext(ctx, "osascript", scriptPath)
	case "powershell":
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
		} else {
			cmd = exec.CommandContext(ctx, "pwsh", "-File", scriptPath)
		}
	case "javascript":
		cmd = exec.CommandContext(ctx, "node", scriptPath)
	case "bash":
		cmd = exec.CommandContext(ctx, "bash", scriptPath)
	default:
		return "", fmt.Errorf("unsupported script language: %s", language)
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (s *OfficeService) parseCommands(script string) []string {
	// Parse the script into individual commands
	lines := strings.Split(script, "\n")
	var commands []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		commands = append(commands, line)
	}

	return commands
}

// Helper methods for version detection
func (s *OfficeService) isApplicationAvailable(appPath string) bool {
	_, err := os.Stat(appPath)
	return err == nil
}

func (s *OfficeService) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func (s *OfficeService) getExcelVersionMac() string {
	return s.getMacOfficeVersion("Excel")
}

func (s *OfficeService) getWordVersionMac() string {
	return s.getMacOfficeVersion("Word")
}

func (s *OfficeService) getPowerPointVersionMac() string {
	return s.getMacOfficeVersion("PowerPoint")
}

func (s *OfficeService) getOutlookVersionMac() string {
	return s.getMacOfficeVersion("Outlook")
}

func (s *OfficeService) getMacOfficeVersion(app string) string {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Microsoft %s" to get version`, app))
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return "Unknown"
}

func (s *OfficeService) getExcelVersionWindows() string {
	return s.getWindowsOfficeVersion("Excel.Application")
}

func (s *OfficeService) getWordVersionWindows() string {
	return s.getWindowsOfficeVersion("Word.Application")
}

func (s *OfficeService) getPowerPointVersionWindows() string {
	return s.getWindowsOfficeVersion("PowerPoint.Application")
}

func (s *OfficeService) getOutlookVersionWindows() string {
	return s.getWindowsOfficeVersion("Outlook.Application")
}

func (s *OfficeService) getWindowsOfficeVersion(progID string) string {
	script := fmt.Sprintf(`$app = New-Object -ComObject %s; $app.Version; $app.Quit()`, progID)
	cmd := exec.Command("powershell", "-Command", script)
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return "Unknown"
}

func (s *OfficeService) getLibreOfficeVersion() string {
	cmd := exec.Command("libreoffice", "--version")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return "Unknown"
}

// Cleanup old scripts
func (s *OfficeService) CleanupOldScripts() error {
	scriptsDir := filepath.Join(os.TempDir(), "tashi_office_scripts")

	// Remove scripts older than 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)

	return filepath.Walk(scriptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		if !info.IsDir() && info.ModTime().Before(cutoff) {
			os.Remove(path)
		}

		return nil
	})
}
