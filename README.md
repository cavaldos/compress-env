# compress-env

A tool for compressing and restoring .env files in your project.

## Features

- Compress all .env files from your project directory into a single backup file
- Restore .env files from the backup file to their original locations

## Usage

```bash
# Compress all .env files in the current directory tree
compress-env compress

# Compress all .env files with a custom output filename
compress-env compress -o my-env-backup.txt

# Restore .env files from the default backup file (env_contents.txt)
compress-env restore

# Restore .env files from a custom backup file
compress-env restore -i my-env-backup.txt
```

## How it Works

When compressing, the tool:
1. Searches for all .env files in your current directory and subdirectories
2. Aggregates the contents into a single file named `env_contents.txt`
3. Stores the relative path information for each .env file

When restoring, the tool:
1. Reads the `env_contents.txt` file
2. Recreates each .env file in its original location with its original content

## Building

```bash
cargo build --release
```

The executable will be available in `./target/release/compress-env`