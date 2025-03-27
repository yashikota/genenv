# genenv

A simple CLI tool to generate `.env` files from template files, automatically filling in placeholders with cryptographically secure random values.

*[日本語](README_ja.md)*

## Features

- Generate `.env` files from template files (e.g., `.env.example`)
- Automatically replace placeholders with cryptographically secure random values
- Preserve existing values in the template
- Option to force overwrite existing `.env` files
- Customizable output file path
- Customizable length and character set of generated values
- Properly handle escaped placeholders (`\${not_a_placeholder}`)
- Interactive mode with field validation and type checking
- Support for comparing with existing `.env` files and adding only new fields
- Option to skip fields that already exist in the `.env` file
- Create a new `.env` file from scratch without a template

## Examples

The `examples` directory contains various examples demonstrating the features of `genenv`:

- **Basic**: Simple placeholder replacement
- **With Metadata**: Field metadata and validation
- **With Types**: Field type validation
- **Complex**: Comprehensive real-world example
- **Escaped Placeholders**: Preserving literal `${...}` syntax
- **Compare with Existing**: Adding new fields to existing `.env` files
- **Custom Character Sets**: Using different character sets and lengths
- **New From Scratch**: Creating a new `.env` file without a template

Each example includes:

- A `.env.example` template file
- A sample `.env` file showing the expected output
- A detailed README explaining the feature and usage

### Running the Examples

Navigate to an example directory and run:

```bash
genenv .env.example
```

### Testing the Examples

You can run automated tests for all examples using the provided test script:

```bash
# Navigate to the examples directory
cd examples

# Build genenv first (if not already built)
go build -o genenv ../cmd/genenv/main.go

# Run the tests
go run test_examples.go
```

All examples are also automatically tested in the CI pipeline to ensure they work correctly.

## Installation

### From GitHub Releases

Download the latest binary for your platform from the [GitHub Releases](https://github.com/yashikota/genenv/releases) page.

### Go install

```bash
go install github.com/yashikota/genenv/cmd/genenv@latest
```

## Usage

Basic usage:

```bash
genenv .env.example
```

This will generate a `.env` file from the `.env.example` template.

Interactive mode (no arguments):

```bash
genenv
```

This will start the interactive mode, prompting for all configuration options and field values.

Create a new `.env` file from scratch:

```bash
genenv -N
```

This will create a new `.env` file with common environment variables without requiring a template file.

### Options

- `-f, --force`: Force overwrite of existing `.env` file
- `-o, --output`: Specify output file path (default: `.env`)
- `-i, --input`: Read template from file instead of command line argument
- `-l, --length`: Length of generated random values (default: 24)
- `-c, --charset`: Character set for generated values (default: alphanumeric)
  - Valid options: `alphanumeric`, `alphabetic`, `uppercase`, `lowercase`, `numeric`
- `-I, --interactive`: Run in interactive mode, prompting for values
- `-C, --compare`: Compare with existing `.env` file and add only new fields
- `-S, --skip-existing`: Skip fields that already exist in the `.env` file
- `-N, --new`: Create a new `.env` file from scratch without a template
- `-h, --help`: Show help information
- `-v, --version`: Show version information

### Example

If your `.env.example` file looks like this:

```txt
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=${db_user}
DB_PASSWORD=${db_password}

# API configuration
API_KEY=${api_key}
API_URL=https://api.example.com
```

Running `genenv .env.example` will generate a `.env` file like:

```txt
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
DB_PASSWORD=dGhpcyBpcyBhIHNlY3VyZSByYW5kb20gdmFsdWU=

# API configuration
API_KEY=q7w8e9r0t1y2u3i4o5p6a7s8d9f0g1h2
API_URL=https://api.example.com
```

Note that:

- Values without placeholders (`localhost`, `5432`, `https://api.example.com`) are preserved
- Placeholders (`${db_user}`, `${db_password}`, `${api_key}`) are replaced with unique, cryptographically secure random values

### Customizing Length and Character Set

You can customize the length and character set of generated values using the `-l` and `-c` options.

```bash
# Generate .env file with custom length (32 characters)
genenv -l 32 .env.example

# Generate .env file with custom character set (numeric only)
genenv -c numeric .env.example

# Generate .env file with custom length and character set (16 uppercase letters)
genenv -l 16 -c uppercase .env.example
```

### Character Sets

genenv supports the following character sets for generated values:

- `alphanumeric` (default): A-Z, a-z, 0-9
- `alphabetic`: A-Z, a-z
- `uppercase`: A-Z
- `lowercase`: a-z
- `numeric`: 0-9

### Interactive Mode

Running `genenv` without any arguments or with the `-I/--interactive` flag will start the interactive mode. In this mode:

1. You'll be prompted for all configuration options (template file, output file, etc.)
2. For each field in the template, you'll be prompted for a value
3. Field prompts include:
   - Field description (if available)
   - Whether the field is required or optional
   - Field type (string, int, bool, etc.)
   - Default value (if available)
   - Current value (if exists in the current .env file)

Example of interactive mode:

```bash
$ genenv

Welcome to genenv interactive mode!
Press Enter to use default values.
Enter template file path (.env.example): 
Enter output file path (.env): 
Enter length for generated values (24): 
Enter charset for generated values (alphanumeric, alphabetic, uppercase, lowercase, numeric) [alphanumeric]: 
Compare with existing .env file? (y/N): y
Skip fields that already exist in the .env file? (y/N): y
Force overwrite of existing .env file? (y/N): y

Configuration summary:
Template file: .env.example
Output file: .env
Value length: 24
Charset: alphanumeric
Compare with existing .env: true
Skip existing fields: true
Force overwrite: true

DB_HOST (Database host, [OPTIONAL], type: string): localhost
DB_PORT (Database port, [OPTIONAL], type: int): 5432
DB_USER (Database username, [REQUIRED], type: string): admin
DB_PASSWORD (Database password, [REQUIRED], type: string): 
Generated random value: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
API_KEY (API key for external service, [REQUIRED], type: string): 
Generated random value: q7w8e9r0t1y2u3i4o5p6a7s8d9f0g1h2

Successfully generated .env from .env.example
```

### Creating a New .env File From Scratch

You can create a new `.env` file from scratch without needing a template file using the `-N/--new` flag:

```bash
genenv -N
```

This will create a temporary template with common environment variables and guide you through filling them in interactively. The temporary template includes:

- Database configuration (host, port, name, user, password)
- API configuration (key, URL)
- Application configuration (environment, debug mode, log level, secret key)

You can also run `genenv` without arguments and choose to create a new file from scratch when prompted.

### Template Metadata Format

You can add metadata to your template file to provide validation and description for fields. The metadata is specified in comments before the field definition:

```txt
# @field_name [required] (type) Description
KEY=${field_name}
```

Example:

```txt
# @db_password [required] (string) Database password
DB_PASSWORD=${db_password}

# @db_port [optional] (int) Database port
DB_PORT=${db_port}

# @debug_mode [optional] (bool) Enable debug mode
DEBUG=${debug_mode}
```

Supported field types:

- `string`: Text value (default)
- `int`/`integer`: Integer value
- `bool`/`boolean`: Boolean value (true/false, yes/no, 1/0)
- `float`/`double`: Floating point value
- `url`: URL value (must start with http:// or https://)
- `email`: Email address
- `ip`: IPv4 address

When running in interactive mode, the tool will validate input based on the field type and show appropriate prompts with the field description.

### Comparing with Existing .env Files

Using the `-C/--compare` flag, genenv will compare the template with an existing `.env` file and only add new fields. This is useful when updating your `.env` file with new fields from an updated template.

```bash
genenv -C .env.example
```

### Skipping Existing Fields

Using the `-S/--skip-existing` flag, genenv will skip fields that already exist in the `.env` file. This is useful when you want to keep existing values for fields that are already defined.

```bash
genenv -S .env.example
```

### Template Format

The template file should be in the format of a standard `.env` file, with placeholders for values that should be generated.

```txt
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=${db_user}
DB_PASSWORD=${db_password}

# API configuration
API_KEY=${api_key}
API_URL=https://api.example.com

# Other settings
DEBUG=true
SECRET_TOKEN=${secret_token}
CACHE_TTL=3600
```

In this example, `${db_user}`, `${db_password}`, `${api_key}`, and `${secret_token}` will be replaced with generated values, while the other values will be preserved.

To include a literal `${...}` in your template without it being replaced, escape it with a backslash: `\${not_a_placeholder}`.
