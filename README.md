# genenv

A simple CLI tool to generate `.env` files from template files, automatically filling in placeholders with cryptographically secure random values.

## Features

- Generate `.env` files from template files (e.g., `.env.example`)
- Automatically replace placeholders with cryptographically secure random values
- Preserve existing values in the template
- Option to force overwrite existing `.env` files
- Customizable output file path
- Customizable length and character set of generated values
- Properly handle escaped placeholders (`\${not_a_placeholder}`)

## Installation

### From GitHub Releases

Download the latest binary for your platform from the [GitHub Releases](https://github.com/yashikota/genenv/releases) page.

### Go install

```bash
go install github.com/yashikota/genenv@latest
```

## Usage

Basic usage:

```bash
genenv .env.example
```

This will generate a `.env` file from the `.env.example` template.

### Options

- `-f, --force`: Force overwrite of existing `.env` file
- `-o, --output`: Specify output file path (default: `.env`)
- `-i, --input`: Read template from file instead of command line argument
- `-l, --length`: Length of generated random values (default: 24)
- `-c, --charset`: Character set for generated values (default: alphanumeric)
  - Valid options: `alphanumeric`, `alphabetic`, `uppercase`, `lowercase`, `numeric`
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
