# Complex Example

This example demonstrates a comprehensive use case for `genenv` with a complex application configuration, combining multiple features.

## Files

- `.env.example`: A complex template file with various field types and metadata
- `.env`: The generated environment file with all values filled in

## Features Demonstrated

- Comprehensive application configuration
- Combination of required and optional fields
- Multiple field types for validation
- Detailed field descriptions
- Nested variable references (`MAIL_FROM_NAME="${APP_NAME}"`)

## Usage

```bash
# Generate .env with interactive mode
genenv -i .env.example

# Or run without arguments to enter fully interactive mode
genenv
```

## Application Structure

This example simulates a full-featured web application with:

1. **Application Settings**: Basic app configuration
2. **Database Settings**: MySQL database connection details
3. **Redis Settings**: Cache configuration
4. **Mail Settings**: Email service configuration
5. **AWS Settings**: Cloud storage configuration
6. **Pusher Settings**: Real-time messaging configuration

## Benefits of This Approach

Using a comprehensive template with metadata:

1. **Documentation**: The template itself serves as documentation for the configuration
2. **Validation**: Field types ensure values are in the correct format
3. **Required Fields**: Critical configuration won't be missed
4. **Consistency**: Ensures all environments have the same configuration structure

This approach is ideal for complex applications where proper configuration is critical for functionality.
