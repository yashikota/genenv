package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yashikota/genenv/internal/generator"
)

const (
	Version = "1.1.0"
)

func main() {
	force := flag.Bool("force", false, "Force regenerate all values including existing ones")
	flag.BoolVar(force, "f", false, "Force regenerate all values including existing ones")

	yes := flag.Bool("yes", false, "Skip confirmation prompt when using --force")
	flag.BoolVar(yes, "y", false, "Skip confirmation prompt when using --force")

	output := flag.String("output", ".env", "Output file path")
	flag.StringVar(output, "o", ".env", "Output file path")

	length := flag.Int("length", 24, "Length of generated random values")
	flag.IntVar(length, "l", 24, "Length of generated random values")

	charset := flag.String("charset", "alphanumeric", "Character set for generated values: alphanumeric, alphabetic, uppercase, lowercase, numeric")
	flag.StringVar(charset, "c", "alphanumeric", "Character set for generated values: alphanumeric, alphabetic, uppercase, lowercase, numeric")

	version := flag.Bool("version", false, "Show version information")
	flag.BoolVar(version, "v", false, "Show version information")

	help := flag.Bool("help", false, "Show help information")
	flag.BoolVar(help, "h", false, "Show help information")

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "genenv - A tool to generate .env files from templates\n\n")
		fmt.Fprintf(os.Stderr, "Usage: genenv [options] <template-file>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  genenv .env.example\n")
		fmt.Fprintf(os.Stderr, "  genenv .env.example --output .env.production\n")
		fmt.Fprintf(os.Stderr, "  genenv .env.example --length 32 --charset numeric\n")
	}

	reorderArgs()

	flag.Parse()

	// Show version if requested
	if *version {
		fmt.Printf("genenv version %s\n", Version)
		os.Exit(0)
	}

	// Show help if requested
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Get template file path from arguments
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(0)
	}
	templatePath := args[0]

	// Validate charset
	charsetType := generator.CharsetType(*charset)
	if !isValidCharset(charsetType) {
		fmt.Printf("Error: Invalid charset '%s'. Valid options are: alphanumeric, alphabetic, uppercase, lowercase, numeric\n", *charset)
		os.Exit(1)
	}

	// Create generator config
	config := generator.Config{
		TemplatePath: templatePath,
		OutputPath:   *output,
		Force:        *force,
		ValueLength:  *length,
		Charset:      charsetType,
	}

	// Prompt for confirmation only when --force is used without --yes
	if config.Force && !*yes && fileExists(config.OutputPath) {
		if !promptOverwrite(config.OutputPath) {
			fmt.Println("Operation cancelled")
			os.Exit(0)
		}
	}

	// Create generator and generate .env file
	gen := generator.New(config)
	if err := gen.Generate(); err != nil {
		fmt.Printf("Error generating .env file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s from %s\n", config.OutputPath, templatePath)
}

// isValidCharset checks if the given charset is valid
func isValidCharset(charset generator.CharsetType) bool {
	validCharsets := map[generator.CharsetType]bool{
		generator.CharsetAlphanumeric: true,
		generator.CharsetAlphabetic:   true,
		generator.CharsetUppercase:    true,
		generator.CharsetLowercase:    true,
		generator.CharsetNumeric:      true,
	}
	return validCharsets[charset]
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// promptOverwrite prompts the user for confirmation to regenerate all values
func promptOverwrite(path string) bool {
	fmt.Printf("File %s already exists. Regenerate all values? (y/N): ", path)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func reorderArgs() {
	if len(os.Args) <= 1 {
		return
	}

	args := os.Args[1:]
	var flags []string
	var positional []string

	// Bool flags that don't take values
	boolFlags := map[string]bool{
		"-f": true, "--force": true,
		"-y": true, "--yes": true,
		"-v": true, "--version": true,
		"-h": true, "--help": true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Check if it's a flag
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)

			// Check if this flag expects a value
			if !boolFlags[arg] && !strings.Contains(arg, "=") {
				// Check if there's a next argument and it's not a flag
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					flags = append(flags, args[i])
				}
			}
		} else {
			// It's a positional argument
			positional = append(positional, arg)
		}
	}

	os.Args = append([]string{os.Args[0]}, append(flags, positional...)...)
}
