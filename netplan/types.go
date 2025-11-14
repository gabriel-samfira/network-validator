package netplan

// Config represents the root netplan configuration
type Config struct {
	Network Network `yaml:"network"`
}

// Network represents the main network configuration block
type Network struct {
	Version   int                  `yaml:"version"`
	Renderer  string               `yaml:"renderer,omitempty"`
	Ethernets map[string]*Ethernet `yaml:"ethernets,omitempty"`
	Wifis     map[string]*Wifi     `yaml:"wifis,omitempty"`
	Bridges   map[string]*Bridge   `yaml:"bridges,omitempty"`
	Bonds     map[string]*Bond     `yaml:"bonds,omitempty"`
	VLANs     map[string]*VLAN     `yaml:"vlans,omitempty"`
	Tunnels   map[string]*Tunnel   `yaml:"tunnels,omitempty"`
	VRFs      map[string]*VRF      `yaml:"vrfs,omitempty"`
	Modems    map[string]*Modem    `yaml:"modems,omitempty"`
}

// CommonInterface contains common network interface properties
type CommonInterface struct {
	// Basic configuration
	DHCP4          *bool           `yaml:"dhcp4,omitempty"`
	DHCP6          *bool           `yaml:"dhcp6,omitempty"`
	IPv6Privacy    *bool           `yaml:"ipv6-privacy,omitempty"`
	LinkLocal      []string        `yaml:"link-local,omitempty"`
	Critical       *bool           `yaml:"critical,omitempty"`
	DHCPIdentifier string          `yaml:"dhcp-identifier,omitempty"`
	DHCP4Overrides *DHCP4Overrides `yaml:"dhcp4-overrides,omitempty"`
	DHCP6Overrides *DHCP6Overrides `yaml:"dhcp6-overrides,omitempty"`
	AcceptRA       *bool           `yaml:"accept-ra,omitempty"`

	// Address configuration
	Addresses   []string     `yaml:"addresses,omitempty"`
	Gateway4    string       `yaml:"gateway4,omitempty"`
	Gateway6    string       `yaml:"gateway6,omitempty"`
	Nameservers *Nameservers `yaml:"nameservers,omitempty"`
	MacAddress  string       `yaml:"macaddress,omitempty"`
	MTU         int          `yaml:"mtu,omitempty"`

	// Advanced configuration
	Optional       *bool           `yaml:"optional,omitempty"`
	ActivationMode string          `yaml:"activation-mode,omitempty"`
	Routes         []Route         `yaml:"routes,omitempty"`
	RoutingPolicy  []RoutingPolicy `yaml:"routing-policy,omitempty"`
	Neigh          []Neighbor      `yaml:"neigh,omitempty"`

	// Matching and device selection
	Match     *Match `yaml:"match,omitempty"`
	SetName   string `yaml:"set-name,omitempty"`
	WakeOnLan *bool  `yaml:"wakeonlan,omitempty"`

	// SR-IOV configuration
	EmbeddedSwitch string `yaml:"embedded-switch,omitempty"`
	SRIOV          *SRIOV `yaml:"sriov,omitempty"`

	// OpenVSwitch configuration
	OpenVSwitch *OpenVSwitch `yaml:"openvswitch,omitempty"`
}

// Ethernet represents ethernet interface configuration
type Ethernet struct {
	CommonInterface `yaml:",inline"`

	// Ethernet-specific configuration
	Link            string           `yaml:"link,omitempty"`
	VirtualFunction *VirtualFunction `yaml:"virtual-function,omitempty"`
}

// Wifi represents wireless interface configuration
type Wifi struct {
	CommonInterface `yaml:",inline"`

	// WiFi-specific configuration
	AccessPoints map[string]*AccessPoint `yaml:"access-points,omitempty"`
	Regulatory   string                  `yaml:"regulatory-domain,omitempty"`
}

// AccessPoint represents a WiFi access point configuration
type AccessPoint struct {
	Password    string `yaml:"password,omitempty"`
	Auth        *Auth  `yaml:"auth,omitempty"`
	Mode        string `yaml:"mode,omitempty"`
	Band        string `yaml:"band,omitempty"`
	Channel     int    `yaml:"channel,omitempty"`
	BSSID       string `yaml:"bssid,omitempty"`
	Hidden      *bool  `yaml:"hidden,omitempty"`
	NetworkName string `yaml:"networkname,omitempty"`
}

// Auth represents authentication configuration for WiFi
type Auth struct {
	KeyManagement     string `yaml:"key-management,omitempty"`
	Method            string `yaml:"method,omitempty"`
	Identity          string `yaml:"identity,omitempty"`
	AnonymousIdentity string `yaml:"anonymous-identity,omitempty"`
	Password          string `yaml:"password,omitempty"`
	CACertificate     string `yaml:"ca-certificate,omitempty"`
	ClientCertificate string `yaml:"client-certificate,omitempty"`
	ClientKey         string `yaml:"client-key,omitempty"`
	ClientKeyPassword string `yaml:"client-key-password,omitempty"`
	Phase2Auth        string `yaml:"phase2-auth,omitempty"`
}

// Bridge represents bridge interface configuration
type Bridge struct {
	CommonInterface `yaml:",inline"`

	// Bridge-specific configuration
	Interfaces []string          `yaml:"interfaces,omitempty"`
	Parameters *BridgeParameters `yaml:"parameters,omitempty"`
}

// BridgeParameters represents bridge-specific parameters
type BridgeParameters struct {
	AgeingTime   int   `yaml:"ageing-time,omitempty"`
	Priority     int   `yaml:"priority,omitempty"`
	PortPriority int   `yaml:"port-priority,omitempty"`
	ForwardDelay int   `yaml:"forward-delay,omitempty"`
	HelloTime    int   `yaml:"hello-time,omitempty"`
	MaxAge       int   `yaml:"max-age,omitempty"`
	PathCost     int   `yaml:"path-cost,omitempty"`
	STP          *bool `yaml:"stp,omitempty"`
}

// Bond represents bond interface configuration
type Bond struct {
	CommonInterface `yaml:",inline"`

	// Bond-specific configuration
	Interfaces []string        `yaml:"interfaces,omitempty"`
	Parameters *BondParameters `yaml:"parameters,omitempty"`
}

// BondParameters represents bond-specific parameters
type BondParameters struct {
	Mode                  string   `yaml:"mode,omitempty"`
	LACPRate              string   `yaml:"lacp-rate,omitempty"`
	MIIMonitorInterval    string   `yaml:"mii-monitor-interval,omitempty"`
	MinLinks              int      `yaml:"min-links,omitempty"`
	TransmitHashPolicy    string   `yaml:"transmit-hash-policy,omitempty"`
	ADSelect              string   `yaml:"ad-select,omitempty"`
	AllSlavesActive       *bool    `yaml:"all-slaves-active,omitempty"`
	ARPInterval           string   `yaml:"arp-interval,omitempty"`
	ARPIPTargets          []string `yaml:"arp-ip-targets,omitempty"`
	ARPValidate           string   `yaml:"arp-validate,omitempty"`
	ARPAllTargets         string   `yaml:"arp-all-targets,omitempty"`
	UpDelay               string   `yaml:"up-delay,omitempty"`
	DownDelay             string   `yaml:"down-delay,omitempty"`
	FailOverMac           string   `yaml:"fail-over-mac,omitempty"`
	GratuitousARP         int      `yaml:"gratuitous-arp,omitempty"`
	PacketsPerSlave       int      `yaml:"packets-per-slave,omitempty"`
	PrimaryReselectPolicy string   `yaml:"primary-reselect-policy,omitempty"`
	ResendIGMP            int      `yaml:"resend-igmp,omitempty"`
	LearnPacketInterval   string   `yaml:"learn-packet-interval,omitempty"`
	Primary               string   `yaml:"primary,omitempty"`
}

// VLAN represents VLAN interface configuration
type VLAN struct {
	CommonInterface `yaml:",inline"`

	// VLAN-specific configuration
	ID   int    `yaml:"id"`
	Link string `yaml:"link"`
}

// Tunnel represents tunnel interface configuration
type Tunnel struct {
	CommonInterface `yaml:",inline"`

	// Tunnel-specific configuration
	Mode   string `yaml:"mode"`
	Local  string `yaml:"local,omitempty"`
	Remote string `yaml:"remote,omitempty"`
	Key    string `yaml:"key,omitempty"`
	Keys   *Keys  `yaml:"keys,omitempty"`
	TTL    int    `yaml:"ttl,omitempty"`
	TOS    int    `yaml:"tos,omitempty"`
	PMTU   int    `yaml:"pmtu-discovery,omitempty"`
}

// Keys represents tunnel key configuration
type Keys struct {
	Input  string `yaml:"input,omitempty"`
	Output string `yaml:"output,omitempty"`
}

// VRF represents VRF (Virtual Routing and Forwarding) configuration
type VRF struct {
	// VRF configuration
	Table         int             `yaml:"table"`
	Interfaces    []string        `yaml:"interfaces,omitempty"`
	Routes        []Route         `yaml:"routes,omitempty"`
	RoutingPolicy []RoutingPolicy `yaml:"routing-policy,omitempty"`
}

// Modem represents modem interface configuration
type Modem struct {
	CommonInterface `yaml:",inline"`

	// Modem-specific configuration
	APN           string `yaml:"apn,omitempty"`
	AutoConfig    *bool  `yaml:"auto-config,omitempty"`
	DeviceID      string `yaml:"device-id,omitempty"`
	NetworkID     string `yaml:"network-id,omitempty"`
	Number        string `yaml:"number,omitempty"`
	Password      string `yaml:"password,omitempty"`
	PIN           string `yaml:"pin,omitempty"`
	SimID         string `yaml:"sim-id,omitempty"`
	SimOperatorID string `yaml:"sim-operator-id,omitempty"`
	Username      string `yaml:"username,omitempty"`
}

// DHCP4Overrides represents DHCP4 override configuration
type DHCP4Overrides struct {
	UseDNS      *bool  `yaml:"use-dns,omitempty"`
	UseDomains  string `yaml:"use-domains,omitempty"`
	UseHostname *bool  `yaml:"use-hostname,omitempty"`
	UseMTU      *bool  `yaml:"use-mtu,omitempty"`
	UseNTP      *bool  `yaml:"use-ntp,omitempty"`
	UseRoutes   *bool  `yaml:"use-routes,omitempty"`
	Hostname    string `yaml:"hostname,omitempty"`
	RouteMetric int    `yaml:"route-metric,omitempty"`
}

// DHCP6Overrides represents DHCP6 override configuration
type DHCP6Overrides struct {
	UseDNS      *bool  `yaml:"use-dns,omitempty"`
	UseDomains  string `yaml:"use-domains,omitempty"`
	UseHostname *bool  `yaml:"use-hostname,omitempty"`
	UseMTU      *bool  `yaml:"use-mtu,omitempty"`
	UseNTP      *bool  `yaml:"use-ntp,omitempty"`
	Hostname    string `yaml:"hostname,omitempty"`
}

// Nameservers represents DNS nameserver configuration
type Nameservers struct {
	Search    []string `yaml:"search,omitempty"`
	Addresses []string `yaml:"addresses,omitempty"`
}

// Route represents a network route
type Route struct {
	To               string `yaml:"to,omitempty"`
	Via              string `yaml:"via,omitempty"`
	From             string `yaml:"from,omitempty"`
	OnLink           *bool  `yaml:"on-link,omitempty"`
	Metric           int    `yaml:"metric,omitempty"`
	Type             string `yaml:"type,omitempty"`
	Scope            string `yaml:"scope,omitempty"`
	Table            int    `yaml:"table,omitempty"`
	MTU              int    `yaml:"mtu,omitempty"`
	CongestionWindow int    `yaml:"congestion-window,omitempty"`
	AdvertisedMSS    int    `yaml:"advertised-mss,omitempty"`
}

// RoutingPolicy represents routing policy configuration
type RoutingPolicy struct {
	From          string `yaml:"from,omitempty"`
	To            string `yaml:"to,omitempty"`
	Table         int    `yaml:"table,omitempty"`
	Priority      int    `yaml:"priority,omitempty"`
	Mark          int    `yaml:"mark,omitempty"`
	TypeOfService int    `yaml:"type-of-service,omitempty"`
}

// Neighbor represents neighbor/ARP configuration
type Neighbor struct {
	To  string `yaml:"to"`
	MAC string `yaml:"macaddress"`
}

// Match represents interface matching criteria
type Match struct {
	Name       string `yaml:"name,omitempty"`
	MacAddress string `yaml:"macaddress,omitempty"`
	Driver     string `yaml:"driver,omitempty"`
	Path       string `yaml:"path,omitempty"`
}

// SRIOV represents SR-IOV configuration
type SRIOV struct {
	TotalVFs int                  `yaml:"total-vfs,omitempty"`
	VFTable  map[string]*VFConfig `yaml:"vf-table,omitempty"`
}

// VFConfig represents virtual function configuration
type VFConfig struct {
	ID         int    `yaml:"id"`
	MacAddress string `yaml:"macaddress,omitempty"`
	VLAN       int    `yaml:"vlan,omitempty"`
	QoS        int    `yaml:"qos,omitempty"`
	SpoofCheck *bool  `yaml:"spoof-check,omitempty"`
	Trust      *bool  `yaml:"trust,omitempty"`
	LinkState  string `yaml:"link-state,omitempty"`
}

// VirtualFunction represents virtual function assignment
type VirtualFunction struct {
	Link string `yaml:"link"`
}

// OpenVSwitch represents Open vSwitch configuration
type OpenVSwitch struct {
	ExternalIDs         map[string]string `yaml:"external-ids,omitempty"`
	OtherConfig         map[string]string `yaml:"other-config,omitempty"`
	Lacp                string            `yaml:"lacp,omitempty"`
	FailMode            string            `yaml:"fail-mode,omitempty"`
	McastSnoopingEnable *bool             `yaml:"mcast-snooping-enable,omitempty"`
	Protocols           []string          `yaml:"protocols,omitempty"`
	RSTPEnable          *bool             `yaml:"rstp-enable,omitempty"`
	Controller          *Controller       `yaml:"controller,omitempty"`
	Ports               [][]interface{}   `yaml:"ports,omitempty"`
	SSL                 *SSL              `yaml:"ssl,omitempty"`
}

// Controller represents OpenVSwitch controller configuration
type Controller struct {
	Addresses      []string `yaml:"addresses,omitempty"`
	ConnectionMode string   `yaml:"connection-mode,omitempty"`
}

// SSL represents SSL configuration for OpenVSwitch
type SSL struct {
	CAFile   string `yaml:"ca-file,omitempty"`
	CertFile string `yaml:"cert-file,omitempty"`
	KeyFile  string `yaml:"key-file,omitempty"`
}

// RendererType represents the network renderer type
type RendererType string

const (
	RendererNetworkd       RendererType = "networkd"
	RendererNetworkManager RendererType = "NetworkManager"
)

// TunnelMode represents tunnel mode types
type TunnelMode string

const (
	TunnelModeGRE    TunnelMode = "gre"
	TunnelModeIPIP   TunnelMode = "ipip"
	TunnelModeIP6IP6 TunnelMode = "ip6ip6"
	TunnelModeIP6GRE TunnelMode = "ip6gre"
	TunnelModeVTI    TunnelMode = "vti"
	TunnelModeVTI6   TunnelMode = "vti6"
	TunnelModeWG     TunnelMode = "wireguard"
)

// BondMode represents bond mode types
type BondMode string

const (
	BondModeRoundRobin   BondMode = "balance-rr"
	BondModeActiveBackup BondMode = "active-backup"
	BondModeBalanceXOR   BondMode = "balance-xor"
	BondModeBroadcast    BondMode = "broadcast"
	BondMode8023AD       BondMode = "802.3ad"
	BondModeBalanceTLB   BondMode = "balance-tlb"
	BondModeBalanceALB   BondMode = "balance-alb"
)

// WiFiMode represents WiFi mode types
type WiFiMode string

const (
	WiFiModeInfrastructure WiFiMode = "infrastructure"
	WiFiModeAdhoc          WiFiMode = "adhoc"
	WiFiModeAP             WiFiMode = "ap"
)

// KeyManagement represents WiFi key management types
type KeyManagement string

const (
	KeyManagementNone  KeyManagement = "none"
	KeyManagementPSK   KeyManagement = "psk"
	KeyManagementEAP   KeyManagement = "eap"
	KeyManagement8021X KeyManagement = "802.1x"
)
