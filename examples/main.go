package main

import (
	"fmt"
	"log"

	"validate/netplan"
)

func main() {
	// Example 1: Simple DHCP configuration
	fmt.Println("=== Example 1: Simple DHCP Configuration ===")
	config1 := netplan.NewConfig()
	config1.Network.Renderer = "networkd"
	config1.AddEthernet("enp3s0", netplan.NewEthernetDHCP())

	yaml1, _ := config1.ToYAML()
	fmt.Println(string(yaml1))

	// Example 2: Static IP configuration
	fmt.Println("\n=== Example 2: Static IP Configuration ===")
	config2 := netplan.NewConfig()
	config2.Network.Renderer = "networkd"

	staticEth := netplan.NewEthernetStatic(
		[]string{"192.168.1.100/24"},
		"192.168.1.1",
		"",
		[]string{"8.8.8.8", "8.8.4.4"},
	)
	config2.AddEthernet("enp3s0", staticEth)

	yaml2, _ := config2.ToYAML()
	fmt.Println(string(yaml2))

	// Example 3: WiFi configuration
	fmt.Println("\n=== Example 3: WiFi Configuration ===")
	config3 := netplan.NewConfig()
	config3.Network.Renderer = "NetworkManager"

	wifi := netplan.NewWifiWPA("MyHomeWiFi", "mypassword123")
	config3.AddWifi("wlp2s0", wifi)

	yaml3, _ := config3.ToYAML()
	fmt.Println(string(yaml3))

	// Example 4: Bridge configuration
	fmt.Println("\n=== Example 4: Bridge Configuration ===")
	config4 := netplan.NewConfig()
	config4.Network.Renderer = "networkd"

	// Add ethernet interface without DHCP (will be part of bridge)
	ethForBridge := &netplan.Ethernet{
		CommonInterface: netplan.CommonInterface{
			DHCP4: netplan.Bool(false),
		},
	}
	config4.AddEthernet("enp3s0", ethForBridge)

	// Add bridge
	bridge := netplan.NewBridge([]string{"enp3s0"})
	config4.AddBridge("br0", bridge)

	yaml4, _ := config4.ToYAML()
	fmt.Println(string(yaml4))

	// Example 5: Bond configuration
	fmt.Println("\n=== Example 5: Bond Configuration ===")
	config5 := netplan.NewConfig()
	config5.Network.Renderer = "networkd"

	// Add ethernet interfaces for bonding
	eth1 := &netplan.Ethernet{
		CommonInterface: netplan.CommonInterface{
			DHCP4: netplan.Bool(false),
		},
	}
	eth2 := &netplan.Ethernet{
		CommonInterface: netplan.CommonInterface{
			DHCP4: netplan.Bool(false),
		},
	}
	config5.AddEthernet("enp3s0", eth1)
	config5.AddEthernet("enp4s0", eth2)

	// Add bond
	bond := netplan.NewBond([]string{"enp3s0", "enp4s0"}, netplan.BondModeActiveBackup)
	config5.AddBond("bond0", bond)

	yaml5, _ := config5.ToYAML()
	fmt.Println(string(yaml5))

	// Example 6: VLAN configuration
	fmt.Println("\n=== Example 6: VLAN Configuration ===")
	config6 := netplan.NewConfig()
	config6.Network.Renderer = "networkd"

	// Add base ethernet interface
	baseEth := &netplan.Ethernet{
		CommonInterface: netplan.CommonInterface{
			DHCP4: netplan.Bool(false),
		},
	}
	config6.AddEthernet("enp3s0", baseEth)

	// Add VLAN
	vlan := netplan.NewVLAN(100, "enp3s0")
	vlan.Addresses = []string{"192.168.100.10/24"}
	vlan.DHCP4 = netplan.Bool(false)
	config6.AddVLAN("enp3s0.100", vlan)

	yaml6, _ := config6.ToYAML()
	fmt.Println(string(yaml6))

	// Example 7: Complex configuration with routes and nameservers
	fmt.Println("\n=== Example 7: Complex Configuration ===")
	config7 := netplan.NewConfig()
	config7.Network.Renderer = "networkd"

	complexEth := &netplan.Ethernet{
		CommonInterface: netplan.CommonInterface{
			Addresses: []string{"10.0.1.100/24", "fd12:3456:789a::100/64"},
			Gateway4:  "10.0.1.1",
			Gateway6:  "fd12:3456:789a::1",
			Nameservers: &netplan.Nameservers{
				Addresses: []string{"1.1.1.1", "8.8.8.8"},
				Search:    []string{"example.com", "local"},
			},
			Routes: []netplan.Route{
				{
					To:     "172.16.0.0/16",
					Via:    "10.0.1.254",
					Metric: 100,
				},
			},
			MTU: 1500,
		},
	}
	config7.AddEthernet("enp3s0", complexEth)

	yaml7, _ := config7.ToYAML()
	fmt.Println(string(yaml7))

	// Example 8: Loading and validating configuration
	fmt.Println("\n=== Example 8: Validation ===")

	// Save a configuration to a file
	filename := "/tmp/01-netcfg.yaml"
	if err := netplan.SaveConfig(config1, filename); err != nil {
		log.Printf("Error saving config: %v", err)
	} else {
		fmt.Printf("Configuration saved to %s\n", filename)
	}

	// Load it back
	loadedConfig, err := netplan.LoadConfig(filename)
	if err != nil {
		log.Printf("Error loading config: %v", err)
	} else {
		fmt.Println("Configuration loaded successfully")

		// Validate the configuration
		if errors := loadedConfig.Validate(); len(errors) > 0 {
			fmt.Println("Validation errors:")
			for _, err := range errors {
				fmt.Printf("  - %v\n", err)
			}
		} else {
			fmt.Println("Configuration is valid")
		}

		// Show interface names
		fmt.Printf("Interfaces defined: %v\n", loadedConfig.GetInterfaceNames())
		fmt.Printf("Has DHCP interfaces: %v\n", loadedConfig.HasDHCP())
	}
}
