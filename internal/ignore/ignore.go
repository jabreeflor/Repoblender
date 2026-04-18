package ignore

import (
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// DefaultSkipDirs are directory names always skipped, regardless of
// .gitignore. They're the usual "build artifacts + VCS + editor" noise.
var DefaultSkipDirs = []string{
	".git",
	".hg",
	".svn",
	"node_modules",
	"__pycache__",
	".venv",
	"venv",
	"env",
	"dist",
	"build",
	"target",
	".next",
	".nuxt",
	".cache",
	".idea",
	".vscode",
	".gradle",
	".pytest_cache",
	".mypy_cache",
	".tox",
	".terraform",
	"vendor",
}

// Options configures matcher construction.
type Options struct {
	// Root is the repo root. Used to locate .gitignore files.
	Root string
	// RespectGitignore controls whether .gitignore at the root is loaded.
	RespectGitignore bool
	// ExtraPatterns is a list of gitignore-style patterns from --exclude.
	ExtraPatterns []string
	// IncludePatterns are gitignore-style patterns; when non-empty, only
	// paths matching at least one pattern are kept.
	IncludePatterns []string
}

// Matcher decides whether a relative path should be ignored.
type Matcher struct {
	skipDirs map[string]struct{}
	gi       *gitignore.GitIgnore
	extra    *gitignore.GitIgnore
	include  *gitignore.GitIgnore
}

// New builds a Matcher for the given options.
func New(opts Options) (*Matcher, error) {
	m := &Matcher{skipDirs: map[string]struct{}{}}
	for _, d := range DefaultSkipDirs {
		m.skipDirs[d] = struct{}{}
	}

	if opts.RespectGitignore && opts.Root != "" {
		path := filepath.Join(opts.Root, ".gitignore")
		if _, err := os.Stat(path); err == nil {
			gi, err := gitignore.CompileIgnoreFile(path)
			if err != nil {
				return nil, err
			}
			m.gi = gi
		}
	}
	if len(opts.ExtraPatterns) > 0 {
		m.extra = gitignore.CompileIgnoreLines(opts.ExtraPatterns...)
	}
	if len(opts.IncludePatterns) > 0 {
		m.include = gitignore.CompileIgnoreLines(opts.IncludePatterns...)
	}
	return m, nil
}

// Match reports whether the given repo-relative path should be skipped.
// isDir indicates whether the path is a directory. If includes are set,
// a file that matches no include is also skipped (directories are still
// traversed so their contents can be checked).
func (m *Matcher) Match(relPath string, isDir bool) bool {
	relPath = filepath.ToSlash(relPath)
	if relPath == "." || relPath == "" {
		return false
	}

	if isDir {
		base := filepath.Base(relPath)
		if _, ok := m.skipDirs[base]; ok {
			return true
		}
	}

	if m.gi != nil && m.gi.MatchesPath(relPath) {
		return true
	}
	if m.extra != nil && m.extra.MatchesPath(relPath) {
		return true
	}

	if !isDir && m.include != nil && !m.include.MatchesPath(relPath) {
		return true
	}
	return false
}

// NormalizePatterns trims whitespace and drops empties/comments. Useful
// when callers accept multiple --exclude flags that may be sloppy.
func NormalizePatterns(in []string) []string {
	out := make([]string, 0, len(in))
	for _, p := range in {
		p = strings.TrimSpace(p)
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		out = append(out, p)
	}
	return out
}
