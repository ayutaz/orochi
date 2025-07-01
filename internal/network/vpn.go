package network

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ayutaz/orochi/internal/errors"
	"github.com/ayutaz/orochi/internal/logger"
)

// VPNConfig represents VPN binding configuration.
type VPNConfig struct {
	Enabled       bool   `json:"enabled"`
	InterfaceName string `json:"interface_name"`
	KillSwitch    bool   `json:"kill_switch"`
}

// Interface represents a network interface.
type Interface struct {
	Name      string   `json:"name"`
	Index     int      `json:"index"`
	Addresses []string `json:"addresses"`
	IsVPN     bool     `json:"is_vpn"`
	IsUp      bool     `json:"is_up"`
}

// Monitor monitors network interfaces and VPN status.
type Monitor struct {
	config        *VPNConfig
	logger        logger.Logger
	checkInterval time.Duration
	stopCh        chan struct{}
	mu            sync.RWMutex
	lastCheck     time.Time
	vpnActive     bool
}

// NewVPNConfig creates a new VPN configuration with defaults.
func NewVPNConfig() *VPNConfig {
	return &VPNConfig{
		Enabled:    false,
		KillSwitch: true, // Enable kill switch by default for safety
	}
}

// Validate validates the VPN configuration.
func (c *VPNConfig) Validate() error {
	if c.Enabled && c.InterfaceName == "" {
		return errors.ValidationErrorf("interface name is required when VPN binding is enabled")
	}
	return nil
}

// GetNetworkInterfaces returns all network interfaces.
func GetNetworkInterfaces() ([]Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var result []Interface
	for _, iface := range interfaces {
		ni := Interface{
			Name:  iface.Name,
			Index: iface.Index,
			IsVPN: IsVPNInterface(iface.Name),
			IsUp:  iface.Flags&net.FlagUp != 0,
		}

		// Get addresses
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				ni.Addresses = append(ni.Addresses, addr.String())
			}
		}

		result = append(result, ni)
	}

	return result, nil
}

// IsVPNInterface checks if an interface name looks like a VPN interface.
func IsVPNInterface(name string) bool {
	// Common VPN interface prefixes
	vpnPrefixes := []string{
		"tun",      // OpenVPN, WireGuard
		"tap",      // OpenVPN
		"wg",       // WireGuard
		"ppp",      // PPTP, L2TP
		"ipsec",    // IPSec
		"vpn",      // Generic VPN
		"nordlynx", // NordVPN
		"proton",   // ProtonVPN
		"mullvad",  // Mullvad
	}

	lowerName := strings.ToLower(name)
	for _, prefix := range vpnPrefixes {
		if strings.HasPrefix(lowerName, prefix) {
			return true
		}
	}

	return false
}

// GetVPNInterfaces returns only VPN interfaces.
func GetVPNInterfaces() ([]Interface, error) {
	interfaces, err := GetNetworkInterfaces()
	if err != nil {
		return nil, err
	}

	var vpnInterfaces []Interface
	for _, iface := range interfaces {
		if iface.IsVPN {
			vpnInterfaces = append(vpnInterfaces, iface)
		}
	}

	return vpnInterfaces, nil
}

// BindToInterface binds network connections to a specific interface.
// This is a placeholder - actual implementation would be platform-specific.
func BindToInterface(interfaceName string) error {
	if interfaceName == "" {
		return errors.ValidationErrorf("interface name cannot be empty")
	}

	// Check if interface exists
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return errors.NotFoundf("interface %q not found", interfaceName)
	}

	// Check if interface is up
	if iface.Flags&net.FlagUp == 0 {
		return errors.ValidationErrorf("interface %q is not up", interfaceName)
	}

	// Note: Actual binding would require platform-specific code
	// and integration with the torrent client library
	return nil
}

// NewNetworkMonitor creates a new network monitor.
func NewNetworkMonitor(config *VPNConfig) *Monitor {
	monitor := &Monitor{
		config:        config,
		checkInterval: 5 * time.Second,
		stopCh:        make(chan struct{}),
	}
	// Set initial state
	monitor.checkVPNStatus()
	return monitor
}

// SetLogger sets the logger for the network monitor.
func (m *Monitor) SetLogger(log logger.Logger) {
	m.logger = log
}

// Start starts monitoring network interfaces.
func (m *Monitor) Start() {
	go m.run()
}

// Stop stops the network monitor.
func (m *Monitor) Stop() {
	close(m.stopCh)
}

// run is the main monitoring loop.
func (m *Monitor) run() {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	// Initial check
	m.checkVPNStatus()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkVPNStatus()
		}
	}
}

// checkVPNStatus checks if the VPN interface is active.
func (m *Monitor) checkVPNStatus() {
	if !m.config.Enabled {
		m.setVPNActive(true) // VPN check disabled, always "active"
		return
	}

	// Check if the configured interface exists and is up
	iface, err := net.InterfaceByName(m.config.InterfaceName)
	if err != nil {
		m.setVPNActive(false)
		if m.logger != nil {
			m.logger.Warn("VPN interface not found",
				logger.String("interface", m.config.InterfaceName),
				logger.Err(err),
			)
		}
		return
	}

	isUp := iface.Flags&net.FlagUp != 0
	m.setVPNActive(isUp)

	if !isUp && m.logger != nil {
		m.logger.Warn("VPN interface is down",
			logger.String("interface", m.config.InterfaceName),
		)
	}
}

// setVPNActive updates the VPN active status.
func (m *Monitor) setVPNActive(active bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.vpnActive = active
	m.lastCheck = time.Now()
}

// IsVPNActive returns whether the VPN is currently active.
func (m *Monitor) IsVPNActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.vpnActive
}

// GetLastCheck returns the time of the last VPN check.
func (m *Monitor) GetLastCheck() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.lastCheck
}

// ShouldAllowConnection checks if a connection should be allowed based on VPN status.
func (m *Monitor) ShouldAllowConnection() bool {
	if !m.config.Enabled {
		return true // VPN binding disabled, allow all
	}

	if !m.config.KillSwitch {
		return true // Kill switch disabled, allow all
	}

	return m.IsVPNActive()
}
