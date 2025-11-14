package netplan

// NewConfig creates a new netplan configuration with default values
func NewConfig() *Config {
	return &Config{
		Network: Network{
			Version: 2,
		},
	}
}

// AddEthernet adds an ethernet interface configuration
func (c *Config) AddEthernet(name string, config *Ethernet) {
	if c.Network.Ethernets == nil {
		c.Network.Ethernets = make(map[string]*Ethernet)
	}
	c.Network.Ethernets[name] = config
}

// AddWifi adds a wifi interface configuration
func (c *Config) AddWifi(name string, config *Wifi) {
	if c.Network.Wifis == nil {
		c.Network.Wifis = make(map[string]*Wifi)
	}
	c.Network.Wifis[name] = config
}

// AddBridge adds a bridge interface configuration
func (c *Config) AddBridge(name string, config *Bridge) {
	if c.Network.Bridges == nil {
		c.Network.Bridges = make(map[string]*Bridge)
	}
	c.Network.Bridges[name] = config
}

// AddBond adds a bond interface configuration
func (c *Config) AddBond(name string, config *Bond) {
	if c.Network.Bonds == nil {
		c.Network.Bonds = make(map[string]*Bond)
	}
	c.Network.Bonds[name] = config
}

// AddVLAN adds a VLAN interface configuration
func (c *Config) AddVLAN(name string, config *VLAN) {
	if c.Network.VLANs == nil {
		c.Network.VLANs = make(map[string]*VLAN)
	}
	c.Network.VLANs[name] = config
}

// AddTunnel adds a tunnel interface configuration
func (c *Config) AddTunnel(name string, config *Tunnel) {
	if c.Network.Tunnels == nil {
		c.Network.Tunnels = make(map[string]*Tunnel)
	}
	c.Network.Tunnels[name] = config
}

// Helper functions for creating common configurations

// NewEthernetDHCP creates an ethernet interface with DHCP configuration
func NewEthernetDHCP() *Ethernet {
	dhcp4 := true
	dhcp6 := true
	return &Ethernet{
		CommonInterface: CommonInterface{
			DHCP4: &dhcp4,
			DHCP6: &dhcp6,
		},
	}
}

// NewEthernetStatic creates an ethernet interface with static IP configuration
func NewEthernetStatic(addresses []string, gateway4, gateway6 string, nameservers []string) *Ethernet {
	eth := &Ethernet{
		CommonInterface: CommonInterface{
			Addresses: addresses,
		},
	}

	if gateway4 != "" {
		eth.Gateway4 = gateway4
	}
	if gateway6 != "" {
		eth.Gateway6 = gateway6
	}

	if len(nameservers) > 0 {
		eth.Nameservers = &Nameservers{
			Addresses: nameservers,
		}
	}

	return eth
}

// NewWifiWPA creates a WiFi interface with WPA/WPA2 configuration
func NewWifiWPA(ssid, password string) *Wifi {
	dhcp4 := true
	return &Wifi{
		CommonInterface: CommonInterface{
			DHCP4: &dhcp4,
		},
		AccessPoints: map[string]*AccessPoint{
			ssid: {
				Password: password,
			},
		},
	}
}

// NewBridge creates a bridge interface
func NewBridge(interfaces []string) *Bridge {
	dhcp4 := true
	return &Bridge{
		CommonInterface: CommonInterface{
			DHCP4: &dhcp4,
		},
		Interfaces: interfaces,
	}
}

// NewBond creates a bond interface
func NewBond(interfaces []string, mode BondMode) *Bond {
	dhcp4 := true
	return &Bond{
		CommonInterface: CommonInterface{
			DHCP4: &dhcp4,
		},
		Interfaces: interfaces,
		Parameters: &BondParameters{
			Mode: string(mode),
		},
	}
}

// NewVLAN creates a VLAN interface
func NewVLAN(id int, link string) *VLAN {
	dhcp4 := true
	return &VLAN{
		CommonInterface: CommonInterface{
			DHCP4: &dhcp4,
		},
		ID:   id,
		Link: link,
	}
}

// Bool is a helper function to create a pointer to a boolean value
func Bool(b bool) *bool {
	return &b
}

// Int is a helper function to create a pointer to an integer value
func Int(i int) *int {
	return &i
}
