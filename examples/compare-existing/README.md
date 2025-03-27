# Compare with Existing .env Example

This example demonstrates how to use `genenv` to add new fields to an existing `.env` file without overwriting existing values.

## Files

- `.env.example`: A template file with both existing and new fields
- `.env`: An existing environment file that will be updated with new fields

## Features Demonstrated

- Comparing with an existing `.env` file
- Adding only new fields to the existing file
- Preserving existing values
- Identifying which fields are new during interactive mode

## Usage

```bash
# Compare with existing .env and add new fields
genenv -c .env.example

# Skip fields that already exist in the .env file
genenv -c -s .env.example

# Interactive mode with comparison
genenv -i -c .env.example

# Or run without arguments to enter fully interactive mode
genenv
```

## How It Works

When using the compare option (`-c` or `--compare`):

1. `genenv` reads the existing `.env` file
2. It identifies which fields from the template are missing in the existing file
3. It generates values only for the missing fields
4. In interactive mode, it marks fields as `[NEW]` if they don't exist in the current `.env`

## Benefits

This feature is particularly useful when:

- Updating an application with new configuration options
- Adding new features that require additional environment variables
- Ensuring your `.env` file is up-to-date with the latest template
- Migrating between versions without losing your existing configuration

Using the compare option ensures you don't lose your custom configuration while still getting all the new fields you need.
