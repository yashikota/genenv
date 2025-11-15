# genenv

https://github.com/user-attachments/assets/1f53164a-d64c-4591-bc31-b4259c70fad8

A simple CLI tool to generate `.env` files from template files, automatically filling in placeholders with cryptographically secure random values  

*[日本語](README_ja.md)*

## Example

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
DB_PASSWORD=dGhpcyBpcyBhIHNlY3VyZSByYW5kb20gdmFsdWU

# API configuration
API_KEY=q7w8e9r0t1y2u3i4o5p6a7s8d9f0g1h2
API_URL=https://api.example.com
```

## Installation

Download the latest binary for your platform from the [GitHub Releases](https://github.com/yashikota/genenv/releases) page.

Or install using Go:

```bash
go install github.com/yashikota/genenv@latest
```

## Usage

```bash
genenv .env.example
```

When an existing `.env` file exists, existing field values are always preserved, and random values are generated only for new fields.

To preserve literal placeholders, escape them with a backslash: `\${not_a_placeholder}`

### Options

- `-f, --force`: Force regenerate all values including existing ones
  - `-y, --yes`: Skip confirmation prompt when using `--force`
- `-o, --output`: Specify output file path (default: `.env`)
- `-l, --length`: Length of generated random values (default: 24)
- `-c, --charset`: Character set for generated values
  - `alphanumeric` (default): A-Z, a-z, 0-9
  - `alphabetic`: A-Z, a-z
  - `uppercase`: A-Z
  - `lowercase`: a-z
  - `numeric`: 0-9
- `-h, --help`: Show help information
- `-v, --version`: Show version information

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

### Regeneration

By default, existing values in your `.env` file are preserved. To regenerate all values including existing ones, use the `--force` flag.  
Use `--yes` or `-y` to skip confirmation.

```bash
genenv --force --yes .env.example
```
