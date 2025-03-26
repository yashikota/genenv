package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yashikota/genenv/pkg/generator"
)

// Version information
const (
	Version = "0.1.0"
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

	// Register both short and long versions of flags
	flag.Var(&forceFlag, "force", "Force overwrite of existing .env file")
	flag.Var(&forceFlag, "f", "Force overwrite of existing .env file (shorthand)")

	flag.Var(&outputFlag, "output", "Output file path")
	flag.Var(&outputFlag, "o", "Output file path (shorthand)")
	outputFlag.value = ".env" // Default value

	flag.Var(&versionFlag, "version", "Show version information")
	flag.Var(&versionFlag, "v", "Show version information (shorthand)")

	flag.Var(&helpFlag, "help", "Show help information")
	flag.Var(&helpFlag, "h", "Show help information (shorthand)")

	flag.Var(&inputFlag, "input", "Read template from file instead of command line argument")
	flag.Var(&inputFlag, "i", "Read template from file instead of command line argument (shorthand)")

	flag.Var(&lengthFlag, "length", "Length of generated random values (default: 24)")
	flag.Var(&lengthFlag, "l", "Length of generated random values (shorthand)")
	lengthFlag.value = 24 // Default value

	flag.Var(&charsetFlag, "charset", "Character set for generated values: alphanumeric, alphabetic, uppercase, lowercase, numeric (default: alphanumeric)")
	flag.Var(&charsetFlag, "c", "Character set for generated values (shorthand)")
	charsetFlag.value = "alphanumeric" // Default value

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "genenv - A tool to generate .env files from templates\n\n")
		fmt.Fprintf(os.Stderr, "Usage: genenv [options] <template-file>\n")
		fmt.Fprintf(os.Stderr, "   or: genenv -i/--input <template-file> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -f, --force           Force overwrite of existing .env file\n")
		fmt.Fprintf(os.Stderr, "  -o, --output string   Output file path (default \".env\")\n")
		fmt.Fprintf(os.Stderr, "  -i, --input string    Read template from file instead of command line argument\n")
		fmt.Fprintf(os.Stderr, "  -l, --length int      Length of generated random values (default: 24)\n")
		fmt.Fprintf(os.Stderr, "  -c, --charset string  Character set for generated values (default: alphanumeric)\n")
		fmt.Fprintf(os.Stderr, "                        Valid options: alphanumeric, alphabetic, uppercase, lowercase, numeric\n")
		fmt.Fprintf(os.Stderr, "  -h, --help            Show help information\n")
		fmt.Fprintf(os.Stderr, "  -v, --version         Show version information\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  genenv .env.example\n")
		fmt.Fprintf(os.Stderr, "  genenv --input .env.example --output .env.production\n")
		fmt.Fprintf(os.Stderr, "  genenv -i .env.example -o .env.production -f\n")
		fmt.Fprintf(os.Stderr, "  genenv .env.example --length 32 --charset numeric\n")
		fmt.Fprintf(os.Stderr, "  genenv .env.example -l 16 -c uppercase\n")
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

	// Get template file path either from -i/--input flag or positional argument
	if inputFlag.set && inputFlag.value != "" {
		templatePath = inputFlag.value
	} else {
		// Get template file path from arguments
		args := flag.Args()
		if len(args) < 1 {
			fmt.Println("Error: Template file path is required")
			flag.Usage()
			os.Exit(1)
		}
		templatePath = args[0]
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
		TemplatePath: templatePath,
		OutputPath:   outputFlag.value,
		Force:        forceFlag.value,
		ValueLength:  lengthFlag.value,
		Charset:      charset,
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
	if err := gen.Generate(); err != nil {
		fmt.Printf("Error generating .env file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s from %s\n", config.OutputPath, templatePath)
}
