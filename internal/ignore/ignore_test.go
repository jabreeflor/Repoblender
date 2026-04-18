package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultSkipDirs(t *testing.T) {
	m, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{".git", "node_modules", "__pycache__", ".venv"} {
		if !m.Match(d, true) {
			t.Errorf("expected %q to be skipped as dir", d)
		}
	}
	if m.Match("src", true) {
		t.Errorf("expected src to not be skipped")
	}
}

func TestGitignore(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".gitignore", "*.log\nsecret.txt\n/build\n")

	m, err := New(Options{Root: dir, RespectGitignore: true})
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]bool{
		"server.log":       true,
		"secret.txt":       true,
		"build/out.txt":    true,
		"src/main.go":      false,
		"README.md":        false,
		"nested/debug.log": true,
	}
	for p, want := range cases {
		if got := m.Match(p, false); got != want {
			t.Errorf("Match(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestExtraAndInclude(t *testing.T) {
	m, err := New(Options{
		ExtraPatterns:   []string{"docs/**"},
		IncludePatterns: []string{"*.go", "*.md"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !m.Match("docs/intro.md", false) {
		t.Error("docs/intro.md should be excluded via --exclude")
	}
	if m.Match("main.go", false) {
		t.Error("main.go should be kept via include")
	}
	if !m.Match("data.json", false) {
		t.Error("data.json should be excluded because it doesn't match any include")
	}
}

func TestNormalizePatterns(t *testing.T) {
	got := NormalizePatterns([]string{"  a ", "", "# comment", "b"})
	want := []string{"a", "b"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("NormalizePatterns = %v, want %v", got, want)
	}
}
