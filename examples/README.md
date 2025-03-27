# genenv Examples

This directory contains various examples demonstrating the features of the `genenv` tool.

## Example Categories

1. [Basic](./basic/): Simple placeholder replacement
2. [With Metadata](./with-metadata/): Field metadata and validation
3. [With Types](./with-types/): Field type validation
4. [Complex](./complex/): Comprehensive real-world example
5. [Escaped Placeholders](./escaped/): Preserving literal `${...}` syntax
6. [Compare with Existing](./compare-existing/): Adding new fields to existing `.env` files
7. [Custom Character Sets](./custom-charset/): Using different character sets and lengths
8. [New From Scratch](./new-from-scratch/): Creating a new `.env` file without a template

## Running the Examples

Each example directory contains:

- A `.env.example` template file
- A sample `.env` file showing the expected output
- A `README.md` with detailed explanations and usage instructions

To try an example, navigate to its directory and run:

```bash
# Basic usage
genenv .env.example

# Interactive mode
genenv -i .env.example

# With specific options (see each example's README for details)
genenv -f -c -l 32 -a alphabetic .env.example
```

## Testing the Examples

You can run automated tests for all examples using the provided test script:

```bash
# Navigate to the examples directory
cd examples

# Build genenv first (if not already built)
go build -o genenv ../cmd/genenv/main.go

# Run the tests
go run test_examples.go
```

The test script will:

1. Test each example with appropriate options
2. Verify that the generated files match expected patterns
3. Report success or failure for each example
4. Provide a summary of test results

Note: Some tests require interactive input and are skipped in the automated tests. You can try these examples manually following the instructions in their README files.
