package network

import (
	"net"
	"testing"
)

func TestLocalIP(t *testing.T) {
	// Test getting IPv4 address
	ipv4, err := LocalIP(IPv4)
	if err != nil {
		t.Logf("Failed to get IPv4: %v", err)
	} else {
		t.Logf("Got IPv4: %v", ipv4)
		if ipv4.To4() == nil {
			t.Errorf("Expected IPv4 address, got: %v", ipv4)
		}
	}

	// Test getting IPv6 address
	// Note: This test might be skipped if no IPv6 is available
	ipv6, err := LocalIP(IPv6)
	if err != nil {
		t.Logf("Failed to get IPv6 (this might be normal if your system doesn't have IPv6): %v", err)
	} else {
		t.Logf("Got IPv6: %v", ipv6)
		if ipv6.To4() != nil {
			t.Errorf("Expected IPv6 address, got: %v", ipv6)
		}
	}

	// Test getting any IP
	ipAny, err := LocalIP(IPAny)
	if err != nil {
		t.Errorf("Failed to get IP: %v", err)
	} else {
		t.Logf("Got IP: %v", ipAny)
	}
}

func TestIsPrivateIP(t *testing.T) {
	testCases := []struct {
		ip       string
		expected bool
	}{
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"8.8.8.8", false},
		{"fd00::1", true},
		{"2001:4860:4860::8888", false}, // Google's public DNS
	}

	for _, tc := range testCases {
		ip := net.ParseIP(tc.ip)
		result := isPrivateIP(ip)
		if result != tc.expected {
			t.Errorf("isPrivateIP(%s) = %v, expected %v", tc.ip, result, tc.expected)
		}
	}
}
