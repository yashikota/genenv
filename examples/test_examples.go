package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// TestResult represents the result of a test
type TestResult struct {
	Name    string
	Success bool
	Message string
}

func main() {
	// Find the genenv binary
	genenvPath, err := findGenenvBinary()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using genenv binary: %s\n", genenvPath)

	// Get the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Testing examples in: %s\n\n", currentDir)

	// Validate the test environment
	if err := validateTestEnvironment(currentDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Please make sure all example directories and files are properly set up.")
		os.Exit(1)
	}

	// Define the examples to test
	examples := []struct {
		name        string
		description string
		testFunc    func(string, string) TestResult
	}{
		{
			name:        "basic",
			description: "Basic placeholder replacement",
			testFunc:    testBasicExample,
		},
		{
			name:        "with-metadata",
			description: "Field metadata and validation",
			testFunc:    testMetadataExample,
		},
		{
			name:        "with-types",
			description: "Field type validation",
			testFunc:    testTypesExample,
		},
		{
			name:        "complex",
			description: "Complex real-world configuration",
			testFunc:    testComplexExample,
		},
		{
			name:        "escaped",
			description: "Escaped placeholders",
			testFunc:    testEscapedExample,
		},
		{
			name:        "compare-existing",
			description: "Compare with existing .env",
			testFunc:    testCompareExample,
		},
		{
			name:        "custom-charset",
			description: "Custom character sets",
			testFunc:    testCharsetExample,
		},
		{
			name:        "new-from-scratch",
			description: "Creating a new .env file",
			testFunc:    testNewFromScratchExample,
		},
	}

	// Run the tests
	results := make([]TestResult, 0, len(examples))
	for _, example := range examples {
		fmt.Printf("Testing %s example (%s)...\n", example.name, example.description)
		exampleDir := filepath.Join(currentDir, example.name)

		// Check if example directory exists
		if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
			result := TestResult{
				Name:    example.name,
				Success: false,
				Message: fmt.Sprintf("Example directory does not exist: %s", exampleDir),
			}
			results = append(results, result)
			fmt.Printf("❌ %s: %s\n\n", example.name, result.Message)
			continue
		}

		result := example.testFunc(exampleDir, genenvPath)
		results = append(results, result)

		if result.Success {
			fmt.Printf("✅ %s: %s\n\n", example.name, result.Message)
		} else {
			fmt.Printf("❌ %s: %s\n\n", example.name, result.Message)
		}
	}

	// Print summary
	fmt.Println("=== Test Summary ===")
	passCount := 0
	for _, result := range results {
		if result.Success {
			fmt.Printf("✅ PASSED: %s\n", result.Name)
			passCount++
		} else {
			fmt.Printf("❌ FAILED: %s\n", result.Name)
		}
	}

	fmt.Printf("\n%d/%d tests passed\n", passCount, len(results))

	// Exit with error code if any tests failed
	if passCount < len(results) {
		os.Exit(1)
	}
}

// findGenenvBinary attempts to find the genenv binary
func findGenenvBinary() (string, error) {
	// Try in the current directory with absolute path
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	genenvPath := filepath.Join(currentDir, "genenv")
	if _, err := os.Stat(genenvPath); err == nil {
		return genenvPath, nil
	}

	genenvPathExe := filepath.Join(currentDir, "genenv.exe")
	if _, err := os.Stat(genenvPathExe); err == nil {
		return genenvPathExe, nil
	}

	// Try in the parent directory with absolute path
	parentDir := filepath.Dir(currentDir)
	parentGenenvPath := filepath.Join(parentDir, "genenv")
	if _, err := os.Stat(parentGenenvPath); err == nil {
		return parentGenenvPath, nil
	}

	parentGenenvPathExe := filepath.Join(parentDir, "genenv.exe")
	if _, err := os.Stat(parentGenenvPathExe); err == nil {
		return parentGenenvPathExe, nil
	}

	// Try to build it with absolute path
	buildPath := filepath.Join(currentDir, "genenv")
	cmd := exec.Command("go", "build", "-o", buildPath, filepath.Join(parentDir, "cmd", "genenv", "main.go"))
	if err := cmd.Run(); err == nil {
		return buildPath, nil
	}

	buildPathExe := filepath.Join(currentDir, "genenv.exe")
	if _, err := os.Stat(buildPathExe); err == nil {
		return buildPathExe, nil
	}

	// Check if it's in PATH
	path, err := exec.LookPath("genenv")
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("genenv binary not found")
}

// validateTestEnvironment checks that all required example directories and files exist
func validateTestEnvironment(examplesDir string) error {
	// List of required example directories
	requiredExamples := []string{
		"basic",
		"with-metadata",
		"with-types",
		"complex",
		"escaped",
		"compare-existing",
		"custom-charset",
		"new-from-scratch",
	}

	// Check that each required example directory exists
	for _, example := range requiredExamples {
		exampleDir := filepath.Join(examplesDir, example)
		if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
			return fmt.Errorf("required example directory not found: %s", exampleDir)
		}

		// Check that each example has a .env.example file
		templatePath := filepath.Join(exampleDir, ".env.example")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return fmt.Errorf("required template file not found: %s", templatePath)
		}
	}

	return nil
}

// testBasicExample tests the basic example
func testBasicExample(exampleDir, genenvPath string) TestResult {
	// Backup existing .env if it exists
	envPath := filepath.Join(exampleDir, ".env")
	backupPath := filepath.Join(exampleDir, ".env.backup")
	if _, err := os.Stat(envPath); err == nil {
		if err := os.Rename(envPath, backupPath); err != nil {
			return TestResult{
				Name:    "basic",
				Success: false,
				Message: fmt.Sprintf("Failed to backup .env: %v", err),
			}
		}
		defer os.Rename(backupPath, envPath) // Restore backup after test
	}

	// Run genenv on the example
	templatePath := filepath.Join(exampleDir, ".env.example")
	cmd := exec.Command(genenvPath, "-force", templatePath)
	cmd.Dir = exampleDir // Set working directory to example directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return TestResult{
			Name:    "basic",
			Success: false,
			Message: fmt.Sprintf("Failed to run genenv: %v\nOutput: %s", err, output),
		}
	}

	// Check if .env was created
	if _, err := os.Stat(envPath); err != nil {
		return TestResult{
			Name:    "basic",
			Success: false,
			Message: fmt.Sprintf(".env file was not created: %v", err),
		}
	}

	// Read the generated .env file
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return TestResult{
			Name:    "basic",
			Success: false,
			Message: fmt.Sprintf("Failed to read .env: %v", err),
		}
	}

	// Check that placeholders were replaced
	if strings.Contains(string(envContent), "${") {
		return TestResult{
			Name:    "basic",
			Success: false,
			Message: "Placeholders were not replaced in the .env file",
		}
	}

	return TestResult{
		Name:    "basic",
		Success: true,
		Message: "Successfully generated .env file with replaced placeholders",
	}
}

// testMetadataExample tests the metadata example
func testMetadataExample(exampleDir, genenvPath string) TestResult {
	// Backup existing .env if it exists
	envPath := filepath.Join(exampleDir, ".env")
	backupPath := filepath.Join(exampleDir, ".env.backup")
	if _, err := os.Stat(envPath); err == nil {
		if err := os.Rename(envPath, backupPath); err != nil {
			return TestResult{
				Name:    "with-metadata",
				Success: false,
				Message: fmt.Sprintf("Failed to backup .env: %v", err),
			}
		}
		defer os.Rename(backupPath, envPath) // Restore backup after test
	}

	// Run genenv on the example (non-interactive mode for testing)
	templatePath := filepath.Join(exampleDir, ".env.example")
	cmd := exec.Command(genenvPath, "-force", templatePath)
	cmd.Dir = exampleDir // Set working directory to example directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return TestResult{
			Name:    "with-metadata",
			Success: false,
			Message: fmt.Sprintf("Failed to run genenv: %v\nOutput: %s", err, output),
		}
	}

	// Check if .env was created
	if _, err := os.Stat(envPath); err != nil {
		return TestResult{
			Name:    "with-metadata",
			Success: false,
			Message: fmt.Sprintf(".env file was not created: %v", err),
		}
	}

	// Read the generated .env file
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return TestResult{
			Name:    "with-metadata",
			Success: false,
			Message: fmt.Sprintf("Failed to read .env: %v", err),
		}
	}

	// Check that placeholders were replaced
	if strings.Contains(string(envContent), "${") {
		return TestResult{
			Name:    "with-metadata",
			Success: false,
			Message: "Placeholders were not replaced in the .env file",
		}
	}

	return TestResult{
		Name:    "with-metadata",
		Success: true,
		Message: "Successfully generated .env file with metadata-based placeholders",
	}
}

// testTypesExample tests the types example with improved validation
func testTypesExample(exampleDir, genenvPath string) TestResult {
	// Backup existing .env if it exists
	envPath := filepath.Join(exampleDir, ".env")
	backupPath := filepath.Join(exampleDir, ".env.backup")
	if _, err := os.Stat(envPath); err == nil {
		if err := os.Rename(envPath, backupPath); err != nil {
			return TestResult{
				Name:    "with-types",
				Success: false,
				Message: fmt.Sprintf("Failed to backup .env: %v", err),
			}
		}
		defer os.Rename(backupPath, envPath) // Restore backup after test
	}

	// Run genenv on the example (non-interactive mode for testing)
	templatePath := filepath.Join(exampleDir, ".env.example")
	cmd := exec.Command(genenvPath, "-force", templatePath)
	cmd.Dir = exampleDir // Set working directory to example directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return TestResult{
			Name:    "with-types",
			Success: false,
			Message: fmt.Sprintf("Failed to run genenv: %v\nOutput: %s", err, output),
		}
	}

	// Check if .env was created
	if _, err := os.Stat(envPath); err != nil {
		return TestResult{
			Name:    "with-types",
			Success: false,
			Message: fmt.Sprintf(".env file was not created: %v", err),
		}
	}

	// Read the generated .env file
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return TestResult{
			Name:    "with-types",
			Success: false,
			Message: fmt.Sprintf("Failed to read .env: %v", err),
		}
	}

	// Check that placeholders were replaced
	if strings.Contains(string(envContent), "${") {
		return TestResult{
			Name:    "with-types",
			Success: false,
			Message: "Placeholders were not replaced in the .env file",
		}
	}

	// Enhanced validation for types
	envLines := strings.Split(string(envContent), "\n")
	for _, line := range envLines {
		if strings.HasPrefix(line, "INT_VALUE=") {
			// Check that INT_VALUE contains only digits
			value := strings.TrimPrefix(line, "INT_VALUE=")
			if value == "" {
				continue
			}
			if !isNumeric(value) {
				return TestResult{
					Name:    "with-types",
					Success: false,
					Message: fmt.Sprintf("INT_VALUE is not a valid integer: %s", value),
				}
			}
		} else if strings.HasPrefix(line, "FLOAT_VALUE=") {
			// Check that FLOAT_VALUE contains a valid float
			value := strings.TrimPrefix(line, "FLOAT_VALUE=")
			if value == "" {
				continue
			}
			if !isFloat(value) {
				return TestResult{
					Name:    "with-types",
					Success: false,
					Message: fmt.Sprintf("FLOAT_VALUE is not a valid float: %s", value),
				}
			}
		} else if strings.HasPrefix(line, "BOOL_VALUE=") {
			// Check that BOOL_VALUE contains a valid boolean
			value := strings.TrimPrefix(line, "BOOL_VALUE=")
			if value == "" {
				continue
			}
			if !isBool(value) {
				return TestResult{
					Name:    "with-types",
					Success: false,
					Message: fmt.Sprintf("BOOL_VALUE is not a valid boolean: %s", value),
				}
			}
		} else if strings.HasPrefix(line, "EMAIL_VALUE=") {
			// Check that EMAIL_VALUE contains a valid email format
			value := strings.TrimPrefix(line, "EMAIL_VALUE=")
			if value == "" {
				continue
			}
			if !isEmail(value) {
				return TestResult{
					Name:    "with-types",
					Success: false,
					Message: fmt.Sprintf("EMAIL_VALUE is not a valid email: %s", value),
				}
			}
		} else if strings.HasPrefix(line, "URL_VALUE=") {
			// Check that URL_VALUE contains a valid URL format
			value := strings.TrimPrefix(line, "URL_VALUE=")
			if value == "" {
				continue
			}
			if !isURL(value) {
				return TestResult{
					Name:    "with-types",
					Success: false,
					Message: fmt.Sprintf("URL_VALUE is not a valid URL: %s", value),
				}
			}
		}
	}

	return TestResult{
		Name:    "with-types",
		Success: true,
		Message: "Successfully generated .env file with properly typed values",
	}
}

// Helper functions for type validation
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func isFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isBool(s string) bool {
	_, err := strconv.ParseBool(s)
	return err == nil
}

func isEmail(s string) bool {
	// Simple email validation
	return strings.Contains(s, "@") && strings.Contains(s, ".")
}

func isURL(s string) bool {
	// Simple URL validation
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// testEscapedExample tests the escaped placeholders example with improved validation
func testEscapedExample(exampleDir, genenvPath string) TestResult {
	// Backup existing .env if it exists
	envPath := filepath.Join(exampleDir, ".env")
	backupPath := filepath.Join(exampleDir, ".env.backup")
	if _, err := os.Stat(envPath); err == nil {
		if err := os.Rename(envPath, backupPath); err != nil {
			return TestResult{
				Name:    "escaped",
				Success: false,
				Message: fmt.Sprintf("Failed to backup .env: %v", err),
			}
		}
		defer os.Rename(backupPath, envPath) // Restore backup after test
	}

	// Run genenv on the example
	templatePath := filepath.Join(exampleDir, ".env.example")
	cmd := exec.Command(genenvPath, "-force", templatePath)
	cmd.Dir = exampleDir // Set working directory to example directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return TestResult{
			Name:    "escaped",
			Success: false,
			Message: fmt.Sprintf("Failed to run genenv: %v\nOutput: %s", err, output),
		}
	}

	// Check if .env was created
	if _, err := os.Stat(envPath); err != nil {
		return TestResult{
			Name:    "escaped",
			Success: false,
			Message: fmt.Sprintf(".env file was not created: %v", err),
		}
	}

	// Read the generated .env file
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return TestResult{
			Name:    "escaped",
			Success: false,
			Message: fmt.Sprintf("Failed to read .env: %v", err),
		}
	}

	// Enhanced validation for escaped placeholders
	envLines := strings.Split(string(envContent), "\n")

	// Create a map of environment variable names to their values
	envVars := make(map[string]string)
	for _, line := range envLines {
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}
	}

	// Check that escaped placeholders were preserved and normal placeholders were replaced
	escapedFound := false
	normalFound := false

	// Check for escaped placeholders
	for _, key := range []string{"TEMPLATE_EXAMPLE", "ANOTHER_EXAMPLE", "ESCAPED_QUOTED"} {
		if value, ok := envVars[key]; ok {
			if strings.Contains(value, "${") {
				escapedFound = true
				break
			}
		}
	}

	// Check for normal placeholders that should be replaced
	for _, key := range []string{"API_KEY", "API_SECRET", "QUOTED_EXAMPLE"} {
		if value, ok := envVars[key]; ok {
			if !strings.Contains(value, "${") && len(value) > 0 {
				normalFound = true
				break
			}
		}
	}

	if !escapedFound {
		return TestResult{
			Name:    "escaped",
			Success: false,
			Message: "Escaped placeholders were not preserved in the .env file",
		}
	}

	if !normalFound {
		return TestResult{
			Name:    "escaped",
			Success: false,
			Message: "Normal placeholders were not replaced in the .env file",
		}
	}

	return TestResult{
		Name:    "escaped",
		Success: true,
		Message: "Successfully generated .env file with preserved escaped placeholders",
	}
}

// testCompareExample tests the compare with existing example
func testCompareExample(exampleDir, genenvPath string) TestResult {
	// Backup existing .env if it exists
	envPath := filepath.Join(exampleDir, ".env")
	backupPath := filepath.Join(exampleDir, ".env.backup")
	if _, err := os.Stat(envPath); err == nil {
		if err := os.Rename(envPath, backupPath); err != nil {
			return TestResult{
				Name:    "compare-existing",
				Success: false,
				Message: fmt.Sprintf("Failed to backup .env: %v", err),
			}
		}
		defer os.Rename(backupPath, envPath) // Restore backup after test
	}

	// Create a minimal .env file with some existing values
	minimalEnv := []byte("APP_NAME=My App\nAPP_ENV=development\nAPP_DEBUG=true\n")
	if err := os.WriteFile(envPath, minimalEnv, 0644); err != nil {
		return TestResult{
			Name:    "compare-existing",
			Success: false,
			Message: fmt.Sprintf("Failed to create minimal .env: %v", err),
		}
	}

	// Run genenv on the example with compare flag
	templatePath := filepath.Join(exampleDir, ".env.example")
	cmd := exec.Command(genenvPath, "-compare", "-force", templatePath)
	cmd.Dir = exampleDir // Set working directory to example directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return TestResult{
			Name:    "compare-existing",
			Success: false,
			Message: fmt.Sprintf("Failed to run genenv: %v\nOutput: %s", err, output),
		}
	}

	// Read the updated .env file
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return TestResult{
			Name:    "compare-existing",
			Success: false,
			Message: fmt.Sprintf("Failed to read .env: %v", err),
		}
	}

	// Check that existing values were preserved
	envStr := string(envContent)
	if !strings.Contains(envStr, "APP_NAME=My App") ||
		!strings.Contains(envStr, "APP_ENV=development") ||
		!strings.Contains(envStr, "APP_DEBUG=true") {
		return TestResult{
			Name:    "compare-existing",
			Success: false,
			Message: "Existing values were not preserved in the .env file",
		}
	}

	// Check that new fields were added
	if !strings.Contains(envStr, "APP_KEY=") ||
		!strings.Contains(envStr, "DB_USERNAME=") ||
		!strings.Contains(envStr, "DB_PASSWORD=") {
		return TestResult{
			Name:    "compare-existing",
			Success: false,
			Message: "New fields were not added to the .env file",
		}
	}

	return TestResult{
		Name:    "compare-existing",
		Success: true,
		Message: "Successfully updated .env file while preserving existing values",
	}
}

// testCharsetExample tests the custom charset example
func testCharsetExample(exampleDir, genenvPath string) TestResult {
	// Backup existing .env if it exists
	envPath := filepath.Join(exampleDir, ".env")
	backupPath := filepath.Join(exampleDir, ".env.backup")
	if _, err := os.Stat(envPath); err == nil {
		if err := os.Rename(envPath, backupPath); err != nil {
			return TestResult{
				Name:    "custom-charset",
				Success: false,
				Message: fmt.Sprintf("Failed to backup .env: %v", err),
			}
		}
		defer os.Rename(backupPath, envPath) // Restore backup after test
	}

	// Test with different character sets
	charsets := []string{"alphanumeric", "alphabetic", "uppercase", "lowercase", "numeric"}

	for _, charset := range charsets {
		// Run genenv with the specific charset
		templatePath := filepath.Join(exampleDir, ".env.example")
		cmd := exec.Command(genenvPath, "-force", "-charset", charset, templatePath)
		cmd.Dir = exampleDir // Set working directory to example directory
		output, err := cmd.CombinedOutput()
		if err != nil {
			return TestResult{
				Name:    "custom-charset",
				Success: false,
				Message: fmt.Sprintf("Failed to run genenv with charset %s: %v\nOutput: %s", charset, err, output),
			}
		}

		// Check if .env was created
		if _, err := os.Stat(envPath); err != nil {
			return TestResult{
				Name:    "custom-charset",
				Success: false,
				Message: fmt.Sprintf(".env file was not created with charset %s: %v", charset, err),
			}
		}

		// Read the generated .env file
		envContent, err := os.ReadFile(envPath)
		if err != nil {
			return TestResult{
				Name:    "custom-charset",
				Success: false,
				Message: fmt.Sprintf("Failed to read .env with charset %s: %v", charset, err),
			}
		}

		// Check that placeholders were replaced
		if strings.Contains(string(envContent), "${") {
			return TestResult{
				Name:    "custom-charset",
				Success: false,
				Message: fmt.Sprintf("Placeholders were not replaced in the .env file with charset %s", charset),
			}
		}
	}

	// Test with custom length
	templatePath := filepath.Join(exampleDir, ".env.example")
	cmd := exec.Command(genenvPath, "-force", "-length", "10", templatePath)
	cmd.Dir = exampleDir // Set working directory to example directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return TestResult{
			Name:    "custom-charset",
			Success: false,
			Message: fmt.Sprintf("Failed to run genenv with custom length: %v\nOutput: %s", err, output),
		}
	}

	// Read the generated .env file
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return TestResult{
			Name:    "custom-charset",
			Success: false,
			Message: fmt.Sprintf("Failed to read .env with custom length: %v", err),
		}
	}

	// Check that values have the correct length
	envLines := strings.Split(string(envContent), "\n")
	for _, line := range envLines {
		if strings.HasPrefix(line, "API_KEY=") {
			value := strings.TrimPrefix(line, "API_KEY=")
			if len(value) != 10 {
				return TestResult{
					Name:    "custom-charset",
					Success: false,
					Message: fmt.Sprintf("Generated value has incorrect length: %d (expected 10)", len(value)),
				}
			}
			break
		}
	}

	return TestResult{
		Name:    "custom-charset",
		Success: true,
		Message: "Successfully tested all character sets and custom length",
	}
}

// testNewFromScratchExample tests the new-from-scratch example
func testNewFromScratchExample(exampleDir, genenvPath string) TestResult {
	// Backup existing .env if it exists
	envPath := filepath.Join(exampleDir, ".env")
	backupPath := filepath.Join(exampleDir, ".env.backup")
	if _, err := os.Stat(envPath); err == nil {
		if err := os.Rename(envPath, backupPath); err != nil {
			return TestResult{
				Name:    "new-from-scratch",
				Success: false,
				Message: fmt.Sprintf("Failed to backup .env: %v", err),
			}
		}
		defer os.Rename(backupPath, envPath) // Restore backup after test
	}

	// Run genenv on the example with the template
	templatePath := filepath.Join(exampleDir, ".env.example")
	cmd := exec.Command(genenvPath, "-force", templatePath)
	cmd.Dir = exampleDir // Set working directory to example directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return TestResult{
			Name:    "new-from-scratch",
			Success: false,
			Message: fmt.Sprintf("Failed to run genenv: %v\nOutput: %s", err, output),
		}
	}

	// Check if .env was created
	if _, err := os.Stat(envPath); err != nil {
		return TestResult{
			Name:    "new-from-scratch",
			Success: false,
			Message: fmt.Sprintf(".env file was not created: %v", err),
		}
	}

	// Read the generated .env file
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return TestResult{
			Name:    "new-from-scratch",
			Success: false,
			Message: fmt.Sprintf("Failed to read .env: %v", err),
		}
	}

	// Check that placeholders were replaced
	if strings.Contains(string(envContent), "${") {
		return TestResult{
			Name:    "new-from-scratch",
			Success: false,
			Message: "Placeholders were not replaced in the .env file",
		}
	}

	return TestResult{
		Name:    "new-from-scratch",
		Success: true,
		Message: "Successfully generated .env file from template",
	}
}

// testComplexExample tests the complex example
func testComplexExample(exampleDir, genenvPath string) TestResult {
	// Backup existing .env if it exists
	envPath := filepath.Join(exampleDir, ".env")
	backupPath := filepath.Join(exampleDir, ".env.backup")
	if _, err := os.Stat(envPath); err == nil {
		if err := os.Rename(envPath, backupPath); err != nil {
			return TestResult{
				Name:    "complex",
				Success: false,
				Message: fmt.Sprintf("Failed to backup .env: %v", err),
			}
		}
		defer os.Rename(backupPath, envPath) // Restore backup after test
	}

	// Run genenv on the example
	templatePath := filepath.Join(exampleDir, ".env.example")
	cmd := exec.Command(genenvPath, "-force", templatePath)
	cmd.Dir = exampleDir // Set working directory to example directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return TestResult{
			Name:    "complex",
			Success: false,
			Message: fmt.Sprintf("Failed to run genenv: %v\nOutput: %s", err, output),
		}
	}

	// Check if .env was created
	if _, err := os.Stat(envPath); err != nil {
		return TestResult{
			Name:    "complex",
			Success: false,
			Message: fmt.Sprintf(".env file was not created: %v", err),
		}
	}

	// Read the generated .env file
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return TestResult{
			Name:    "complex",
			Success: false,
			Message: fmt.Sprintf("Failed to read .env: %v", err),
		}
	}

	// Check that placeholders were replaced
	if strings.Contains(string(envContent), "${") {
		return TestResult{
			Name:    "complex",
			Success: false,
			Message: "Placeholders were not replaced in the .env file",
		}
	}

	return TestResult{
		Name:    "complex",
		Success: true,
		Message: "Successfully generated .env file with complex configuration",
	}
}
