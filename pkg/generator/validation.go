package generator

import (
	"net"
	"strconv"
	"strings"
)

// ValidateFieldValue validates a field value based on its type
func ValidateFieldValue(input string, fieldType string) bool {
	input = strings.TrimSpace(input)
	if input == "" {
		return true // Empty input is valid, will use default later
	}

	switch fieldType {
	case "int", "integer":
		_, err := strconv.Atoi(input)
		return err == nil
	case "bool", "boolean":
		lower := strings.ToLower(input)
		return lower == "true" || lower == "false" || lower == "1" || lower == "0" || lower == "yes" || lower == "no"
	case "float", "double":
		_, err := strconv.ParseFloat(input, 64)
		return err == nil
	case "url":
		// Simple URL validation
		return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
	case "email":
		// Simple email validation
		return strings.Contains(input, "@") && strings.Contains(input, ".") && !strings.Contains(input, " ")
	case "ip":
		// Validate both IPv4 and IPv6
		ip := net.ParseIP(input)
		return ip != nil
	case "ipv4":
		// Validate IPv4 only
		ip := net.ParseIP(input)
		return ip != nil && ip.To4() != nil
	case "ipv6":
		// Validate IPv6 only
		ip := net.ParseIP(input)
		return ip != nil && ip.To4() == nil
	default:
		return true // Default to valid for unknown types
	}
}

// NormalizeFieldValue normalizes a field value based on its type
func NormalizeFieldValue(input string, fieldType string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return input
	}

	switch fieldType {
	case "bool", "boolean":
		lower := strings.ToLower(input)
		if lower == "1" || lower == "yes" {
			return "true"
		} else if lower == "0" || lower == "no" {
			return "false"
		}
	}

	return input
}
