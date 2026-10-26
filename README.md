# compress-env

Pack your `.env` files into one protected file, and restore them anywhere.

## Install

Download the latest build for your OS:

| OS | Intel / AMD | ARM (Apple Silicon, etc.) |
|----|-------------|---------------------------|
| macOS | [compress-env-darwin-amd64](https://github.com/cavaldos/compress-env/releases/latest/download/compress-env-darwin-amd64) | [compress-env-darwin-arm64](https://github.com/cavaldos/compress-env/releases/latest/download/compress-env-darwin-arm64) |
| Linux | [compress-env-linux-amd64](https://github.com/cavaldos/compress-env/releases/latest/download/compress-env-linux-amd64) | [compress-env-linux-arm64](https://github.com/cavaldos/compress-env/releases/latest/download/compress-env-linux-arm64) |
| Windows | [compress-env-windows-amd64.exe](https://github.com/cavaldos/compress-env/releases/latest/download/compress-env-windows-amd64.exe) | [compress-env-windows-arm64.exe](https://github.com/cavaldos/compress-env/releases/latest/download/compress-env-windows-arm64.exe) |

macOS / Linux: make it executable and put it on your PATH.

```bash
chmod +x compress-env-*
sudo mv compress-env-* /usr/local/bin/compress-env
```

Windows: put the `.exe` in a folder on your PATH.

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
