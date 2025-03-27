# Field Types Example

This example demonstrates the various field types supported by `genenv` for validation in interactive mode.

## Files

- `.env.example`: A template file with different field types
- `.env`: The generated environment file with validated values

## Features Demonstrated

- Field type validation for various data types
- Type-specific prompts and validation in interactive mode

## Supported Field Types

`genenv` supports the following field types:

- `string`: Text values (default)
- `int`/`integer`: Integer values
- `bool`/`boolean`: Boolean values (true/false, yes/no, 1/0)
- `float`/`double`: Floating-point values
- `url`: URL values (must start with http:// or https://)
- `email`: Email addresses
- `ip`: IPv4 addresses

## Usage

```bash
# Generate .env with interactive mode to leverage type validation
genenv -i .env.example

# Or run without arguments to enter fully interactive mode
genenv
```

## Type Validation Examples

When using interactive mode, the tool will validate input based on the field type:

- For `int` fields: Only numeric values are accepted
- For `bool` fields: Only true/false, yes/no, or 1/0 are accepted
- For `float` fields: Only numeric values with optional decimal points are accepted
- For `url` fields: Must start with http:// or https://
- For `email` fields: Must contain @ and . characters without spaces
- For `ip` fields: Must be a valid IPv4 address (e.g., 192.168.1.1)

This ensures that your `.env` file contains properly formatted values for each field.
