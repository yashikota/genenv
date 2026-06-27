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

// EnvLineType represents the type of line in an env file
type EnvLineType int

const (
	// LineTypeComment represents a comment or empty line
	LineTypeComment EnvLineType = iota
	// LineTypeKeyValue represents a key-value assignment
	LineTypeKeyValue
)

// EnvLine represents a single line in an env file with its structure
type EnvLine struct {
	Type  EnvLineType
	Raw   string // The original line content
	Key   string // Only set for LineTypeKeyValue (trimmed)
	Value string // Only set for LineTypeKeyValue (original format)
}

// TemplateInfo holds information about a key from the template
type TemplateInfo struct {
	Line           string // Original line from template
	Value          string // Value part (may contain placeholders)
	HasPlaceholder bool   // Whether value contains ${...}
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
//
// NEW DESIGN PRINCIPLES:
// 1. The .env file is the primary source - it's always fully preserved
// 2. The .env.example defines which keys should exist
// 3. Missing keys from .env.example are added to .env with their comment groups
// 4. Placeholders ${...} in new keys trigger value generation
// 5. --force flag regenerates values for existing keys with placeholders in template
func (g *Generator) Generate() error {
	// STEP 1: Read template lines
	templateLines, err := g.readTemplateFile()
	if err != nil {
		return err
	}

	// STEP 2: Parse template to extract key information and line indices
	templateInfo := g.parseTemplateInfo(templateLines)

	// STEP 3: Check if .env file exists
	existingLines, err := g.readEnvFileWithStructure(g.config.OutputPath)
	outputExists := (err == nil)

	// Shared placeholder values across all operations
	placeholderValues := make(map[string]string)

	if !outputExists {
		// No existing .env file - create from template
		return g.generateFromTemplate(templateLines, templateInfo, placeholderValues)
	}

	// STEP 4: .env exists - preserve it and add missing keys
	// Build a map of existing keys
	existingKeys := make(map[string]bool)
	for _, envLine := range existingLines {
		if envLine.Type == LineTypeKeyValue {
			existingKeys[envLine.Key] = true
		}
	}

	// STEP 5: Find missing keys from template
	missingKeys := g.findMissingKeys(templateLines, existingKeys)

	// STEP 6: Handle --force flag - regenerate values for existing keys with placeholders
	var outputLines []string
	if g.config.Force {
		for _, envLine := range existingLines {
			if envLine.Type == LineTypeComment {
				outputLines = append(outputLines, envLine.Raw)
				continue
			}

			// Check if this key has a placeholder in template
			templateEntry, inTemplate := templateInfo[envLine.Key]
			if inTemplate && templateEntry.HasPlaceholder {
				// Regenerate value
				newValue, err := g.generateValueFromTemplate(templateEntry.Value, placeholderValues)
				if err != nil {
					return err
				}
				outputLines = append(outputLines, replaceValueInLine(envLine.Raw, newValue))
			} else {
				// Preserve as-is
				outputLines = append(outputLines, envLine.Raw)
			}
		}
	} else {
		// No force - preserve existing file completely
		for _, envLine := range existingLines {
			outputLines = append(outputLines, envLine.Raw)
		}
	}

	// STEP 7: Add missing keys with their comment groups
	if len(missingKeys) > 0 {
		// Add a separator if the file doesn't end with an empty line
		if len(outputLines) > 0 && strings.TrimSpace(outputLines[len(outputLines)-1]) != "" {
			outputLines = append(outputLines, "")
		}

		for _, key := range missingKeys {
			keyGroup := g.buildKeyGroupFromTemplate(templateLines, key, templateInfo, placeholderValues)
			outputLines = append(outputLines, keyGroup...)
		}
	}

	// STEP 8: Write output
	return g.writeOutputFile(outputLines)
}

// generateFromTemplate generates a new .env file from template (when .env doesn't exist)
func (g *Generator) generateFromTemplate(templateLines []string, templateInfo map[string]TemplateInfo, placeholderValues map[string]string) error {
	var outputLines []string

	for _, line := range templateLines {
		if isCommentOrEmpty(line) {
			outputLines = append(outputLines, line)
			continue
		}

		key, _, ok := parseKeyValue(line)
		if !ok {
			outputLines = append(outputLines, line)
			continue
		}

		templateEntry := templateInfo[key]
		if templateEntry.HasPlaceholder {
			// Generate value for placeholder
			newValue, err := g.generateValueFromTemplate(templateEntry.Value, placeholderValues)
			if err != nil {
				return err
			}
			outputLines = append(outputLines, replaceValueInLine(line, newValue))
		} else {
			// Use line as-is
			outputLines = append(outputLines, line)
		}
	}

	return g.writeOutputFile(outputLines)
}

// findMissingKeys returns keys that are in template but not in existing .env
func (g *Generator) findMissingKeys(templateLines []string, existingKeys map[string]bool) []string {
	var missingKeys []string
	seenKeys := make(map[string]bool)

	for _, line := range templateLines {
		if isCommentOrEmpty(line) {
			continue
		}

		key, _, ok := parseKeyValue(line)
		if !ok || seenKeys[key] {
			continue
		}

		seenKeys[key] = true
		if !existingKeys[key] {
			missingKeys = append(missingKeys, key)
		}
	}

	return missingKeys
}

// buildKeyGroupFromTemplate builds a group of lines for a key including its comment group
func (g *Generator) buildKeyGroupFromTemplate(templateLines []string, targetKey string, templateInfo map[string]TemplateInfo, placeholderValues map[string]string) []string {
	var result []string

	// Find the line index for this key
	keyLineIndex := -1
	for i, line := range templateLines {
		if isCommentOrEmpty(line) {
			continue
		}

		key, _, ok := parseKeyValue(line)
		if ok && key == targetKey {
			keyLineIndex = i
			break
		}
	}

	if keyLineIndex == -1 {
		return result
	}

	// Extract comment group before this key
	commentGroup := g.extractCommentGroup(templateLines, keyLineIndex)
	result = append(result, commentGroup...)

	// Add the key line with generated value if it has placeholder
	keyLine := templateLines[keyLineIndex]
	templateEntry := templateInfo[targetKey]

	if templateEntry.HasPlaceholder {
		newValue, err := g.generateValueFromTemplate(templateEntry.Value, placeholderValues)
		if err != nil {
			// If generation fails, use the line as-is
			result = append(result, keyLine)
		} else {
			result = append(result, replaceValueInLine(keyLine, newValue))
		}
	} else {
		result = append(result, keyLine)
	}

	return result
}

// extractCommentGroup extracts comment lines that belong to a key
// It looks backward from the key line to find all related comments
func (g *Generator) extractCommentGroup(lines []string, keyLineIndex int) []string {
	var comments []string

	// Look backward from the key line
	for i := keyLineIndex - 1; i >= 0; i-- {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			// Empty line - stop if we already have comments, otherwise skip
			if len(comments) > 0 {
				break
			}
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			// Comment line - add to the beginning
			comments = append([]string{line}, comments...)
		} else {
			// Non-comment, non-empty line - stop looking
			break
		}
	}

	return comments
}

// parseTemplateInfo parses template lines and extracts key information
func (g *Generator) parseTemplateInfo(lines []string) map[string]TemplateInfo {
	templateInfo := make(map[string]TemplateInfo)
	placeholderRe := regexp.MustCompile(`\${[^}]+}`)

	for _, line := range lines {
		if isCommentOrEmpty(line) {
			continue
		}

		key, value, ok := parseKeyValue(line)
		if !ok {
			continue
		}

		templateInfo[key] = TemplateInfo{
			Line:           line,
			Value:          value,
			HasPlaceholder: placeholderRe.MatchString(value),
		}
	}

	return templateInfo
}

// generateValueFromTemplate generates a value by replacing placeholders
func (g *Generator) generateValueFromTemplate(templateValue string, placeholderValues map[string]string) (string, error) {
	const escapeMarker = "##ESCAPED##"

	// Handle escaped placeholders
	value := strings.ReplaceAll(templateValue, `\${`, escapeMarker)

	placeholderRe := regexp.MustCompile(`\${([^}]+)}`)

	// Replace all placeholders
	result := placeholderRe.ReplaceAllStringFunc(value, func(match string) string {
		placeholderName := placeholderRe.FindStringSubmatch(match)[1]

		// Reuse existing value for same placeholder name
		if existingValue, exists := placeholderValues[placeholderName]; exists {
			return existingValue
		}

		// Generate new value
		newValue, err := g.generateSecureValue()
		if err != nil {
			return match // Keep placeholder on error
		}

		placeholderValues[placeholderName] = newValue
		return newValue
	})

	// Restore escaped placeholders
	result = strings.ReplaceAll(result, escapeMarker, `${`)

	return result, nil
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

// readEnvFileWithStructure reads an env file and returns structured line information
func (g *Generator) readEnvFileWithStructure(path string) ([]EnvLine, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []EnvLine
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		rawLine := scanner.Text()

		if isCommentOrEmpty(rawLine) {
			lines = append(lines, EnvLine{
				Type: LineTypeComment,
				Raw:  rawLine,
			})
		} else if key, value, ok := parseKeyValue(rawLine); ok {
			lines = append(lines, EnvLine{
				Type:  LineTypeKeyValue,
				Raw:   rawLine,
				Key:   key,
				Value: value,
			})
		} else {
			// Treat unparseable lines as comments
			lines = append(lines, EnvLine{
				Type: LineTypeComment,
				Raw:  rawLine,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// isCommentOrEmpty checks if a line is a comment or empty
func isCommentOrEmpty(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "" || strings.HasPrefix(trimmed, "#")
}

// parseKeyValue parses a key-value assignment from a line
// Preserves the original formatting of the value (including quotes and spaces)
func parseKeyValue(line string) (key, value string, ok bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	// Trim only the key, preserve the value as-is
	return strings.TrimSpace(parts[0]), parts[1], true
}

// replaceValueInLine replaces the value part of a key=value line while preserving the original format
func replaceValueInLine(originalLine, newValue string) string {
	// Find the first = sign
	idx := strings.Index(originalLine, "=")
	if idx == -1 {
		return originalLine
	}
	// Return key part (including =) + new value
	return originalLine[:idx+1] + newValue
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
	charset := getCharset(g.config.Charset)
	length := g.config.ValueLength

	result := make([]byte, length)
	randomBytes := make([]byte, length)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
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
