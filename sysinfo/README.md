# System Information Web Server

A comprehensive Go package that provides system information through a REST API with a web dashboard.

## Features

### 🖥️ System Information Collection
- **Operating System**: OS name, version, kernel, architecture from `/etc/os-release`
- **CPU Information**: Model, vendor, cores, threads, frequency, cache, flags from `/proc/cpuinfo`
- **Memory Details**: Total, available, used, free memory with percentages from `/proc/meminfo`
- **Network Interfaces**: IP addresses, MAC addresses, MTU, interface status
- **System Uptime**: Boot time, uptime in days/hours/minutes from `/proc/uptime`
- **Hostname**: System hostname

### 🌐 Web Server
- **REST API**: Multiple endpoints for different information types
- **HTML Dashboard**: Interactive web interface with auto-refresh
- **JSON Responses**: Well-structured JSON data for all endpoints
- **CORS Support**: Cross-origin requests enabled
- **Request Logging**: HTTP request logging with timing
- **Health Check**: Service health monitoring endpoint

## Package Structure

```
sysinfo/
├── info.go          # System information collection
├── server.go        # Web server and HTTP handlers
└── sysinfo_test.go  # Comprehensive test suite
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /` | HTML dashboard with live system information |
| `GET /api/sysinfo` | Complete system information (JSON) |
| `GET /api/sysinfo/os` | Operating system information only |
| `GET /api/sysinfo/cpu` | CPU information only |
| `GET /api/sysinfo/memory` | Memory information only |
| `GET /api/sysinfo/network` | Network interfaces only |
| `GET /api/sysinfo/uptime` | System uptime only |
| `GET /api/health` | Health check endpoint |

## Usage

### Command Line Application

```bash
# Show system information and exit
go run cmd/sysinfo_app.go -info

# Start web server on default port (8080)
go run cmd/sysinfo_app.go

# Start web server on custom port
go run cmd/sysinfo_app.go -port 3000

# Show help
go run cmd/sysinfo_app.go -help
```

### As a Package

```go
package main

import (
    "fmt"
    "log"
    "validate/sysinfo"
)

func main() {
    // Get complete system information
    info, err := sysinfo.GetSystemInfo()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Hostname: %s\n", info.Hostname)
    fmt.Printf("OS: %s %s\n", info.OS.Name, info.OS.Version)
    fmt.Printf("CPU: %s (%d cores)\n", info.CPU.Model, info.CPU.Cores)
    fmt.Printf("Memory: %.2f GB total\n", info.Memory.TotalGB)

    // Start web server
    log.Fatal(sysinfo.RunServer(8080))
}
```

### Individual Information Functions

```go
// Get specific information types
osInfo, err := sysinfo.GetOSRelease()
cpuInfo, err := getCPUInfo()
memInfo, err := getMemoryInfo()
netInfo, err := getNetworkInfo()
uptimeInfo, err := getUptimeInfo()
```

## Data Structures

### SystemInfo
Complete system information structure containing all subsystems.

### OSInfo
```go
type OSInfo struct {
    Name         string `json:"name"`
    Version      string `json:"version"`
    ID           string `json:"id"`
    PrettyName   string `json:"pretty_name"`
    Kernel       string `json:"kernel"`
    Architecture string `json:"architecture"`
    // ... additional fields
}
```

### CPUInfo
```go
type CPUInfo struct {
    Model     string  `json:"model"`
    Vendor    string  `json:"vendor"`
    Cores     int     `json:"cores"`
    Threads   int     `json:"threads"`
    MHz       float64 `json:"mhz"`
    CacheSize string  `json:"cache_size"`
    Flags     []string `json:"flags"`
}
```

### MemoryInfo
```go
type MemoryInfo struct {
    TotalGB      float64 `json:"total_gb"`
    TotalBytes   uint64  `json:"total_bytes"`
    AvailableGB  float64 `json:"available_gb"`
    UsedGB       float64 `json:"used_gb"`
    FreeGB       float64 `json:"free_gb"`
    UsedPercent  float64 `json:"used_percent"`
}
```

### NetworkInfo
```go
type NetworkInfo struct {
    Interfaces []InterfaceInfo `json:"interfaces"`
    Hostname   string          `json:"hostname"`
}

type InterfaceInfo struct {
    Name         string   `json:"name"`
    IPAddresses  []string `json:"ip_addresses"`
    MACAddress   string   `json:"mac_address"`
    MTU          int      `json:"mtu"`
    IsUp         bool     `json:"is_up"`
    IsLoopback   bool     `json:"is_loopback"`
    // ... additional fields
}
```

## Example API Responses

### Complete System Information
```json
{
  "hostname": "myserver",
  "os": {
    "name": "Ubuntu",
    "version": "24.04.3 LTS",
    "pretty_name": "Ubuntu 24.04.3 LTS",
    "kernel": "6.14.0-33-generic",
    "architecture": "x86_64"
  },
  "cpu": {
    "model": "AMD Ryzen 7 5700G with Radeon Graphics",
    "vendor": "AuthenticAMD",
    "cores": 8,
    "threads": 16,
    "mhz": 2519.724
  },
  "memory": {
    "total_gb": 121.73,
    "available_gb": 62.77,
    "used_percent": 48.44
  },
  "network": {
    "interfaces": [
      {
        "name": "enp3s0",
        "ip_addresses": ["192.168.1.140"],
        "mac_address": "a8:a1:59:af:67:5f",
        "mtu": 1500,
        "is_up": true
      }
    ]
  },
  "uptime": {
    "uptime_string": "13d 0h 3m",
    "boot_time": "2025-10-10T09:33:15Z"
  },
  "timestamp": "2025-10-23T09:36:30Z"
}
```

## Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./sysinfo -v

# Run benchmarks
go test ./sysinfo -bench=.

# Test API endpoints
bash test_api.sh
```

## Requirements

- **Linux System**: Designed for Linux systems with `/proc` filesystem
- **Root Access**: Some information requires root privileges for complete data
- **Go 1.21+**: Modern Go version for best compatibility

## Security Considerations

- **Network Binding**: Server binds to all interfaces (`0.0.0.0`)
- **CORS Enabled**: Cross-origin requests are allowed
- **No Authentication**: API endpoints are publicly accessible
- **System Information**: Exposes detailed system information

For production use, consider:
- Adding authentication/authorization
- Limiting network binding to specific interfaces
- Implementing rate limiting
- Filtering sensitive information

## Performance

- **Fast Response**: Most endpoints respond in < 5ms
- **Concurrent Safe**: All functions are safe for concurrent use
- **Low Memory**: Minimal memory footprint
- **No Dependencies**: Uses only Go standard library

## Error Handling

The package gracefully handles common scenarios:
- Missing `/proc` files (falls back to defaults)
- Permission errors (returns partial information)
- Malformed system files (skips invalid entries)
- Network interface errors (continues with available interfaces)

## Use Cases

- **System Monitoring**: Real-time system status monitoring
- **Infrastructure Management**: Automated system inventory
- **Health Checks**: Service health monitoring
- **Development**: Local development environment information
- **Debugging**: Quick system information gathering
- **Documentation**: Automated system documentation generation

## Integration Examples

### Docker Health Check
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:8080/api/health || exit 1
```

### Monitoring Integration
```bash
# Prometheus metrics endpoint
curl http://localhost:8080/api/sysinfo | jq '.memory.used_percent'

# Grafana dashboard data source
http://localhost:8080/api/sysinfo
```

### CI/CD Pipeline
```yaml
- name: Gather System Info
  run: |
    go run cmd/sysinfo_app.go -info > system-info.txt
    cat system-info.txt
```