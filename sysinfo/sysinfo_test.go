package sysinfo

import (
	"testing"
	"time"
)

func TestGetSystemInfo(t *testing.T) {
	info, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("Failed to get system info: %v", err)
	}

	// Basic validation
	if info.Hostname == "" {
		t.Error("Hostname should not be empty")
	}

	if info.OS.Name == "" {
		t.Error("OS name should not be empty")
	}

	if info.CPU.Model == "" {
		t.Error("CPU model should not be empty")
	}

	if info.Memory.TotalBytes == 0 {
		t.Error("Total memory should not be zero")
	}

	if len(info.Network.Interfaces) == 0 {
		t.Error("Should have at least one network interface")
	}

	if info.Uptime.Seconds <= 0 {
		t.Error("Uptime should be positive")
	}

	// Check timestamp is recent
	if time.Since(info.Timestamp) > time.Minute {
		t.Error("Timestamp should be recent")
	}
}

func TestGetOSRelease(t *testing.T) {
	osRelease, err := GetOSRelease()
	if err != nil {
		t.Fatalf("Failed to get OS release: %v", err)
	}

	// Should have at least some basic fields
	if len(osRelease) == 0 {
		t.Error("OS release should not be empty")
	}

	// Check for common fields
	commonFields := []string{"NAME", "VERSION", "ID"}
	for _, field := range commonFields {
		if _, exists := osRelease[field]; !exists {
			t.Logf("Warning: OS release missing common field: %s", field)
		}
	}
}

func TestGetProcessorInfo(t *testing.T) {
	procInfo, err := GetProcessorInfo()
	if err != nil {
		t.Fatalf("Failed to get processor info: %v", err)
	}

	if count, exists := procInfo["processor_count"]; !exists || count == 0 {
		t.Error("Should have at least one processor")
	}

	if processors, exists := procInfo["processors"]; !exists {
		t.Error("Should have processors array")
	} else {
		if procs, ok := processors.([]map[string]string); ok {
			if len(procs) == 0 {
				t.Error("Processors array should not be empty")
			}
		}
	}
}

func TestMemoryCalculations(t *testing.T) {
	memInfo, err := getMemoryInfo()
	if err != nil {
		t.Fatalf("Failed to get memory info: %v", err)
	}

	// Validate memory calculations
	if memInfo.TotalBytes < memInfo.UsedBytes+memInfo.FreeBytes {
		t.Error("Total memory should be >= used + free")
	}

	if memInfo.UsedPercent < 0 || memInfo.UsedPercent > 100 {
		t.Errorf("Used percent should be 0-100, got %.2f", memInfo.UsedPercent)
	}

	// GB calculations should be consistent with byte values
	expectedTotalGB := float64(memInfo.TotalBytes) / (1024 * 1024 * 1024)
	if abs(memInfo.TotalGB-expectedTotalGB) > 0.01 {
		t.Errorf("Total GB calculation inconsistent: expected %.2f, got %.2f", expectedTotalGB, memInfo.TotalGB)
	}
}

func TestNetworkInterfaces(t *testing.T) {
	netInfo, err := getNetworkInfo()
	if err != nil {
		t.Fatalf("Failed to get network info: %v", err)
	}

	// Should have at least loopback interface
	hasLoopback := false
	for _, iface := range netInfo.Interfaces {
		if iface.Name == "lo" || iface.IsLoopback {
			hasLoopback = true
			break
		}
	}

	if !hasLoopback {
		t.Error("Should have loopback interface")
	}

	// Validate interface data
	for _, iface := range netInfo.Interfaces {
		if iface.Name == "" {
			t.Error("Interface name should not be empty")
		}

		if iface.MTU <= 0 {
			t.Errorf("Interface %s MTU should be positive, got %d", iface.Name, iface.MTU)
		}
	}
}

func TestUptimeInfo(t *testing.T) {
	uptimeInfo, err := getUptimeInfo()
	if err != nil {
		t.Fatalf("Failed to get uptime info: %v", err)
	}

	if uptimeInfo.Seconds <= 0 {
		t.Error("Uptime seconds should be positive")
	}

	if uptimeInfo.Uptime == "" {
		t.Error("Uptime string should not be empty")
	}

	// Boot time should be in the past
	if uptimeInfo.BootTime.After(time.Now()) {
		t.Error("Boot time should be in the past")
	}

	// Validate uptime calculations
	totalMinutes := int(uptimeInfo.Seconds) / 60
	expectedDays := totalMinutes / (24 * 60)
	expectedHours := (totalMinutes % (24 * 60)) / 60
	expectedMins := totalMinutes % 60

	if uptimeInfo.Days != expectedDays {
		t.Errorf("Days calculation incorrect: expected %d, got %d", expectedDays, uptimeInfo.Days)
	}
	if uptimeInfo.Hours != expectedHours {
		t.Errorf("Hours calculation incorrect: expected %d, got %d", expectedHours, uptimeInfo.Hours)
	}
	if uptimeInfo.Minutes != expectedMins {
		t.Errorf("Minutes calculation incorrect: expected %d, got %d", expectedMins, uptimeInfo.Minutes)
	}
}

func TestCPUInfo(t *testing.T) {
	cpuInfo, err := getCPUInfo()
	if err != nil {
		t.Fatalf("Failed to get CPU info: %v", err)
	}

	if cpuInfo.Model == "" {
		t.Error("CPU model should not be empty")
	}

	if cpuInfo.Cores <= 0 {
		t.Error("CPU cores should be positive")
	}

	if cpuInfo.Threads < cpuInfo.Cores {
		t.Error("CPU threads should be >= cores")
	}

	if cpuInfo.MHz <= 0 {
		t.Error("CPU MHz should be positive")
	}
}

// Helper function to calculate absolute difference
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests
func BenchmarkGetSystemInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetSystemInfo()
		if err != nil {
			b.Fatalf("Failed to get system info: %v", err)
		}
	}
}

func BenchmarkGetOSInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := getOSInfo()
		if err != nil {
			b.Fatalf("Failed to get OS info: %v", err)
		}
	}
}

func BenchmarkGetMemoryInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := getMemoryInfo()
		if err != nil {
			b.Fatalf("Failed to get memory info: %v", err)
		}
	}
}

func BenchmarkGetNetworkInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := getNetworkInfo()
		if err != nil {
			b.Fatalf("Failed to get network info: %v", err)
		}
	}
}
