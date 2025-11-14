package aggregator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"validate/agent"
	"validate/database"
	"validate/sysinfo"
)

// Aggregator represents an aggregator server
type Aggregator struct {
	port   int
	db     *database.DB
	server *http.Server
}

// NewAggregator creates a new aggregator server
func NewAggregator(port int, dbPath string) (*Aggregator, error) {
	db, err := database.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	return &Aggregator{
		port: port,
		db:   db,
	}, nil
}

// Start starts the aggregator server
func (a *Aggregator) Start() error {
	mux := http.NewServeMux()

	// Register routes using Go 1.22+ enhanced routing
	mux.HandleFunc("GET /", a.handleRoot)

	// System info endpoints (agent mode endpoints)
	mux.HandleFunc("GET /api/sysinfo", a.handleSystemInfo)
	mux.HandleFunc("GET /api/health", a.handleHealth)

	// Aggregator-specific endpoints
	mux.HandleFunc("POST /api/server", a.handleServerRegistration)
	mux.HandleFunc("GET /api/servers", a.handleGetServers)
	mux.HandleFunc("POST /api/test-results", a.handleTestResults)
	mux.HandleFunc("GET /api/test-results", a.handleGetTestResults)
	mux.HandleFunc("POST /api/run-tests", a.handleRunTests)

	a.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.port),
		Handler:      a.loggingMiddleware(a.corsMiddleware(mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting aggregator server on port %d", a.port)
	log.Printf("Available endpoints:")
	log.Printf("  GET / - HTML dashboard")
	log.Printf("  GET /api/sysinfo - System information")
	log.Printf("  GET /api/health - Health check")
	log.Printf("  POST /api/server - Server registration")
	log.Printf("  GET /api/servers - List registered servers")
	log.Printf("  POST /api/test-results - Submit test results")
	log.Printf("  GET /api/test-results - Get test results")
	log.Printf("  POST /api/run-tests - Trigger connectivity tests")

	return a.server.ListenAndServe()
}

// Stop stops the aggregator server
func (a *Aggregator) Stop() error {
	if a.server != nil {
		return a.server.Close()
	}
	return nil
}

// Close closes the aggregator and its database connection
func (a *Aggregator) Close() error {
	if err := a.Stop(); err != nil {
		return err
	}
	return a.db.Close()
}

// Middleware functions
func (a *Aggregator) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func (a *Aggregator) corsMiddleware(next http.Handler) http.Handler {
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

// Handler for server registration
func (a *Aggregator) handleServerRegistration(w http.ResponseWriter, r *http.Request) {
	var payload agent.RegistrationPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if payload.Hostname == "" {
		http.Error(w, "hostname is required", http.StatusBadRequest)
		return
	}

	if payload.IPAddress == "" {
		http.Error(w, "ip_address is required", http.StatusBadRequest)
		return
	}

	// Register the server in the database
	if err := a.db.RegisterServer(payload.Hostname, payload.IPAddress, payload.SystemInfo, payload.Bonds); err != nil {
		log.Printf("Failed to register server %s: %v", payload.Hostname, err)
		http.Error(w, fmt.Sprintf("Failed to register server: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Server registered: %s (%s) with bonds: %v", payload.Hostname, payload.IPAddress, payload.Bonds)

	response := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Server %s registered successfully", payload.Hostname),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Handler to get all registered servers
func (a *Aggregator) handleGetServers(w http.ResponseWriter, r *http.Request) {
	servers, err := a.db.GetAllServers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get servers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

// Handler to receive test results
func (a *Aggregator) handleTestResults(w http.ResponseWriter, r *http.Request) {
	var payload agent.TestResultPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Save each test result to the database
	for _, result := range payload.Results {
		dbResult := database.TestResult{
			SourceHostname: payload.SourceHostname,
			TargetHostname: result.TargetHostname,
			TargetIP:       result.TargetIP,
			SourceIP:       result.SourceIP,
			BondName:       result.BondName,
			TestType:       result.TestType,
			Success:        result.Success,
			ResponseTime:   result.ResponseTimeMS,
			ErrorMessage:   result.ErrorMessage,
			TestedAt:       payload.TestedAt,
		}

		if err := a.db.SaveTestResult(dbResult); err != nil {
			log.Printf("Failed to save test result: %v", err)
			continue
		}
	}

	log.Printf("Received %d test results from %s", len(payload.Results), payload.SourceHostname)

	response := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Received %d test results", len(payload.Results)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handler to get test results
func (a *Aggregator) handleGetTestResults(w http.ResponseWriter, r *http.Request) {
	// Get limit from query parameter, default to 0 (unlimited)
	limit := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	// Get source hostname filter if provided
	source := r.URL.Query().Get("source")

	var results []database.TestResult
	var err error

	if source != "" {
		results, err = a.db.GetTestResultsBySource(source, limit)
	} else {
		results, err = a.db.GetTestResults(limit)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get test results: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// Handler to trigger connectivity tests
func (a *Aggregator) handleRunTests(w http.ResponseWriter, r *http.Request) {
	// Trigger connectivity tests on all registered agents...
	log.Println("Triggering connectivity tests on all agents...")

	// Clear existing test results before running new tests
	if err := a.db.ClearTestResults(); err != nil {
		log.Printf("Warning: Failed to clear test results: %v", err)
		// Continue anyway - don't fail the request
	} else {
		log.Println("Cleared all previous test results")
	}

	servers, err := a.db.GetAllServers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get servers: %v", err), http.StatusInternalServerError)
		return
	}

	if len(servers) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "No servers registered to test",
			"count":   0,
		})
		return
	}

	// Build test targets from registered servers
	allTargets := make(map[string]agent.TargetInfo)

	for _, server := range servers {
		var bonds map[string][]string
		if err := json.Unmarshal([]byte(server.Bonds), &bonds); err != nil {
			log.Printf("Failed to unmarshal bonds for %s: %v", server.Hostname, err)
			continue
		}

		allTargets[server.Hostname] = agent.TargetInfo{
			Links: bonds,
		}
	}

	// Trigger tests on each agent asynchronously
	type triggerResult struct {
		hostname string
		ipAddr   string
		success  bool
		err      error
	}

	resultsChan := make(chan triggerResult, len(servers))

	for _, server := range servers {
		// Build targets for this agent (exclude itself)
		targets := make(map[string]agent.TargetInfo)
		for hostname, info := range allTargets {
			if hostname != server.Hostname {
				targets[hostname] = info
			}
		}

		testRequest := agent.TestRequest{
			Targets: targets,
		}

		// Send test request to agent using its IP address
		agentURL := fmt.Sprintf("http://%s:8080/api/run-tests", server.IPAddress)

		go func(url, hostname, ipAddr string, req agent.TestRequest) {
			reqBody, _ := json.Marshal(req)
			resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
			if err != nil {
				log.Printf("Failed to trigger tests on %s (%s): %v", hostname, ipAddr, err)
				resultsChan <- triggerResult{hostname: hostname, ipAddr: ipAddr, success: false, err: err}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
				log.Printf("Successfully triggered tests on %s (%s)", hostname, ipAddr)
				resultsChan <- triggerResult{hostname: hostname, ipAddr: ipAddr, success: true}
			} else {
				errMsg := fmt.Errorf("status %d", resp.StatusCode)
				log.Printf("Agent %s (%s) returned status %d", hostname, ipAddr, resp.StatusCode)
				resultsChan <- triggerResult{hostname: hostname, ipAddr: ipAddr, success: false, err: errMsg}
			}
		}(agentURL, server.Hostname, server.IPAddress, testRequest)
	}

	// Wait briefly for all trigger acknowledgments (not test results)
	successCount := 0
	failedAgents := []string{}
	timeout := time.After(2 * time.Second)

	for i := 0; i < len(servers); i++ {
		select {
		case result := <-resultsChan:
			if result.success {
				successCount++
			} else {
				failedAgents = append(failedAgents, fmt.Sprintf("%s (%s): %v", result.hostname, result.ipAddr, result.err))
			}
		case <-timeout:
			remaining := len(servers) - i
			if remaining > 0 {
				log.Printf("Timeout waiting for %d agent acknowledgments", remaining)
				failedAgents = append(failedAgents, fmt.Sprintf("%d agents timed out", remaining))
			}
			goto done
		}
	}

done:
	// Return results
	response := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Test requests sent to %d/%d agent(s). Results will be posted back.", successCount, len(servers)),
		"count":   successCount,
		"total":   len(servers),
	}

	if len(failedAgents) > 0 {
		response["failed_agents"] = failedAgents
		response["message"] = fmt.Sprintf("Tests triggered on %d/%d agent(s). %d failed.", successCount, len(servers), len(failedAgents))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handler for system info (this server's info)
func (a *Aggregator) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	info, err := sysinfo.GetSystemInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting system info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// Handler for health check
func (a *Aggregator) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"mode":      "aggregator",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// Handler for the root dashboard
func (a *Aggregator) handleRoot(w http.ResponseWriter, r *http.Request) {
	html := a.getDashboardHTML()
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// getDashboardHTML returns the HTML for the aggregator dashboard
func (a *Aggregator) getDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Network Aggregator Dashboard</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            font-size: 16px;
            line-height: 1.6;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }

        .container { max-width: 1400px; margin: 0 auto; }

        .header {
            background: white;
            color: #2c3e50;
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 30px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.1);
        }

        .header h1 { font-size: 2.5rem; margin-bottom: 10px; }
        .header p { font-size: 1.2rem; color: #7f8c8d; }

        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .stat-card {
            background: white;
            padding: 25px;
            border-radius: 12px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }

        .stat-card h3 { font-size: 0.9rem; color: #7f8c8d; margin-bottom: 10px; text-transform: uppercase; }
        .stat-card .value { font-size: 2.5rem; font-weight: bold; color: #2c3e50; }

        .card {
            background: white;
            padding: 30px;
            margin: 20px 0;
            border-radius: 12px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.08);
        }

        .card h2 { font-size: 1.8rem; margin-bottom: 20px; color: #2c3e50; }

        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }

        th, td {
            padding: 15px;
            text-align: left;
            border-bottom: 1px solid #ecf0f1;
        }

        th {
            background: #f8f9fa;
            font-weight: 600;
            color: #2c3e50;
        }

        tr:hover {
            background: #f8f9fa;
        }

        .success { color: #27ae60; font-weight: bold; }
        .failure { color: #e74c3c; font-weight: bold; }

        .refresh-btn {
            background: linear-gradient(135deg, #3498db, #2980b9);
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            transition: all 0.3s ease;
            margin-right: 10px;
        }

        .refresh-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 15px rgba(52, 152, 219, 0.4);
        }

        .run-tests-btn {
            background: linear-gradient(135deg, #e74c3c, #c0392b);
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            transition: all 0.3s ease;
        }

        .run-tests-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 15px rgba(231, 76, 60, 0.4);
        }

        .run-tests-btn:disabled {
            background: #95a5a6;
            cursor: not-allowed;
            transform: none;
        }

        .filter-btn {
            background: linear-gradient(135deg, #3498db, #2980b9);
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 600;
            transition: all 0.3s ease;
        }

        .filter-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 15px rgba(52, 152, 219, 0.4);
        }

        .filter-btn.active {
            background: linear-gradient(135deg, #e74c3c, #c0392b);
        }

        .button-group {
            margin-bottom: 20px;
        }

        .status-message {
            padding: 10px;
            margin: 10px 0;
            border-radius: 4px;
            display: none;
        }

        .status-message.success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }

        .status-message.error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌐 Network Aggregator Dashboard</h1>
            <p>Centralized server and connectivity monitoring</p>
        </div>

        <div class="stats">
            <div class="stat-card">
                <h3>Registered Servers</h3>
                <div class="value" id="server-count">0</div>
            </div>
            <div class="stat-card">
                <h3>Total Tests</h3>
                <div class="value" id="test-count">0</div>
            </div>
            <div class="stat-card">
                <h3>Success Rate</h3>
                <div class="value" id="success-rate">0%</div>
            </div>
        </div>

        <div class="card">
            <h2>📡 Registered Servers</h2>
            <div class="button-group">
                <button class="refresh-btn" onclick="refreshData()">🔄 Refresh</button>
                <button class="run-tests-btn" onclick="runAllTests()" id="run-tests-btn">🚀 Run Connectivity Tests</button>
            </div>
            <div id="test-status" class="status-message"></div>
            <table id="servers-table">
                <thead>
                    <tr>
                        <th>Hostname</th>
                        <th>IP Address</th>
                        <th>Bonds</th>
                        <th>Last Seen</th>
                    </tr>
                </thead>
                <tbody id="servers-body">
                    <tr><td colspan="4">Loading...</td></tr>
                </tbody>
            </table>
        </div>

        <div class="card">
            <h2>🔍 Recent Connectivity Tests</h2>
            <div class="button-group">
                <button class="filter-btn active" onclick="toggleFilter()" id="filter-btn">✓ Show All</button>
            </div>
            <table id="tests-table">
                <thead>
                    <tr>
                        <th>Source</th>
                        <th>Source IP</th>
                        <th>Target</th>
                        <th>Target IP</th>
                        <th>Bond</th>
                        <th>Test Type</th>
                        <th>Status</th>
                        <th>Response Time</th>
                        <th>Tested At</th>
                    </tr>
                </thead>
                <tbody id="tests-body">
                    <tr><td colspan="9">Loading...</td></tr>
                </tbody>
            </table>
        </div>
    </div>

    <script>
        async function refreshData() {
            await Promise.all([loadServers(), loadTestResults()]);
        }

        async function loadServers() {
            try {
                const response = await fetch('/api/servers');
                const servers = await response.json();

                document.getElementById('server-count').textContent = servers.length;

                const tbody = document.getElementById('servers-body');
                if (servers.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="4">No servers registered</td></tr>';
                    return;
                }

                tbody.innerHTML = servers.map(server => {
                    const bonds = JSON.parse(server.bonds);
                    const bondList = Object.keys(bonds).join(', ') || 'None';
                    const lastSeen = new Date(server.last_seen).toLocaleString();

                    return ` + "`" + `
                        <tr>
                            <td>${server.hostname}</td>
                            <td>${server.ip_address}</td>
                            <td>${bondList}</td>
                            <td>${lastSeen}</td>
                        </tr>
                    ` + "`" + `;
                }).join('');
            } catch (error) {
                console.error('Failed to load servers:', error);
            }
        }

        let allTestResults = [];
        let showFailedOnly = true;  // Default to showing only failed tests

        async function loadTestResults() {
            try {
                const response = await fetch('/api/test-results');
                const results = await response.json();

                allTestResults = results;
                renderTestResults();
            } catch (error) {
                console.error('Failed to load test results:', error);
            }
        }

        function renderTestResults() {
            const results = showFailedOnly
                ? allTestResults.filter(r => !r.success)
                : allTestResults;

            const totalTests = allTestResults.length;
            const successCount = allTestResults.filter(r => r.success).length;
            const failedCount = totalTests - successCount;

            // Display test count as "x out of x failed"
            const testCountText = totalTests > 0
                ? failedCount + ' out of ' + totalTests + ' failed'
                : '0 tests';
            document.getElementById('test-count').textContent = testCountText;

            // Calculate success rate as float without rounding
            const successRate = totalTests > 0
                ? ((successCount / totalTests) * 100).toFixed(2)
                : '0.00';
            document.getElementById('success-rate').textContent = successRate + '%';            const tbody = document.getElementById('tests-body');
            if (results.length === 0) {
                const message = showFailedOnly ? 'No failed tests' : 'No test results yet';
                tbody.innerHTML = ` + "`<tr><td colspan=\"9\">${message}</td></tr>`" + `;
                return;
            }

            tbody.innerHTML = results.map(result => {
                const status = result.success
                    ? '<span class="success">✓ Success</span>'
                    : '<span class="failure">✗ Failed</span>';
                const responseTime = result.success
                    ? ` + "`" + `${result.response_time_ms}ms` + "`" + `
                    : result.error_message;
                const testedAt = new Date(result.tested_at).toLocaleString();
                const testType = result.test_type ? result.test_type.toUpperCase() : 'N/A';

                return ` + "`" + `
                    <tr>
                        <td>${result.source_hostname}</td>
                        <td>${result.source_ip}</td>
                        <td>${result.target_hostname}</td>
                        <td>${result.target_ip}</td>
                        <td>${result.bond_name}</td>
                        <td>${testType}</td>
                        <td>${status}</td>
                        <td>${responseTime}</td>
                        <td>${testedAt}</td>
                    </tr>
                ` + "`" + `;
            }).join('');
        }

        function toggleFilter() {
            showFailedOnly = !showFailedOnly;
            const btn = document.getElementById('filter-btn');

            if (showFailedOnly) {
                btn.textContent = '✓ Show All';
                btn.classList.add('active');
            } else {
                btn.textContent = '❌ Show Failed Only';
                btn.classList.remove('active');
            }

            renderTestResults();
        }

        async function runAllTests() {
            const btn = document.getElementById('run-tests-btn');
            const statusDiv = document.getElementById('test-status');

            try {
                btn.disabled = true;
                btn.textContent = '⏳ Running tests...';
                statusDiv.style.display = 'none';

                // Trigger tests on the aggregator (it will coordinate with agents)
                const response = await fetch('/api/run-tests', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' }
                });

                if (!response.ok) {
                    throw new Error('Failed to trigger tests');
                }

                const result = await response.json();

                // Show message with failed agents if any
                let message = result.message;
                if (result.failed_agents && result.failed_agents.length > 0) {
                    message += '\\n\\nFailed to trigger on:\\n' + result.failed_agents.join('\\n');
                    showStatus(message, 'error');
                } else {
                    showStatus(message, 'success');
                }

                // Refresh results after a delay to see the test results
                setTimeout(() => {
                    refreshData();
                }, 3000);

            } catch (error) {
                showStatus(` + "`Error: ${error.message}`" + `, 'error');
                console.error('Error running tests:', error);
            } finally {
                btn.disabled = false;
                btn.textContent = '🚀 Run Connectivity Tests';
            }
        }

        function showStatus(message, type) {
            const statusDiv = document.getElementById('test-status');
            statusDiv.textContent = message;
            statusDiv.className = ` + "`status-message ${type}`" + `;
            statusDiv.style.display = 'block';

            // Auto-hide after 5 seconds
            setTimeout(() => {
                statusDiv.style.display = 'none';
            }, 5000);
        }

        // Load data on page load
        refreshData();

        // Auto-refresh every 30 seconds
        setInterval(refreshData, 30000);
    </script>
</body>
</html>`
}
