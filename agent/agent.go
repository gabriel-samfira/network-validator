package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"validate/netplan"
	"validate/sysinfo"
)

// Agent represents an agent that registers with an aggregator
type Agent struct {
	aggregatorURL string
	httpClient    *http.Client
	hostname      string
}

// RegistrationPayload is the data sent when registering with the aggregator
type RegistrationPayload struct {
	Hostname   string              `json:"hostname"`
	IPAddress  string              `json:"ip_address"`
	SystemInfo interface{}         `json:"system_info"`
	Bonds      map[string][]string `json:"bonds"`
}

// TestRequest represents a test request from the aggregator
type TestRequest struct {
	Targets map[string]TargetInfo `json:"targets"`
}

// TargetInfo contains information about target servers and their links
type TargetInfo struct {
	Links map[string][]string `json:"links"` // bond -> IPs mapping
}

// TestResultPayload is the result of connectivity tests
type TestResultPayload struct {
	SourceHostname string       `json:"source_hostname"`
	Results        []TestResult `json:"results"`
	TestedAt       time.Time    `json:"tested_at"`
}

// TestResult represents a single connectivity test result
type TestResult struct {
	TargetHostname string `json:"target_hostname"`
	TargetIP       string `json:"target_ip"`
	SourceIP       string `json:"source_ip"`
	BondName       string `json:"bond_name"`
	TestType       string `json:"test_type"` // "arp" or "http"
	Success        bool   `json:"success"`
	ResponseTimeMS int64  `json:"response_time_ms"`
	ErrorMessage   string `json:"error_message,omitempty"`
}

// NewAgent creates a new agent
func NewAgent(aggregatorURL string) (*Agent, error) {
	hostname, err := sysinfo.GetHostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	return &Agent{
		aggregatorURL: aggregatorURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		hostname: hostname,
	}, nil
}

// Register registers this agent with the aggregator
func (a *Agent) Register() error {
	// Get system info
	systemInfo, err := sysinfo.GetSystemInfo()
	if err != nil {
		return fmt.Errorf("failed to get system info: %w", err)
	}

	// Get main IP address
	ipAddr, err := sysinfo.GetMainIPAddress()
	if err != nil {
		return fmt.Errorf("failed to get main IP address: %w", err)
	}

	// Get bond IP addresses
	bonds, err := a.getBondIPAddresses()
	if err != nil {
		return fmt.Errorf("failed to get bond IP addresses: %w", err)
	}

	payload := RegistrationPayload{
		Hostname:   a.hostname,
		IPAddress:  ipAddr,
		SystemInfo: systemInfo,
		Bonds:      bonds,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/server", a.aggregatorURL)
	resp, err := a.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	return nil
}

// getBondIPAddresses gets all bond IP addresses from the system
func (a *Agent) getBondIPAddresses() (map[string][]string, error) {
	// Try to load netplan configurations
	configs, err := netplan.LoadNetplanConfigsFromDir("/etc/netplan")
	if err != nil {
		// If netplan fails, return empty map (not all systems use netplan)
		return make(map[string][]string), nil
	}

	allBonds := make(map[string][]string)

	// Iterate through all configs and all bonds
	for _, config := range configs {
		if config.Network.Bonds == nil {
			continue
		}

		for bondName := range config.Network.Bonds {
			bondIPs := config.GetBondIPAddresses(bondName)

			// Flatten the map - we want bond -> all IPs across all interfaces
			var ips []string
			for _, addrs := range bondIPs {
				ips = append(ips, addrs...)
			}

			if len(ips) > 0 {
				allBonds[bondName] = ips
			}
		}
	}

	return allBonds, nil
}

// getBondIPAddressesWithMask returns IP addresses with CIDR notation for subnet matching
func (a *Agent) getBondIPAddressesWithMask() ([]netplan.IPWithMask, error) {
	// Try to load netplan configurations
	configs, err := netplan.LoadNetplanConfigsFromDir("/etc/netplan")
	if err != nil {
		// If netplan fails, return empty slice
		return []netplan.IPWithMask{}, nil
	}

	var allIPs []netplan.IPWithMask

	// Iterate through all configs and all bonds
	for _, config := range configs {
		if config.Network.Bonds == nil {
			continue
		}

		for bondName := range config.Network.Bonds {
			bondIPs := config.GetBondIPAddressesWithMask(bondName)
			allIPs = append(allIPs, bondIPs...)
		}
	}

	return allIPs, nil
}

// RunConnectivityTests performs connectivity tests to the specified targets
// Only tests connectivity to targets where this agent has an interface in the same subnet
// Posts results immediately after each test instead of batching
func (a *Agent) RunConnectivityTests(targets map[string]TargetInfo) {
	// Get this agent's IP addresses with CIDR notation for subnet matching
	myIPs, err := a.getBondIPAddressesWithMask()
	if err != nil {
		fmt.Printf("Warning: Failed to get local IP configuration: %v\n", err)
		myIPs = []netplan.IPWithMask{}
	}

	fmt.Printf("Starting connectivity tests to %d targets\n", len(targets))
	fmt.Printf("My IPs with subnets: ")
	for _, ip := range myIPs {
		fmt.Printf("%s ", ip.CIDR)
	}
	fmt.Printf("\n")

	testCount := 0
	for targetHostname, targetInfo := range targets {
		for bondName, ips := range targetInfo.Links {
			fmt.Printf("Checking %s via bond %s (%d IPs)\n", targetHostname, bondName, len(ips))

			for _, targetIP := range ips {
				// Check if this agent has an IP in the same subnet as the target
				inSameSubnet := false
				var matchingLocalIP string
				var matchingInterface string

				for _, myIP := range myIPs {
					if netplan.InSameSubnet(myIP.CIDR, targetIP) {
						inSameSubnet = true
						matchingLocalIP = myIP.IP
						matchingInterface = myIP.BondName
						break
					}
				}

				if !inSameSubnet {
					fmt.Printf("  Skipping %s - no local interface in same subnet\n", targetIP)
					continue
				}

				fmt.Printf("  Testing %s (local IP %s on %s is in same subnet)\n", targetIP, matchingLocalIP, matchingInterface)
				results := a.testConnectivity(targetHostname, targetIP, bondName, matchingLocalIP, matchingInterface)

				// Submit each result immediately (ARP and HTTP)
				for _, result := range results {
					fmt.Printf("  -> %s [%s]: %vms (success=%v)\n", targetIP, result.TestType, result.ResponseTimeMS, result.Success)
					if err := a.SubmitSingleTestResult(result); err != nil {
						fmt.Printf("  Failed to submit %s result: %v\n", result.TestType, err)
					} else {
						testCount++
					}
				}
			}
		}
	}

	fmt.Printf("Completed and submitted %d connectivity tests\n", testCount)
}

// SubmitSingleTestResult submits a single test result immediately to the aggregator
func (a *Agent) SubmitSingleTestResult(result TestResult) error {
	// Wrap the single result in an array and reuse existing SubmitTestResults
	return a.SubmitTestResults([]TestResult{result})
}

// testConnectivity tests connectivity to a specific IP address using both arping and HTTP
// Returns two results: one for ARP test and one for HTTP test
func (a *Agent) testConnectivity(targetHostname, targetIP, bondName, sourceIP, sourceInterface string) []TestResult {
	var results []TestResult

	// Test 1: ARP connectivity
	arpResult := TestResult{
		TargetHostname: targetHostname,
		TargetIP:       targetIP,
		SourceIP:       sourceIP,
		BondName:       bondName,
		TestType:       "arp",
	}

	arpStart := time.Now()
	arpCmd := exec.Command("arping", "-W", "0.5", "-c", "3", "-I", sourceInterface, targetIP)
	arpErr := arpCmd.Run()
	arpElapsed := time.Since(arpStart)

	arpResult.ResponseTimeMS = arpElapsed.Milliseconds()
	if arpErr != nil {
		arpResult.Success = false
		arpResult.ErrorMessage = fmt.Sprintf("ARP ping failed: %v", arpErr)
	} else {
		arpResult.Success = true
	}
	results = append(results, arpResult)

	// Test 2: HTTP connectivity (always run, regardless of ARP result)
	httpResult := TestResult{
		TargetHostname: targetHostname,
		TargetIP:       targetIP,
		SourceIP:       sourceIP,
		BondName:       bondName,
		TestType:       "http",
	}

	url := fmt.Sprintf("http://%s:8080/api/sysinfo", targetIP)

	httpStart := time.Now()
	resp, err := a.httpClient.Get(url)
	httpElapsed := time.Since(httpStart)

	httpResult.ResponseTimeMS = httpElapsed.Milliseconds()

	if err != nil {
		httpResult.Success = false
		httpResult.ErrorMessage = err.Error()
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			httpResult.Success = true
		} else {
			httpResult.Success = false
			httpResult.ErrorMessage = fmt.Sprintf("HTTP status %d", resp.StatusCode)
		}
	}
	results = append(results, httpResult)

	return results
}

// SubmitTestResults submits test results back to the aggregator
func (a *Agent) SubmitTestResults(results []TestResult) error {
	payload := TestResultPayload{
		SourceHostname: a.hostname,
		Results:        results,
		TestedAt:       time.Now(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	url := fmt.Sprintf("%s/api/test-results", a.aggregatorURL)
	fmt.Printf("Submitting %d test results to %s\n", len(results), url)

	resp, err := a.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to submit results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("submission failed with status %d", resp.StatusCode)
	}

	fmt.Printf("Successfully submitted test results, got status %d\n", resp.StatusCode)

	return nil
}

// StartPeriodicRegistration starts periodic registration with the aggregator
func (a *Agent) StartPeriodicRegistration(interval time.Duration, stopChan <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Register immediately
	if err := a.Register(); err != nil {
		fmt.Printf("Initial registration failed: %v\n", err)
	} else {
		fmt.Printf("Successfully registered with aggregator at %s\n", a.aggregatorURL)
	}

	for {
		select {
		case <-ticker.C:
			if err := a.Register(); err != nil {
				fmt.Printf("Registration failed: %v\n", err)
			} else {
				fmt.Printf("Registration renewed at %s\n", time.Now().Format(time.RFC3339))
			}
		case <-stopChan:
			fmt.Println("Stopping periodic registration")
			return
		}
	}
}
