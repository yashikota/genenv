package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/yashikota/genenv/pkg/generator"
)

// Version information
const (
	Version = "0.4.1"
)

// Define custom flag types to support both short and long options
type stringFlag struct {
	set   bool
	value string
}

func (sf *stringFlag) String() string {
	return sf.value
}

func (sf *stringFlag) Set(value string) error {
	sf.value = value
	sf.set = true
	return nil
}

type boolFlag struct {
	set   bool
	value bool
}

func (bf *boolFlag) String() string {
	if bf.value {
		return "true"
	}
	return "false"
}

func (bf *boolFlag) Set(value string) error {
	if value == "true" || value == "" {
		bf.value = true
	} else {
		bf.value = false
	}
	bf.set = true
	return nil
}

func (bf *boolFlag) IsBoolFlag() bool {
	return true
}

type intFlag struct {
	set   bool
	value int
}

func (inf *intFlag) String() string {
	return fmt.Sprintf("%d", inf.value)
}

func (inf *intFlag) Set(value string) error {
	var err error
	inf.value, err = parseInt(value)
	if err != nil {
		return err
	}
	inf.set = true
	return nil
}

// parseInt parses a string to an int
func parseInt(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func main() {
	// Define command line flags with both short and long options
	var forceFlag boolFlag
	var outputFlag stringFlag
	var versionFlag boolFlag
	var helpFlag boolFlag
	var inputFlag stringFlag
	var lengthFlag intFlag
	var charsetFlag stringFlag
	var interactiveFlag boolFlag
	var compareWithEnvFlag boolFlag
	var skipExistingFlag boolFlag

	// Register both short and long versions of flags
	flag.Var(&forceFlag, "force", "Force overwrite of existing .env file")
	flag.Var(&forceFlag, "f", "Force overwrite of existing .env file")

	flag.Var(&outputFlag, "output", "Output file path (default \".env\")")
	flag.Var(&outputFlag, "o", "Output file path (default \".env\")")
	outputFlag.value = ".env" // Default value

	flag.Var(&versionFlag, "version", "Show version information")
	flag.Var(&versionFlag, "v", "Show version information")

	flag.Var(&helpFlag, "help", "Show help information")
	flag.Var(&helpFlag, "h", "Show help information")

	flag.Var(&inputFlag, "input", "Read template from file instead of command line argument")
	flag.Var(&inputFlag, "i", "Read template from file instead of command line argument")

	flag.Var(&lengthFlag, "length", "Length of generated random values (default: 24)")
	flag.Var(&lengthFlag, "l", "Length of generated random values (default: 24)")
	lengthFlag.value = 24 // Default value

	flag.Var(&charsetFlag, "charset", "Character set for generated values: alphanumeric, alphabetic, uppercase, lowercase, numeric (default: alphanumeric)")
	flag.Var(&charsetFlag, "c", "Character set for generated values: alphanumeric, alphabetic, uppercase, lowercase, numeric (default: alphanumeric)")
	charsetFlag.value = "alphanumeric" // Default value

	flag.Var(&interactiveFlag, "interactive", "Run in interactive mode, prompting for values")
	flag.Var(&interactiveFlag, "I", "Run in interactive mode, prompting for values")

	flag.Var(&compareWithEnvFlag, "compare", "Compare with existing .env file and add only new fields")
	flag.Var(&compareWithEnvFlag, "C", "Compare with existing .env file and add only new fields")

	flag.Var(&skipExistingFlag, "skip-existing", "Skip fields that already exist in the .env file")
	flag.Var(&skipExistingFlag, "S", "Skip fields that already exist in the .env file")

	// Custom usage function
	oldUsage := flag.Usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "genenv - A tool to generate .env files from templates\n\n")
		fmt.Fprintf(os.Stderr, "Usage: genenv [options] <template-file>\n")
		fmt.Fprintf(os.Stderr, "   or: genenv -i/--input <template-file> [options]\n")
		fmt.Fprintf(os.Stderr, "   or: genenv (runs in interactive mode)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		oldUsage()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  genenv .env.example\n")
		fmt.Fprintf(os.Stderr, "  genenv --input .env.example --output .env.production\n")
		fmt.Fprintf(os.Stderr, "  genenv -i .env.example -o .env.production -f\n")
		fmt.Fprintf(os.Stderr, "  genenv .env.example --length 32 --charset numeric\n")
		fmt.Fprintf(os.Stderr, "  genenv .env.example -l 16 -c uppercase\n")
		fmt.Fprintf(os.Stderr, "  genenv -I -i .env.example (interactive mode)\n")
		fmt.Fprintf(os.Stderr, "  genenv -C -i .env.example (compare with existing .env)\n")
		fmt.Fprintf(os.Stderr, "  genenv -S -i .env.example (skip existing fields)\n\n")
		fmt.Fprintf(os.Stderr, "Template Format:\n")
		fmt.Fprintf(os.Stderr, "  # @field_name [required] (type) Description\n")
		fmt.Fprintf(os.Stderr, "  KEY=${field_name}\n\n")
		fmt.Fprintf(os.Stderr, "  Example:\n")
		fmt.Fprintf(os.Stderr, "  # @db_password [required] (string) Database password\n")
		fmt.Fprintf(os.Stderr, "  DB_PASSWORD=${db_password}\n")
	}

	// Parse flags
	flag.Parse()

	// Show version if requested
	if versionFlag.set && versionFlag.value {
		fmt.Printf("genenv version %s\n", Version)
		os.Exit(0)
	}

	// Show help if requested
	if helpFlag.set && helpFlag.value {
		flag.Usage()
		os.Exit(0)
	}

	var templatePath string
	runInteractive := false

	// Get template file path either from -i/--input flag or positional argument
	if inputFlag.set && inputFlag.value != "" {
		templatePath = inputFlag.value
	} else {
		// Get template file path from arguments
		args := flag.Args()
		if len(args) < 1 {
			// If no template file is provided, enter fully interactive mode
			runInteractive = true
			interactiveFlag.value = true

			// In fully interactive mode, prompt for all configuration options
			reader := bufio.NewReader(os.Stdin)

			fmt.Println("Welcome to genenv interactive mode!")
			fmt.Println("Press Enter to use default values.")

			// Prompt for template file
			fmt.Print("Enter template file path (.env.example): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				os.Exit(1)
			}

			templatePath = strings.TrimSpace(input)
			if templatePath == "" {
				templatePath = ".env.example"
			}

			// Prompt for output file
			fmt.Printf("Enter output file path (%s): ", outputFlag.value)
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				os.Exit(1)
			}

			outputValue := strings.TrimSpace(input)
			if outputValue != "" {
				outputFlag.value = outputValue
			}

			// Prompt for value length
			fmt.Printf("Enter length for generated values (%d): ", lengthFlag.value)
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				os.Exit(1)
			}

			lengthValue := strings.TrimSpace(input)
			if lengthValue != "" {
				length, err := strconv.Atoi(lengthValue)
				if err != nil {
					fmt.Printf("Invalid length value, using default: %d\n", lengthFlag.value)
				} else {
					lengthFlag.value = length
				}
			}

			// Prompt for charset
			fmt.Printf("Enter charset for generated values (alphanumeric, alphabetic, uppercase, lowercase, numeric) [%s]: ", charsetFlag.value)
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				os.Exit(1)
			}

			charsetValue := strings.TrimSpace(input)
			if charsetValue != "" {
				charsetFlag.value = charsetValue
			}

			// Prompt for compare with env
			fmt.Print("Compare with existing .env file? (y/N): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				os.Exit(1)
			}

			compareValue := strings.TrimSpace(strings.ToLower(input))
			compareWithEnvFlag.value = compareValue == "y" || compareValue == "yes"

			// Prompt for skip existing
			fmt.Print("Skip fields that already exist in the .env file? (y/N): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				os.Exit(1)
			}

			skipValue := strings.TrimSpace(strings.ToLower(input))
			skipExistingFlag.value = skipValue == "y" || skipValue == "yes"

			// Prompt for force overwrite
			fmt.Print("Force overwrite of existing .env file? (y/N): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				os.Exit(1)
			}

			forceValue := strings.TrimSpace(strings.ToLower(input))
			forceFlag.value = forceValue == "y" || forceValue == "yes"

			fmt.Println("\nConfiguration summary:")
			fmt.Printf("Template file: %s\n", templatePath)
			fmt.Printf("Output file: %s\n", outputFlag.value)
			fmt.Printf("Value length: %d\n", lengthFlag.value)
			fmt.Printf("Charset: %s\n", charsetFlag.value)
			fmt.Printf("Compare with existing .env: %t\n", compareWithEnvFlag.value)
			fmt.Printf("Skip existing fields: %t\n", skipExistingFlag.value)
			fmt.Printf("Force overwrite: %t\n", forceFlag.value)
			fmt.Println()
		} else {
			templatePath = args[0]
		}
	}

	// Validate charset
	charset := generator.CharsetType(charsetFlag.value)
	validCharsets := map[generator.CharsetType]bool{
		generator.CharsetAlphanumeric: true,
		generator.CharsetAlphabetic:   true,
		generator.CharsetUppercase:    true,
		generator.CharsetLowercase:    true,
		generator.CharsetNumeric:      true,
	}

	if !validCharsets[charset] {
		fmt.Printf("Error: Invalid charset '%s'. Valid options are: alphanumeric, alphabetic, uppercase, lowercase, numeric\n", charsetFlag.value)
		os.Exit(1)
	}

	// Create generator config
	config := generator.Config{
		TemplatePath:   templatePath,
		OutputPath:     outputFlag.value,
		Force:          forceFlag.value,
		ValueLength:    lengthFlag.value,
		Charset:        charset,
		Interactive:    interactiveFlag.value || runInteractive,
		CompareWithEnv: compareWithEnvFlag.value,
		SkipExisting:   skipExistingFlag.value,
	}

	// Create generator
	gen := generator.New(config)

	// Check if output file exists and prompt for confirmation if not forced
	if _, err := os.Stat(config.OutputPath); err == nil && !config.Force {
		fmt.Printf("File %s already exists. Overwrite? (y/N): ", config.OutputPath)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Operation cancelled")
			os.Exit(0)
		}

		// Set force to true to skip the check in the generator
		config.Force = true
		gen = generator.New(config)
	}

	// Generate .env file
	var err error
	if config.Interactive {
		err = gen.RunInteractive()
	} else {
		err = gen.Generate()
	}

	if err != nil {
		fmt.Printf("Error generating .env file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s from %s\n", config.OutputPath, templatePath)
}

// createTemporaryTemplate creates a temporary template file with common environment variables
func createTemporaryTemplate() (string, error) {
	tempFile, err := os.CreateTemp("", "genenv-template-*.env")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary template file: %w", err)
	}
	defer tempFile.Close()

	// Write common environment variables to the template
	templateContent := `# Database configuration
# @db_host [optional] (string) Database host
DB_HOST=${db_host}

# @db_port [optional] (int) Database port
DB_PORT=${db_port}

# @db_name [required] (string) Database name
DB_NAME=${db_name}

# @db_user [required] (string) Database username
DB_USER=${db_user}

# @db_password [required] (string) Database password
DB_PASSWORD=${db_password}

# API configuration
# @api_key [optional] (string) API key for external service
API_KEY=${api_key}

# @api_url [optional] (url) API URL
API_URL=${api_url}

# Application configuration
# @app_env [optional] (string) Application environment (development, production, testing)
APP_ENV=${app_env}

# @debug [optional] (bool) Enable debug mode
DEBUG=${debug}

# @log_level [optional] (string) Log level (debug, info, warn, error)
LOG_LEVEL=${log_level}

# @secret_key [required] (string) Secret key for encryption/signing
SECRET_KEY=${secret_key}

# Add your custom environment variables below
`

	if _, err := tempFile.WriteString(templateContent); err != nil {
		return "", fmt.Errorf("failed to write to temporary template file: %w", err)
	}

	return tempFile.Name(), nil
}
