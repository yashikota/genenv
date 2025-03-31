package generator

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/yashikota/genenv/pkg/network"
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
	TemplatePath   string
	OutputPath     string
	Force          bool
	ValueLength    int
	Charset        CharsetType
	Interactive    bool
	CompareWithEnv bool
	SkipExisting   bool
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

	// Read existing .env file if needed
	existingValues := make(map[string]string)
	if g.config.CompareWithEnv || g.config.SkipExisting {
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
	// Parse template metadata to get field types
	fields, _ := g.ParseTemplateMetadata()

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

				// Check if this field exists in the existing .env file and we should skip it
				if existingValue, exists := existingValues[key]; exists && g.config.SkipExisting {
					// Use existing value
					generatedValues[placeholderName] = existingValue
				} else {
					// Generate a value if not already generated
					if _, exists := generatedValues[placeholderName]; !exists {
						// Check if field has a specific type
						fieldType := "string"
						if field, exists := fields[placeholderName]; exists && field.Type != "" {
							fieldType = field.Type
						}

						// Generate value based on field type
						secureValue, err := g.generateValueForField(fieldType)
						if err != nil {
							return nil, err
						}
						generatedValues[placeholderName] = secureValue
					}
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
			// Restore escaped placeholders
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

// generateValueForField generates a value for a field based on its type
func (g *Generator) generateValueForField(fieldType string) (string, error) {
	// For IP types, try to get the actual IP address
	switch fieldType {
	case "ip":
		// Get any IP (prefer IPv4)
		ip, err := network.LocalIP(network.IPAny)
		if err == nil {
			return ip.String(), nil
		}
		// Fallback to generate a random value
		return g.generateSecureValue()
	case "ipv4":
		// Get IPv4 address
		ip, err := network.LocalIP(network.IPv4)
		if err == nil {
			return ip.String(), nil
		}
		// Fallback to generate a random value
		return g.generateSecureValue()
	case "ipv6":
		// Get IPv6 address
		ip, err := network.LocalIP(network.IPv6)
		if err == nil {
			return ip.String(), nil
		}
		// Fallback to generate a random value
		return g.generateSecureValue()
	default:
		// For all other types, generate a secure random value
		return g.generateSecureValue()
	}
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

// TemplateField represents a field in the template with validation information
type TemplateField struct {
	Key          string
	DefaultValue string
	Required     bool
	Type         string // string, int, bool, etc.
	Description  string
}

// ParseTemplateMetadata parses the template file for metadata about fields
func (g *Generator) ParseTemplateMetadata() (map[string]TemplateField, error) {
	file, err := os.Open(g.config.TemplatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template file: %w", err)
	}
	defer file.Close()

	fields := make(map[string]TemplateField)
	scanner := bufio.NewScanner(file)

	// Regular expression to match ${variable} pattern
	placeholderRe := regexp.MustCompile(`\${([^}]+)}`)

	// Regular expression to match metadata in comments: # @field_name [required] (type) description
	metadataRe := regexp.MustCompile(`#\s*@([a-zA-Z0-9_]+)\s*(?:\[([^\]]+)\])?\s*(?:\(([^)]+)\))?\s*(.*)`)

	var currentKey string
	var currentField TemplateField

	for scanner.Scan() {
		line := scanner.Text()

		// Check if line is a metadata comment
		if matches := metadataRe.FindStringSubmatch(line); len(matches) > 1 {
			key := matches[1]
			currentKey = key
			currentField = TemplateField{
				Key:         key,
				Required:    strings.Contains(strings.ToLower(matches[2]), "required"),
				Type:        strings.ToLower(matches[3]),
				Description: strings.TrimSpace(matches[4]),
			}

			// Set default type if not specified
			if currentField.Type == "" {
				currentField.Type = "string" // Default type
			}

			// Normalize type names
			switch currentField.Type {
			case "integer":
				currentField.Type = "int"
			case "boolean":
				currentField.Type = "bool"
			case "double":
				currentField.Type = "float"
			}

			// Validate supported types
			validTypes := map[string]bool{
				"string": true,
				"int":    true,
				"bool":   true,
				"float":  true,
				"url":    true,
				"email":  true,
				"ip":     true,
				"ipv4":   true,
				"ipv6":   true,
			}

			if !validTypes[currentField.Type] {
				fmt.Printf("Warning: Unsupported type '%s' for field '%s', defaulting to 'string'\n",
					currentField.Type, key)
				currentField.Type = "string"
			}

			fields[key] = currentField
			continue
		}

		// Check if line contains a variable assignment
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Check if this is a placeholder
				if matches := placeholderRe.FindStringSubmatch(value); len(matches) > 1 {
					placeholderName := matches[1]

					// If we have metadata for this field, update it with the key
					if field, exists := fields[placeholderName]; exists {
						field.Key = key
						fields[placeholderName] = field
					} else {
						// Create a new field if no metadata was found
						fields[placeholderName] = TemplateField{
							Key:      key,
							Required: false,
							Type:     "string",
						}
					}
				} else if currentKey != "" {
					// Update default value for the current field
					field := fields[currentKey]
					field.DefaultValue = value
					fields[currentKey] = field
					currentKey = "" // Reset current key
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading template file: %w", err)
	}

	return fields, nil
}

// RunInteractive runs the generator in interactive mode
func (g *Generator) RunInteractive() error {
	// Parse template metadata
	fields, err := g.ParseTemplateMetadata()
	if err != nil {
		return err
	}

	// Check if we need to compare with existing .env file
	existingValues := make(map[string]string)
	if g.config.CompareWithEnv {
		if _, err := os.Stat(g.config.OutputPath); err == nil {
			// Read existing .env file
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

	// Process template interactively
	outputLines, err := g.processTemplateInteractively(templateLines, fields, existingValues)
	if err != nil {
		return err
	}

	// Write output file
	return g.writeOutputFile(outputLines)
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
			envVars[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return envVars, nil
}

// processTemplateInteractively processes template lines with user input
func (g *Generator) processTemplateInteractively(lines []string, fields map[string]TemplateField, existingValues map[string]string) ([]string, error) {
	// Keep track of generated values to ensure they're different
	generatedValues := make(map[string]string)

	// Regular expression to match ${variable} pattern
	re := regexp.MustCompile(`\${([^}]+)}`)

	result := make([]string, 0, len(lines))

	reader := bufio.NewReader(os.Stdin)

	// If we're comparing with existing .env, show which fields are new
	if g.config.CompareWithEnv {
		fmt.Println("Fields marked with [NEW] are not present in the existing .env file.")
	}

	// Validate user input based on field type
	validateInput := func(input string, fieldType string) (string, bool) {
		input = strings.TrimSpace(input)
		if input == "" {
			return input, true // Empty input is valid, will use default later
		}

		return input, ValidateFieldValue(input, fieldType)
	}

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

				// Check if this field exists in the existing .env file
				existingValue, existsInEnv := existingValues[key]

				// Skip if configured to skip existing values
				if existsInEnv && g.config.SkipExisting {
					generatedValues[placeholderName] = existingValue
					newValue := re.ReplaceAllString(escapedValue, existingValue)
					newValue = strings.ReplaceAll(newValue, tempMarker, `${`)
					result = append(result, fmt.Sprintf("%s=%s", key, newValue))
					fmt.Printf("Using existing value for %s: %s\n", key, existingValue)
					continue
				}

				// Get field metadata
				field, hasMetadata := fields[placeholderName]

				// Generate or prompt for value
				var fieldValue string

				if g.config.Interactive {
					// Prompt user for input
					var prompt string
					if hasMetadata && field.Description != "" {
						prompt = fmt.Sprintf("%s (%s", key, field.Description)
					} else {
						prompt = key
					}

					// Add required/optional indicator
					if hasMetadata {
						if field.Required {
							prompt += ", [REQUIRED]"
						} else {
							prompt += ", [OPTIONAL]"
						}
					}

					// Add type information
					if hasMetadata && field.Type != "" && field.Type != "string" {
						prompt += fmt.Sprintf(", type: %s", field.Type)
					}

					// Show if field is new compared to existing .env
					if g.config.CompareWithEnv && !existsInEnv {
						prompt += " [NEW]"
					}

					// Show default value if exists
					if hasMetadata && field.DefaultValue != "" {
						prompt += fmt.Sprintf(", default: %s", field.DefaultValue)
					} else if existsInEnv {
						prompt += fmt.Sprintf(", current: %s", existingValue)
					}

					// For IP fields, suggest auto-detection
					var suggestIP string
					if hasMetadata {
						switch field.Type {
						case "ip":
							if ip, err := network.LocalIP(network.IPAny); err == nil {
								suggestIP = ip.String()
								prompt += fmt.Sprintf(" (detected: %s)", suggestIP)
							}
						case "ipv4":
							if ip, err := network.LocalIP(network.IPv4); err == nil {
								suggestIP = ip.String()
								prompt += fmt.Sprintf(" (detected: %s)", suggestIP)
							}
						case "ipv6":
							if ip, err := network.LocalIP(network.IPv6); err == nil {
								suggestIP = ip.String()
								prompt += fmt.Sprintf(" (detected: %s)", suggestIP)
							}
						}
					}

					prompt += ": "

					// Loop until valid input is provided
					for {
						fmt.Print(prompt)
						input, err := reader.ReadString('\n')
						if err != nil {
							return nil, fmt.Errorf("error reading input: %w", err)
						}

						fieldValue = strings.TrimSpace(input)

						// Use detected IP if input is empty and we have a suggested IP
						if fieldValue == "" && suggestIP != "" {
							fieldValue = suggestIP
							fmt.Printf("Using detected IP: %s\n", fieldValue)
							break
						}

						// Validate input based on field type
						if hasMetadata && field.Type != "" {
							validatedValue, isValid := validateInput(fieldValue, field.Type)
							if !isValid {
								continue
							}
							fieldValue = validatedValue
						}

						// Use default value if input is empty
						if fieldValue == "" {
							if hasMetadata && field.DefaultValue != "" {
								fieldValue = field.DefaultValue
								fmt.Printf("Using default value: %s\n", fieldValue)
							} else if existsInEnv {
								fieldValue = existingValue
								fmt.Printf("Using current value: %s\n", fieldValue)
							} else if hasMetadata && field.Required {
								// If required field is empty, prompt again
								fmt.Println("This field is required. Please enter a value.")
								continue
							} else {
								// Generate a value based on field type
								fieldType := "string"
								if hasMetadata {
									fieldType = field.Type
								}

								secureValue, err := g.generateValueForField(fieldType)
								if err != nil {
									return nil, err
								}
								fieldValue = secureValue
								fmt.Printf("Generated value: %s\n", fieldValue)
							}
						}

						break
					}
				} else {
					// Generate a value if not already generated
					if _, exists := generatedValues[placeholderName]; !exists {
						// Determine field type
						fieldType := "string"
						if hasMetadata && field.Type != "" {
							fieldType = field.Type
						}

						// Generate value based on field type
						secureValue, err := g.generateValueForField(fieldType)
						if err != nil {
							return nil, err
						}
						fieldValue = secureValue
					} else {
						fieldValue = generatedValues[placeholderName]
					}
				}

				// Store the value
				generatedValues[placeholderName] = fieldValue

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
			// Restore escaped placeholders
			restoredValue := strings.ReplaceAll(escapedValue, tempMarker, `${`)
			result = append(result, fmt.Sprintf("%s=%s", key, restoredValue))
		}
	}

	return result, nil
}
