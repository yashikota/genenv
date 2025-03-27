# Creating a New .env File From Scratch

This example demonstrates how to use `genenv` to create a new `.env` file from scratch without an existing template.

## Files

- `.env`: A sample environment file created from scratch
- `.env.example`: An example template showing the format with placeholders

## Features Demonstrated

- Creating a new `.env` file without a template
- Interactive creation of environment variables
- Using the built-in template for common environment variables

## Usage

```bash
# Create a new .env file from scratch in interactive mode
genenv -N

# Or run without arguments to enter fully interactive mode
genenv
# Then select "new" when prompted for the template file
```

## How It Works

When using the new file option (`-N` or `--new`):

1. `genenv` creates a temporary template with common environment variables
2. It prompts you for values for each variable in interactive mode
3. You can add or remove variables as needed
4. The tool generates a new `.env` file with your specified values

## Interactive Process

In interactive mode, you'll be prompted for:

1. The output file path
2. Whether to force overwrite if the file exists
3. The length of generated values
4. The character set to use
5. Values for each environment variable

## Benefits

This feature is particularly useful when:

- Starting a new project without an existing template
- Creating configuration files for different environments
- Setting up a new development environment
- You need a quick way to generate a standard `.env` file

Creating a new `.env` file from scratch ensures you have all the common environment variables you need without having to manually create a template first.
