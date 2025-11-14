package sysinfo

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Server represents the system info web server
type Server struct {
	port   int
	server *http.Server
}

// NewServer creates a new system info server
func NewServer(port int) *Server {
	return &Server{
		port: port,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register routes using Go 1.22+ enhanced routing with HTTP methods
	mux.HandleFunc("GET /", s.handleRoot)

	// System info endpoints
	mux.HandleFunc("GET /api/sysinfo", s.handleSystemInfo)
	mux.HandleFunc("POST /api/sysinfo", s.handleSystemInfoPost)
	mux.HandleFunc("GET /api/sysinfo/os", s.handleOSInfo)
	mux.HandleFunc("GET /api/sysinfo/cpu", s.handleCPUInfo)
	mux.HandleFunc("GET /api/sysinfo/memory", s.handleMemoryInfo)
	mux.HandleFunc("GET /api/sysinfo/network", s.handleNetworkInfo)
	mux.HandleFunc("GET /api/sysinfo/uptime", s.handleUptimeInfo)

	// Health endpoints
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("POST /api/health", s.handleHealthPost)

	// Configuration endpoints (CRUD operations)
	mux.HandleFunc("GET /api/config", s.handleConfigGet)
	mux.HandleFunc("POST /api/config", s.handleConfigPost)
	mux.HandleFunc("PUT /api/config", s.handleConfigPut)
	mux.HandleFunc("DELETE /api/config", s.handleConfigDelete)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.loggingMiddleware(s.corsMiddleware(mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting system info server on port %d", s.port)
	log.Printf("Available endpoints:")
	log.Printf("  GET / - HTML dashboard")
	log.Printf("  GET /api/sysinfo - Complete system information")
	log.Printf("  POST /api/sysinfo - Get filtered system information")
	log.Printf("  GET /api/sysinfo/os - OS information only")
	log.Printf("  GET /api/sysinfo/cpu - CPU information only")
	log.Printf("  GET /api/sysinfo/memory - Memory information only")
	log.Printf("  GET /api/sysinfo/network - Network information only")
	log.Printf("  GET /api/sysinfo/uptime - Uptime information only")
	log.Printf("  GET /api/health - Health check")
	log.Printf("  POST /api/health - Health check with parameters")
	log.Printf("  GET /api/config - Get server configuration")
	log.Printf("  POST /api/config - Create new configuration")
	log.Printf("  PUT /api/config - Update configuration")
	log.Printf("  DELETE /api/config - Reset configuration to defaults")

	return s.server.ListenAndServe()
}

// Stop stops the web server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Middleware functions
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Route handlers
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>System Information Dashboard</title>
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            font-size: 16px;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f8f9fa;
            color: #333;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 30px;
            text-align: center;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }

        .header h1 {
            font-size: 2.5rem;
            margin-bottom: 10px;
            font-weight: 700;
        }

        .header p {
            font-size: 1.2rem;
            opacity: 0.9;
            font-weight: 300;
        }

        .card {
            background: white;
            padding: 30px;
            margin: 20px 0;
            border-radius: 12px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.08);
            border: 1px solid #e9ecef;
        }

        .card h2 {
            font-size: 1.8rem;
            margin-bottom: 20px;
            color: #2c3e50;
            font-weight: 600;
        }

        .endpoint {
            background: #f8f9fa;
            padding: 15px;
            margin: 10px 0;
            border-radius: 8px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 14px;
            border-left: 4px solid #3498db;
            transition: all 0.2s ease;
        }

        .endpoint:hover {
            background: #e9ecef;
            transform: translateX(5px);
        }

        .endpoint a {
            color: #2980b9;
            text-decoration: none;
            font-weight: 600;
        }

        .endpoint a:hover {
            text-decoration: underline;
            color: #1abc9c;
        }

        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 20px;
        }

        pre {
            background: #2c3e50;
            color: #ecf0f1;
            padding: 20px;
            border-radius: 8px;
            overflow-x: auto;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 13px;
            line-height: 1.4;
            border: 1px solid #34495e;
            max-height: 500px;
            overflow-y: auto;
        }

        .refresh-btn {
            background: linear-gradient(135deg, #3498db, #2980b9);
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            margin-bottom: 20px;
            transition: all 0.3s ease;
            box-shadow: 0 2px 5px rgba(52, 152, 219, 0.3);
        }

        .refresh-btn:hover {
            background: linear-gradient(135deg, #2980b9, #1abc9c);
            transform: translateY(-2px);
            box-shadow: 0 4px 15px rgba(52, 152, 219, 0.4);
        }

        .refresh-btn:active {
            transform: translateY(0);
        }

        .status-indicator {
            display: inline-block;
            width: 12px;
            height: 12px;
            background: #27ae60;
            border-radius: 50%;
            margin-right: 8px;
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0% { opacity: 1; }
            50% { opacity: 0.5; }
            100% { opacity: 1; }
        }

        @media (max-width: 768px) {
            body { padding: 10px; }
            .header h1 { font-size: 2rem; }
            .header p { font-size: 1rem; }
            .card { padding: 20px; }
            .grid { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🖥️ System Information Dashboard</h1>
            <p><span class="status-indicator"></span>Real-time system information and monitoring</p>
        </div>

        <div class="card">
            <h2>📊 API Endpoints</h2>
            <div class="grid">
                <div>
                    <div class="endpoint"><a href="/api/sysinfo">/api/sysinfo</a> - Complete system information</div>
                    <div class="endpoint"><a href="/api/sysinfo/os">/api/sysinfo/os</a> - Operating system details</div>
                    <div class="endpoint"><a href="/api/sysinfo/cpu">/api/sysinfo/cpu</a> - CPU information</div>
                    <div class="endpoint"><a href="/api/health">/api/health</a> - Health check endpoint</div>
                </div>
                <div>
                    <div class="endpoint"><a href="/api/sysinfo/memory">/api/sysinfo/memory</a> - Memory usage</div>
                    <div class="endpoint"><a href="/api/sysinfo/network">/api/sysinfo/network</a> - Network interfaces</div>
                    <div class="endpoint"><a href="/api/sysinfo/uptime">/api/sysinfo/uptime</a> - System uptime</div>
                    <div class="endpoint"><a href="/api/config">/api/config</a> - Server configuration</div>
                </div>
            </div>
        </div>

        <div class="card">
            <h2>🔍 Live System Information</h2>
            <button class="refresh-btn" onclick="refreshSystemInfo()">🔄 Refresh</button>
            <pre id="sysinfo-content">Loading system information...</pre>
        </div>
    </div>

    <script>
        async function refreshSystemInfo() {
            const content = document.getElementById('sysinfo-content');
            const btn = document.querySelector('.refresh-btn');

            try {
                btn.textContent = '⏳ Loading...';
                btn.disabled = true;

                const response = await fetch('/api/sysinfo');
                const data = await response.json();
                content.textContent = JSON.stringify(data, null, 2);

                btn.textContent = '🔄 Refresh';
                btn.disabled = false;
            } catch (error) {
                content.textContent = 'Error loading system information: ' + error.message;
                btn.textContent = '❌ Error - Retry';
                btn.disabled = false;
            }
        }

        // Load system info on page load
        document.addEventListener('DOMContentLoaded', refreshSystemInfo);

        // Auto-refresh every 30 seconds
        setInterval(refreshSystemInfo, 30000);

        // Add keyboard shortcut for refresh (Ctrl+R or Cmd+R)
        document.addEventListener('keydown', function(e) {
            if ((e.ctrlKey || e.metaKey) && e.key === 'r') {
                e.preventDefault();
                refreshSystemInfo();
            }
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (s *Server) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	info, err := GetSystemInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting system info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (s *Server) handleOSInfo(w http.ResponseWriter, r *http.Request) {
	osInfo, err := getOSInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting OS info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(osInfo)
}

func (s *Server) handleCPUInfo(w http.ResponseWriter, r *http.Request) {
	cpuInfo, err := getCPUInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting CPU info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cpuInfo)
}

func (s *Server) handleMemoryInfo(w http.ResponseWriter, r *http.Request) {
	memInfo, err := getMemoryInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting memory info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memInfo)
}

func (s *Server) handleNetworkInfo(w http.ResponseWriter, r *http.Request) {
	netInfo, err := getNetworkInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting network info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(netInfo)
}

func (s *Server) handleUptimeInfo(w http.ResponseWriter, r *http.Request) {
	uptimeInfo, err := getUptimeInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting uptime info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uptimeInfo)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    time.Since(serverStartTime),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// POST handler for system info - demonstrates receiving JSON data
func (s *Server) handleSystemInfoPost(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Fields []string `json:"fields"`
		Format string   `json:"format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get full system info
	info, err := GetSystemInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting system info: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter fields if requested
	response := make(map[string]interface{})

	if len(request.Fields) == 0 {
		// Return all info if no fields specified
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
		return
	}

	// Return only requested fields
	for _, field := range request.Fields {
		switch field {
		case "hostname":
			response["hostname"] = info.Hostname
		case "os":
			response["os"] = info.OS
		case "cpu":
			response["cpu"] = info.CPU
		case "memory":
			response["memory"] = info.Memory
		case "network":
			response["network"] = info.Network
		case "uptime":
			response["uptime"] = info.Uptime
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST handler for health endpoint
func (s *Server) handleHealthPost(w http.ResponseWriter, r *http.Request) {
	var request struct {
		CheckType string `json:"check_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	health := map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now(),
		"uptime":     time.Since(serverStartTime),
		"check_type": request.CheckType,
	}

	// Add additional checks based on check_type
	switch request.CheckType {
	case "detailed":
		// Add memory usage check
		memInfo, _ := getMemoryInfo()
		health["memory_usage_percent"] = memInfo.UsedPercent
		health["disk_space"] = "OK" // placeholder
	case "minimal":
		health = map[string]interface{}{
			"status": "healthy",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// Configuration management handlers (demonstration of CRUD operations)
type ServerConfig struct {
	LogLevel    string `json:"log_level"`
	RefreshRate int    `json:"refresh_rate"`
	EnableCORS  bool   `json:"enable_cors"`
}

var currentConfig = ServerConfig{
	LogLevel:    "info",
	RefreshRate: 30,
	EnableCORS:  true,
}

func (s *Server) handleConfigGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentConfig)
}

func (s *Server) handleConfigPost(w http.ResponseWriter, r *http.Request) {
	var newConfig ServerConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate config
	if newConfig.RefreshRate < 1 || newConfig.RefreshRate > 300 {
		http.Error(w, "refresh_rate must be between 1 and 300 seconds", http.StatusBadRequest)
		return
	}

	currentConfig = newConfig

	response := map[string]interface{}{
		"message": "Configuration created successfully",
		"config":  currentConfig,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleConfigPut(w http.ResponseWriter, r *http.Request) {
	var updateConfig ServerConfig
	if err := json.NewDecoder(r.Body).Decode(&updateConfig); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update only provided fields (partial update)
	if updateConfig.LogLevel != "" {
		currentConfig.LogLevel = updateConfig.LogLevel
	}
	if updateConfig.RefreshRate > 0 {
		if updateConfig.RefreshRate < 1 || updateConfig.RefreshRate > 300 {
			http.Error(w, "refresh_rate must be between 1 and 300 seconds", http.StatusBadRequest)
			return
		}
		currentConfig.RefreshRate = updateConfig.RefreshRate
	}
	currentConfig.EnableCORS = updateConfig.EnableCORS

	response := map[string]interface{}{
		"message": "Configuration updated successfully",
		"config":  currentConfig,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleConfigDelete(w http.ResponseWriter, r *http.Request) {
	// Reset to default configuration
	currentConfig = ServerConfig{
		LogLevel:    "info",
		RefreshRate: 30,
		EnableCORS:  true,
	}

	response := map[string]interface{}{
		"message": "Configuration reset to defaults",
		"config":  currentConfig,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var serverStartTime = time.Now()

// RunServer is a convenience function to start the server
func RunServer(port int) error {
	server := NewServer(port)
	return server.Start()
}
