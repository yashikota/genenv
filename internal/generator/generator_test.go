package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// parseEnvFile is a test helper to parse env file content into a map
func parseEnvFile(content string) map[string]string {
	envVars := make(map[string]string)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle inline comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
			line = strings.TrimSpace(line)
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envVars[key] = value
		}
	}

	return envVars
}

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

	// Read and parse generated file
	generatedContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	envVars := parseEnvFile(string(generatedContent))

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
	templateContent := `TEST_SECRET=${secret}
NEW_VALUE=${new_value}`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Create an existing output file
	outputPath := filepath.Join(tempDir, ".env")
	existingContent := `TEST_SECRET=existing_value`
	if err := os.WriteFile(outputPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Test 1: Without force flag, existing values should be preserved
	config := Config{
		TemplatePath: templatePath,
		OutputPath:   outputPath,
		Force:        false,
	}
	gen := New(config)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate .env file: %v", err)
	}

	generatedContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	envVars := parseEnvFile(string(generatedContent))

	// Existing value should be preserved
	if envVars["TEST_SECRET"] != "existing_value" {
		t.Errorf("Existing value was not preserved, got %s, want existing_value", envVars["TEST_SECRET"])
	}

	// New value should be generated
	if envVars["NEW_VALUE"] == "" || envVars["NEW_VALUE"] == "${new_value}" {
		t.Error("New value was not generated")
	}

	// Test 2: With force flag, all values should be regenerated
	config.Force = true
	gen = New(config)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate .env file with force flag: %v", err)
	}

	generatedContent, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	envVars = parseEnvFile(string(generatedContent))

	// Value should be regenerated (different from existing value)
	if envVars["TEST_SECRET"] == "existing_value" {
		t.Error("Value was not regenerated with force flag")
	}

	// Value should match secure pattern
	secureValuePattern := regexp.MustCompile(`^[A-Za-z0-9]{24}$`)
	if !secureValuePattern.MatchString(envVars["TEST_SECRET"]) {
		t.Errorf("Generated value doesn't match expected pattern: %s", envVars["TEST_SECRET"])
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

	// Read and parse generated file
	generatedContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	envVars := parseEnvFile(string(generatedContent))

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

			// Read and parse generated file
			generatedContent, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}
			envVars := parseEnvFile(string(generatedContent))

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

			// Read and parse generated file
			generatedContent, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}
			envVars := parseEnvFile(string(generatedContent))

			// Check that values match the expected pattern
			for key, value := range envVars {
				if !tc.pattern.MatchString(value) {
					t.Errorf("Generated value for %s doesn't match expected pattern for charset %s: %s", key, tc.charset, value)
				}
			}
		})
	}
}

// TestGeneratorPreservesEnvStructure tests that existing .env structure is fully preserved
func TestGeneratorPreservesEnvStructure(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .env.example with some keys
	templatePath := filepath.Join(tempDir, ".env.example")
	templateContent := `# Template configuration
DB_HOST=localhost
DB_PASSWORD=${db_password}
NEW_KEY=${new_key}`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Create existing .env with custom structure
	outputPath := filepath.Join(tempDir, ".env")
	existingContent := `# My custom configuration
# These are my personal settings
DB_HOST=my-custom-host
DB_PASSWORD=my-secret-password

# Custom section
CUSTOM_KEY=custom_value`
	if err := os.WriteFile(outputPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing .env file: %v", err)
	}

	// Run generator
	config := Config{
		TemplatePath: templatePath,
		OutputPath:   outputPath,
		Force:        false,
	}
	gen := New(config)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate .env file: %v", err)
	}

	// Read the result
	result, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	resultStr := string(result)

	// Verify that existing structure is preserved
	if !strings.Contains(resultStr, "# My custom configuration") {
		t.Error("Original comments were not preserved")
	}
	if !strings.Contains(resultStr, "# These are my personal settings") {
		t.Error("Original comments were not preserved")
	}
	if !strings.Contains(resultStr, "DB_HOST=my-custom-host") {
		t.Error("Original DB_HOST value was not preserved")
	}
	if !strings.Contains(resultStr, "DB_PASSWORD=my-secret-password") {
		t.Error("Original DB_PASSWORD value was not preserved")
	}
	if !strings.Contains(resultStr, "CUSTOM_KEY=custom_value") {
		t.Error("Custom key was not preserved")
	}

	// Verify that new key was added
	envVars := parseEnvFile(resultStr)
	if _, exists := envVars["NEW_KEY"]; !exists {
		t.Error("NEW_KEY was not added from template")
	}
	if envVars["NEW_KEY"] == "${new_key}" {
		t.Error("NEW_KEY placeholder was not replaced")
	}
}

// TestGeneratorAddsCommentGroups tests that missing keys are added with their comment groups
func TestGeneratorAddsCommentGroups(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .env.example with comment groups
	templatePath := filepath.Join(tempDir, ".env.example")
	templateContent := `# Existing section
EXISTING_KEY=value1

# Database configuration
# Connection settings for the database
DB_HOST=localhost
DB_PASSWORD=${db_password}

# API configuration
# External API settings
API_KEY=${api_key}
API_URL=https://api.example.com`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Create existing .env with only one key
	outputPath := filepath.Join(tempDir, ".env")
	existingContent := `EXISTING_KEY=my_value`
	if err := os.WriteFile(outputPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing .env file: %v", err)
	}

	// Run generator
	config := Config{
		TemplatePath: templatePath,
		OutputPath:   outputPath,
		Force:        false,
	}
	gen := New(config)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate .env file: %v", err)
	}

	// Read the result
	result, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	resultStr := string(result)

	// Verify that existing key is preserved
	if !strings.Contains(resultStr, "EXISTING_KEY=my_value") {
		t.Error("Existing key was not preserved")
	}

	// Verify that comment groups were added
	if !strings.Contains(resultStr, "# Database configuration") {
		t.Error("Database comment group was not added")
	}
	if !strings.Contains(resultStr, "# Connection settings for the database") {
		t.Error("Database comment group was not added")
	}
	if !strings.Contains(resultStr, "# API configuration") {
		t.Error("API comment group was not added")
	}
	if !strings.Contains(resultStr, "# External API settings") {
		t.Error("API comment group was not added")
	}

	// Verify that new keys were added
	envVars := parseEnvFile(resultStr)
	if _, exists := envVars["DB_HOST"]; !exists {
		t.Error("DB_HOST was not added")
	}
	if _, exists := envVars["DB_PASSWORD"]; !exists {
		t.Error("DB_PASSWORD was not added")
	}
	if _, exists := envVars["API_KEY"]; !exists {
		t.Error("API_KEY was not added")
	}
	if _, exists := envVars["API_URL"]; !exists {
		t.Error("API_URL was not added")
	}

	// Verify that placeholders were replaced
	if envVars["DB_PASSWORD"] == "${db_password}" {
		t.Error("DB_PASSWORD placeholder was not replaced")
	}
	if envVars["API_KEY"] == "${api_key}" {
		t.Error("API_KEY placeholder was not replaced")
	}
}

// TestGeneratorForceRegeneratesValues tests that --force regenerates values for keys with placeholders
func TestGeneratorForceRegeneratesValues(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .env.example
	templatePath := filepath.Join(tempDir, ".env.example")
	templateContent := `SECRET=${secret}
FIXED_VALUE=fixed
NO_PLACEHOLDER=value`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Create existing .env
	outputPath := filepath.Join(tempDir, ".env")
	existingContent := `SECRET=old_secret
FIXED_VALUE=my_fixed_value
NO_PLACEHOLDER=my_value`
	if err := os.WriteFile(outputPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing .env file: %v", err)
	}

	// Run generator with force flag
	config := Config{
		TemplatePath: templatePath,
		OutputPath:   outputPath,
		Force:        true,
	}
	gen := New(config)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate .env file: %v", err)
	}

	// Read the result
	result, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	envVars := parseEnvFile(string(result))

	// SECRET should be regenerated (different from old_secret)
	if envVars["SECRET"] == "old_secret" {
		t.Error("SECRET was not regenerated with --force flag")
	}
	if envVars["SECRET"] == "${secret}" {
		t.Error("SECRET placeholder was not replaced")
	}

	// FIXED_VALUE should be preserved (no placeholder in template)
	if envVars["FIXED_VALUE"] != "my_fixed_value" {
		t.Errorf("FIXED_VALUE should be preserved, got %s", envVars["FIXED_VALUE"])
	}

	// NO_PLACEHOLDER should be preserved (no placeholder in template)
	if envVars["NO_PLACEHOLDER"] != "my_value" {
		t.Errorf("NO_PLACEHOLDER should be preserved, got %s", envVars["NO_PLACEHOLDER"])
	}
}
