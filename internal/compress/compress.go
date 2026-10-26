package compress

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Encoder turns the aggregated .env content into the bytes written to disk.
type Encoder func(content string) ([]byte, error)

// Result describes what Run did.
type Result struct {
	Path     string   // output file
	Files    int      // number of .env files encoded
	Warnings []string // files that were skipped
}

// Run collects .env files (the given ones, or every .env under the current
// directory) and writes them, encoded, to outputFile.
func Run(outputFile string, specificFiles []string, encode Encoder) (Result, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return Result{}, err
	}
	res := Result{Path: filepath.Join(currentDir, outputFile)}

	var aggregated strings.Builder

	if len(specificFiles) > 0 {
		for _, filePath := range specificFiles {
			if info, err := os.Stat(filePath); err == nil && info.Mode().IsRegular() {
				if err := processEnvFile(filePath, currentDir, &aggregated); err != nil {
					return res, err
				}
				res.Files++
			} else {
				res.Warnings = append(res.Warnings, fmt.Sprintf("File not found or not a regular file: %s", filePath))
			}
		}
	} else {
		if res.Files, err = findEnvFiles(currentDir, currentDir, &aggregated); err != nil {
			return res, err
		}
	}

	encoded, err := encode(aggregated.String())
	if err != nil {
		return res, err
	}
	return res, os.WriteFile(res.Path, encoded, 0o666)
}

// findEnvFiles walks dir recursively and appends every .env file it finds. It
// returns how many files were added.
func findEnvFiles(dir, currentDir string, aggregated *strings.Builder) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			n, err := findEnvFiles(path, currentDir, aggregated)
			if err != nil {
				return count, err
			}
			count += n
		} else if entry.Name() == ".env" {
			if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
				if err := processEnvFile(path, currentDir, aggregated); err != nil {
					return count, err
				}
				count++
			}
		}
	}
	return count, nil
}

// processEnvFile appends one .env file, headed by its path relative to
// currentDir, to the aggregated content.
func processEnvFile(path, currentDir string, aggregated *strings.Builder) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", path, err)
	}
	if !utf8.Valid(data) {
		return fmt.Errorf("reading file %s: stream did not contain valid UTF-8", path)
	}

	fmt.Fprintf(aggregated, "--- %s ---\n", relativeTo(currentDir, path))
	aggregated.Write(data)
	aggregated.WriteString("\n\n")
	return nil
}

// relativeTo strips currentDir from path when path lies inside it and leaves
// it untouched otherwise.
func relativeTo(currentDir, path string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	rel, err := filepath.Rel(currentDir, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}
	return rel
}
