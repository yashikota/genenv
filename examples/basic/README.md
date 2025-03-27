# Basic Example

This example demonstrates the basic usage of `genenv` to generate a `.env` file from a template.

## Files

- `.env.example`: A template file with placeholders
- `.env`: The generated environment file with placeholders replaced by random values

## Features Demonstrated

- Basic placeholder replacement (`${placeholder_name}`)
- Preservation of fixed values
- Comment preservation

## Usage

```bash
# Generate .env from .env.example
genenv .env.example

# Generate with a custom output path
genenv .env.example .env.production

# Force overwrite if .env already exists
genenv -f .env.example
```

## Template Format

The basic template format uses `${placeholder_name}` syntax for values that should be replaced with random strings:

```
APP_KEY=${app_key}
DB_USERNAME=${db_username}
DB_PASSWORD=${db_password}
```

When processed, these placeholders will be replaced with cryptographically secure random values:

```
APP_KEY=a1b2c3d4e5f6g7h8i9j0k1l2m3
DB_USERNAME=dbuser123
DB_PASSWORD=p@ssw0rd123!
```

Fixed values without placeholders will be preserved as-is:

```
APP_NAME=My App
APP_ENV=development
APP_DEBUG=true
```
