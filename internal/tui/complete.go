package tui

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cavaldos/compress-env/internal/pubkey"
)

// completions lists what the path being typed can be completed to: directories
// and files that compress-env can restore. Directories end with a separator.
func completions(dir, value string) []string {
	cut := strings.LastIndexAny(value, "/"+string(filepath.Separator)) + 1
	typedDir, prefix := value[:cut], value[cut:]

	listDir := filepath.Join(dir, typedDir)
	if filepath.IsAbs(typedDir) {
		listDir = typedDir
	}
	entries, err := os.ReadDir(listDir)
	if err != nil {
		return nil
	}

	var dirs, files []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		isDir := e.IsDir()
		if info, err := os.Stat(filepath.Join(listDir, name)); err == nil && info.IsDir() {
			isDir = true // symlink to a directory
		}

		switch {
		case isDir:
			if name == "node_modules" || (strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".")) {
				continue
			}
			dirs = append(dirs, typedDir+name+string(filepath.Separator))
		case isBackup(filepath.Join(listDir, name)):
			files = append(files, typedDir+name)
		}
	}
	sort.Strings(dirs)
	sort.Strings(files)
	return append(dirs, files...)
}

// isBackup reports whether path starts like a file written by compress.
func isBackup(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	head := make([]byte, len("age-encryption.org/v1\n"))
	n, _ := io.ReadFull(f, head)
	head = head[:n]
	return bytes.HasPrefix(head, []byte("ENVENC0")) || pubkey.IsEncrypted(head)
}

func commonPrefix(items []string) string {
	if len(items) == 0 {
		return ""
	}
	prefix := items[0]
	for _, s := range items[1:] {
		for !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

// complete handles Tab on the restore file prompt: it first extends the text as
// far as all matches agree, then steps through the matches.
func (m *model) complete(backwards bool) {
	n := len(m.matches)
	if n == 0 {
		return
	}

	value := m.input.Value()
	if common := commonPrefix(m.matches); m.matchIdx < 0 && len(common) > len(value) {
		m.input.SetValue(common)
		m.input.CursorEnd()
		m.refreshMatches()
		return
	}

	switch {
	case m.matchIdx < 0 && backwards:
		m.matchIdx = n - 1
	case m.matchIdx < 0:
		m.matchIdx = 0
	case backwards:
		m.matchIdx = (m.matchIdx + n - 1) % n
	default:
		m.matchIdx = (m.matchIdx + 1) % n
	}
	m.input.SetValue(m.matches[m.matchIdx])
	m.input.CursorEnd()
}

func (m *model) refreshMatches() {
	m.matches = completions(m.dir, m.input.Value())
	m.matchIdx = -1
}
