package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/yashikota/genenv/internal/generator"
)

// buildBinary builds the genenv binary for testing
func buildBinary(t *testing.T) (binaryPath string, cleanup func()) {
	t.Helper()

	tmpDir := t.TempDir()
	binaryPath = filepath.Join(tmpDir, "genenv")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build genenv binary: %v\nOutput: %s", err, output)
	}

	cleanup = func() {
		os.Remove(binaryPath)
	}

	return binaryPath, cleanup
}

// runGenenv executes the genenv binary with the given arguments
func runGenenv(t *testing.T, binary string, args ...string) (exitCode int, stdout string, stderr string) {
	t.Helper()

	cmd := exec.Command(binary, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	} else {
		exitCode = 0
	}

	return exitCode, stdout, stderr
}

// createTempTemplate creates a temporary template file with the given content
func createTempTemplate(t *testing.T, content string) (path string) {
	t.Helper()

	tmpDir := t.TempDir()
	path = filepath.Join(tmpDir, "template.env")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp template: %v", err)
	}

	return path
}

// readOutputFile reads the content of an output file
func readOutputFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	return string(content)
}

// parseEnvFile parses an env file and returns a map of key-value pairs
func parseEnvFile(content string) map[string]string {
	envVars := make(map[string]string)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, "\"")
			envVars[key] = value
		}
	}

	return envVars
}

// assertFileExists checks if a file exists
func assertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", path)
	}
}

// assertFileNotExists checks if a file does not exist
func assertFileNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected file to not exist: %s", path)
	}
}

// assertValueLength checks if a value has the expected length
func assertValueLength(t *testing.T, value string, expectedLength int) {
	t.Helper()

	if len(value) != expectedLength {
		t.Errorf("Expected value length %d, got %d: %s", expectedLength, len(value), value)
	}
}

// assertCharsetMatch checks if a value matches the expected charset pattern
func assertCharsetMatch(t *testing.T, value string, charset generator.CharsetType) {
	t.Helper()

	var pattern string
	switch charset {
	case generator.CharsetAlphanumeric:
		pattern = `^[A-Za-z0-9]+$`
	case generator.CharsetAlphabetic:
		pattern = `^[A-Za-z]+$`
	case generator.CharsetUppercase:
		pattern = `^[A-Z]+$`
	case generator.CharsetLowercase:
		pattern = `^[a-z]+$`
	case generator.CharsetNumeric:
		pattern = `^[0-9]+$`
	default:
		t.Fatalf("Unknown charset: %s", charset)
	}

	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		t.Fatalf("Failed to compile regex: %v", err)
	}

	if !matched {
		t.Errorf("Value %s does not match charset %s (pattern: %s)", value, charset, pattern)
	}
}

// assertExitCode checks if the exit code matches the expected value
func assertExitCode(t *testing.T, got, expected int) {
	t.Helper()

	if got != expected {
		t.Errorf("Expected exit code %d, got %d", expected, got)
	}
}

// assertContains checks if a string contains a substring
func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()

	if !strings.Contains(haystack, needle) {
		t.Errorf("Expected string to contain %q, got: %s", needle, haystack)
	}
}

// assertNotContains checks if a string does not contain a substring
func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()

	if strings.Contains(haystack, needle) {
		t.Errorf("Expected string to not contain %q, got: %s", needle, haystack)
	}
}

// Test cases

// TestArgumentOrder tests that arguments work in any order
func TestArgumentOrder(t *testing.T) {
	testCases := []struct {
		name string
		args func(output, template string) []string
	}{
		{
			name: "FlagsBeforeTemplate",
			args: func(output, template string) []string {
				return []string{"-o", output, template}
			},
		},
		{
			name: "TemplateBeforeFlags",
			args: func(output, template string) []string {
				return []string{template, "-o", output}
			},
		},
		{
			name: "MixedFlags",
			args: func(output, template string) []string {
				return []string{"-o", output, template, "-l", "16", "-c", "numeric"}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			binary, cleanup := buildBinary(t)
			defer cleanup()

			template := createTempTemplate(t, "TEST_KEY=${test}")
			tmpDir := filepath.Dir(template)
			output := filepath.Join(tmpDir, "output.env")

			args := tc.args(output, template)
			exitCode, stdout, _ := runGenenv(t, binary, args...)

			assertExitCode(t, exitCode, 0)
			assertContains(t, stdout, "Successfully generated")
			assertFileExists(t, output)

			// For MixedFlags, verify length and charset
			if tc.name == "MixedFlags" {
				content := readOutputFile(t, output)
				envVars := parseEnvFile(content)
				assertValueLength(t, envVars["TEST_KEY"], 16)
				assertCharsetMatch(t, envVars["TEST_KEY"], generator.CharsetNumeric)
			}
		})
	}
}

// TestOutputOption tests the -o/--output flag
func TestOutputOption(t *testing.T) {
	testCases := []struct {
		name        string
		outputName  string
		args        func(output, template string) []string
		expectInMsg string
	}{
		{
			name:       "ShortForm",
			outputName: "custom.env",
			args: func(output, template string) []string {
				return []string{"-o", output, template}
			},
			expectInMsg: "custom.env",
		},
		{
			name:       "LongForm",
			outputName: "custom.env",
			args: func(output, template string) []string {
				return []string{"--output", output, template}
			},
			expectInMsg: "custom.env",
		},
		{
			name:       "AfterTemplate",
			outputName: "after.env",
			args: func(output, template string) []string {
				return []string{template, "-o", output}
			},
			expectInMsg: "after.env",
		},
		{
			name:       "BeforeTemplate",
			outputName: "before.env",
			args: func(output, template string) []string {
				return []string{"-o", output, template}
			},
			expectInMsg: "before.env",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			binary, cleanup := buildBinary(t)
			defer cleanup()

			template := createTempTemplate(t, "TEST_KEY=${test}")
			tmpDir := filepath.Dir(template)
			output := filepath.Join(tmpDir, tc.outputName)

			args := tc.args(output, template)
			exitCode, stdout, _ := runGenenv(t, binary, args...)

			assertExitCode(t, exitCode, 0)
			assertContains(t, stdout, tc.expectInMsg)
			assertFileExists(t, output)
		})
	}
}

func TestOutputOption_DefaultValue(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)

	// Change to the template directory to test default .env output
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	exitCode, stdout, _ := runGenenv(t, binary, filepath.Base(template))

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, ".env")
	assertFileExists(t, filepath.Join(tmpDir, ".env"))
}

// TestLengthOption tests the -l/--length flag
func TestLengthOption(t *testing.T) {
	testCases := []struct {
		name   string
		flag   string
		length string
	}{
		{"ShortForm_32", "-l", "32"},
		{"LongForm_16", "--length", "16"},
		{"Length8", "-l", "8"},
		{"Length16", "-l", "16"},
		{"Length32", "-l", "32"},
		{"Length64", "-l", "64"},
		{"Length128", "-l", "128"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			binary, cleanup := buildBinary(t)
			defer cleanup()

			template := createTempTemplate(t, "TEST_KEY=${test}")
			tmpDir := filepath.Dir(template)
			output := filepath.Join(tmpDir, "output.env")

			exitCode, stdout, _ := runGenenv(t, binary, tc.flag, tc.length, "-o", output, template)

			assertExitCode(t, exitCode, 0)
			assertContains(t, stdout, "Successfully generated")

			content := readOutputFile(t, output)
			envVars := parseEnvFile(content)

			expectedLength := 0
			fmt.Sscanf(tc.length, "%d", &expectedLength)
			assertValueLength(t, envVars["TEST_KEY"], expectedLength)
		})
	}
}

func TestLengthOption_AfterTemplate(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, _ := runGenenv(t, binary, template, "-l", "48", "-o", output)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)
	assertValueLength(t, envVars["TEST_KEY"], 48)
}

func TestLengthOption_DefaultValue(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, _, _ := runGenenv(t, binary, "-o", output, template)

	assertExitCode(t, exitCode, 0)

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)
	// Default length is 24
	assertValueLength(t, envVars["TEST_KEY"], 24)
}

// TestCharsetOption tests the -c/--charset flag
func TestCharsetOption(t *testing.T) {
	testCases := []struct {
		name        string
		flag        string
		charset     string
		charsetType generator.CharsetType
	}{
		{"ShortForm_Numeric", "-c", "numeric", generator.CharsetNumeric},
		{"LongForm_Alphabetic", "--charset", "alphabetic", generator.CharsetAlphabetic},
		{"Alphanumeric", "-c", "alphanumeric", generator.CharsetAlphanumeric},
		{"Alphabetic", "-c", "alphabetic", generator.CharsetAlphabetic},
		{"Uppercase", "-c", "uppercase", generator.CharsetUppercase},
		{"Lowercase", "-c", "lowercase", generator.CharsetLowercase},
		{"Numeric", "-c", "numeric", generator.CharsetNumeric},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			binary, cleanup := buildBinary(t)
			defer cleanup()

			template := createTempTemplate(t, "TEST_KEY=${test}")
			tmpDir := filepath.Dir(template)
			output := filepath.Join(tmpDir, "output.env")

			exitCode, stdout, _ := runGenenv(t, binary, tc.flag, tc.charset, "-o", output, template)

			assertExitCode(t, exitCode, 0)
			assertContains(t, stdout, "Successfully generated")

			content := readOutputFile(t, output)
			envVars := parseEnvFile(content)
			assertCharsetMatch(t, envVars["TEST_KEY"], tc.charsetType)
		})
	}
}

func TestCharsetOption_AfterTemplate(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, _ := runGenenv(t, binary, template, "-c", "uppercase", "-o", output)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)
	assertCharsetMatch(t, envVars["TEST_KEY"], generator.CharsetUppercase)
}

func TestCharsetOption_DefaultValue(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, _, _ := runGenenv(t, binary, "-o", output, template)

	assertExitCode(t, exitCode, 0)

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)
	// Default charset is alphanumeric
	assertCharsetMatch(t, envVars["TEST_KEY"], generator.CharsetAlphanumeric)
}

// TestForceOption tests the -f/--force and -y/--yes flags
func TestForceOption_ShortForm(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	// Create an existing file
	os.WriteFile(output, []byte("TEST_KEY=old_value\n"), 0644)

	// Run with -f -y flags
	exitCode, stdout, _ := runGenenv(t, binary, "-f", "-y", "-o", output, template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)

	// Value should be regenerated (not "old_value")
	if envVars["TEST_KEY"] == "old_value" {
		t.Error("Expected value to be regenerated with --force flag")
	}
}

func TestForceOption_LongForm(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	// Create an existing file
	os.WriteFile(output, []byte("TEST_KEY=old_value\n"), 0644)

	// Run with --force --yes flags
	exitCode, stdout, _ := runGenenv(t, binary, "--force", "--yes", "-o", output, template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)

	// Value should be regenerated (not "old_value")
	if envVars["TEST_KEY"] == "old_value" {
		t.Error("Expected value to be regenerated with --force flag")
	}
}

func TestForceOption_WithoutYes(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	// Create an existing file
	os.WriteFile(output, []byte("TEST_KEY=old_value\n"), 0644)

	// Run with --force but without --yes (should prompt, but stdin is empty so it will fail or default to No)
	cmd := exec.Command(binary, "--force", "-o", output, template)
	cmd.Stdin = strings.NewReader("n\n")
	err := cmd.Run()

	// Should exit with 0 (operation cancelled)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 0 means operation cancelled successfully
			assertExitCode(t, exitErr.ExitCode(), 0)
		}
	}

	// Value should remain unchanged
	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)

	if envVars["TEST_KEY"] != "old_value" {
		t.Errorf("Expected value to remain unchanged without --yes confirmation, got: %s", envVars["TEST_KEY"])
	}
}

func TestForceOption_AfterTemplate(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	// Create an existing file
	os.WriteFile(output, []byte("TEST_KEY=old_value\n"), 0644)

	// Run with flags after template
	exitCode, stdout, _ := runGenenv(t, binary, template, "-f", "-y", "-o", output)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)

	// Value should be regenerated
	if envVars["TEST_KEY"] == "old_value" {
		t.Error("Expected value to be regenerated with --force flag")
	}
}

func TestForceOption_PreservesNonPlaceholderValues(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}\nFIXED_KEY=fixed_value")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	// Create an existing file
	os.WriteFile(output, []byte("TEST_KEY=old_value\nFIXED_KEY=my_fixed_value\n"), 0644)

	// Run with --force
	exitCode, _, _ := runGenenv(t, binary, "--force", "--yes", "-o", output, template)

	assertExitCode(t, exitCode, 0)

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)

	// TEST_KEY should be regenerated (has placeholder)
	if envVars["TEST_KEY"] == "old_value" {
		t.Error("Expected TEST_KEY to be regenerated")
	}

	// FIXED_KEY should remain unchanged (no placeholder in template)
	if envVars["FIXED_KEY"] != "my_fixed_value" {
		t.Errorf("Expected FIXED_KEY to remain unchanged, got: %s", envVars["FIXED_KEY"])
	}
}

// TestVersionOption tests the -v/--version flag
func TestVersionOption_ShortForm(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	exitCode, stdout, _ := runGenenv(t, binary, "-v")

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "genenv version")
	assertContains(t, stdout, "1.1.0")
}

func TestVersionOption_LongForm(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	exitCode, stdout, _ := runGenenv(t, binary, "--version")

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "genenv version")
	assertContains(t, stdout, "1.1.0")
}

// TestHelpOption tests the -h/--help flag
func TestHelpOption_ShortForm(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	exitCode, _, stderr := runGenenv(t, binary, "-h")

	assertExitCode(t, exitCode, 0)
	assertContains(t, stderr, "Usage:")
	assertContains(t, stderr, "genenv")
	assertContains(t, stderr, "Options:")
}

func TestHelpOption_LongForm(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	exitCode, _, stderr := runGenenv(t, binary, "--help")

	assertExitCode(t, exitCode, 0)
	assertContains(t, stderr, "Usage:")
	assertContains(t, stderr, "genenv")
	assertContains(t, stderr, "Options:")
}

func TestHelpOption_NoArgs(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	exitCode, _, stderr := runGenenv(t, binary)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stderr, "Usage:")
}

// TestMultipleOptions tests combinations of multiple options
func TestMultipleOptions_AllShortForm(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, _ := runGenenv(t, binary, "-o", output, "-l", "32", "-c", "numeric", template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)
	assertValueLength(t, envVars["TEST_KEY"], 32)
	assertCharsetMatch(t, envVars["TEST_KEY"], generator.CharsetNumeric)
}

func TestMultipleOptions_AllLongForm(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, _ := runGenenv(t, binary, "--output", output, "--length", "16", "--charset", "uppercase", template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)
	assertValueLength(t, envVars["TEST_KEY"], 16)
	assertCharsetMatch(t, envVars["TEST_KEY"], generator.CharsetUppercase)
}

func TestMultipleOptions_Mixed(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, _ := runGenenv(t, binary, "-o", output, "--length", "48", "-c", "lowercase", template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)
	assertValueLength(t, envVars["TEST_KEY"], 48)
	assertCharsetMatch(t, envVars["TEST_KEY"], generator.CharsetLowercase)
}

func TestMultipleOptions_BeforeAndAfterTemplate(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	// Some flags before, some after template
	exitCode, stdout, _ := runGenenv(t, binary, "-o", output, template, "-l", "64", "-c", "alphabetic")

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)
	assertValueLength(t, envVars["TEST_KEY"], 64)
	assertCharsetMatch(t, envVars["TEST_KEY"], generator.CharsetAlphabetic)
}

func TestMultipleOptions_WithForce(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	// Create an existing file
	os.WriteFile(output, []byte("TEST_KEY=old_value\n"), 0644)

	exitCode, stdout, _ := runGenenv(t, binary, "-f", "-y", "-o", output, "-l", "16", "-c", "numeric", template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)

	// Value should be regenerated
	if envVars["TEST_KEY"] == "old_value" {
		t.Error("Expected value to be regenerated")
	}

	assertValueLength(t, envVars["TEST_KEY"], 16)
	assertCharsetMatch(t, envVars["TEST_KEY"], generator.CharsetNumeric)
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCase_NoTemplate(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	exitCode, _, stderr := runGenenv(t, binary)

	assertExitCode(t, exitCode, 0) // Shows usage
	assertContains(t, stderr, "Usage:")
}

func TestEdgeCase_InvalidCharset(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, stderr := runGenenv(t, binary, "-c", "invalid_charset", "-o", output, template)

	// Should exit with error
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid charset")
	}

	// Error message could be in either stdout or stderr
	combinedOutput := stdout + stderr
	assertContains(t, combinedOutput, "Invalid charset")
}

func TestEdgeCase_NonExistentTemplate(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "output.env")
	nonExistentTemplate := filepath.Join(tmpDir, "non-existent.env")

	exitCode, stdout, stderr := runGenenv(t, binary, "-o", output, nonExistentTemplate)

	// Should exit with error
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for non-existent template")
	}

	// Error message could be in either stdout or stderr
	combinedOutput := stdout + stderr
	assertContains(t, combinedOutput, "Error")
}

func TestEdgeCase_InvalidOutputDirectory(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	nonExistentDir := filepath.Join(t.TempDir(), "non-existent-dir", "deep", "path", "output.env")

	exitCode, stdout, stderr := runGenenv(t, binary, "-o", nonExistentDir, template)

	// Should exit with error
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid output directory")
	}

	// Error message could be in either stdout or stderr
	combinedOutput := stdout + stderr
	assertContains(t, combinedOutput, "Error")
}

func TestEdgeCase_EmptyTemplate(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, _ := runGenenv(t, binary, "-o", output, template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")
	assertFileExists(t, output)

	content := readOutputFile(t, output)
	// Should create empty or minimal file
	if len(strings.TrimSpace(content)) > 0 {
		t.Logf("Generated content from empty template: %s", content)
	}
}

func TestEdgeCase_OnlyComments(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "# This is a comment\n# Another comment\n")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, _ := runGenenv(t, binary, "-o", output, template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")
	assertFileExists(t, output)

	content := readOutputFile(t, output)
	assertContains(t, content, "# This is a comment")
}

func TestEdgeCase_MultiplePlaceholdersSameName(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "KEY1=${test}\nKEY2=${test}\nKEY3=${other}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, _ := runGenenv(t, binary, "-o", output, template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)

	// KEY1 and KEY2 should have the same value (same placeholder name)
	if envVars["KEY1"] != envVars["KEY2"] {
		t.Errorf("Expected KEY1 and KEY2 to have the same value, got KEY1=%s, KEY2=%s", envVars["KEY1"], envVars["KEY2"])
	}

	// KEY3 should have a different value
	if envVars["KEY1"] == envVars["KEY3"] {
		t.Error("Expected KEY3 to have a different value from KEY1")
	}
}

func TestEdgeCase_EscapedPlaceholder(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, `KEY1=${test}
KEY2=\${not_replaced}`)
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, _ := runGenenv(t, binary, "-o", output, template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)

	// KEY2 should contain the literal ${not_replaced}
	assertContains(t, content, "${not_replaced}")
}

func TestEdgeCase_SpecialCharactersInValue(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "KEY1=value_with_spaces and = signs\nKEY2=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, stdout, _ := runGenenv(t, binary, "-o", output, template)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Successfully generated")

	content := readOutputFile(t, output)
	assertContains(t, content, "value_with_spaces and = signs")
}

func TestEdgeCase_VeryLongLength(t *testing.T) {
	binary, cleanup := buildBinary(t)
	defer cleanup()

	template := createTempTemplate(t, "TEST_KEY=${test}")
	tmpDir := filepath.Dir(template)
	output := filepath.Join(tmpDir, "output.env")

	exitCode, _, _ := runGenenv(t, binary, "-l", "256", "-o", output, template)

	assertExitCode(t, exitCode, 0)

	content := readOutputFile(t, output)
	envVars := parseEnvFile(content)
	assertValueLength(t, envVars["TEST_KEY"], 256)
}
