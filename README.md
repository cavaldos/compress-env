# compress-env

Pack your `.env` files into one protected file, and restore them anywhere.

## Install

Download the binary for your machine from the [Releases](https://github.com/cavaldos/compress-env/releases) page, or:

```bash
os=$(uname -s | tr '[:upper:]' '[:lower:]')                  # linux | darwin
arch=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')   # amd64 | arm64
curl -L "https://github.com/cavaldos/compress-env/releases/latest/download/compress-env-${os}-${arch}" -o compress-env
chmod +x compress-env && sudo mv compress-env /usr/local/bin/
```

Windows: download `compress-env-windows-amd64.exe` from the Releases page.

Or build it (needs Go 1.25+): `go build -o compress-env .`

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
