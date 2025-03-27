# Escaped Placeholders Example

This example demonstrates how to use escaped placeholders in your template files when you want to include `${...}` syntax in your actual values without having them replaced.

## Files

- `.env.example`: A template file with both normal and escaped placeholders
- `.env`: The generated environment file showing how escaped placeholders are preserved

## Features Demonstrated

- Normal placeholder replacement
- Escaped placeholder preservation using backslash (`\${...}`)
- Mixed usage of both types in the same file or even the same line

## Usage

```bash
# Generate .env from the template
genenv .env.example

# Or with interactive mode
genenv -i .env.example
```

## Escaping Syntax

To include a literal `${...}` in your output file without having it replaced by a random value, escape it with a backslash:

```
# This will be replaced with a random value
API_KEY=${api_key}

# This will be preserved as-is in the output
TEMPLATE_EXAMPLE=This is a \${placeholder} that will not be replaced
```

## Use Cases

Escaped placeholders are useful when:

1. You're working with templates that need to include the `${...}` syntax
2. You're generating configuration for systems that use similar placeholder syntax
3. You need to include example placeholders in documentation strings
4. You're working with dollar signs in strings (like currency values)

This feature ensures you can generate `.env` files with the exact syntax you need, even when it includes placeholder-like patterns.
