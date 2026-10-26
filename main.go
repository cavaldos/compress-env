package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cavaldos/compress-env/internal/compress"
	"github.com/cavaldos/compress-env/internal/hash"
	"github.com/cavaldos/compress-env/internal/help"
	"github.com/cavaldos/compress-env/internal/pubkey"
	"github.com/cavaldos/compress-env/internal/restore"
	"github.com/cavaldos/compress-env/internal/tui"
)

const defaultEnvFile = ".env.bin"

// version can be overridden at build time with -ldflags "-X main.version=...".
var version = "0.0.1"

type options struct {
	file     string   // -o for compress, -i for restore
	files    []string // -f (compress only)
	password string
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		if err := tui.Run(defaultEnvFile); err != nil {
			fatal(err)
		}
		return
	}

	switch args[0] {
	case "-v", "--version":
		help.PrintVersion(os.Stdout, version)
	case "-h", "--help":
		help.PrintUsage(os.Stdout, defaultEnvFile)
	case "compress":
		fmt.Println("Compressing and encoding .env files...")

		opts, err := parseOptions("compress", args[1:])
		if err != nil {
			usageError(err)
		}
		if opts.file == "" {
			opts.file = defaultEnvFile
		}
		runCompress(opts)
	case "restore":
		fmt.Println("Decoding and restoring .env files from backup...")

		opts, err := parseOptions("restore", args[1:])
		if err != nil {
			usageError(err)
		}
		if opts.file == "" {
			opts.file = defaultEnvFile
		}
		runRestore(opts)
	default:
		help.PrintUsage(os.Stdout, defaultEnvFile)
		os.Exit(1)
	}
}

func runCompress(opts options) {
	if len(opts.files) > 0 {
		fmt.Println("Compressing specified .env files...")
	} else {
		dir, _ := os.Getwd()
		fmt.Printf("Searching for .env files in: %s\n", dir)
	}

	encode := func(content string) ([]byte, error) {
		return hash.Encode(content, opts.password), nil
	}
	res, err := compress.Run(opts.file, opts.files, encode)
	for _, w := range res.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
	}
	if err != nil {
		fatal(err)
	}

	if res.Files == 0 {
		fmt.Println("No .env files found in the directory tree.")
	} else {
		fmt.Printf("Encoded .env contents saved to %s\n", res.Path)
	}
}

func runRestore(opts options) {
	fmt.Printf("Reading from: %s\n", opts.file)

	decode := func(data []byte) (string, error) {
		if pubkey.IsEncrypted(data) {
			return "", fmt.Errorf("this file is encrypted with a public key; run compress-env without arguments to restore it with your private key")
		}
		return hash.Decode(data, opts.password)
	}
	restored, err := restore.Run(opts.file, decode)
	if err != nil {
		fatal(err)
	}

	if len(restored) == 0 {
		fmt.Println("No .env files were restored.")
		return
	}
	fmt.Printf("Restored %d .env files:\n", len(restored))
	for _, file := range restored {
		fmt.Printf("  - %s\n", file)
	}
}

// parseOptions reads the options of the compress or restore command.
func parseOptions(cmd string, args []string) (options, error) {
	var opts options
	fileFlag := "-o"
	if cmd == "restore" {
		fileFlag = "-i"
	}

	for i := 0; i < len(args); {
		arg := args[i]
		switch {
		case arg == fileFlag:
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s option requires a filename", arg)
			}
			opts.file = args[i+1]
			i += 2
		case arg == "-f" && cmd == "compress":
			i++
			for i < len(args) && !strings.HasPrefix(args[i], "-") {
				opts.files = append(opts.files, args[i])
				i++
			}
		case arg == "-p":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s option requires a password", arg)
			}
			opts.password = args[i+1]
			i += 2
		case strings.HasPrefix(arg, "-"):
			return opts, fmt.Errorf("unknown option: %s", arg)
		default:
			i++
		}
	}
	return opts, nil
}

func usageError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	help.PrintUsage(os.Stdout, defaultEnvFile)
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
