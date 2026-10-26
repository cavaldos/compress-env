// Package config keeps the list of files to compress in .compress-env.yaml and
// finds the candidates to choose from.
package config

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const FileName = ".compress-env.yaml"

type config struct {
	Files []string `yaml:"files"`
}

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, ".idea": true, ".vscode": true,
	".venv": true, "__pycache__": true,
}

// Load returns the files saved in dir's config. The bool is false when the
// config does not exist.
func Load(dir string) ([]string, bool, error) {
	data, err := os.ReadFile(filepath.Join(dir, FileName))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var c config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, false, err
	}
	for i, f := range c.Files {
		c.Files[i] = filepath.FromSlash(f)
	}
	return c.Files, true, nil
}

// Save writes files (relative to dir) to dir's config.
func Save(dir string, files []string) error {
	c := config{Files: make([]string, len(files))}
	for i, f := range files {
		c.Files[i] = filepath.ToSlash(f)
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	data := buf.Bytes()
	return os.WriteFile(filepath.Join(dir, FileName), data, 0o666)
}

// Candidates lists, sorted and relative to dir, every file that could be
// compressed. Version control and dependency directories are skipped.
func Candidates(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != dir && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == FileName || !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	sort.Strings(files)
	return files, err
}

// IsDefault reports whether the file is one that is compressed when nothing
// else is chosen: a file named exactly ".env".
func IsDefault(path string) bool {
	return filepath.Base(path) == ".env"
}

// IsEnvLike reports whether the file looks like an environment file
// (.env, .env.local, prod.env, ...). Encoded output (*.bin) never does.
func IsEnvLike(path string) bool {
	name := filepath.Base(path)
	if strings.HasSuffix(name, ".bin") {
		return false
	}
	return name == ".env" || strings.HasPrefix(name, ".env.") || strings.HasSuffix(name, ".env")
}
