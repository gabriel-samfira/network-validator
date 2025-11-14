package netplan

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// stripCIDR removes the CIDR notation (e.g., /24) from an IP address
func stripCIDR(ip string) string {
	if idx := strings.Index(ip, "/"); idx != -1 {
		return ip[:idx]
	}
	return ip
}

// IPWithMask represents an IP address with its subnet mask
type IPWithMask struct {
	IP       string
	CIDR     string // Full CIDR notation (e.g., "10.150.0.1/22")
	IPNet    *net.IPNet
	BondName string
}

// InSameSubnet checks if two IP addresses are in the same subnet
func InSameSubnet(ip1CIDR, ip2 string) bool {
	// Parse the first IP with its CIDR notation
	_, ipNet1, err := net.ParseCIDR(ip1CIDR)
	if err != nil {
		return false
	}

	// Parse the second IP (which might not have CIDR notation)
	ip2Addr := net.ParseIP(stripCIDR(ip2))
	if ip2Addr == nil {
		return false
	}

	// Check if ip2 is in the subnet of ip1
	return ipNet1.Contains(ip2Addr)
}

// LoadConfig loads a netplan configuration from a file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// LoadConfigFromBytes loads a netplan configuration from byte data
func LoadConfigFromBytes(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

// SaveConfig saves a netplan configuration to a file
func SaveConfig(config *Config, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Create directory if it doesn't exist
	if dir := filepath.Dir(filename); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}

// LoadAllNetplanConfigs loads all netplan configuration files from /etc/netplan
func LoadAllNetplanConfigs() ([]*Config, error) {
	return LoadNetplanConfigsFromDir("/etc/netplan")
}

// LoadNetplanConfigsFromDir loads all netplan configuration files from a directory
func LoadNetplanConfigsFromDir(dir string) ([]*Config, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob yaml files in %s: %w", dir, err)
	}

	// Also check for .yml files
	ymlFiles, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob yml files in %s: %w", dir, err)
	}

	files = append(files, ymlFiles...)

	var configs []*Config
	for _, file := range files {
		config, err := LoadConfig(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", file, err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// Validate performs basic validation of the netplan configuration
func (c *Config) Validate() []error {
	var errors []error

	// Check version
	if c.Network.Version != 2 {
		errors = append(errors, fmt.Errorf("unsupported network version: %d (only version 2 is supported)", c.Network.Version))
	}

	// Check renderer
	if c.Network.Renderer != "" {
		validRenderers := []string{"networkd", "NetworkManager"}
		valid := false
		for _, renderer := range validRenderers {
			if c.Network.Renderer == renderer {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, fmt.Errorf("invalid renderer: %s (must be one of: %s)", c.Network.Renderer, strings.Join(validRenderers, ", ")))
		}
	}

	// Validate ethernet interfaces
	for name, eth := range c.Network.Ethernets {
		if err := validateInterfaceName(name); err != nil {
			errors = append(errors, fmt.Errorf("ethernet %s: %w", name, err))
		}
		if errs := validateCommonInterface(&eth.CommonInterface); len(errs) > 0 {
			for _, err := range errs {
				errors = append(errors, fmt.Errorf("ethernet %s: %w", name, err))
			}
		}
	}

	// Validate wifi interfaces
	for name, wifi := range c.Network.Wifis {
		if err := validateInterfaceName(name); err != nil {
			errors = append(errors, fmt.Errorf("wifi %s: %w", name, err))
		}
		if errs := validateCommonInterface(&wifi.CommonInterface); len(errs) > 0 {
			for _, err := range errs {
				errors = append(errors, fmt.Errorf("wifi %s: %w", name, err))
			}
		}
	}

	// Validate bridges
	for name, bridge := range c.Network.Bridges {
		if err := validateInterfaceName(name); err != nil {
			errors = append(errors, fmt.Errorf("bridge %s: %w", name, err))
		}
		if errs := validateCommonInterface(&bridge.CommonInterface); len(errs) > 0 {
			for _, err := range errs {
				errors = append(errors, fmt.Errorf("bridge %s: %w", name, err))
			}
		}
	}

	// Validate bonds
	for name, bond := range c.Network.Bonds {
		if err := validateInterfaceName(name); err != nil {
			errors = append(errors, fmt.Errorf("bond %s: %w", name, err))
		}
		if errs := validateCommonInterface(&bond.CommonInterface); len(errs) > 0 {
			for _, err := range errs {
				errors = append(errors, fmt.Errorf("bond %s: %w", name, err))
			}
		}
	}

	// Validate VLANs
	for name, vlan := range c.Network.VLANs {
		if err := validateInterfaceName(name); err != nil {
			errors = append(errors, fmt.Errorf("vlan %s: %w", name, err))
		}
		if vlan.ID < 1 || vlan.ID > 4094 {
			errors = append(errors, fmt.Errorf("vlan %s: invalid VLAN ID %d (must be 1-4094)", name, vlan.ID))
		}
		if vlan.Link == "" {
			errors = append(errors, fmt.Errorf("vlan %s: link is required", name))
		}
	}

	return errors
}

// validateInterfaceName validates interface names
func validateInterfaceName(name string) error {
	if name == "" {
		return fmt.Errorf("interface name cannot be empty")
	}
	if len(name) > 15 {
		return fmt.Errorf("interface name too long: %s (max 15 characters)", name)
	}
	return nil
}

// validateCommonInterface validates common interface properties
func validateCommonInterface(iface *CommonInterface) []error {
	var errors []error

	// Validate addresses
	for _, addr := range iface.Addresses {
		if !strings.Contains(addr, "/") {
			errors = append(errors, fmt.Errorf("address %s must include subnet mask", addr))
		}
	}

	// Validate MTU
	if iface.MTU != 0 && (iface.MTU < 68 || iface.MTU > 65536) {
		errors = append(errors, fmt.Errorf("invalid MTU %d (must be 68-65536)", iface.MTU))
	}

	return errors
}

// GetInterfaceNames returns all interface names defined in the configuration
func (c *Config) GetInterfaceNames() []string {
	var names []string

	for name := range c.Network.Ethernets {
		names = append(names, name)
	}
	for name := range c.Network.Wifis {
		names = append(names, name)
	}
	for name := range c.Network.Bridges {
		names = append(names, name)
	}
	for name := range c.Network.Bonds {
		names = append(names, name)
	}
	for name := range c.Network.VLANs {
		names = append(names, name)
	}
	for name := range c.Network.Tunnels {
		names = append(names, name)
	}
	for name := range c.Network.VRFs {
		names = append(names, name)
	}
	for name := range c.Network.Modems {
		names = append(names, name)
	}

	return names
}

// HasDHCP returns true if any interface is configured for DHCP
func (c *Config) HasDHCP() bool {
	checkDHCP := func(iface *CommonInterface) bool {
		return (iface.DHCP4 != nil && *iface.DHCP4) || (iface.DHCP6 != nil && *iface.DHCP6)
	}

	for _, eth := range c.Network.Ethernets {
		if checkDHCP(&eth.CommonInterface) {
			return true
		}
	}
	for _, wifi := range c.Network.Wifis {
		if checkDHCP(&wifi.CommonInterface) {
			return true
		}
	}
	for _, bridge := range c.Network.Bridges {
		if checkDHCP(&bridge.CommonInterface) {
			return true
		}
	}
	for _, bond := range c.Network.Bonds {
		if checkDHCP(&bond.CommonInterface) {
			return true
		}
	}

	return false
}

// ToYAML converts the configuration to YAML format
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	data, err := c.ToYAML()
	if err != nil {
		return fmt.Sprintf("Error marshaling config: %v", err)
	}
	return string(data)
}

// GetBondIPAddresses returns a map of interface names to their IP addresses
// for all interfaces that involve the specified bond (including VLANs and bridges)
func (c *Config) GetBondIPAddresses(bondName string) map[string][]string {
	result := make(map[string][]string)

	// Check if the bond exists
	bond, exists := c.Network.Bonds[bondName]
	if !exists {
		return result
	}

	// 1. Get IP addresses directly assigned to the bond
	if len(bond.Addresses) > 0 {
		for _, addr := range bond.Addresses {
			result[bondName] = append(result[bondName], stripCIDR(addr))
		}
	}

	// 2. Find VLANs that use this bond as their link
	for vlanName, vlan := range c.Network.VLANs {
		if vlan.Link == bondName && len(vlan.Addresses) > 0 {
			for _, addr := range vlan.Addresses {
				result[vlanName] = append(result[vlanName], stripCIDR(addr))
			}
		}
	}

	// 3. Find bridges that include this bond in their interfaces
	for bridgeName, bridge := range c.Network.Bridges {
		for _, iface := range bridge.Interfaces {
			if iface == bondName && len(bridge.Addresses) > 0 {
				for _, addr := range bridge.Addresses {
					result[bridgeName] = append(result[bridgeName], stripCIDR(addr))
				}
				break
			}
		}
	}

	// 4. Find VLANs that are built on top of bridges that include this bond
	for vlanName, vlan := range c.Network.VLANs {
		if bridge, bridgeExists := c.Network.Bridges[vlan.Link]; bridgeExists {
			// Check if this bridge includes our bond
			for _, iface := range bridge.Interfaces {
				if iface == bondName && len(vlan.Addresses) > 0 {
					for _, addr := range vlan.Addresses {
						result[vlanName] = append(result[vlanName], stripCIDR(addr))
					}
					break
				}
			}
		}
	}

	// 5. Find bridges that include VLANs built on top of this bond
	for bridgeName, bridge := range c.Network.Bridges {
		for _, iface := range bridge.Interfaces {
			if vlan, vlanExists := c.Network.VLANs[iface]; vlanExists {
				if vlan.Link == bondName && len(bridge.Addresses) > 0 {
					for _, addr := range bridge.Addresses {
						result[bridgeName] = append(result[bridgeName], stripCIDR(addr))
					}
					break
				}
			}
		}
	}

	// 6. Handle nested scenarios: find tunnels that might use the bond or its derived interfaces
	for tunnelName, tunnel := range c.Network.Tunnels {
		if len(tunnel.Addresses) > 0 {
			// Check if tunnel references the bond directly or any of its derived interfaces
			if c.isBondRelated(bondName, tunnelName) {
				for _, addr := range tunnel.Addresses {
					result[tunnelName] = append(result[tunnelName], stripCIDR(addr))
				}
			}
		}
	}

	return result
}

// GetBondIPAddressesWithMask returns bond IP addresses with their CIDR notation intact
// This is used for subnet matching when testing connectivity
// Includes IPs from: bond itself, VLANs on bond, bridges with bond, VLANs on bridges, tunnels
func (c *Config) GetBondIPAddressesWithMask(bondName string) []IPWithMask {
	var result []IPWithMask

	// Check if the bond exists
	bond, exists := c.Network.Bonds[bondName]
	if !exists {
		return result
	}

	// Helper to create IPWithMask from address
	addIPWithMask := func(addr, interfaceName string) {
		_, ipNet, err := net.ParseCIDR(addr)
		if err != nil {
			return
		}
		result = append(result, IPWithMask{
			IP:       stripCIDR(addr),
			CIDR:     addr,
			IPNet:    ipNet,
			BondName: interfaceName,
		})
	}

	// 1. Get IP addresses directly assigned to the bond
	for _, addr := range bond.Addresses {
		addIPWithMask(addr, bondName)
	}

	// 2. Find VLANs that use this bond as their link
	for vlanName, vlan := range c.Network.VLANs {
		if vlan.Link == bondName {
			for _, addr := range vlan.Addresses {
				addIPWithMask(addr, vlanName)
			}
		}
	}

	// 3. Find bridges that include this bond in their interfaces
	for bridgeName, bridge := range c.Network.Bridges {
		for _, iface := range bridge.Interfaces {
			if iface == bondName {
				for _, addr := range bridge.Addresses {
					addIPWithMask(addr, bridgeName)
				}
				break
			}
		}
	}

	// 4. Find VLANs that are built on top of bridges that include this bond
	for vlanName, vlan := range c.Network.VLANs {
		if bridge, bridgeExists := c.Network.Bridges[vlan.Link]; bridgeExists {
			for _, iface := range bridge.Interfaces {
				if iface == bondName {
					for _, addr := range vlan.Addresses {
						addIPWithMask(addr, vlanName)
					}
					break
				}
			}
		}
	}

	// 5. Find bridges that include VLANs built on top of this bond
	for bridgeName, bridge := range c.Network.Bridges {
		for _, iface := range bridge.Interfaces {
			if vlan, vlanExists := c.Network.VLANs[iface]; vlanExists {
				if vlan.Link == bondName {
					for _, addr := range bridge.Addresses {
						addIPWithMask(addr, bridgeName)
					}
					break
				}
			}
		}
	}

	// 6. Handle nested scenarios: find tunnels that might use the bond or its derived interfaces
	for tunnelName, tunnel := range c.Network.Tunnels {
		if len(tunnel.Addresses) > 0 {
			// Check if tunnel references the bond directly or any of its derived interfaces
			if c.isBondRelated(bondName, tunnelName) {
				for _, addr := range tunnel.Addresses {
					addIPWithMask(addr, tunnelName)
				}
			}
		}
	}

	return result
}

// isBondRelated checks if an interface name is related to the specified bond
// This includes the bond itself, VLANs on the bond, bridges containing the bond, etc.
func (c *Config) isBondRelated(bondName, interfaceName string) bool {
	// Direct match
	if interfaceName == bondName {
		return true
	}

	// Check if it's a VLAN on the bond
	if vlan, exists := c.Network.VLANs[interfaceName]; exists {
		if vlan.Link == bondName {
			return true
		}
		// Check if it's a VLAN on a bridge that contains the bond
		if bridge, bridgeExists := c.Network.Bridges[vlan.Link]; bridgeExists {
			for _, iface := range bridge.Interfaces {
				if iface == bondName {
					return true
				}
			}
		}
	}

	// Check if it's a bridge that contains the bond
	if bridge, exists := c.Network.Bridges[interfaceName]; exists {
		for _, iface := range bridge.Interfaces {
			if iface == bondName {
				return true
			}
		}
	}

	return false
}

// GetAllBondRelatedInterfaces returns all interface names that are related to the specified bond
func (c *Config) GetAllBondRelatedInterfaces(bondName string) []string {
	var interfaces []string

	// Check if the bond exists
	if _, exists := c.Network.Bonds[bondName]; !exists {
		return interfaces
	}

	// Add the bond itself
	interfaces = append(interfaces, bondName)

	// Find VLANs that use this bond
	for vlanName, vlan := range c.Network.VLANs {
		if vlan.Link == bondName {
			interfaces = append(interfaces, vlanName)
		}
	}

	// Find bridges that include this bond
	for bridgeName, bridge := range c.Network.Bridges {
		for _, iface := range bridge.Interfaces {
			if iface == bondName {
				interfaces = append(interfaces, bridgeName)
				break
			}
		}
	}

	// Find VLANs on bridges that include this bond
	for vlanName, vlan := range c.Network.VLANs {
		if bridge, bridgeExists := c.Network.Bridges[vlan.Link]; bridgeExists {
			for _, iface := range bridge.Interfaces {
				if iface == bondName {
					interfaces = append(interfaces, vlanName)
					break
				}
			}
		}
	}

	// Find tunnels that might reference bond-related interfaces
	for tunnelName := range c.Network.Tunnels {
		if c.isBondRelated(bondName, tunnelName) {
			interfaces = append(interfaces, tunnelName)
		}
	}

	return interfaces
}

// GetBondIPAddresses loads netplan configs from a directory and returns
// a map of bond names to their associated IP addresses
func GetBondIPAddresses(netplanDir string) (map[string][]string, error) {
	configs, err := LoadNetplanConfigsFromDir(netplanDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load netplan configs: %w", err)
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

// GetBondIPAddressesWithMask loads netplan configs from a directory and returns
// all IP addresses with their CIDR notation for subnet matching
func GetBondIPAddressesWithMask(netplanDir string) (map[string][]IPWithMask, error) {
	configs, err := LoadNetplanConfigsFromDir(netplanDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load netplan configs: %w", err)
	}

	allBonds := make(map[string][]IPWithMask)

	// Iterate through all configs and all bonds
	for _, config := range configs {
		if config.Network.Bonds == nil {
			continue
		}

		for bondName := range config.Network.Bonds {
			bondIPs := config.GetBondIPAddressesWithMask(bondName)
			if len(bondIPs) > 0 {
				allBonds[bondName] = bondIPs
			}
		}
	}

	return allBonds, nil
}
