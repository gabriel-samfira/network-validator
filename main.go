package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"validate/agent"
	"validate/aggregator"
	"validate/config"
	"validate/sysinfo"
)

func main() {
	configFile := flag.String("config", "config.toml", "Path to configuration file")
	generateConfig := flag.String("generate-config", "", "Generate a default config file (aggregator or agent)")
	flag.Parse()

	// Generate config if requested
	if *generateConfig != "" {
		if err := config.GenerateDefaultConfig(*configFile, *generateConfig); err != nil {
			log.Fatalf("Failed to generate config: %v", err)
		}
		fmt.Printf("Generated %s config file at %s\n", *generateConfig, *configFile)
		return
	}

	// Load config
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting in %s mode", cfg.Mode)

	if cfg.Mode == "aggregator" {
		runAggregator(cfg)
	} else {
		runAgent(cfg)
	}
}

func runAggregator(cfg *config.Config) {
	agg, err := aggregator.NewAggregator(cfg.Aggregator.Port, cfg.Aggregator.Database)
	if err != nil {
		log.Fatalf("Failed to create aggregator: %v", err)
	}
	defer agg.Close()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down aggregator...")
		agg.Stop()
		os.Exit(0)
	}()

	log.Fatal(agg.Start())
}

func runAgent(cfg *config.Config) {
	// Create agent
	ag, err := agent.NewAgent(cfg.Agent.AggregatorURL)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Start periodic registration in background
	stopChan := make(chan struct{})
	go ag.StartPeriodicRegistration(time.Duration(cfg.Agent.RegisterInterval)*time.Second, stopChan)

	// Start HTTP server for receiving test requests
	mux := http.NewServeMux()

	// Endpoint for system info
	mux.HandleFunc("GET /api/sysinfo", handleSystemInfo)

	// Endpoint for health check
	mux.HandleFunc("GET /api/health", handleHealth)

	// Endpoint for running connectivity tests
	mux.HandleFunc("POST /api/run-tests", func(w http.ResponseWriter, r *http.Request) {
		handleRunTests(w, r, ag)
	})

	server := &http.Server{
		Addr:         cfg.Agent.ListenAddr,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down agent...")
		close(stopChan)
		server.Close()
		os.Exit(0)
	}()

	log.Printf("Agent HTTP server listening on %s", cfg.Agent.ListenAddr)
	log.Fatal(server.ListenAndServe())
}

// Agent HTTP handlers
func handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	info, err := sysinfo.GetSystemInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting system info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"mode":      "agent",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func handleRunTests(w http.ResponseWriter, r *http.Request, ag *agent.Agent) {
	var testReq agent.TestRequest

	if err := json.NewDecoder(r.Body).Decode(&testReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Received request to run connectivity tests to %d targets", len(testReq.Targets))

	// Run tests asynchronously in background
	// Results are now submitted as each test completes
	go func() {
		log.Printf("Starting connectivity tests in background")
		ag.RunConnectivityTests(testReq.Targets)
		log.Printf("Connectivity tests completed")
	}()

	// Return immediately
	response := map[string]interface{}{
		"status":  "accepted",
		"message": fmt.Sprintf("Tests queued for %d targets", len(testReq.Targets)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
