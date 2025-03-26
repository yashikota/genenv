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
	// Check if output file exists
	if _, err := os.Stat(g.config.OutputPath); err == nil && !g.config.Force {
		return fmt.Errorf("output file %s already exists, use --force to overwrite", g.config.OutputPath)
	}

	// Read template file
	templateLines, err := g.readTemplateFile()
	if err != nil {
		return err
	}

	// Process template lines
	outputLines, err := g.processTemplateLines(templateLines)
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
func (g *Generator) processTemplateLines(lines []string) ([]string, error) {
	// Keep track of generated values to ensure they're different
	generatedValues := make(map[string]string)

	// Regular expression to match ${variable} pattern
	re := regexp.MustCompile(`\${([^}]+)}`)

	result := make([]string, 0, len(lines))

	for _, line := range lines {
		// Skip empty lines and comments
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			result = append(result, line)
			continue
		}

		// Check if line contains a variable assignment
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			result = append(result, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Handle escaped placeholders: replace \${...} with a temporary marker
		tempMarker := "##ESCAPED_PLACEHOLDER##"
		escapedValue := strings.ReplaceAll(value, `\${`, tempMarker)

		// Check if value contains a placeholder
		if re.MatchString(escapedValue) {
			matches := re.FindStringSubmatch(escapedValue)
			if len(matches) > 1 {
				placeholderName := matches[1]

				// Generate a secure random value if not already generated
				if _, exists := generatedValues[placeholderName]; !exists {
					secureValue, err := g.generateSecureValue()
					if err != nil {
						return nil, fmt.Errorf("failed to generate secure value: %w", err)
					}
					generatedValues[placeholderName] = secureValue
				}

				// Replace placeholder with generated value
				newValue := re.ReplaceAllString(escapedValue, generatedValues[matches[1]])

				// Restore escaped placeholders
				newValue = strings.ReplaceAll(newValue, tempMarker, `${`)

				result = append(result, fmt.Sprintf("%s=%s", key, newValue))
			} else {
				// Restore escaped placeholders
				restoredValue := strings.ReplaceAll(escapedValue, tempMarker, `${`)
				result = append(result, fmt.Sprintf("%s=%s", key, restoredValue))
			}
		} else {
			// No placeholder, keep the line as is but restore escaped placeholders
			restoredValue := strings.ReplaceAll(escapedValue, tempMarker, `${`)
			result = append(result, fmt.Sprintf("%s=%s", key, restoredValue))
		}
	}

	return result, nil
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
