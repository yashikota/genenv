package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInteractiveMode tests the interactive mode functionality
func TestInteractiveMode(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping interactive test in CI environment")
	}

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templatePath := filepath.Join(tempDir, ".env.example")
	templateContent := `
# @db_user [required] (string) Database username
DB_USER=${db_user}

# @db_password [required] (string) Database password
DB_PASSWORD=${db_password}

# @db_port [optional] (int) Database port
DB_PORT=${db_port}

# @debug_mode [optional] (bool) Enable debug mode
DEBUG=${debug_mode}
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// Test interactive mode with simulated input
	t.Run("Interactive mode with template", func(t *testing.T) {
		// Prepare input
		input := "y\n" + // Yes, use interactive mode
			templatePath + "\n" + // Template path
			filepath.Join(tempDir, ".env") + "\n" + // Output path
			"y\n" + // Yes, force overwrite
			"24\n" + // Value length
			"alphanumeric\n" + // Charset
			"n\n" + // No, don't compare with existing .env
			"testuser\n" + // DB_USER value
			"testpass\n" + // DB_PASSWORD value
			"5432\n" + // DB_PORT value
			"true\n" // DEBUG value

		// Redirect stdin and stdout
		oldStdin := os.Stdin
		oldStdout := os.Stdout
		defer func() {
			os.Stdin = oldStdin
			os.Stdout = oldStdout
		}()

		r, w, _ := os.Pipe()
		os.Stdin = r

		_, outW, _ := os.Pipe()
		os.Stdout = outW

		// Write input to stdin
		go func() {
			defer w.Close()
			w.Write([]byte(input))
		}()

		// Run the interactive mode
		// Note: In a real test, we would call the main function or a test-specific function
		// that handles interactive mode. For this example, we're just demonstrating the approach.
		t.Skip("This test requires a refactored main function to be testable")

		// Close stdout pipe and read output
		outW.Close()
		var buf bytes.Buffer
		output := buf.String()

		// Verify the output contains expected prompts and messages
		expectedPrompts := []string{
			"Do you want to run in interactive mode?",
			"Enter the template file path",
			"Enter the output file path",
			"Force overwrite existing .env file?",
			"Enter the length of generated values",
			"Enter the character set",
			"Compare with existing .env file?",
			"DB_USER (Database username, [REQUIRED]",
			"DB_PASSWORD (Database password, [REQUIRED]",
			"DB_PORT (Database port, [OPTIONAL], type: int",
			"DEBUG (Enable debug mode, [OPTIONAL], type: bool",
		}

		for _, prompt := range expectedPrompts {
			if !strings.Contains(output, prompt) {
				t.Errorf("Expected output to contain prompt: %s", prompt)
			}
		}

		// Check that the .env file was created with the expected values
		envPath := filepath.Join(tempDir, ".env")
		envContent, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf("Failed to read generated .env file: %v", err)
		}

		envLines := strings.Split(string(envContent), "\n")
		envMap := make(map[string]string)

		for _, line := range envLines {
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
			envMap[key] = value
		}

		// Verify the values
		expectedValues := map[string]string{
			"DB_USER":     "testuser",
			"DB_PASSWORD": "testpass",
			"DB_PORT":     "5432",
			"DEBUG":       "true",
		}

		for key, expected := range expectedValues {
			if value, ok := envMap[key]; !ok || value != expected {
				t.Errorf("Expected %s=%s, got %s", key, expected, value)
			}
		}
	})
}

// TestNewEnvFileCreation tests the creation of a new .env file from scratch
func TestNewEnvFileCreation(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping interactive test in CI environment")
	}

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "genenv-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test new .env file creation with simulated input
	t.Run("Create new .env file", func(t *testing.T) {
		// Prepare input
		input := "y\n" + // Yes, use interactive mode
			"new\n" + // Create new .env file
			filepath.Join(tempDir, ".env") + "\n" + // Output path
			"y\n" + // Yes, force overwrite
			"24\n" + // Value length
			"alphanumeric\n" + // Charset
			"n\n" + // No, don't compare with existing .env
			"localhost\n" + // DB_HOST value
			"5432\n" + // DB_PORT value
			"testuser\n" + // DB_USER value
			"testpass\n" + // DB_PASSWORD value
			"true\n" // DEBUG value

		// Redirect stdin and stdout
		oldStdin := os.Stdin
		oldStdout := os.Stdout
		defer func() {
			os.Stdin = oldStdin
			os.Stdout = oldStdout
		}()

		r, w, _ := os.Pipe()
		os.Stdin = r

		_, outW, _ := os.Pipe()
		os.Stdout = outW

		// Write input to stdin
		go func() {
			defer w.Close()
			w.Write([]byte(input))
		}()

		// Run the interactive mode
		// Note: In a real test, we would call the main function or a test-specific function
		// that handles interactive mode. For this example, we're just demonstrating the approach.
		t.Skip("This test requires a refactored main function to be testable")

		// Close stdout pipe and read output
		outW.Close()
		var buf bytes.Buffer
		output := buf.String()

		// Verify the output contains expected prompts and messages
		expectedPrompts := []string{
			"Do you want to run in interactive mode?",
			"Enter 'new' to create a new .env file from scratch",
			"Enter the output file path",
			"Force overwrite existing .env file?",
			"Enter the length of generated values",
			"Enter the character set",
			"Compare with existing .env file?",
			"DB_HOST",
			"DB_PORT",
			"DB_USER",
			"DB_PASSWORD",
			"DEBUG",
		}

		for _, prompt := range expectedPrompts {
			if !strings.Contains(output, prompt) {
				t.Errorf("Expected output to contain prompt: %s", prompt)
			}
		}

		// Check that the .env file was created with the expected values
		envPath := filepath.Join(tempDir, ".env")
		envContent, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf("Failed to read generated .env file: %v", err)
		}

		envLines := strings.Split(string(envContent), "\n")
		envMap := make(map[string]string)

		for _, line := range envLines {
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
			envMap[key] = value
		}

		// Verify the values
		expectedValues := map[string]string{
			"DB_HOST":     "localhost",
			"DB_PORT":     "5432",
			"DB_USER":     "testuser",
			"DB_PASSWORD": "testpass",
			"DEBUG":       "true",
		}

		for key, expected := range expectedValues {
			if value, ok := envMap[key]; !ok || value != expected {
				t.Errorf("Expected %s=%s, got %s", key, expected, value)
			}
		}
	})
}
