<div align="center">

# 🔐 compress-env

**Secure .env File Management Made Simple**

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/cavaldos/compress-env)](https://github.com/cavaldos/compress-env/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub stars](https://img.shields.io/github/stars/cavaldos/compress-env?style=social)](https://github.com/cavaldos/compress-env/stargazers)
[![GitHub issues](https://img.shields.io/github/issues/cavaldos/compress-env)](https://github.com/cavaldos/compress-env/issues)

<p align="center">
  <i>A lightweight Go tool for securely compressing, encoding, and managing .env files in your projects</i>
</p>

</div>

## 📋 Table of Contents

- [Features](#features)
- [Why compress-env?](#why-compress-env)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Uninstallation](#uninstallation)
- [Usage](#usage)
- [How it Works](#how-it-works)
- [Security](#security)
- [Tips](#tips)
- [License](#license)
- [Contributing](#contributing)
- [Contact & Support](#contact--support)

## Why compress-env?

Managing environment variables across different development environments, staging servers, and production deployments can be challenging. compress-env solves this problem by:

- **Simplifying backup and restoration** of .env files with a single command
- **Enhancing security** with optional password-based encryption
- **Preserving directory structures** when working with complex projects
- **Reducing the risk of exposing sensitive credentials** in your version control system
- **Working seamlessly** across different operating systems

Whether you're a solo developer or part of a team, compress-env helps you manage your environment variables efficiently and securely.

## Quick Start

```bash
# Install compress-env (picks the right build for your machine)
os=$(uname -s | tr '[:upper:]' '[:lower:]')                  # linux | darwin
arch=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')   # amd64 | arm64
curl -L "https://github.com/cavaldos/compress-env/releases/latest/download/compress-env-${os}-${arch}" -o compress-env
chmod +x compress-env

# Open the interactive menu (compress / restore / generate keys)
./compress-env

# Or use the commands directly
./compress-env compress
./compress-env restore
```

## Features

- **Secure Encoding**: Compress and encode .env files with optional password protection
- **Directory Structure Preservation**: Maintains the original directory structure when restoring
- **Selective Compression**: Compress specific .env files or all .env files in a directory tree
- **Simple Restoration**: Easily restore .env files to their original locations
- **Password Protection**: Add an extra layer of security with password-based encryption

## Installation

### Option 1: Download and Install Pre-built Binary

1. Download the pre-built binary for your platform from the [Release page](https://github.com/cavaldos/compress-env/releases). Files are named `compress-env-<os>-<arch>` (`linux`, `darwin`, `windows` × `amd64`, `arm64`), and `checksums.txt` lists their SHA-256.

   **For Linux/macOS:**
   ```bash
os=$(uname -s | tr '[:upper:]' '[:lower:]')                  # linux | darwin
arch=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')   # amd64 | arm64
curl -L "https://github.com/cavaldos/compress-env/releases/latest/download/compress-env-${os}-${arch}" -o compress-env
chmod +x compress-env

   # Or pin a version (replace v0.0.1 with the desired version)
   # curl -L https://github.com/cavaldos/compress-env/releases/download/v0.0.1/compress-env-${os}-${arch} -o compress-env
   ```

   **For Windows (using PowerShell):**
   ```powershell
   # Download the latest release (use arm64 instead of amd64 on ARM machines)
   curl.exe -L https://github.com/cavaldos/compress-env/releases/latest/download/compress-env-windows-amd64.exe -o compress-env.exe

   # Or pin a version (replace v0.0.1 with the desired version)
   # curl.exe -L https://github.com/cavaldos/compress-env/releases/download/v0.0.1/compress-env-windows-amd64.exe -o compress-env.exe
   ```

2. Make the binary executable:
   ```bash
   chmod +x compress-env
   ```

3. Move the binary to a directory in your system PATH:

   **For Linux/macOS:**
   ```bash
   # Move to /usr/local/bin (requires sudo)
   sudo mv compress-env /usr/local/bin/

   # Or move to ~/.local/bin (doesn't require sudo)
   mkdir -p ~/.local/bin
   mv compress-env ~/.local/bin/

   # Make sure ~/.local/bin is in your PATH
   echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
   # Or if you use zsh
   echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
   ```

   **For Windows:**
   ```powershell
   # Create a directory for binaries if it doesn't exist
   mkdir -p $env:USERPROFILE\bin

   # Move the binary
   move compress-env.exe $env:USERPROFILE\bin\

   # Add to PATH (run in PowerShell as Administrator)
   [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\bin", "User")
   ```

4. Verify the installation:
   ```bash
   compress-env --version
   ```

### Option 2: Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/cavaldos/compress-env.git
   cd compress-env
   ```

2. Build the binary (requires [Go](https://go.dev/dl/) 1.25 or newer):
   ```bash
   go build -o compress-env .
   ```

3. Move the binary to a directory in your system PATH:

   **For Linux/macOS:**
   ```bash
   # Move to /usr/local/bin (requires sudo)
   sudo cp compress-env /usr/local/bin/

   # Or move to ~/.local/bin (doesn't require sudo)
   mkdir -p ~/.local/bin
   cp compress-env ~/.local/bin/
   ```

   **For Windows:**
   ```powershell
   # Create a directory for binaries if it doesn't exist
   mkdir -p $env:USERPROFILE\bin

   # Copy the binary
   copy compress-env.exe $env:USERPROFILE\bin\

   # Add to PATH (run in PowerShell as Administrator)
   [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\bin", "User")
   ```

4. Verify the installation:
   ```bash
   compress-env --version
   ```

## Uninstallation

### For Linux/macOS:

1. If you installed the binary to `/usr/local/bin`:
   ```bash
   sudo rm /usr/local/bin/compress-env
   ```

2. If you installed the binary to `~/.local/bin`:
   ```bash
   rm ~/.local/bin/compress-env
   ```

### For Windows:

1. If you installed the binary to your user bin directory:
   ```powershell
   # Remove the binary
   Remove-Item $env:USERPROFILE\bin\compress-env.exe
   ```

2. Verify the uninstallation:
   ```bash
   compress-env --version
   ```
   If the uninstallation was successful, you should see a "command not found" error.

## Usage

### Interactive menu

Run `compress-env` without arguments to get a menu, so you don't have to remember any option:

1. **Compress .env files**
   - Pick the files to compress in a file tree, like neo-tree. Choose nothing and every `.env` file is compressed.

     | Key | Action |
     |-----|--------|
     | `↑` `↓` / `j` `k` | move (`g` / `G` first / last) |
     | `→` `←` / `l` `h` | open / close a directory (`←` on a file jumps to its directory) |
     | `enter` | open / close a directory, select a file |
     | `space` | select a file, or every file of a directory |
     | `a` | select all / none |
     | `t` | switch between env files only and all files |
     | `c` | continue |
     | `esc` | cancel |
   - If your choice differs from the default, it is saved to `.compress-env.yaml` and pre-selected the next time:
     ```yaml
     files:
       - .env
       - services/api/.env
     ```
   - Protect the result with either a **password**, or a **public / private key**: the public key encrypts, only the private key can restore.
2. **Restore .env files** — asks for the file (`tab` completes the path and lists the backups it finds, `shift+tab` goes back), detects how it was protected, and then asks for the password, or the private key (paste it or give the path to the key file).
3. **Generate a key pair** — writes `compress-env.key` (private) and `compress-env.key.pub` (public) to the current directory. Keep the private key secret; it is git-ignored by default.

Public / private keys use [age](https://github.com/FiloSottile/age) (X25519), so keys made with `age-keygen` work too. Files protected this way can only be restored from the menu.

### Command line

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
3. Applies encoding (password-based, or encrypted with a public key)
4. Stores the encoded data in a binary file (default: `.env.bin`)

### Restoration Process

1. Reads the encoded binary file
2. Decodes the content (using the provided password if encrypted)
3. Parses the file paths and contents
4. Recreates each .env file in its original location with its original content

## Security

When using the password option (`-p`), compress-env applies a simple but effective encryption to your environment variables. This adds an extra layer of protection for sensitive information like API keys, database credentials, and other secrets.



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

### Development

```bash
scripts/test.sh              # gofmt, vet, tests, build (what CI runs)
scripts/build.sh v0.0.1      # cross-compile into dist/ with checksums
scripts/release.sh v0.0.1    # tag and push; GitHub Actions publishes the release
```

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Contact & Support

- **GitHub**: [github.com/cavaldos/compress-env](https://github.com/cavaldos/compress-env)
- **Author**: [Nguyễn Ngọc Khánh](https://github.com/cavaldos)
- **Report Issues**: [Submit an Issue](https://github.com/cavaldos/compress-env/issues)
- **Feature Requests**: [Submit a Feature Request](https://github.com/cavaldos/compress-env/issues/new)

### ⭐ Star the Project

If you find compress-env useful, please consider giving it a star on GitHub! It helps make the project more visible and encourages continued development.

[![Star on GitHub](https://img.shields.io/github/stars/cavaldos/compress-env?style=social)](https://github.com/cavaldos/compress-env/stargazers)