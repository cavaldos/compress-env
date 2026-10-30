// Package config keeps the list of files to compress in .compress-env.yaml and
// finds the candidates to choose from.
package config

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

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

// IsDefault reports whether the file is always compressed when nothing else
// is chosen: a file named exactly ".env".
func IsDefault(path string) bool {
	return filepath.Base(path) == ".env"
}

// maxDefaultSize is the largest git-ignored file chosen by default.
const maxDefaultSize = 1 << 20

// DefaultFiles picks, among candidates (relative to dir), the files compressed
// when nothing else is chosen: every ".env", plus every text file git ignores.
// Inside directories ignored as a whole (dist/, .next/, ...) only env files
// count, so build output is left out. The private key and encoded output
// (*.bin) are never chosen.
func DefaultFiles(dir string, candidates []string, ig *Ignored) []string {
	var files []string
	for _, f := range candidates {
		name := filepath.Base(f)
		if name == PrivateKeyFile || strings.HasSuffix(name, ".bin") {
			continue
		}
		if IsDefault(f) {
			files = append(files, f)
			continue
		}
		if (ig.File(f) || ig.InDir(f) && IsEnvLike(f)) && isSmallText(filepath.Join(dir, f)) {
			files = append(files, f)
		}
	}
	return files
}

// PrivateKeyFile is the private key written by the key pair generator.
const PrivateKeyFile = "compress-env.key"

// isSmallText reports whether the file is valid UTF-8 text no larger than
// maxDefaultSize.
func isSmallText(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxDefaultSize {
		return false
	}
	data, err := os.ReadFile(path)
	return err == nil && utf8.Valid(data) && !bytes.ContainsRune(data, 0)
}

// Ignored is what git ignores under a directory. A nil *Ignored means the
// directory is not in a git repository (or git is not installed): nothing is
// ignored.
type Ignored struct {
	files map[string]bool // ignored files, relative to the directory
	dirs  []string        // directories ignored as a whole, ending with a separator
}

// LoadIgnored asks git which files under dir it ignores.
func LoadIgnored(dir string) *Ignored {
	cmd := exec.Command("git", "ls-files", "--others", "--ignored", "--exclude-standard", "--directory", "-z")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	ig := &Ignored{files: map[string]bool{}}
	for _, p := range strings.Split(string(out), "\x00") {
		switch {
		case p == "":
		case strings.HasSuffix(p, "/"):
			ig.dirs = append(ig.dirs, filepath.FromSlash(p))
		default:
			ig.files[filepath.FromSlash(p)] = true
		}
	}
	return ig
}

// File reports whether git ignores the file itself.
func (ig *Ignored) File(path string) bool {
	return ig != nil && ig.files[path]
}

// InDir reports whether path is inside a directory git ignores as a whole.
// Pass a directory with a trailing separator to ask about the directory.
func (ig *Ignored) InDir(path string) bool {
	if ig == nil {
		return false
	}
	for _, d := range ig.dirs {
		if strings.HasPrefix(path, d) {
			return true
		}
	}
	return false
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
