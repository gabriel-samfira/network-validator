package sysinfo

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// SystemInfo represents comprehensive system information
type SystemInfo struct {
	Hostname    string            `json:"hostname"`
	OS          OSInfo            `json:"os"`
	CPU         CPUInfo           `json:"cpu"`
	Memory      MemoryInfo        `json:"memory"`
	Network     NetworkInfo       `json:"network"`
	Uptime      UptimeInfo        `json:"uptime"`
	Timestamp   time.Time         `json:"timestamp"`
	Environment map[string]string `json:"environment,omitempty"`
}

// OSInfo contains operating system information
type OSInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	ID           string `json:"id"`
	IDLike       string `json:"id_like,omitempty"`
	PrettyName   string `json:"pretty_name"`
	VersionID    string `json:"version_id"`
	HomeURL      string `json:"home_url,omitempty"`
	BugReportURL string `json:"bug_report_url,omitempty"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
}

// CPUInfo contains CPU information
type CPUInfo struct {
	Model     string   `json:"model"`
	Vendor    string   `json:"vendor"`
	Cores     int      `json:"cores"`
	Threads   int      `json:"threads"`
	MHz       float64  `json:"mhz"`
	CacheSize string   `json:"cache_size,omitempty"`
	Flags     []string `json:"flags,omitempty"`
}

// MemoryInfo contains memory information
type MemoryInfo struct {
	TotalGB        float64 `json:"total_gb"`
	TotalBytes     uint64  `json:"total_bytes"`
	AvailableGB    float64 `json:"available_gb"`
	AvailableBytes uint64  `json:"available_bytes"`
	UsedGB         float64 `json:"used_gb"`
	UsedBytes      uint64  `json:"used_bytes"`
	FreeGB         float64 `json:"free_gb"`
	FreeBytes      uint64  `json:"free_bytes"`
	UsedPercent    float64 `json:"used_percent"`
}

// NetworkInfo contains network information
type NetworkInfo struct {
	Interfaces []InterfaceInfo `json:"interfaces"`
	Hostname   string          `json:"hostname"`
}

// InterfaceInfo represents a network interface
type InterfaceInfo struct {
	Name        string   `json:"name"`
	IPAddresses []string `json:"ip_addresses"`
	MACAddress  string   `json:"mac_address,omitempty"`
	MTU         int      `json:"mtu"`
	IsUp        bool     `json:"is_up"`
	IsLoopback  bool     `json:"is_loopback"`
	IsMulticast bool     `json:"is_multicast"`
	IsBroadcast bool     `json:"is_broadcast"`
}

// UptimeInfo contains system uptime information
type UptimeInfo struct {
	Seconds  float64   `json:"seconds"`
	Days     int       `json:"days"`
	Hours    int       `json:"hours"`
	Minutes  int       `json:"minutes"`
	Uptime   string    `json:"uptime_string"`
	BootTime time.Time `json:"boot_time"`
}

// GetSystemInfo gathers comprehensive system information
func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		Timestamp: time.Now(),
	}

	var err error

	// Get hostname
	info.Hostname, err = os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Get OS information
	info.OS, err = getOSInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get OS info: %w", err)
	}

	// Get CPU information
	info.CPU, err = getCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	// Get memory information
	info.Memory, err = getMemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	// Get network information
	info.Network, err = getNetworkInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %w", err)
	}

	// Get uptime information
	info.Uptime, err = getUptimeInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get uptime info: %w", err)
	}

	return info, nil
}

// GetHostname returns just the hostname of the system
func GetHostname() (string, error) {
	return os.Hostname()
}

// GetMainIPAddress gets the source IP used to reach the default gateway in table 254
// This is typically the primary IP address of the server
func GetMainIPAddress() (string, error) {
	// Try to read from route table 254
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return "", fmt.Errorf("failed to open route table: %w", err)
	}
	defer file.Close()

	// Find default gateway route
	scanner := bufio.NewScanner(file)
	scanner.Scan() // Skip header

	var defaultIface string
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 8 {
			continue
		}
		// fields[0] = interface, fields[1] = destination, fields[2] = gateway
		// Destination 00000000 means default route
		if fields[1] == "00000000" {
			defaultIface = fields[0]
			break
		}
	}

	if defaultIface == "" {
		return "", fmt.Errorf("no default route found")
	}

	// Get IP address for this interface
	iface, err := net.InterfaceByName(defaultIface)
	if err != nil {
		return "", fmt.Errorf("failed to get interface %s: %w", defaultIface, err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get addresses for %s: %w", defaultIface, err)
	}

	// Return the first IPv4 address
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no IPv4 address found on interface %s", defaultIface)
}

// getOSInfo reads OS information from /etc/os-release and other sources
func getOSInfo() (OSInfo, error) {
	osInfo := OSInfo{}

	// Read /etc/os-release
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return osInfo, fmt.Errorf("failed to open /etc/os-release: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			value := strings.Trim(parts[1], `"`)

			switch key {
			case "NAME":
				osInfo.Name = value
			case "VERSION":
				osInfo.Version = value
			case "ID":
				osInfo.ID = value
			case "ID_LIKE":
				osInfo.IDLike = value
			case "PRETTY_NAME":
				osInfo.PrettyName = value
			case "VERSION_ID":
				osInfo.VersionID = value
			case "HOME_URL":
				osInfo.HomeURL = value
			case "BUG_REPORT_URL":
				osInfo.BugReportURL = value
			}
		}
	}

	// Get kernel version
	if data, err := os.ReadFile("/proc/version"); err == nil {
		osInfo.Kernel = strings.Fields(string(data))[2]
	}

	// Get architecture
	if data, err := os.ReadFile("/proc/sys/kernel/arch"); err == nil {
		osInfo.Architecture = strings.TrimSpace(string(data))
	} else {
		// Fallback: try uname -m approach by parsing /proc/version
		if strings.Contains(osInfo.Kernel, "x86_64") {
			osInfo.Architecture = "x86_64"
		} else if strings.Contains(osInfo.Kernel, "aarch64") {
			osInfo.Architecture = "aarch64"
		}
	}

	return osInfo, scanner.Err()
}

// getCPUInfo reads CPU information from /proc/cpuinfo
func getCPUInfo() (CPUInfo, error) {
	cpuInfo := CPUInfo{}

	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return cpuInfo, fmt.Errorf("failed to open /proc/cpuinfo: %w", err)
	}
	defer file.Close()

	cores := make(map[int]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "model name":
				if cpuInfo.Model == "" {
					cpuInfo.Model = value
				}
			case "vendor_id":
				if cpuInfo.Vendor == "" {
					cpuInfo.Vendor = value
				}
			case "core id":
				if coreID, err := strconv.Atoi(value); err == nil {
					cores[coreID] = true
				}
			case "siblings":
				if threads, err := strconv.Atoi(value); err == nil {
					cpuInfo.Threads = threads
				}
			case "cpu MHz":
				if mhz, err := strconv.ParseFloat(value, 64); err == nil {
					cpuInfo.MHz = mhz
				}
			case "cache size":
				cpuInfo.CacheSize = value
			case "flags":
				if len(cpuInfo.Flags) == 0 {
					cpuInfo.Flags = strings.Fields(value)
				}
			}
		}
	}

	cpuInfo.Cores = len(cores)
	if cpuInfo.Cores == 0 {
		cpuInfo.Cores = 1 // Fallback
	}

	return cpuInfo, scanner.Err()
}

// getMemoryInfo reads memory information from /proc/meminfo
func getMemoryInfo() (MemoryInfo, error) {
	memInfo := MemoryInfo{}

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return memInfo, fmt.Errorf("failed to open /proc/meminfo: %w", err)
	}
	defer file.Close()

	memData := make(map[string]uint64)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ":") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				key := strings.TrimSuffix(parts[0], ":")
				if value, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					// Convert from kB to bytes
					memData[key] = value * 1024
				}
			}
		}
	}

	// Calculate memory values
	memInfo.TotalBytes = memData["MemTotal"]
	memInfo.TotalGB = float64(memInfo.TotalBytes) / (1024 * 1024 * 1024)

	memInfo.FreeBytes = memData["MemFree"]
	memInfo.FreeGB = float64(memInfo.FreeBytes) / (1024 * 1024 * 1024)

	memInfo.AvailableBytes = memData["MemAvailable"]
	if memInfo.AvailableBytes == 0 {
		memInfo.AvailableBytes = memInfo.FreeBytes + memData["Buffers"] + memData["Cached"]
	}
	memInfo.AvailableGB = float64(memInfo.AvailableBytes) / (1024 * 1024 * 1024)

	memInfo.UsedBytes = memInfo.TotalBytes - memInfo.AvailableBytes
	memInfo.UsedGB = float64(memInfo.UsedBytes) / (1024 * 1024 * 1024)

	if memInfo.TotalBytes > 0 {
		memInfo.UsedPercent = (float64(memInfo.UsedBytes) / float64(memInfo.TotalBytes)) * 100
	}

	return memInfo, scanner.Err()
}

// getNetworkInfo gathers network interface information
func getNetworkInfo() (NetworkInfo, error) {
	netInfo := NetworkInfo{}

	// Get hostname
	hostname, _ := os.Hostname()
	netInfo.Hostname = hostname

	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return netInfo, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		ifaceInfo := InterfaceInfo{
			Name:        iface.Name,
			MTU:         iface.MTU,
			IsUp:        iface.Flags&net.FlagUp != 0,
			IsLoopback:  iface.Flags&net.FlagLoopback != 0,
			IsMulticast: iface.Flags&net.FlagMulticast != 0,
			IsBroadcast: iface.Flags&net.FlagBroadcast != 0,
		}

		// Get MAC address
		if iface.HardwareAddr != nil {
			ifaceInfo.MACAddress = iface.HardwareAddr.String()
		}

		// Get IP addresses
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok {
					ifaceInfo.IPAddresses = append(ifaceInfo.IPAddresses, ipNet.IP.String())
				}
			}
		}

		netInfo.Interfaces = append(netInfo.Interfaces, ifaceInfo)
	}

	return netInfo, nil
}

// getUptimeInfo reads system uptime from /proc/uptime
func getUptimeInfo() (UptimeInfo, error) {
	uptimeInfo := UptimeInfo{}

	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return uptimeInfo, fmt.Errorf("failed to read /proc/uptime: %w", err)
	}

	parts := strings.Fields(string(data))
	if len(parts) < 1 {
		return uptimeInfo, fmt.Errorf("invalid uptime format")
	}

	uptimeSeconds, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return uptimeInfo, fmt.Errorf("failed to parse uptime: %w", err)
	}

	uptimeInfo.Seconds = uptimeSeconds

	// Calculate human-readable uptime
	totalMinutes := int(uptimeSeconds) / 60
	uptimeInfo.Days = totalMinutes / (24 * 60)
	uptimeInfo.Hours = (totalMinutes % (24 * 60)) / 60
	uptimeInfo.Minutes = totalMinutes % 60

	// Format uptime string
	if uptimeInfo.Days > 0 {
		uptimeInfo.Uptime = fmt.Sprintf("%dd %dh %dm", uptimeInfo.Days, uptimeInfo.Hours, uptimeInfo.Minutes)
	} else if uptimeInfo.Hours > 0 {
		uptimeInfo.Uptime = fmt.Sprintf("%dh %dm", uptimeInfo.Hours, uptimeInfo.Minutes)
	} else {
		uptimeInfo.Uptime = fmt.Sprintf("%dm", uptimeInfo.Minutes)
	}

	// Calculate boot time
	uptimeInfo.BootTime = time.Now().Add(-time.Duration(uptimeSeconds) * time.Second)

	return uptimeInfo, nil
}

// GetOSRelease returns a map of key-value pairs from /etc/os-release
func GetOSRelease() (map[string]string, error) {
	result := make(map[string]string)

	file, err := os.Open("/etc/os-release")
	if err != nil {
		return result, fmt.Errorf("failed to open /etc/os-release: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			value := strings.Trim(parts[1], `"`)
			result[key] = value
		}
	}

	return result, scanner.Err()
}

// GetProcessorInfo returns detailed processor information
func GetProcessorInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Read /proc/cpuinfo
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return result, fmt.Errorf("failed to open /proc/cpuinfo: %w", err)
	}
	defer file.Close()

	processors := []map[string]string{}
	currentProcessor := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			if len(currentProcessor) > 0 {
				processors = append(processors, currentProcessor)
				currentProcessor = make(map[string]string)
			}
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			currentProcessor[key] = value
		}
	}

	// Add the last processor if exists
	if len(currentProcessor) > 0 {
		processors = append(processors, currentProcessor)
	}

	result["processors"] = processors
	result["processor_count"] = len(processors)

	return result, scanner.Err()
}
