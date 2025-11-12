package generator

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// CharsetType defines the type of character set to use for random values
type CharsetType string

const (
	// CharsetAlphanumeric includes uppercase, lowercase letters and numbers
	CharsetAlphanumeric CharsetType = "alphanumeric"
	// CharsetAlphabetic includes uppercase and lowercase letters only
	CharsetAlphabetic CharsetType = "alphabetic"
	// CharsetUppercase includes uppercase letters only
	CharsetUppercase CharsetType = "uppercase"
	// CharsetLowercase includes lowercase letters only
	CharsetLowercase CharsetType = "lowercase"
	// CharsetNumeric includes numbers only
	CharsetNumeric CharsetType = "numeric"

	// Default length for generated values
	DefaultValueLength = 24
)

// Config holds configuration for the env generator
type Config struct {
	TemplatePath string
	OutputPath   string
	Force        bool
	ValueLength  int
	Charset      CharsetType
}

// Generator is responsible for generating .env files
type Generator struct {
	config Config
}

// New creates a new Generator instance
func New(config Config) *Generator {
	// Set default values if not specified
	if config.ValueLength <= 0 {
		config.ValueLength = DefaultValueLength
	}

	if config.Charset == "" {
		config.Charset = CharsetAlphanumeric
	}

	return &Generator{
		config: config,
	}
}

// Generate creates a .env file from a template
func (g *Generator) Generate() error {
	// Read existing .env file to preserve existing values (unless --force is used)
	existingValues := make(map[string]string)
	if !g.config.Force {
		if _, err := os.Stat(g.config.OutputPath); err == nil {
			var err error
			existingValues, err = g.readEnvFile(g.config.OutputPath)
			if err != nil {
				return fmt.Errorf("failed to read existing .env file: %w", err)
			}
		}
	}

	// Read template file
	templateLines, err := g.readTemplateFile()
	if err != nil {
		return err
	}

	// Process template lines
	outputLines, err := g.processTemplateLines(templateLines, existingValues)
	if err != nil {
		return err
	}

	// Write output file
	return g.writeOutputFile(outputLines)
}

// readTemplateFile reads the template file
func (g *Generator) readTemplateFile() ([]string, error) {
	file, err := os.Open(g.config.TemplatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading template file: %w", err)
	}

	return lines, nil
}

// processTemplateLines processes template lines and replaces placeholders
func (g *Generator) processTemplateLines(lines []string, existingValues map[string]string) ([]string, error) {
	// Keep track of generated values to ensure consistency across placeholders
	generatedValues := make(map[string]string)
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		processedLine, err := g.processLine(line, existingValues, generatedValues)
		if err != nil {
			return nil, err
		}
		result = append(result, processedLine)
	}

	return result, nil
}

// processLine processes a single template line
func (g *Generator) processLine(line string, existingValues, generatedValues map[string]string) (string, error) {
	// Skip empty lines and comments
	if isCommentOrEmpty(line) {
		return line, nil
	}

	// Parse key-value assignment
	key, value, ok := parseKeyValue(line)
	if !ok {
		return line, nil
	}

	// Process placeholders in the value
	processedValue, err := g.processPlaceholders(key, value, existingValues, generatedValues)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s=%s", key, processedValue), nil
}

// processPlaceholders replaces placeholders in a value with generated or existing values
func (g *Generator) processPlaceholders(key, value string, existingValues, generatedValues map[string]string) (string, error) {
	const tempMarker = "##ESCAPED_PLACEHOLDER##"

	// Temporarily replace escaped placeholders
	escapedValue := strings.ReplaceAll(value, `\${`, tempMarker)

	// Regular expression to match ${variable} pattern
	re := regexp.MustCompile(`\${([^}]+)}`)

	// Replace all placeholders
	result := re.ReplaceAllStringFunc(escapedValue, func(match string) string {
		placeholderName := re.FindStringSubmatch(match)[1]

		// Use existing value if available
		if existingValue, exists := existingValues[key]; exists {
			generatedValues[placeholderName] = existingValue
			return existingValue
		}

		// Use or generate a value for this placeholder
		if value, exists := generatedValues[placeholderName]; exists {
			return value
		}

		// Generate new secure random value
		secureValue, err := g.generateSecureValue()
		if err != nil {
			return match // Keep original on error
		}
		generatedValues[placeholderName] = secureValue
		return secureValue
	})

	// Restore escaped placeholders
	return strings.ReplaceAll(result, tempMarker, `${`), nil
}

// isCommentOrEmpty checks if a line is a comment or empty
func isCommentOrEmpty(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "" || strings.HasPrefix(trimmed, "#")
}

// parseKeyValue parses a key-value assignment from a line
func parseKeyValue(line string) (key, value string, ok bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
}

// writeOutputFile writes processed lines to the output file
func (g *Generator) writeOutputFile(lines []string) error {
	file, err := os.Create(g.config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to output file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush output file: %w", err)
	}

	return nil
}

// generateSecureValue generates a cryptographically secure random value
func (g *Generator) generateSecureValue() (string, error) {
	// Get the appropriate charset based on the configuration
	charset := getCharset(g.config.Charset)
	length := g.config.ValueLength

	// Create a byte slice for the result
	result := make([]byte, length)

	// Generate random bytes
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Map random bytes to characters in the charset
	for i := 0; i < length; i++ {
		// Use modulo to map the random byte to a character in the charset
		// This ensures uniform distribution
		result[i] = charset[int(randomBytes[i])%len(charset)]
	}

	return string(result), nil
}

// getCharset returns the appropriate character set based on the CharsetType
func getCharset(charsetType CharsetType) string {
	switch charsetType {
	case CharsetAlphabetic:
		return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case CharsetUppercase:
		return "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case CharsetLowercase:
		return "abcdefghijklmnopqrstuvwxyz"
	case CharsetNumeric:
		return "0123456789"
	case CharsetAlphanumeric:
		fallthrough
	default:
		return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}
}

// readEnvFile reads an env file and returns a map of key-value pairs
func (g *Generator) readEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	envVars := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if isCommentOrEmpty(line) {
			continue
		}

		// Parse key-value pairs
		if key, value, ok := parseKeyValue(line); ok {
			envVars[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return envVars, nil
}
