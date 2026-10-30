// Package tui is the interactive menu: compress, restore, generate a key pair.
package tui

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cavaldos/compress-env/internal/compress"
	"github.com/cavaldos/compress-env/internal/config"
	"github.com/cavaldos/compress-env/internal/hash"
	"github.com/cavaldos/compress-env/internal/pubkey"
	"github.com/cavaldos/compress-env/internal/restore"
)

const (
	privateKeyFile = config.PrivateKeyFile
	publicKeyFile  = "compress-env.key.pub"
)

type screen int

const (
	screenMenu screen = iota
	screenPicker
	screenMethod
	screenPassword
	screenPasswordConfirm
	screenPublicKey
	screenOutput
	screenRestoreFile
	screenRestorePassword
	screenRestoreKey
	screenResult
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	cursorStyle = lipgloss.NewStyle().Background(lipgloss.Color("237")).Foreground(lipgloss.Color("39")).Bold(true)
	dirStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("179")).Bold(true)
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("179"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	okStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

var (
	menuItems   = []string{"Compress .env files", "Restore .env files", "Generate a key pair", "Quit"}
	methodItems = []string{"Password", "Public / private key"}
)

type model struct {
	defaultFile   string
	dir           string
	width, height int

	screen screen
	cursor int // menu selection
	method int // 0: password, 1: key pair
	input  textinput.Model
	err    string

	picker picker

	// compress
	files       []string
	usePassword bool
	password    string
	recipients  []age.Recipient
	savedConfig bool

	// restore
	inputFile string
	data      []byte
	matches   []string // Tab completions for the file prompt
	matchIdx  int      // the one currently inserted, -1 for none

	result []string
	failed bool
}

// Run shows the menu until the user quits. defaultFile is the encoded file
// name offered when none is typed.
func Run(defaultFile string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	ti := textinput.New()
	ti.CharLimit = 4096

	m := model{defaultFile: defaultFile, dir: dir, input: ti, width: 80, height: 24}
	_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.Width = max(20, min(msg.Width-6, 100))
		m.picker.clamp(m.pickerRows())
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.screen {
		case screenMenu:
			return m.updateMenu(msg)
		case screenPicker:
			return m.updatePicker(msg)
		case screenMethod:
			return m.updateMethod(msg)
		case screenResult:
			return m.updateResult(msg)
		default:
			return m.updateInput(msg)
		}
	}

	if m.isInputScreen() {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) isInputScreen() bool {
	switch m.screen {
	case screenMenu, screenPicker, screenMethod, screenResult:
		return false
	}
	return true
}

// ask switches to an input screen.
func (m model) ask(s screen, value string, echo textinput.EchoMode) (model, tea.Cmd) {
	m.screen = s
	m.err = ""
	m.input.EchoMode = echo
	m.input.Placeholder = ""
	m.input.SetValue(value)
	m.input.CursorEnd()
	m.matches = nil
	if s == screenRestoreFile {
		m.input.Placeholder = m.defaultFile
		m.refreshMatches()
	}
	return m, m.input.Focus()
}

func (m model) finish(failed bool, lines ...string) model {
	m.screen = screenResult
	m.failed = failed
	m.result = lines
	m.input.Blur()
	return m
}

// --- menu ---------------------------------------------------------------

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.cursor = (m.cursor + len(menuItems) - 1) % len(menuItems)
	case "down", "j":
		m.cursor = (m.cursor + 1) % len(menuItems)
	case "q", "esc":
		return m, tea.Quit
	case "enter":
		switch m.cursor {
		case 0:
			p, err := newPicker(m.dir)
			if err != nil {
				return m.finish(true, err.Error()), nil
			}
			m.picker = p
			m.screen = screenPicker
			m.picker.clamp(m.pickerRows())
		case 1:
			return m.ask(screenRestoreFile, "", textinput.EchoNormal)
		case 2:
			return m.generateKeys(), nil
		case 3:
			return m, tea.Quit
		}
	}
	return m, nil
}

// --- file picker --------------------------------------------------------

func (m model) pickerRows() int { return max(5, m.height-10) }

func (m model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.pickerRows()
	p := &m.picker
	switch msg.String() {
	case "up", "k":
		p.move(-1, rows)
	case "down", "j":
		p.move(1, rows)
	case "pgup":
		p.move(-rows, rows)
	case "pgdown":
		p.move(rows, rows)
	case "g", "home":
		p.move(-len(p.rows), rows)
	case "G", "end":
		p.move(len(p.rows), rows)
	case "left", "h":
		p.left()
		p.clamp(rows)
	case "right", "l":
		p.right()
		p.clamp(rows)
	case "enter":
		p.enter()
		p.clamp(rows)
	case " ", "x":
		p.toggleSelect()
		p.rebuild()
	case "a":
		p.toggleAll()
		p.rebuild()
	case "t":
		p.toggleView()
	case "esc":
		m.screen = screenMenu
	case "c":
		files, save := p.outcome()
		if len(files) == 0 {
			return m.finish(true, "No .env files found in the directory tree."), nil
		}
		m.files = files
		m.savedConfig = false
		if save {
			if err := config.Save(m.dir, files); err != nil {
				return m.finish(true, "Cannot save "+config.FileName+": "+err.Error()), nil
			}
			m.savedConfig = true
		}
		m.screen = screenMethod
	}
	return m, nil
}

// --- encryption method --------------------------------------------------

func (m model) updateMethod(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		m.method = 1 - m.method
	case "esc":
		m.screen = screenMenu
	case "enter":
		m.usePassword = m.method == 0
		if m.usePassword {
			return m.ask(screenPassword, "", textinput.EchoPassword)
		}
		return m.ask(screenPublicKey, "", textinput.EchoNormal)
	}
	return m, nil
}

// --- text input screens -------------------------------------------------

func (m model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenMenu
		m.input.Blur()
		return m, nil
	case "ctrl+r":
		if m.isSecret() {
			if m.input.EchoMode == textinput.EchoPassword {
				m.input.EchoMode = textinput.EchoNormal
			} else {
				m.input.EchoMode = textinput.EchoPassword
			}
		}
		return m, nil
	case "tab", "shift+tab":
		if m.screen == screenRestoreFile {
			m.complete(msg.String() == "shift+tab")
		}
		return m, nil
	case "enter":
		return m.submit(strings.TrimSpace(m.input.Value()))
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.screen == screenRestoreFile {
		m.refreshMatches()
	}
	return m, cmd
}

func (m model) isSecret() bool {
	switch m.screen {
	case screenPassword, screenPasswordConfirm, screenRestorePassword, screenRestoreKey:
		return true
	}
	return false
}

func (m model) submit(value string) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenPassword:
		if value == "" {
			m.err = "The password cannot be empty."
			return m, nil
		}
		m.password = m.input.Value()
		return m.ask(screenPasswordConfirm, "", textinput.EchoPassword)

	case screenPasswordConfirm:
		if m.input.Value() != m.password {
			m, cmd := m.ask(screenPassword, "", textinput.EchoPassword)
			m.err = "The passwords do not match, try again."
			return m, cmd
		}
		return m.ask(screenOutput, m.defaultFile, textinput.EchoNormal)

	case screenPublicKey:
		recipients, err := pubkey.ParseRecipients(value)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.recipients = recipients
		return m.ask(screenOutput, m.defaultFile, textinput.EchoNormal)

	case screenOutput:
		if value == "" {
			value = m.defaultFile
		}
		return m.compress(value), nil

	case screenRestoreFile:
		if value == "" {
			value = m.defaultFile
		}
		return m.openBackup(value)

	case screenRestorePassword:
		text, err := hash.Decode(m.data, m.input.Value())
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		return m.apply(text), nil

	case screenRestoreKey:
		identities, err := pubkey.ParseIdentities(value)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		text, err := pubkey.Decrypt(m.data, identities)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		return m.apply(text), nil
	}
	return m, nil
}

// --- actions ------------------------------------------------------------

func (m model) compress(output string) model {
	var encode compress.Encoder
	if m.usePassword {
		password := m.password
		encode = func(content string) ([]byte, error) { return hash.Encode(content, password), nil }
	} else {
		recipients := m.recipients
		encode = func(content string) ([]byte, error) { return pubkey.Encrypt(content, recipients) }
	}

	res, err := compress.Run(output, m.files, encode)
	if err != nil {
		return m.finish(true, err.Error())
	}

	lines := []string{fmt.Sprintf("Encoded %d file(s) into %s", res.Files, res.Path)}
	for _, f := range m.files {
		lines = append(lines, "  - "+f)
	}
	if m.savedConfig {
		lines = append(lines, "", "File list saved to "+config.FileName)
	}
	for _, w := range res.Warnings {
		lines = append(lines, "", "Warning: "+w)
	}
	return m.finish(false, lines...)
}

func (m model) openBackup(name string) (tea.Model, tea.Cmd) {
	path := name
	if !filepath.IsAbs(path) {
		path = filepath.Join(m.dir, name)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		m.err = err.Error()
		return m, nil
	}

	m.inputFile, m.data = name, data
	switch {
	case pubkey.IsEncrypted(data):
		return m.ask(screenRestoreKey, "", textinput.EchoPassword)
	case bytes.HasPrefix(data, []byte("ENVENC0")):
		return m.ask(screenRestorePassword, "", textinput.EchoPassword)
	}
	m.err = "Invalid file format: not an encoded .env file"
	return m, nil
}

func (m model) apply(text string) model {
	restored, err := restore.Apply(m.dir, text)
	if err != nil {
		return m.finish(true, err.Error())
	}
	if len(restored) == 0 {
		return m.finish(false, "No .env files were restored.")
	}

	lines := []string{fmt.Sprintf("Restored %d .env file(s):", len(restored))}
	for _, f := range restored {
		if rel, err := filepath.Rel(m.dir, f); err == nil {
			f = rel
		}
		lines = append(lines, "  - "+f)
	}
	return m.finish(false, lines...)
}

func (m model) generateKeys() model {
	public, private, err := pubkey.GenerateKeyPair()
	if err != nil {
		return m.finish(true, err.Error())
	}

	privatePath := filepath.Join(m.dir, privateKeyFile)
	publicPath := filepath.Join(m.dir, publicKeyFile)
	if err := writeNew(privatePath, private+"\n", 0o600); err != nil {
		return m.finish(true, err.Error())
	}
	if err := writeNew(publicPath, public+"\n", 0o644); err != nil {
		return m.finish(true, err.Error())
	}

	return m.finish(false,
		"Key pair created.",
		"",
		"Public key (use it to compress):",
		"  "+public,
		"",
		"Saved to:",
		"  "+publicPath,
		"  "+privatePath+"  (private, needed to restore)",
		"",
		"Keep the private key secret and never commit it.",
	)
}

// writeNew creates a file and refuses to overwrite an existing one.
func writeNew(path, content string, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("%s already exists, remove or rename it first", path)
	}
	if err != nil {
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (m model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "enter", "esc":
		m.screen = screenMenu
	}
	return m, nil
}

// --- view ---------------------------------------------------------------

func (m model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("compress-env") + "\n\n")

	switch m.screen {
	case screenMenu:
		b.WriteString("What do you want to do?\n\n")
		b.WriteString(renderList(menuItems, m.cursor))
		b.WriteString("\n" + dimStyle.Render("↑/↓ move · enter select · q quit"))

	case screenPicker:
		b.WriteString(m.viewPicker())

	case screenMethod:
		b.WriteString("How should the files be protected?\n\n")
		b.WriteString(renderList(methodItems, m.method))
		b.WriteString("\n" + dimStyle.Render("↑/↓ move · enter select · esc cancel"))

	case screenResult:
		style := okStyle
		if m.failed {
			style = errStyle
		}
		for i, line := range m.result {
			if i == 0 {
				line = style.Render(line)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n" + dimStyle.Render("enter back to menu · q quit"))

	default:
		b.WriteString(m.prompt() + "\n\n")
		b.WriteString(m.input.View() + "\n")
		if m.screen == screenRestoreFile {
			b.WriteString(m.viewMatches())
		}
		if m.err != "" {
			b.WriteString("\n" + errStyle.Render(m.err) + "\n")
		}
		hint := "enter confirm · esc cancel"
		if m.screen == screenRestoreFile {
			hint += " · tab complete"
		}
		if m.isSecret() {
			hint += " · ctrl+r show/hide"
		}
		b.WriteString("\n" + dimStyle.Render(hint))
	}
	return b.String()
}

func (m model) viewMatches() string {
	const maxShown = 8
	if len(m.matches) == 0 {
		return ""
	}

	// keep the highlighted match inside the window
	start := 0
	if m.matchIdx >= maxShown {
		start = m.matchIdx - maxShown + 1
	}
	end := min(len(m.matches), start+maxShown)

	var b strings.Builder
	b.WriteString("\n")
	for i := start; i < end; i++ {
		if i == m.matchIdx {
			b.WriteString(cursorStyle.Render("▸ "+m.matches[i]) + "\n")
		} else {
			b.WriteString(dimStyle.Render("  "+m.matches[i]) + "\n")
		}
	}
	if end < len(m.matches) {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  … %d more", len(m.matches)-end)) + "\n")
	}
	return b.String()
}

func (m model) prompt() string {
	switch m.screen {
	case screenPassword:
		return "Enter a password:"
	case screenPasswordConfirm:
		return "Repeat the password:"
	case screenPublicKey:
		return "Paste the public key (age1...) or enter the path to a public key file:"
	case screenOutput:
		return "Output file name:"
	case screenRestoreFile:
		return "File to restore from:"
	case screenRestorePassword:
		return "Enter the password (leave empty if none was used):"
	case screenRestoreKey:
		return "This file was encrypted with a public key.\nPaste the private key (AGE-SECRET-KEY-...) or enter the path to the private key file:"
	}
	return ""
}

func (m model) viewPicker() string {
	var b strings.Builder
	p := m.picker

	view, fallback := "env files", "all .env files"
	if p.ignored != nil {
		view, fallback = "env + ignored files", ".env + ignored files"
	}
	if p.showAll {
		view = "all files"
	}
	b.WriteString("Choose the files to compress\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("Showing %s · %d selected · none = %s", view, len(p.selection()), fallback)) + "\n\n")
	b.WriteString(dirStyle.Render(displayPath(m.dir)) + "\n")

	if len(p.rows) == 0 {
		b.WriteString(dimStyle.Render("  (no env files here, press t to show all files)") + "\n")
	}

	width := max(20, m.width-2)
	end := min(len(p.rows), p.offset+m.pickerRows())
	for i := p.offset; i < end; i++ {
		b.WriteString(m.renderRow(p, p.rows[i], i == p.cursor, width) + "\n")
	}

	b.WriteString("\n" + dimStyle.Render("↑/↓ move · ←/→ close/open · space select · enter open/select"))
	b.WriteString("\n" + dimStyle.Render("a all · t short/full list · c continue · esc cancel"))
	return b.String()
}

func (m model) renderRow(p picker, r row, cursor bool, width int) string {
	n := r.n
	selected, total := p.state(n)

	box := "[ ]"
	switch {
	case total > 0 && selected == total:
		box = "[x]"
	case selected > 0:
		box = "[-]"
	}

	arrow, name := "  ", n.name
	if n.isDir {
		arrow = "▸ "
		if p.expanded[n.path] {
			arrow = "▾ "
		}
		name += "/"
	}

	if cursor {
		line := r.prefix + arrow + box + " " + name
		return cursorStyle.Render(line + strings.Repeat(" ", max(0, width-lipgloss.Width(line))))
	}

	boxStyle, nameStyle := dimStyle, lipgloss.NewStyle()
	switch {
	case selected > 0 && selected == total:
		boxStyle, nameStyle = okStyle, okStyle
	case selected > 0:
		boxStyle = warnStyle
	}
	if n.isDir {
		nameStyle = dirStyle
	}
	return dimStyle.Render(r.prefix) + dirStyle.Render(arrow) + boxStyle.Render(box) + " " + nameStyle.Render(name)
}

// displayPath shortens the home directory to ~.
func displayPath(dir string) string {
	if home, err := os.UserHomeDir(); err == nil {
		if rel, err := filepath.Rel(home, dir); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.Join("~", rel)
		}
	}
	return dir
}

func renderList(items []string, cursor int) string {
	var b strings.Builder
	for i, item := range items {
		if i == cursor {
			b.WriteString(cursorStyle.Render("▸ "+item) + "\n")
		} else {
			b.WriteString("  " + item + "\n")
		}
	}
	return b.String()
}
