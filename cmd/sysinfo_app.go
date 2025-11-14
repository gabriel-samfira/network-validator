package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"validate/sysinfo"
)

func main() {
	// Command line flags
	port := flag.Int("port", 8080, "Port to run the server on")
	showInfo := flag.Bool("info", false, "Show system information and exit")
	help := flag.Bool("help", false, "Show help message")
	flag.Parse()

	if *help {
		fmt.Println("System Information Web Server")
		fmt.Println("Usage:")
		fmt.Println("  -port int    Port to run the server on (default 8080)")
		fmt.Println("  -info        Show system information and exit")
		fmt.Println("  -help        Show this help message")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run sysinfo_app.go -port 3000")
		fmt.Println("  go run sysinfo_app.go -info")
		return
	}

	if *showInfo {
		// Just show system information and exit
		info, err := sysinfo.GetSystemInfo()
		if err != nil {
			log.Fatalf("Failed to get system info: %v", err)
		}

		fmt.Println("=== System Information ===")
		fmt.Printf("Hostname: %s\n", info.Hostname)
		fmt.Printf("OS: %s %s (%s)\n", info.OS.Name, info.OS.Version, info.OS.Architecture)
		fmt.Printf("Kernel: %s\n", info.OS.Kernel)
		fmt.Printf("CPU: %s (%d cores, %d threads)\n", info.CPU.Model, info.CPU.Cores, info.CPU.Threads)
		fmt.Printf("Memory: %.2f GB total, %.2f GB available (%.1f%% used)\n", 
			info.Memory.TotalGB, info.Memory.AvailableGB, info.Memory.UsedPercent)
		fmt.Printf("Uptime: %s\n", info.Uptime.Uptime)
		
		fmt.Println("\n=== Network Interfaces ===")
		for _, iface := range info.Network.Interfaces {
			if len(iface.IPAddresses) > 0 {
				fmt.Printf("  %s: %v (MAC: %s, MTU: %d)\n", 
					iface.Name, iface.IPAddresses, iface.MACAddress, iface.MTU)
			}
		}
		return
	}

	// Start the web server
	log.Printf("Starting system information web server...")
	
	// Handle graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Shutting down gracefully...")
		os.Exit(0)
	}()

	// Start server
	if err := sysinfo.RunServer(*port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}