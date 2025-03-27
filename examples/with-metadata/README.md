# Metadata Example

This example demonstrates how to use metadata in your template files to provide additional information about fields, including whether they are required, their data type, and descriptions.

## Files

- `.env.example`: A template file with metadata comments
- `.env`: The generated environment file

## Features Demonstrated

- Field metadata using special comment format
- Required/optional field specification
- Field type specification
- Field descriptions

## Usage

```bash
# Generate .env with interactive mode to leverage metadata
genenv -i .env.example

# Or run without arguments to enter fully interactive mode
genenv
```

## Metadata Format

Metadata is specified in comments above each field using the following format:

```
# @field_name [required|optional] (type) Description
KEY=${field_name}
```

For example:

```
# @api_key [required] (string) API key for external service
API_KEY=${api_key}

# @debug_mode [optional] (bool) Enable debug mode
DEBUG=${debug_mode}
```

## Benefits of Using Metadata

When using metadata with interactive mode:

1. The tool will prompt you with the field description
2. Required fields must be filled in (can't be left empty)
3. The tool will validate input based on the specified type
4. Default values can be specified and used when input is empty

This makes it easier to create properly formatted `.env` files with valid values for each field.
