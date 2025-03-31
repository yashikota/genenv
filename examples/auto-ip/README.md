# Automatic IP Address Detection

This example demonstrates genenv's ability to automatically detect and insert your local IP addresses into `.env` files.

## Features Shown

- Automatic detection and insertion of local IP addresses in templates
- Support for both IPv4 and IPv6 address detection
- Field type validation for IP addresses
- Metadata with field descriptions

## How It Works

When using a field with the type `ip`, `ipv4`, or `ipv6`, genenv will automatically detect your local IP address and use it to fill in the placeholder.

### IP Field Types

- `ip`: Detects and uses any available IP address (prefers IPv4)
- `ipv4`: Specifically detects and uses IPv4 addresses only
- `ipv6`: Specifically detects and uses IPv6 addresses only

## Template Format

The template file contains fields with IP metadata types:

```bash
# @server_ip [required] (ip) Main server IP address (IPv4 or IPv6)
SERVER_IP=${server_ip}

# @server_ipv4 [required] (ipv4) Server IPv4 address
SERVER_IPV4=${server_ipv4}

# @server_ipv6 [optional] (ipv6) Server IPv6 address
SERVER_IPV6=${server_ipv6}
```

## Usage

```bash
genenv .env.example
```

This will generate a `.env` file with your actual local IP addresses inserted in place of the placeholders.

## Interactive Mode

In interactive mode, genenv will detect your IP addresses and suggest them as defaults:

```bash
genenv -I .env.example
```

When prompted for an IP address field, you can simply press Enter to use the detected IP address.
