package restore

import (
	"os"
	"path/filepath"
	"strings"
)

// Decoder turns the bytes of an encoded file back into the aggregated content.
type Decoder func(data []byte) (string, error)

// Run decodes inputFile and recreates the .env files it contains, relative to
// the current directory. It returns the files that were written.
func Run(inputFile string, decode Decoder) ([]string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	encoded, err := os.ReadFile(filepath.Join(currentDir, inputFile))
	if err != nil {
		return nil, err
	}

	decoded, err := decode(encoded)
	if err != nil {
		return nil, err
	}
	return Apply(currentDir, decoded)
}

// Apply writes the files described by the decoded content under dir.
func Apply(dir, decoded string) ([]string, error) {
	var (
		currentFile string
		content     strings.Builder
		restored    []string
	)

	// flush writes the file collected so far, if it has any content.
	flush := func() error {
		if currentFile == "" || content.Len() == 0 {
			return nil
		}
		if err := writeEnvFile(currentFile, content.String()); err != nil {
			return err
		}
		restored = append(restored, currentFile)
		content.Reset()
		return nil
	}

	for _, line := range strings.Split(decoded, "\n") {
		line = strings.TrimSuffix(line, "\r")

		if strings.HasPrefix(line, "--- ") && strings.HasSuffix(line, " ---") {
			if err := flush(); err != nil {
				return restored, err
			}
			path := line
			for strings.HasPrefix(path, "--- ") {
				path = strings.TrimPrefix(path, "--- ")
			}
			for strings.HasSuffix(path, " ---") {
				path = strings.TrimSuffix(path, " ---")
			}
			currentFile = filepath.Join(dir, path)
		} else if currentFile != "" && line != "" {
			content.WriteString(line)
			content.WriteByte('\n')
		}
	}
	return restored, flush()
}

func writeEnvFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o666)
}
