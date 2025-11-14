package netplan

import (
	"testing"
)

// Sample netplan YAML configurations for testing
var sampleConfigs = map[string]string{
	"simple-dhcp": `network:
  version: 2
  ethernets:
    enp3s0:
      dhcp4: true`,

	"static-ip": `network:
  version: 2
  ethernets:
    enp3s0:
      addresses:
        - 192.168.1.100/24
      gateway4: 192.168.1.1
      nameservers:
        addresses:
          - 8.8.8.8
          - 8.8.4.4`,

	"wifi": `network:
  version: 2
  wifis:
    wlp2s0:
      dhcp4: true
      access-points:
        "MyNetwork":
          password: "password123"`,

	"bridge": `network:
  version: 2
  ethernets:
    enp3s0:
      dhcp4: false
  bridges:
    br0:
      interfaces: [enp3s0]
      dhcp4: true`,

	"bond": `network:
  version: 2
  ethernets:
    enp3s0:
      dhcp4: false
    enp4s0:
      dhcp4: false
  bonds:
    bond0:
      interfaces: [enp3s0, enp4s0]
      parameters:
        mode: active-backup
      dhcp4: true`,

	"vlan": `network:
  version: 2
  ethernets:
    enp3s0:
      dhcp4: false
  vlans:
    enp3s0.100:
      id: 100
      link: enp3s0
      addresses:
        - 192.168.100.10/24`,

	"complex": `network:
  version: 2
  renderer: networkd
  ethernets:
    enp3s0:
      addresses:
        - 10.0.1.100/24
        - "fd12:3456:789a::100/64"
      gateway4: 10.0.1.1
      gateway6: "fd12:3456:789a::1"
      nameservers:
        addresses:
          - 1.1.1.1
          - 8.8.8.8
        search:
          - example.com
          - local
      routes:
        - to: 172.16.0.0/16
          via: 10.0.1.254
          metric: 100
      mtu: 1500`,
}

func TestLoadConfigFromBytes(t *testing.T) {
	for name, yamlContent := range sampleConfigs {
		t.Run(name, func(t *testing.T) {
			config, err := LoadConfigFromBytes([]byte(yamlContent))
			if err != nil {
				t.Fatalf("Failed to load config %s: %v", name, err)
			}

			if config.Network.Version != 2 {
				t.Errorf("Expected version 2, got %d", config.Network.Version)
			}

			// Test that we can marshal it back
			_, err = config.ToYAML()
			if err != nil {
				t.Errorf("Failed to marshal config back to YAML: %v", err)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid-config",
			config: &Config{
				Network: Network{
					Version: 2,
					Ethernets: map[string]*Ethernet{
						"eth0": NewEthernetDHCP(),
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid-version",
			config: &Config{
				Network: Network{
					Version: 1,
				},
			},
			expectError: true,
		},
		{
			name: "invalid-renderer",
			config: &Config{
				Network: Network{
					Version:  2,
					Renderer: "invalid",
				},
			},
			expectError: true,
		},
		{
			name: "invalid-vlan-id",
			config: &Config{
				Network: Network{
					Version: 2,
					VLANs: map[string]*VLAN{
						"vlan0": {
							CommonInterface: CommonInterface{},
							ID:              5000, // Invalid VLAN ID
							Link:            "eth0",
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.expectError {
				t.Errorf("Expected error: %v, got errors: %v", tt.expectError, errors)
			}
		})
	}
}

func TestBuilders(t *testing.T) {
	config := NewConfig()
	config.Network.Renderer = "networkd"

	// Test ethernet
	eth := NewEthernetDHCP()
	config.AddEthernet("eth0", eth)

	// Test wifi
	wifi := NewWifiWPA("TestSSID", "password")
	config.AddWifi("wlan0", wifi)

	// Test bridge
	bridge := NewBridge([]string{"eth1"})
	config.AddBridge("br0", bridge)

	// Test bond
	bond := NewBond([]string{"eth2", "eth3"}, BondModeActiveBackup)
	config.AddBond("bond0", bond)

	// Test VLAN
	vlan := NewVLAN(100, "eth0")
	config.AddVLAN("eth0.100", vlan)

	// Verify all interfaces were added
	interfaces := config.GetInterfaceNames()
	expectedInterfaces := []string{"eth0", "wlan0", "br0", "bond0", "eth0.100"}

	for _, expected := range expectedInterfaces {
		found := false
		for _, actual := range interfaces {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected interface %s not found in %v", expected, interfaces)
		}
	}

	// Test DHCP detection
	if !config.HasDHCP() {
		t.Error("Expected DHCP to be detected")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test Bool helper
	b := Bool(true)
	if b == nil || *b != true {
		t.Error("Bool helper failed")
	}

	// Test Int helper
	i := Int(42)
	if i == nil || *i != 42 {
		t.Error("Int helper failed")
	}
}

func TestInterfaceTypes(t *testing.T) {
	// Test that we can create all interface types
	config := NewConfig()

	// Ethernet
	config.AddEthernet("eth0", &Ethernet{
		CommonInterface: CommonInterface{
			DHCP4: Bool(true),
		},
	})

	// WiFi with enterprise auth
	config.AddWifi("wlan0", &Wifi{
		CommonInterface: CommonInterface{
			DHCP4: Bool(true),
		},
		AccessPoints: map[string]*AccessPoint{
			"Enterprise": {
				Auth: &Auth{
					KeyManagement: string(KeyManagement8021X),
					Method:        "peap",
					Identity:      "user@example.com",
					Password:      "password",
				},
			},
		},
	})

	// Bridge with parameters
	config.AddBridge("br0", &Bridge{
		CommonInterface: CommonInterface{
			DHCP4: Bool(true),
		},
		Interfaces: []string{"eth1"},
		Parameters: &BridgeParameters{
			STP:      Bool(true),
			Priority: 32768,
		},
	})

	// Bond with advanced parameters
	config.AddBond("bond0", &Bond{
		CommonInterface: CommonInterface{
			DHCP4: Bool(true),
		},
		Interfaces: []string{"eth2", "eth3"},
		Parameters: &BondParameters{
			Mode:               string(BondMode8023AD),
			LACPRate:           "fast",
			MIIMonitorInterval: "100",
		},
	})

	// VLAN
	config.AddVLAN("vlan100", &VLAN{
		CommonInterface: CommonInterface{
			Addresses: []string{"192.168.100.1/24"},
		},
		ID:   100,
		Link: "eth0",
	})

	// Tunnel
	config.AddTunnel("tun0", &Tunnel{
		CommonInterface: CommonInterface{
			Addresses: []string{"10.0.0.1/30"},
		},
		Mode:   string(TunnelModeGRE),
		Local:  "192.168.1.1",
		Remote: "192.168.2.1",
	})

	// Validate the configuration
	errors := config.Validate()
	if len(errors) > 0 {
		for _, err := range errors {
			t.Logf("Validation error: %v", err)
		}
	}

	// Convert to YAML and back
	yamlData, err := config.ToYAML()
	if err != nil {
		t.Fatalf("Failed to convert to YAML: %v", err)
	}

	loadedConfig, err := LoadConfigFromBytes(yamlData)
	if err != nil {
		t.Fatalf("Failed to load from YAML: %v", err)
	}

	// Compare interface counts
	originalInterfaces := config.GetInterfaceNames()
	loadedInterfaces := loadedConfig.GetInterfaceNames()

	if len(originalInterfaces) != len(loadedInterfaces) {
		t.Errorf("Interface count mismatch: original=%d, loaded=%d",
			len(originalInterfaces), len(loadedInterfaces))
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	renderers := []RendererType{RendererNetworkd, RendererNetworkManager}
	if len(renderers) != 2 {
		t.Error("Missing renderer constants")
	}

	tunnelModes := []TunnelMode{TunnelModeGRE, TunnelModeIPIP, TunnelModeIP6IP6,
		TunnelModeIP6GRE, TunnelModeVTI, TunnelModeVTI6, TunnelModeWG}
	if len(tunnelModes) != 7 {
		t.Error("Missing tunnel mode constants")
	}

	bondModes := []BondMode{BondModeRoundRobin, BondModeActiveBackup,
		BondModeBalanceXOR, BondModeBroadcast, BondMode8023AD,
		BondModeBalanceTLB, BondModeBalanceALB}
	if len(bondModes) != 7 {
		t.Error("Missing bond mode constants")
	}

	wifiModes := []WiFiMode{WiFiModeInfrastructure, WiFiModeAdhoc, WiFiModeAP}
	if len(wifiModes) != 3 {
		t.Error("Missing WiFi mode constants")
	}

	keyMgmt := []KeyManagement{KeyManagementNone, KeyManagementPSK,
		KeyManagementEAP, KeyManagement8021X}
	if len(keyMgmt) != 4 {
		t.Error("Missing key management constants")
	}
}

func TestGetBondIPAddresses(t *testing.T) {
	config := NewConfig()

	// Create a bond with IP addresses
	bond := &Bond{
		CommonInterface: CommonInterface{
			Addresses: []string{"10.0.1.100/24", "192.168.1.100/24"},
		},
		Interfaces: []string{"eth0", "eth1"},
	}
	config.AddBond("bond0", bond)

	// Create VLANs on the bond
	vlan100 := &VLAN{
		CommonInterface: CommonInterface{
			Addresses: []string{"10.100.1.1/24"},
		},
		ID:   100,
		Link: "bond0",
	}
	config.AddVLAN("bond0.100", vlan100)

	// Create a bridge that includes the bond
	bridge := &Bridge{
		CommonInterface: CommonInterface{
			Addresses: []string{"172.16.1.1/24"},
		},
		Interfaces: []string{"bond0"},
	}
	config.AddBridge("br0", bridge)

	// Test GetBondIPAddresses
	result := config.GetBondIPAddresses("bond0")

	// Verify bond IP addresses
	bondIPs, exists := result["bond0"]
	if !exists {
		t.Error("Expected bond0 to have IP addresses")
	}
	if len(bondIPs) != 2 {
		t.Errorf("Expected 2 IP addresses for bond0, got %d", len(bondIPs))
	}

	// Verify VLAN IP addresses
	vlanIPs, exists := result["bond0.100"]
	if !exists {
		t.Error("Expected bond0.100 to have IP addresses")
	}
	if len(vlanIPs) != 1 {
		t.Errorf("Expected 1 IP address for bond0.100, got %d", len(vlanIPs))
	}

	// Verify bridge IP addresses
	bridgeIPs, exists := result["br0"]
	if !exists {
		t.Error("Expected br0 to have IP addresses")
	}
	if len(bridgeIPs) != 1 {
		t.Errorf("Expected 1 IP address for br0, got %d", len(bridgeIPs))
	}

	// Test non-existent bond
	emptyResult := config.GetBondIPAddresses("nonexistent")
	if len(emptyResult) != 0 {
		t.Error("Expected empty result for non-existent bond")
	}
}

func TestGetAllBondRelatedInterfaces(t *testing.T) {
	config := NewConfig()

	// Create a bond
	bond := &Bond{
		CommonInterface: CommonInterface{
			Addresses: []string{"10.0.1.100/24"},
		},
		Interfaces: []string{"eth0", "eth1"},
	}
	config.AddBond("bond0", bond)

	// Create VLAN on bond
	vlan := &VLAN{
		CommonInterface: CommonInterface{
			Addresses: []string{"10.100.1.1/24"},
		},
		ID:   100,
		Link: "bond0",
	}
	config.AddVLAN("bond0.100", vlan)

	// Create bridge with bond
	bridge := &Bridge{
		CommonInterface: CommonInterface{
			Addresses: []string{"172.16.1.1/24"},
		},
		Interfaces: []string{"bond0"},
	}
	config.AddBridge("br0", bridge)

	// Test GetAllBondRelatedInterfaces
	interfaces := config.GetAllBondRelatedInterfaces("bond0")

	expectedInterfaces := []string{"bond0", "bond0.100", "br0"}
	if len(interfaces) != len(expectedInterfaces) {
		t.Errorf("Expected %d interfaces, got %d: %v", len(expectedInterfaces), len(interfaces), interfaces)
	}

	// Check that all expected interfaces are present
	for _, expected := range expectedInterfaces {
		found := false
		for _, actual := range interfaces {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected interface %s not found in result: %v", expected, interfaces)
		}
	}

	// Test non-existent bond
	emptyResult := config.GetAllBondRelatedInterfaces("nonexistent")
	if len(emptyResult) != 0 {
		t.Error("Expected empty result for non-existent bond")
	}
}
