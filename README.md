# compress-env

[![GitHub Actions](https://github.com/hatoo/oha/workflows/CI/badge.svg)](https://github.com/hatoo/oha/actions?query=workflow%3ACI)
[![Crates.io](https://img.shields.io/crates/v/oha.svg)](https://crates.io/crates/oha)
[![Arch Linux](https://img.shields.io/archlinux/v/extra/x86_64/oha)](https://archlinux.org/packages/extra/x86_64/oha/)
[![Homebrew](https://img.shields.io/homebrew/v/oha)](https://formulae.brew.sh/formula/oha)

compress-env is a lightweight tool for securely compressing, encoding, and managing .env files in your projects. It allows you to easily backup and restore environment variables across different environments while maintaining their directory structure.



## Features

- **Secure Encoding**: Compress and encode .env files with optional password protection
- **Directory Structure Preservation**: Maintains the original directory structure when restoring
- **Selective Compression**: Compress specific .env files or all .env files in a directory tree
- **Simple Restoration**: Easily restore .env files to their original locations
- **Password Protection**: Add an extra layer of security with password-based encryption



## Download pre-built binary

You can download pre-built binary from [Release page](https://github.com/cavaldos/compress-env/releases).

## Usage

```bash
Usage: compress-env [COMMAND] [OPTIONS]
Commands:
  compress          - Compress and encode all .env files into a single encrypted binary file
  restore           - Decode and restore .env files from the encrypted binary file
Options:
  -o <filename>     - Specify the output filename for compression (default: .env.bin)
  -i <filename>     - Specify the input filename for restoration (default: .env.bin)
  -f <file1> <file2> ... - Specify specific .env files to compress
  -p <password>     - Specify a password for encryption/decryption (optional)
  -h, --help        - Show this help message
  -v, --version     - Show version information
```

### Examples

```bash
# Compress all .env files in the current directory tree
compress-env compress

# Compress all .env files with a custom output filename
compress-env compress -o secrets.bin

# Compress specific .env files with password protection
compress-env compress -f ./project1/.env ./project2/.env -o project-secrets.bin -p mypassword

# Restore .env files from the default backup file (.env.bin)
compress-env restore

# Restore .env files from a custom backup file with password
compress-env restore -i secrets.bin -p mypassword
```

## How it Works

### Compression Process

1. Searches for all .env files in your current directory and subdirectories (or uses specified files)
2. Aggregates the contents with path information into a single data structure
3. Applies encoding (with optional password-based encryption)
4. Stores the encoded data in a binary file (default: `.env.bin`)

### Restoration Process

1. Reads the encoded binary file
2. Decodes the content (using the provided password if encrypted)
3. Parses the file paths and contents
4. Recreates each .env file in its original location with its original content

## Security

When using the password option (`-p`), compress-env applies a simple but effective encryption to your environment variables. This adds an extra layer of protection for sensitive information like API keys, database credentials, and other secrets.

## Building from Source

```bash
# Clone the repository
git clone https://github.com/cavaldos/compress-env.git
cd compress-env

# Build the release version
cargo build --release

# The executable will be available at
./target/release/compress-env
```

## Tips

### Managing Multiple Environments

You can use compress-env to manage different environment configurations:

```bash
# Save production environment variables
compress-env compress -o prod.env.bin -p production-password

# Save development environment variables
compress-env compress -o dev.env.bin -p development-password

# Restore the appropriate environment as needed
compress-env restore -i prod.env.bin -p production-password
```

### Backup Before Deployment

Add compress-env to your deployment workflow to automatically backup environment variables:

```bash
# In your deployment script
compress-env compress -o backup-$(date +%Y%m%d).env.bin -p "$BACKUP_PASSWORD"
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request