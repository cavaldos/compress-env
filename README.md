# compress-env

Pack your `.env` files into one protected file, and restore them anywhere.

## Install

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.ps1 | iex
```

Install a specific version: `VERSION=v0.0.1` before `sh` (macOS / Linux), or `$env:VERSION='v0.0.1'` before `irm` (Windows).

Or build it yourself (needs Go 1.25+): `go build -o compress-env .`

## Use

Run it with no arguments and pick from the menu:

```bash
compress-env
```

- **Compress** – choose the files in a tree (nothing chosen = all `.env` files), then protect them with a **password** or a **public key**.
- **Restore** – pick the file, then enter the password or the private key.
- **Generate a key pair** – creates `compress-env.key` (private) and `compress-env.key.pub` (public).

Or use the commands directly (password only):

```bash
compress-env compress -p mypassword      # writes .env.bin
compress-env restore -p mypassword
```

Options: `-o <file>` output, `-i <file>` input, `-f <files...>` specific files, `-p <password>`.

## License

MIT
