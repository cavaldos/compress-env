package help

import (
	"fmt"
	"io"
)

func PrintUsage(w io.Writer, defaultFile string) {
	fmt.Fprintln(w, "Usage: compress-env [COMMAND] [OPTIONS]")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  compress          - Compress and encode all .env files into a single encrypted binary file")
	fmt.Fprintln(w, "  restore           - Decode and restore .env files from the encrypted binary file")
	fmt.Fprintln(w, "Options:")
	fmt.Fprintf(w, "  -o <filename>     - Specify the output filename for compression (default: %s)\n", defaultFile)
	fmt.Fprintf(w, "  -i <filename>     - Specify the input filename for restoration (default: %s)\n", defaultFile)
	fmt.Fprintln(w, "  -f <file1> <file2> ... - Specify specific .env files to compress")
	fmt.Fprintln(w, "  -p <password>     - Specify a password for encryption/decryption (optional)")
	fmt.Fprintln(w, "  -h, --help        - Show this help message")
	fmt.Fprintln(w, "  -v, --version     - Show version information")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  compress-env compress -o secrets.bin")
	fmt.Fprintln(w, "  compress-env compress -f ./project1/.env -o project1-secrets.bin -p mypassword")
	fmt.Fprintln(w, "  compress-env restore -i secrets.bin -p mypassword")
}

func PrintVersion(w io.Writer, version string) {
	fmt.Fprintf(w, "compress-env v%s\n", version)
	fmt.Fprintln(w, "A lightweight tool for securely compressing, encoding, and managing .env files")
	fmt.Fprintln(w, "Copyright (c) 2025 Nguyễn Ngọc Khánh")
	fmt.Fprintln(w, "License: MIT")
}
