package network

import (
	"net"
	"testing"
)

func TestVPNBinding(t *testing.T) {
	t.Run("GetNetworkInterfaces", func(t *testing.T) {
		interfaces, err := GetNetworkInterfaces()
		if err != nil {
			t.Fatalf("Failed to get network interfaces: %v", err)
		}

		// Should have at least one interface (loopback)
		if len(interfaces) == 0 {
			t.Error("No network interfaces found")
		}

		// Check that each interface has required fields
		for _, iface := range interfaces {
			if iface.Name == "" {
				t.Error("Interface has empty name")
			}
			if iface.Index == 0 {
				t.Error("Interface has zero index")
			}
		}
	})

	t.Run("IsVPNInterface", func(t *testing.T) {
		tests := []struct {
			name     string
			ifaceName string
			expected bool
		}{
			{"TAP interface", "tap0", true},
			{"TUN interface", "tun0", true},
			{"WireGuard interface", "wg0", true},
			{"OpenVPN TAP", "tap-vpn", true},
			{"OpenVPN TUN", "tun-vpn", true},
			{"PPTP interface", "ppp0", true},
			{"L2TP interface", "ppp1", true},
			{"Ethernet interface", "eth0", false},
			{"WiFi interface", "wlan0", false},
			{"Loopback interface", "lo", false},
			{"Docker interface", "docker0", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsVPNInterface(tt.ifaceName)
				if result != tt.expected {
					t.Errorf("IsVPNInterface(%q) = %v, want %v", tt.ifaceName, result, tt.expected)
				}
			})
		}
	})

	t.Run("GetVPNInterfaces", func(t *testing.T) {
		interfaces, err := GetVPNInterfaces()
		if err != nil {
			t.Fatalf("Failed to get VPN interfaces: %v", err)
		}

		// Check that all returned interfaces are VPN interfaces
		for _, iface := range interfaces {
			if !IsVPNInterface(iface.Name) {
				t.Errorf("Non-VPN interface returned: %s", iface.Name)
			}
		}
	})

	t.Run("BindToInterface", func(t *testing.T) {
		// This test requires root privileges and an actual interface
		// So we'll test error handling instead
		
		// Test with invalid interface name
		err := BindToInterface("nonexistent-interface")
		if err == nil {
			t.Error("Expected error for non-existent interface")
		}

		// Test with empty interface name
		err = BindToInterface("")
		if err == nil {
			t.Error("Expected error for empty interface name")
		}
	})
}

func TestVPNConfig(t *testing.T) {
	t.Run("NewVPNConfig", func(t *testing.T) {
		config := NewVPNConfig()
		
		if config.Enabled {
			t.Error("VPN should be disabled by default")
		}
		
		if config.InterfaceName != "" {
			t.Error("Interface name should be empty by default")
		}
		
		if !config.KillSwitch {
			t.Error("Kill switch should be enabled by default for safety")
		}
	})

	t.Run("Validate", func(t *testing.T) {
		tests := []struct {
			name    string
			config  *VPNConfig
			wantErr bool
		}{
			{
				name: "Valid disabled config",
				config: &VPNConfig{
					Enabled: false,
				},
				wantErr: false,
			},
			{
				name: "Valid enabled config",
				config: &VPNConfig{
					Enabled:       true,
					InterfaceName: "tun0",
				},
				wantErr: false,
			},
			{
				name: "Invalid - enabled without interface",
				config: &VPNConfig{
					Enabled:       true,
					InterfaceName: "",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.config.Validate()
				if (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

func TestNetworkMonitor(t *testing.T) {
	t.Run("NewNetworkMonitor", func(t *testing.T) {
		config := &VPNConfig{
			Enabled:       true,
			InterfaceName: "tun0",
			KillSwitch:    true,
		}
		
		monitor := NewNetworkMonitor(config)
		
		if monitor == nil {
			t.Fatal("NewNetworkMonitor returned nil")
		}
		
		if monitor.config != config {
			t.Error("Config not set correctly")
		}
		
		if monitor.checkInterval == 0 {
			t.Error("Check interval not set")
		}
	})

	t.Run("IsVPNActive", func(t *testing.T) {
		config := &VPNConfig{
			Enabled:       false,
			InterfaceName: "",
		}
		
		monitor := NewNetworkMonitor(config)
		
		// When VPN is disabled, should always return true
		if !monitor.IsVPNActive() {
			t.Error("IsVPNActive should return true when VPN is disabled")
		}
		
		// Enable VPN binding
		config.Enabled = true
		config.InterfaceName = "nonexistent-vpn"
		
		// Create new monitor with updated config
		monitor2 := NewNetworkMonitor(config)
		
		// Should return false for non-existent interface
		if monitor2.IsVPNActive() {
			t.Error("IsVPNActive should return false for non-existent interface")
		}
	})
}

// Helper function to check if running with sufficient privileges
func hasPrivileges() bool {
	// Try to create a raw socket (requires privileges)
	conn, err := net.Dial("ip4:icmp", "127.0.0.1")
	if err != nil {
		return false
	}
	conn.Close()
	return true
}