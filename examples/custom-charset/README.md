# Custom Character Sets Example

This example demonstrates how to use `genenv` with custom character sets and value lengths for generated values.

## Files

- `.env.example`: A template file with placeholders for different character sets
- `.env`: The generated environment file with values using different character sets

## Features Demonstrated

- Custom character sets for generated values
- Custom lengths for generated values
- Different character set options for different security needs

## Usage

```bash
# Generate with alphanumeric character set (default)
genenv .env.example

# Generate with alphabetic character set (letters only)
genenv -a alphabetic .env.example

# Generate with uppercase character set
genenv -a uppercase .env.example

# Generate with lowercase character set
genenv -a lowercase .env.example

# Generate with numeric character set
genenv -a numeric .env.example

# Generate with custom length (10 characters)
genenv -l 10 .env.example

# Combine character set and length
genenv -a uppercase -l 32 .env.example

# Interactive mode with custom options
genenv -i .env.example
```

## Available Character Sets

`genenv` supports the following character sets:

- `alphanumeric`: Uppercase letters, lowercase letters, and numbers (default)
- `alphabetic`: Uppercase and lowercase letters only
- `uppercase`: Uppercase letters only
- `lowercase`: Lowercase letters only
- `numeric`: Numbers only

## Use Cases

Custom character sets are useful when:

1. You need to generate values for systems with specific character requirements
2. You want to create more readable keys (alphabetic only)
3. You need numeric-only values for certain systems
4. You need shorter or longer values for specific security requirements

This feature gives you flexibility in generating values that meet your specific needs.
