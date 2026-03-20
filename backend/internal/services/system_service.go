package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type SystemService struct {
	systemInfo *SystemInfo
}

type SystemInfo struct {
	Hardware    HardwareInfo    `json:"hardware"`
	OS          OSInfo          `json:"os"`
	Network     NetworkInfo     `json:"network"`
	User        UserInfo        `json:"user"`
	Location    LocationInfo    `json:"location"`
	Environment EnvironmentInfo `json:"environment"`
	Timestamp   time.Time       `json:"timestamp"`
}

type HardwareInfo struct {
	CPU      CPUInfo       `json:"cpu"`
	Memory   MemoryInfo    `json:"memory"`
	Storage  []StorageInfo `json:"storage"`
	Platform string        `json:"platform"`
	Arch     string        `json:"architecture"`
}

type CPUInfo struct {
	Model     string  `json:"model"`
	Cores     int     `json:"cores"`
	Threads   int     `json:"threads"`
	Frequency float64 `json:"frequency_mhz"`
	Usage     float64 `json:"usage_percent"`
}

type MemoryInfo struct {
	Total        uint64  `json:"total_bytes"`
	Available    uint64  `json:"available_bytes"`
	Used         uint64  `json:"used_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

type StorageInfo struct {
	Device       string  `json:"device"`
	Mountpoint   string  `json:"mountpoint"`
	Filesystem   string  `json:"filesystem"`
	Total        uint64  `json:"total_bytes"`
	Free         uint64  `json:"free_bytes"`
	Used         uint64  `json:"used_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

type OSInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
	Hostname     string `json:"hostname"`
	Uptime       uint64 `json:"uptime_seconds"`
	ProcessCount uint64 `json:"process_count"`
}

type NetworkInfo struct {
	Hostname   string          `json:"hostname"`
	Interfaces []InterfaceInfo `json:"interfaces"`
	ExternalIP string          `json:"external_ip"`
	LocalIPs   []string        `json:"local_ips"`
	DNSServers []string        `json:"dns_servers"`
}

type InterfaceInfo struct {
	Name  string   `json:"name"`
	MAC   string   `json:"mac"`
	IPs   []string `json:"ips"`
	IsUp  bool     `json:"is_up"`
	Speed uint64   `json:"speed_mbps"`
}

type UserInfo struct {
	Username   string   `json:"username"`
	HomeDir    string   `json:"home_directory"`
	CurrentDir string   `json:"current_directory"`
	UID        string   `json:"uid"`
	GID        string   `json:"gid"`
	Shell      string   `json:"shell"`
	Groups     []string `json:"groups"`
}

type LocationInfo struct {
	Timezone  string  `json:"timezone"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	ISP       string  `json:"isp,omitempty"`
}

type EnvironmentInfo struct {
	Variables map[string]string `json:"variables"`
	PATH      []string          `json:"path"`
	Languages []string          `json:"languages"`
	Locale    string            `json:"locale"`
}

func NewSystemService() *SystemService {
	return &SystemService{}
}

func (s *SystemService) GatherSystemInfo(ctx context.Context) (*SystemInfo, error) {
	info := &SystemInfo{
		Timestamp: time.Now(),
	}

	// Gather hardware info
	hwInfo, err := s.gatherHardwareInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to gather hardware info: %w", err)
	}
	info.Hardware = *hwInfo

	// Gather OS info
	osInfo, err := s.gatherOSInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to gather OS info: %w", err)
	}
	info.OS = *osInfo

	// Gather network info
	netInfo, err := s.gatherNetworkInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to gather network info: %w", err)
	}
	info.Network = *netInfo

	// Gather user info
	userInfo, err := s.gatherUserInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to gather user info: %w", err)
	}
	info.User = *userInfo

	// Gather location info
	locInfo, err := s.gatherLocationInfo(ctx)
	if err != nil {
		// Location info is optional, log but don't fail
		fmt.Printf("Warning: failed to gather location info: %v\n", err)
		locInfo = &LocationInfo{}
	}
	info.Location = *locInfo

	// Gather environment info
	envInfo, err := s.gatherEnvironmentInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to gather environment info: %w", err)
	}
	info.Environment = *envInfo

	s.systemInfo = info
	return info, nil
}

func (s *SystemService) gatherHardwareInfo(ctx context.Context) (*HardwareInfo, error) {
	hwInfo := &HardwareInfo{
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
	}

	// Basic CPU info
	hwInfo.CPU.Threads = runtime.NumCPU()
	hwInfo.CPU.Cores = runtime.NumCPU() // Simplified assumption
	hwInfo.CPU.Model = s.getCPUModel()

	// Basic memory info (simplified)
	hwInfo.Memory = s.getMemoryInfo()

	// Basic storage info
	hwInfo.Storage = s.getStorageInfo()

	return hwInfo, nil
}

func (s *SystemService) getCPUModel() string {
	switch runtime.GOOS {
	case "darwin":
		if output, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	case "linux":
		if output, err := exec.Command("cat", "/proc/cpuinfo").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "model name") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}
	case "windows":
		if output, err := exec.Command("wmic", "cpu", "get", "name", "/value").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Name=") {
					return strings.TrimSpace(strings.TrimPrefix(line, "Name="))
				}
			}
		}
	}
	return "Unknown CPU"
}

func (s *SystemService) getMemoryInfo() MemoryInfo {
	memInfo := MemoryInfo{}

	switch runtime.GOOS {
	case "darwin":
		// Get total memory
		if output, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
			if total, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64); err == nil {
				memInfo.Total = total
			}
		}
		// Get memory pressure (simplified)
		if output, err := exec.Command("vm_stat").Output(); err == nil {
			// Parse vm_stat output for available memory estimation
			lines := strings.Split(string(output), "\n")
			var free, inactive uint64
			for _, line := range lines {
				if strings.Contains(line, "Pages free:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						if val, err := strconv.ParseUint(strings.TrimSuffix(parts[2], "."), 10, 64); err == nil {
							free = val * 4096 // Page size
						}
					}
				}
				if strings.Contains(line, "Pages inactive:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						if val, err := strconv.ParseUint(strings.TrimSuffix(parts[2], "."), 10, 64); err == nil {
							inactive = val * 4096 // Page size
						}
					}
				}
			}
			memInfo.Available = free + inactive
			memInfo.Used = memInfo.Total - memInfo.Available
			if memInfo.Total > 0 {
				memInfo.UsagePercent = float64(memInfo.Used) / float64(memInfo.Total) * 100
			}
		}
	case "linux":
		if output, err := exec.Command("cat", "/proc/meminfo").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					key := strings.TrimSuffix(fields[0], ":")
					if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						val *= 1024 // Convert KB to bytes
						switch key {
						case "MemTotal":
							memInfo.Total = val
						case "MemAvailable":
							memInfo.Available = val
						}
					}
				}
			}
			memInfo.Used = memInfo.Total - memInfo.Available
			if memInfo.Total > 0 {
				memInfo.UsagePercent = float64(memInfo.Used) / float64(memInfo.Total) * 100
			}
		}
	case "windows":
		if output, err := exec.Command("wmic", "OS", "get", "TotalVisibleMemorySize,FreePhysicalMemory", "/value").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "TotalVisibleMemorySize=") {
					if val, err := strconv.ParseUint(strings.TrimPrefix(line, "TotalVisibleMemorySize="), 10, 64); err == nil {
						memInfo.Total = val * 1024 // Convert KB to bytes
					}
				} else if strings.HasPrefix(line, "FreePhysicalMemory=") {
					if val, err := strconv.ParseUint(strings.TrimPrefix(line, "FreePhysicalMemory="), 10, 64); err == nil {
						memInfo.Available = val * 1024 // Convert KB to bytes
					}
				}
			}
			memInfo.Used = memInfo.Total - memInfo.Available
			if memInfo.Total > 0 {
				memInfo.UsagePercent = float64(memInfo.Used) / float64(memInfo.Total) * 100
			}
		}
	}

	return memInfo
}

func (s *SystemService) getStorageInfo() []StorageInfo {
	var storageInfos []StorageInfo

	switch runtime.GOOS {
	case "darwin":
		if output, err := exec.Command("df", "-k").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for i, line := range lines {
				if i == 0 { // Skip header
					continue
				}
				fields := strings.Fields(line)
				if len(fields) >= 6 {
					if total, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						if used, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
							if available, err := strconv.ParseUint(fields[3], 10, 64); err == nil {
								storage := StorageInfo{
									Device:       fields[0],
									Mountpoint:   fields[5],
									Total:        total * 1024, // Convert KB to bytes
									Used:         used * 1024,
									Free:         available * 1024,
									UsagePercent: float64(used) / float64(total) * 100,
								}
								storageInfos = append(storageInfos, storage)
							}
						}
					}
				}
			}
		}
	case "linux":
		if output, err := exec.Command("df", "-k").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for i, line := range lines {
				if i == 0 { // Skip header
					continue
				}
				fields := strings.Fields(line)
				if len(fields) >= 6 {
					if total, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						if used, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
							if available, err := strconv.ParseUint(fields[3], 10, 64); err == nil {
								storage := StorageInfo{
									Device:       fields[0],
									Mountpoint:   fields[5],
									Total:        total * 1024, // Convert KB to bytes
									Used:         used * 1024,
									Free:         available * 1024,
									UsagePercent: float64(used) / float64(total) * 100,
								}
								storageInfos = append(storageInfos, storage)
							}
						}
					}
				}
			}
		}
	case "windows":
		if output, err := exec.Command("wmic", "logicaldisk", "get", "size,freespace,caption", "/value").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			var caption, size, freespace string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Caption=") {
					caption = strings.TrimPrefix(line, "Caption=")
				} else if strings.HasPrefix(line, "FreeSpace=") {
					freespace = strings.TrimPrefix(line, "FreeSpace=")
				} else if strings.HasPrefix(line, "Size=") {
					size = strings.TrimPrefix(line, "Size=")
					// Process when we have all three values
					if caption != "" && size != "" && freespace != "" {
						if totalBytes, err := strconv.ParseUint(size, 10, 64); err == nil {
							if freeBytes, err := strconv.ParseUint(freespace, 10, 64); err == nil {
								usedBytes := totalBytes - freeBytes
								storage := StorageInfo{
									Device:       caption,
									Mountpoint:   caption,
									Total:        totalBytes,
									Used:         usedBytes,
									Free:         freeBytes,
									UsagePercent: float64(usedBytes) / float64(totalBytes) * 100,
								}
								storageInfos = append(storageInfos, storage)
							}
						}
						// Reset for next disk
						caption, size, freespace = "", "", ""
					}
				}
			}
		}
	}

	return storageInfos
}

func (s *SystemService) gatherOSInfo(ctx context.Context) (*OSInfo, error) {
	osInfo := &OSInfo{
		Name:         runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		osInfo.Hostname = hostname
	}

	// Get OS version
	switch runtime.GOOS {
	case "darwin":
		if output, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			osInfo.Version = strings.TrimSpace(string(output))
		}
		if output, err := exec.Command("uname", "-r").Output(); err == nil {
			osInfo.Kernel = strings.TrimSpace(string(output))
		}
	case "linux":
		if output, err := exec.Command("uname", "-r").Output(); err == nil {
			osInfo.Kernel = strings.TrimSpace(string(output))
		}
		// Try to get distribution info
		if output, err := exec.Command("lsb_release", "-d").Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Description:") {
					osInfo.Version = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
				}
			}
		}
	case "windows":
		if output, err := exec.Command("ver").Output(); err == nil {
			osInfo.Version = strings.TrimSpace(string(output))
		}
	}

	return osInfo, nil
}

func (s *SystemService) gatherNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	netInfo := &NetworkInfo{}

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		netInfo.Hostname = hostname
	}

	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			ifaceInfo := InterfaceInfo{
				Name: iface.Name,
				IsUp: iface.Flags&net.FlagUp != 0,
			}

			// Get MAC address
			if iface.HardwareAddr != nil {
				ifaceInfo.MAC = iface.HardwareAddr.String()
			}

			// Get IP addresses
			addrs, err := iface.Addrs()
			if err == nil {
				for _, addr := range addrs {
					ifaceInfo.IPs = append(ifaceInfo.IPs, addr.String())

					// Add to local IPs list
					if ipNet, ok := addr.(*net.IPNet); ok {
						ip := ipNet.IP
						if !ip.IsLoopback() && !ip.IsLinkLocalMulticast() && !ip.IsLinkLocalUnicast() {
							netInfo.LocalIPs = append(netInfo.LocalIPs, ip.String())
						}
					}
				}
			}

			netInfo.Interfaces = append(netInfo.Interfaces, ifaceInfo)
		}
	}

	// Try to get external IP
	netInfo.ExternalIP = s.getExternalIP()

	return netInfo, nil
}

func (s *SystemService) gatherUserInfo(ctx context.Context) (*UserInfo, error) {
	userInfo := &UserInfo{}

	if currentUser, err := user.Current(); err == nil {
		userInfo.Username = currentUser.Username
		userInfo.HomeDir = currentUser.HomeDir
		userInfo.UID = currentUser.Uid
		userInfo.GID = currentUser.Gid
	}

	if wd, err := os.Getwd(); err == nil {
		userInfo.CurrentDir = wd
	}

	// Get shell
	if shell := os.Getenv("SHELL"); shell != "" {
		userInfo.Shell = shell
	}

	// Get groups (Unix-like systems)
	if runtime.GOOS != "windows" {
		if currentUser, err := user.Current(); err == nil {
			if groupIds, err := currentUser.GroupIds(); err == nil {
				for _, gid := range groupIds {
					if group, err := user.LookupGroupId(gid); err == nil {
						userInfo.Groups = append(userInfo.Groups, group.Name)
					}
				}
			}
		}
	}

	return userInfo, nil
}

func (s *SystemService) gatherLocationInfo(ctx context.Context) (*LocationInfo, error) {
	locInfo := &LocationInfo{}

	// Get timezone
	tz, _ := time.Now().Zone()
	locInfo.Timezone = tz

	// Note: For privacy and security, we'll only get basic timezone info
	// Advanced location services would require user permission and external APIs

	return locInfo, nil
}

func (s *SystemService) gatherEnvironmentInfo(ctx context.Context) (*EnvironmentInfo, error) {
	envInfo := &EnvironmentInfo{
		Variables: make(map[string]string),
	}

	// Get all environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envInfo.Variables[parts[0]] = parts[1]
		}
	}

	// Parse PATH
	if pathVar := os.Getenv("PATH"); pathVar != "" {
		separator := ":"
		if runtime.GOOS == "windows" {
			separator = ";"
		}
		envInfo.PATH = strings.Split(pathVar, separator)
	}

	// Get locale
	if locale := os.Getenv("LANG"); locale != "" {
		envInfo.Locale = locale
	} else if locale := os.Getenv("LC_ALL"); locale != "" {
		envInfo.Locale = locale
	}

	// Detect available languages/tools
	envInfo.Languages = s.detectAvailableLanguages()

	return envInfo, nil
}

func (s *SystemService) detectAvailableLanguages() []string {
	var languages []string

	// Common programming languages and tools to check
	tools := []string{
		"python", "python3", "node", "go", "java", "git", "docker",
		"npm", "yarn", "ruby", "php", "rust", "swift", "bash", "zsh",
		"powershell", "excel", "word", "powerpoint",
	}

	for _, tool := range tools {
		if s.isCommandAvailable(tool) {
			languages = append(languages, tool)
		}
	}

	return languages
}

func (s *SystemService) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func (s *SystemService) getExternalIP() string {
	// Try multiple services for reliability
	services := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, service := range services {
		if resp, err := client.Get(service); err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				if body, err := io.ReadAll(resp.Body); err == nil {
					ip := strings.TrimSpace(string(body))
					if net.ParseIP(ip) != nil {
						return ip
					}
				}
			}
		}
	}

	return ""
}

func (s *SystemService) GetCachedSystemInfo() *SystemInfo {
	return s.systemInfo
}

func (s *SystemService) GetOptimalSettings() map[string]interface{} {
	if s.systemInfo == nil {
		return map[string]interface{}{}
	}

	settings := make(map[string]interface{})

	// CPU-based recommendations
	if s.systemInfo.Hardware.CPU.Cores > 8 {
		settings["max_concurrent_operations"] = 8
		settings["parallel_processing"] = true
	} else if s.systemInfo.Hardware.CPU.Cores > 4 {
		settings["max_concurrent_operations"] = 4
		settings["parallel_processing"] = true
	} else {
		settings["max_concurrent_operations"] = 2
		settings["parallel_processing"] = false
	}

	// Memory-based recommendations
	memoryGB := float64(s.systemInfo.Hardware.Memory.Total) / (1024 * 1024 * 1024)
	if memoryGB > 16 {
		settings["memory_intensive_operations"] = true
		settings["max_file_size_mb"] = 1000
	} else if memoryGB > 8 {
		settings["memory_intensive_operations"] = true
		settings["max_file_size_mb"] = 500
	} else {
		settings["memory_intensive_operations"] = false
		settings["max_file_size_mb"] = 100
	}

	// Storage-based recommendations
	for _, storage := range s.systemInfo.Hardware.Storage {
		if storage.Mountpoint == s.systemInfo.User.HomeDir ||
			storage.Mountpoint == "/" ||
			storage.Mountpoint == "C:\\" {
			if storage.UsagePercent > 90 {
				settings["storage_warning"] = "Low disk space"
				settings["cleanup_recommended"] = true
			}
			break
		}
	}

	// Platform-specific settings
	switch s.systemInfo.Hardware.Platform {
	case "darwin":
		settings["platform_features"] = []string{"spotlight_integration", "quicklook_support", "applescript_support"}
		if s.systemInfo.Hardware.Arch == "arm64" {
			settings["apple_silicon"] = true
			settings["native_performance"] = true
		}
	case "windows":
		settings["platform_features"] = []string{"explorer_integration", "powershell_support", "registry_access"}
	case "linux":
		settings["platform_features"] = []string{"desktop_integration", "package_manager_support", "shell_integration"}
	}

	return settings
}

func (s *SystemService) SaveSystemInfo(filePath string) error {
	if s.systemInfo == nil {
		return fmt.Errorf("no system info to save")
	}

	data, err := json.MarshalIndent(s.systemInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal system info: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write system info file: %w", err)
	}

	return nil
}

func (s *SystemService) LoadSystemInfo(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read system info file: %w", err)
	}

	var info SystemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("failed to unmarshal system info: %w", err)
	}

	s.systemInfo = &info
	return nil
}
