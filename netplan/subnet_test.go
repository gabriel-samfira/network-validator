package netplan

import "testing"

func TestInSameSubnet(t *testing.T) {
	tests := []struct {
		name     string
		ip1CIDR  string
		ip2      string
		expected bool
	}{
		{
			name:     "Same /22 subnet - within range",
			ip1CIDR:  "10.150.0.1/22",
			ip2:      "10.150.0.2",
			expected: true,
		},
		{
			name:     "Same /22 subnet - at upper boundary",
			ip1CIDR:  "10.150.0.1/22",
			ip2:      "10.150.3.253",
			expected: true,
		},
		{
			name:     "Different /22 subnet",
			ip1CIDR:  "10.150.0.1/22",
			ip2:      "10.150.4.1",
			expected: false,
		},
		{
			name:     "Same /24 subnet",
			ip1CIDR:  "192.168.1.10/24",
			ip2:      "192.168.1.20",
			expected: true,
		},
		{
			name:     "Different /24 subnet",
			ip1CIDR:  "192.168.1.10/24",
			ip2:      "192.168.2.10",
			expected: false,
		},
		{
			name:     "IP2 with CIDR notation",
			ip1CIDR:  "10.150.0.1/22",
			ip2:      "10.150.0.2/22",
			expected: true,
		},
		{
			name:     "Same /16 subnet",
			ip1CIDR:  "172.16.0.1/16",
			ip2:      "172.16.255.254",
			expected: true,
		},
		{
			name:     "Different /16 subnet",
			ip1CIDR:  "172.16.0.1/16",
			ip2:      "172.17.0.1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InSameSubnet(tt.ip1CIDR, tt.ip2)
			if result != tt.expected {
				t.Errorf("InSameSubnet(%s, %s) = %v, want %v", tt.ip1CIDR, tt.ip2, result, tt.expected)
			}
		})
	}
}
