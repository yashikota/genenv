package generator

import (
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
		// Simple IP validation
		parts := strings.Split(input, ".")
		if len(parts) != 4 {
			return false
		}
		for _, part := range parts {
			num, err := strconv.Atoi(part)
			if err != nil || num < 0 || num > 255 {
				return false
			}
		}
		return true
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
