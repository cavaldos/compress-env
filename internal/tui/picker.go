package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cavaldos/compress-env/internal/config"
)

// node is a file or directory of the tree.
type node struct {
	name     string
	path     string // relative to the working directory; "" for the root
	isDir    bool
	children []*node
	parent   *node
}

// row is one visible line of the tree.
type row struct {
	n      *node
	prefix string // branch lines: "│ ├ "
}

// picker is the tree of files to compress.
type picker struct {
	root     *node
	all      []string // every file, sorted
	selected map[string]bool
	base     []string // what the selection is compared with: the saved config, or the default
	defaults []string // compressed when nothing is selected
	ignored  *config.Ignored
	showAll  bool // false: only env-like and git-ignored files
	expanded map[string]bool
	rows     []row
	cursor   int
	offset   int
}

func newPicker(dir string) (picker, error) {
	candidates, err := config.Candidates(dir)
	if err != nil {
		return picker{}, err
	}

	saved, hasConfig, err := config.Load(dir)
	if err != nil {
		return picker{}, err
	}

	ignored := config.LoadIgnored(dir)
	defaults := config.DefaultFiles(dir, candidates, ignored)

	var base []string
	if hasConfig {
		for _, f := range saved {
			if info, err := os.Stat(filepath.Join(dir, f)); err == nil && info.Mode().IsRegular() {
				base = append(base, f)
			}
		}
	} else {
		base = defaults
	}

	p := picker{selected: map[string]bool{}, base: base, defaults: defaults, ignored: ignored}
	known := map[string]bool{}
	for _, f := range candidates {
		known[f] = true
	}
	p.all = candidates
	for _, f := range base {
		p.selected[f] = true
		if !known[f] {
			p.all = append(p.all, f)
		}
	}
	sort.Strings(p.all)

	p.root = buildTree(p.all)
	p.resetExpansion()
	p.rebuild()
	return p, nil
}

func buildTree(files []string) *node {
	root := &node{isDir: true}
	dirs := map[string]*node{"": root}

	for _, file := range files {
		parent := root
		parts := strings.Split(file, string(filepath.Separator))
		for i, part := range parts {
			path := filepath.Join(parts[:i+1]...)
			if i == len(parts)-1 {
				parent.children = append(parent.children, &node{name: part, path: path, parent: parent})
				break
			}
			dir, ok := dirs[path]
			if !ok {
				dir = &node{name: part, path: path, isDir: true, parent: parent}
				dirs[path] = dir
				parent.children = append(parent.children, dir)
			}
			parent = dir
		}
	}

	var order func(n *node)
	order = func(n *node) {
		sort.SliceStable(n.children, func(i, j int) bool {
			a, b := n.children[i], n.children[j]
			if a.isDir != b.isDir {
				return a.isDir // directories first
			}
			return a.name < b.name
		})
		for _, c := range n.children {
			order(c)
		}
	}
	order(root)
	return root
}

// isVisible: a file is shown by the current filter (or when selected); a
// directory when anything inside it is.
func (p *picker) isVisible(n *node) bool {
	if !n.isDir {
		return p.showAll || p.selected[n.path] || config.IsEnvLike(n.path) ||
			p.ignored.File(n.path) || p.ignored.InDir(n.path)
	}
	for _, c := range n.children {
		if p.isVisible(c) {
			return true
		}
	}
	return false
}

// resetExpansion opens every directory when only env and ignored files are
// shown, and just the ones holding a selected file when browsing all files.
// Directories git ignores as a whole (dist/, .next/, ...) stay closed unless
// they hold a selected file.
func (p *picker) resetExpansion() {
	p.expanded = map[string]bool{}
	var visit func(n *node) bool // reports whether n holds a selected file
	visit = func(n *node) bool {
		has := false
		for _, c := range n.children {
			if c.isDir {
				has = visit(c) || has
			} else if p.selected[c.path] {
				has = true
			}
		}
		if n.isDir && n.path != "" {
			whole := p.ignored.InDir(n.path + string(filepath.Separator))
			p.expanded[n.path] = has || (!p.showAll && !whole && p.isVisible(n))
		}
		return has
	}
	visit(p.root)
}

// rebuild flattens the open part of the tree into rows.
func (p *picker) rebuild() {
	p.rows = p.rows[:0]
	p.walk(p.root, "", 0)
	if p.cursor > len(p.rows)-1 {
		p.cursor = len(p.rows) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

func (p *picker) walk(parent *node, guide string, depth int) {
	var kids []*node
	for _, c := range parent.children {
		if p.isVisible(c) {
			kids = append(kids, c)
		}
	}

	for i, c := range kids {
		last := i == len(kids)-1
		prefix, childGuide := "", ""
		if depth > 0 {
			connector, line := "├ ", "│ "
			if last {
				connector, line = "└ ", "  "
			}
			prefix, childGuide = guide+connector, guide+line
		}
		p.rows = append(p.rows, row{n: c, prefix: prefix})
		if c.isDir && p.expanded[c.path] {
			p.walk(c, childGuide, depth+1)
		}
	}
}

// files lists the visible files under n.
func (p *picker) files(n *node) []string {
	if !n.isDir {
		if p.isVisible(n) {
			return []string{n.path}
		}
		return nil
	}
	var out []string
	for _, c := range n.children {
		out = append(out, p.files(c)...)
	}
	return out
}

// state reports how many of the visible files under n are selected.
func (p *picker) state(n *node) (selected, total int) {
	for _, f := range p.files(n) {
		total++
		if p.selected[f] {
			selected++
		}
	}
	return selected, total
}

func (p *picker) selection() []string {
	var files []string
	for _, f := range p.all {
		if p.selected[f] {
			files = append(files, f)
		}
	}
	return files
}

func (p *picker) current() *node {
	if p.cursor < len(p.rows) {
		return p.rows[p.cursor].n
	}
	return nil
}

func (p *picker) move(delta, height int) {
	p.cursor += delta
	if p.cursor > len(p.rows)-1 {
		p.cursor = len(p.rows) - 1
	}
	p.clamp(height)
}

// clamp keeps the cursor inside the window of height rows.
func (p *picker) clamp(height int) {
	if p.cursor < 0 {
		p.cursor = 0
	}
	if p.cursor < p.offset {
		p.offset = p.cursor
	}
	if p.cursor >= p.offset+height {
		p.offset = p.cursor - height + 1
	}
}

// toggleSelect selects the file under the cursor, or every file of the
// directory (clearing them when they are all selected already).
func (p *picker) toggleSelect() {
	n := p.current()
	if n == nil {
		return
	}
	selected, total := p.state(n)
	for _, f := range p.files(n) {
		p.selected[f] = selected != total
	}
}

// toggleAll selects every visible file, or clears them if all are selected.
func (p *picker) toggleAll() {
	selected, total := p.state(p.root)
	for _, f := range p.files(p.root) {
		p.selected[f] = selected != total
	}
}

// open expands the directory under the cursor, or collapses it.
func (p *picker) open(expand bool) {
	if n := p.current(); n != nil && n.isDir {
		p.expanded[n.path] = expand
		p.rebuild()
	}
}

// enter toggles a directory open or closed, and selects a file.
func (p *picker) enter() {
	n := p.current()
	switch {
	case n == nil:
	case n.isDir:
		p.open(!p.expanded[n.path])
	default:
		p.toggleSelect()
	}
}

// left collapses an open directory, otherwise jumps to the parent.
func (p *picker) left() {
	n := p.current()
	if n == nil {
		return
	}
	if n.isDir && p.expanded[n.path] {
		p.open(false)
		return
	}
	for i := p.cursor - 1; i >= 0; i-- {
		if p.rows[i].n == n.parent {
			p.cursor = i
			return
		}
	}
}

// right opens a closed directory, otherwise steps into it.
func (p *picker) right() {
	n := p.current()
	if n == nil || !n.isDir {
		return
	}
	if !p.expanded[n.path] {
		p.open(true)
	} else if p.cursor+1 < len(p.rows) {
		p.cursor++
	}
}

func (p *picker) toggleView() {
	p.showAll = !p.showAll
	p.cursor, p.offset = 0, 0
	p.resetExpansion()
	p.rebuild()
}

// outcome decides what to compress and whether the choice should be saved.
// Choosing nothing falls back to the default files and saves nothing.
func (p *picker) outcome() (files []string, save bool) {
	files = p.selection()
	if len(files) == 0 {
		return p.defaults, false
	}
	return files, !equal(files, p.base)
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := map[string]bool{}
	for _, f := range a {
		set[f] = true
	}
	for _, f := range b {
		if !set[f] {
			return false
		}
	}
	return true
}
