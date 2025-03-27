package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGenerator(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templatePath := filepath.Join(tempDir, ".env.example")
	templateContent := `
# Test configuration
TEST_VALUE=fixed
TEST_SECRET=${secret}
ANOTHER_SECRET=${another_secret}
FIXED_VALUE=1234
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Create generator
	outputPath := filepath.Join(tempDir, ".env")
	config := Config{
		TemplatePath: templatePath,
		OutputPath:   outputPath,
		Force:        true,
	}
	gen := New(config)

	// Generate .env file
	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate .env file: %v", err)
	}

	// Read generated file
	generatedContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Check generated content
	lines := strings.Split(string(generatedContent), "\n")

	// Create a map to store key-value pairs
	envVars := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envVars[key] = value
	}

	// Check fixed values are preserved
	if envVars["TEST_VALUE"] != "fixed" {
		t.Errorf("Fixed value was not preserved, got %s, want %s", envVars["TEST_VALUE"], "fixed")
	}

	if envVars["FIXED_VALUE"] != "1234" {
		t.Errorf("Fixed value was not preserved, got %s, want %s", envVars["FIXED_VALUE"], "1234")
	}

	// Check placeholders were replaced
	if envVars["TEST_SECRET"] == "${secret}" {
		t.Error("Placeholder was not replaced")
	}

	if envVars["ANOTHER_SECRET"] == "${another_secret}" {
		t.Error("Placeholder was not replaced")
	}

	// Check that generated values are different
	if envVars["TEST_SECRET"] == envVars["ANOTHER_SECRET"] {
		t.Error("Generated values should be different")
	}

	// Check that generated values look like secure random values (alphanumeric only)
	secureValuePattern := regexp.MustCompile(`^[A-Za-z0-9]{24}$`)
	if !secureValuePattern.MatchString(envVars["TEST_SECRET"]) {
		t.Errorf("Generated value doesn't match expected pattern: %s", envVars["TEST_SECRET"])
	}

	if !secureValuePattern.MatchString(envVars["ANOTHER_SECRET"]) {
		t.Errorf("Generated value doesn't match expected pattern: %s", envVars["ANOTHER_SECRET"])
	}
}

func TestGeneratorExistingFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templatePath := filepath.Join(tempDir, ".env.example")
	templateContent := `TEST_SECRET=${secret}`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Create an existing output file
	outputPath := filepath.Join(tempDir, ".env")
	existingContent := `TEST_SECRET=existing_value`
	if err := os.WriteFile(outputPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Create generator without force flag
	config := Config{
		TemplatePath: templatePath,
		OutputPath:   outputPath,
		Force:        false,
	}
	gen := New(config)

	// Generate should fail because file exists
	if err := gen.Generate(); err == nil {
		t.Fatal("Generate should fail when output file exists and force is false")
	}

	// Create generator with force flag
	config.Force = true
	gen = New(config)

	// Generate should succeed
	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate .env file with force flag: %v", err)
	}

	// Read generated file
	generatedContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Check generated content
	if string(generatedContent) == existingContent {
		t.Error("Output file was not overwritten")
	}
}

func TestGeneratorComplexTemplate(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file with more complex patterns
	templatePath := filepath.Join(tempDir, ".env.example")
	templateContent := `
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=${db_user}
DB_PASSWORD=${db_password}

# API configuration
API_KEY=${api_key}
API_URL=https://api.example.com

# Empty values and comments
EMPTY_VALUE=
# Comment line
COMMENTED_VALUE=${commented_value} # Inline comment
MULTIPLE_PLACEHOLDERS=${first}_${second}

# Special cases
QUOTED_VALUE="quoted ${quoted_value}"
DOLLAR_SIGN=$$
ESCAPED_PLACEHOLDER=\${not_a_placeholder}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Create generator
	outputPath := filepath.Join(tempDir, ".env")
	config := Config{
		TemplatePath: templatePath,
		OutputPath:   outputPath,
		Force:        true,
	}
	gen := New(config)

	// Generate .env file
	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate .env file: %v", err)
	}

	// Read generated file
	generatedContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Parse generated content
	lines := strings.Split(string(generatedContent), "\n")
	envVars := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle inline comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envVars[key] = value
	}

	// Test fixed values
	if envVars["DB_HOST"] != "localhost" {
		t.Errorf("Fixed value was not preserved, got %s, want %s", envVars["DB_HOST"], "localhost")
	}

	if envVars["DB_PORT"] != "5432" {
		t.Errorf("Fixed value was not preserved, got %s, want %s", envVars["DB_PORT"], "5432")
	}

	// Test placeholder replacements
	secureValuePattern := regexp.MustCompile(`^[A-Za-z0-9]{24}$`)

	if !secureValuePattern.MatchString(envVars["DB_USER"]) {
		t.Errorf("DB_USER doesn't match expected pattern: %s", envVars["DB_USER"])
	}

	if !secureValuePattern.MatchString(envVars["DB_PASSWORD"]) {
		t.Errorf("DB_PASSWORD doesn't match expected pattern: %s", envVars["DB_PASSWORD"])
	}

	if !secureValuePattern.MatchString(envVars["API_KEY"]) {
		t.Errorf("API_KEY doesn't match expected pattern: %s", envVars["API_KEY"])
	}

	// Test special cases
	if envVars["EMPTY_VALUE"] != "" {
		t.Errorf("Empty value was not preserved, got %s, want empty string", envVars["EMPTY_VALUE"])
	}

	// Check that escaped placeholder wasn't replaced
	if envVars["ESCAPED_PLACEHOLDER"] != "${not_a_placeholder}" {
		t.Errorf("Escaped placeholder was incorrectly replaced: %s", envVars["ESCAPED_PLACEHOLDER"])
	}

	// Check that dollar sign was preserved
	if envVars["DOLLAR_SIGN"] != "$$" {
		t.Errorf("Dollar sign was not preserved, got %s, want $$", envVars["DOLLAR_SIGN"])
	}

	// Check that quoted placeholder was replaced
	if strings.Contains(envVars["QUOTED_VALUE"], "${quoted_value}") {
		t.Errorf("Quoted placeholder was not replaced: %s", envVars["QUOTED_VALUE"])
	}
}

func TestGeneratorErrorCases(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Non-existent template file
	config := Config{
		TemplatePath: filepath.Join(tempDir, "non-existent-file"),
		OutputPath:   filepath.Join(tempDir, ".env"),
		Force:        true,
	}
	gen := New(config)

	if err := gen.Generate(); err == nil {
		t.Error("Generate should fail with non-existent template file")
	}

	// Test case 2: Invalid output directory
	config = Config{
		TemplatePath: filepath.Join(tempDir, ".env.example"),
		OutputPath:   filepath.Join(tempDir, "non-existent-dir", ".env"),
		Force:        true,
	}

	// Create a valid template file
	templateContent := `KEY=${value}`
	if err := os.WriteFile(config.TemplatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	gen = New(config)
	if err := gen.Generate(); err == nil {
		t.Error("Generate should fail with invalid output directory")
	}

	// Test case 3: Template file with invalid syntax (unclosed placeholder)
	invalidTemplatePath := filepath.Join(tempDir, ".env.invalid")
	invalidTemplateContent := `KEY=${unclosed`
	if err := os.WriteFile(invalidTemplatePath, []byte(invalidTemplateContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid template file: %v", err)
	}

	config = Config{
		TemplatePath: invalidTemplatePath,
		OutputPath:   filepath.Join(tempDir, ".env"),
		Force:        true,
	}
	gen = New(config)

	// This should still work, as we're just looking for ${...} patterns
	if err := gen.Generate(); err != nil {
		t.Errorf("Generate failed with invalid template syntax: %v", err)
	}
}

func TestGeneratorCustomValueLength(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templatePath := filepath.Join(tempDir, ".env.example")
	templateContent := `
SECRET_KEY=${secret_key}
API_TOKEN=${api_token}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Test different value lengths
	testLengths := []int{8, 16, 32, 64}

	for _, length := range testLengths {
		t.Run(fmt.Sprintf("Length_%d", length), func(t *testing.T) {
			// Create generator with custom value length
			outputPath := filepath.Join(tempDir, fmt.Sprintf(".env.length_%d", length))
			config := Config{
				TemplatePath: templatePath,
				OutputPath:   outputPath,
				Force:        true,
				ValueLength:  length,
				Charset:      CharsetAlphanumeric,
			}
			gen := New(config)

			// Generate .env file
			if err := gen.Generate(); err != nil {
				t.Fatalf("Failed to generate .env file: %v", err)
			}

			// Read generated file
			generatedContent, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}

			// Parse generated content
			lines := strings.Split(string(generatedContent), "\n")
			envVars := make(map[string]string)

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}

				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				envVars[key] = value
			}

			// Check that values have the correct length
			for _, value := range envVars {
				if len(value) != length {
					t.Errorf("Generated value length is %d, want %d: %s", len(value), length, value)
				}
			}
		})
	}
}

func TestGeneratorCustomCharset(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templatePath := filepath.Join(tempDir, ".env.example")
	templateContent := `
SECRET_KEY=${secret_key}
API_TOKEN=${api_token}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Test different charsets
	testCases := []struct {
		charset CharsetType
		pattern *regexp.Regexp
	}{
		{
			charset: CharsetAlphanumeric,
			pattern: regexp.MustCompile(`^[A-Za-z0-9]+$`),
		},
		{
			charset: CharsetAlphabetic,
			pattern: regexp.MustCompile(`^[A-Za-z]+$`),
		},
		{
			charset: CharsetUppercase,
			pattern: regexp.MustCompile(`^[A-Z]+$`),
		},
		{
			charset: CharsetLowercase,
			pattern: regexp.MustCompile(`^[a-z]+$`),
		},
		{
			charset: CharsetNumeric,
			pattern: regexp.MustCompile(`^[0-9]+$`),
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.charset), func(t *testing.T) {
			// Create generator with custom charset
			outputPath := filepath.Join(tempDir, fmt.Sprintf(".env.%s", tc.charset))
			config := Config{
				TemplatePath: templatePath,
				OutputPath:   outputPath,
				Force:        true,
				ValueLength:  16,
				Charset:      tc.charset,
			}
			gen := New(config)

			// Generate .env file
			if err := gen.Generate(); err != nil {
				t.Fatalf("Failed to generate .env file: %v", err)
			}

			// Read generated file
			generatedContent, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}

			// Parse generated content
			lines := strings.Split(string(generatedContent), "\n")
			envVars := make(map[string]string)

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}

				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				envVars[key] = value
			}

			// Check that values match the expected pattern
			for key, value := range envVars {
				if !tc.pattern.MatchString(value) {
					t.Errorf("Generated value for %s doesn't match expected pattern for charset %s: %s", key, tc.charset, value)
				}
			}
		})
	}
}

func TestFieldValidation(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		input     string
		valid     bool
	}{
		// String validation
		{"Valid string", "string", "hello", true},
		{"Empty string", "string", "", true},
		
		// Integer validation
		{"Valid integer", "int", "123", true},
		{"Valid negative integer", "int", "-123", true},
		{"Invalid integer (float)", "int", "123.45", false},
		{"Invalid integer (text)", "int", "abc", false},
		{"Invalid integer (mixed)", "int", "123abc", false},
		
		// Boolean validation
		{"Valid boolean (true)", "bool", "true", true},
		{"Valid boolean (false)", "bool", "false", true},
		{"Valid boolean (yes)", "bool", "yes", true},
		{"Valid boolean (no)", "bool", "no", true},
		{"Valid boolean (1)", "bool", "1", true},
		{"Valid boolean (0)", "bool", "0", true},
		{"Invalid boolean", "bool", "maybe", false},
		
		// Float validation
		{"Valid float", "float", "123.45", true},
		{"Valid float (integer)", "float", "123", true},
		{"Valid float (negative)", "float", "-123.45", true},
		{"Invalid float", "float", "abc", false},
		{"Invalid float (mixed)", "float", "123.45abc", false},
		
		// URL validation
		{"Valid URL (http)", "url", "http://example.com", true},
		{"Valid URL (https)", "url", "https://example.com/path?query=value", true},
		{"Invalid URL (no protocol)", "url", "example.com", false},
		{"Invalid URL (wrong protocol)", "url", "ftp://example.com", false},
		
		// Email validation
		{"Valid email", "email", "user@example.com", true},
		{"Valid email (subdomain)", "email", "user@sub.example.com", true},
		{"Invalid email (no @)", "email", "userexample.com", false},
		{"Invalid email (no domain)", "email", "user@", false},
		{"Invalid email (spaces)", "email", "user @example.com", false},
		
		// IP validation
		{"Valid IP", "ip", "192.168.1.1", true},
		{"Valid IP (zeros)", "ip", "0.0.0.0", true},
		{"Valid IP (max)", "ip", "255.255.255.255", true},
		{"Invalid IP (out of range)", "ip", "256.256.256.256", false},
		{"Invalid IP (wrong format)", "ip", "192.168.1", false},
		{"Invalid IP (letters)", "ip", "192.168.1.a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateFieldValue(tt.input, tt.fieldType)
			
			if result != tt.valid {
				t.Errorf("ValidateFieldValue(%q, %q) = %v, want %v", tt.input, tt.fieldType, result, tt.valid)
			}
		})
	}
}

func TestNormalizeFieldValue(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		input     string
		expected  string
	}{
		{"Boolean yes", "bool", "yes", "true"},
		{"Boolean 1", "bool", "1", "true"},
		{"Boolean no", "bool", "no", "false"},
		{"Boolean 0", "bool", "0", "false"},
		{"Boolean true", "bool", "true", "true"},
		{"Boolean false", "bool", "false", "false"},
		{"Non-boolean", "string", "test", "test"},
		{"Empty string", "bool", "", ""},
		{"Whitespace", "bool", "  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeFieldValue(tt.input, tt.fieldType)
			
			if result != tt.expected {
				t.Errorf("NormalizeFieldValue(%q, %q) = %q, want %q", tt.input, tt.fieldType, result, tt.expected)
			}
		})
	}
}

func TestParseTemplateMetadata(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file with metadata
	templatePath := filepath.Join(tempDir, ".env.example")
	templateContent := `
# @db_password [required] (string) Database password
DB_PASSWORD=${db_password}

# @db_port [optional] (int) Database port
DB_PORT=${db_port}

# @debug_mode [optional] (bool) Enable debug mode
DEBUG=${debug_mode}

# @rate_limit [optional] (float) API rate limit
RATE_LIMIT=${rate_limit}

# @api_url [optional] (url) API URL
API_URL=${api_url}

# @admin_email [optional] (email) Admin email
ADMIN_EMAIL=${admin_email}

# @server_ip [optional] (ip) Server IP address
SERVER_IP=${server_ip}

# No metadata
SIMPLE_VALUE=${simple_value}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Create generator
	config := Config{
		TemplatePath: templatePath,
		OutputPath:   filepath.Join(tempDir, ".env"),
		Force:        true,
	}
	gen := New(config)

	// Parse metadata
	fields, err := gen.ParseTemplateMetadata()
	if err != nil {
		t.Fatalf("Failed to parse template metadata: %v", err)
	}

	// Check parsed metadata
	expectedFields := map[string]struct {
		key      string
		required bool
		fieldType string
		description string
	}{
		"db_password": {"DB_PASSWORD", true, "string", "Database password"},
		"db_port":     {"DB_PORT", false, "int", "Database port"},
		"debug_mode":  {"DEBUG", false, "bool", "Enable debug mode"},
		"rate_limit":  {"RATE_LIMIT", false, "float", "API rate limit"},
		"api_url":     {"API_URL", false, "url", "API URL"},
		"admin_email": {"ADMIN_EMAIL", false, "email", "Admin email"},
		"server_ip":   {"SERVER_IP", false, "ip", "Server IP address"},
		"simple_value": {"SIMPLE_VALUE", false, "string", ""},
	}

	// Check that all expected fields are present
	for name, expected := range expectedFields {
		field, ok := fields[name]
		if !ok {
			t.Errorf("Expected field %s not found", name)
			continue
		}

		if field.Key != expected.key {
			t.Errorf("Field %s: expected key %s, got %s", name, expected.key, field.Key)
		}

		if field.Required != expected.required {
			t.Errorf("Field %s: expected required %v, got %v", name, expected.required, field.Required)
		}

		if field.Type != expected.fieldType {
			t.Errorf("Field %s: expected type %s, got %s", name, expected.fieldType, field.Type)
		}

		if field.Description != expected.description {
			t.Errorf("Field %s: expected description %s, got %s", name, expected.description, field.Description)
		}
	}

	// Check that there are no unexpected fields
	if len(fields) != len(expectedFields) {
		t.Errorf("Expected %d fields, got %d", len(expectedFields), len(fields))
	}
}
