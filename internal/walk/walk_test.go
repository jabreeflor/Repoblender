package walk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jabreeflor/repoblender/internal/ignore"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWalkFiltersBinaryIgnoredOversized(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "README.md"), "# hi\n")
	writeFile(t, filepath.Join(dir, "src", "main.go"), "package main\n")
	writeFile(t, filepath.Join(dir, "src", "big.txt"), "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	writeFile(t, filepath.Join(dir, "node_modules", "pkg", "index.js"), "x")
	writeFile(t, filepath.Join(dir, "image.png"), "not really png but ext wins")
	writeFile(t, filepath.Join(dir, "bin", "blob"), "abc\x00def")
	writeFile(t, filepath.Join(dir, ".gitignore"), "*.log\n")
	writeFile(t, filepath.Join(dir, "debug.log"), "ignored by gitignore")

	m, err := ignore.New(ignore.Options{Root: dir, RespectGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := Walk(dir, Options{Matcher: m, MaxFileSize: 16})
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]bool{}
	for _, e := range entries {
		got[e.RelPath] = true
	}

	want := []string{".gitignore", "README.md", "src/main.go"}
	for _, w := range want {
		if !got[w] {
			t.Errorf("expected entry %q, got %v", w, got)
		}
	}
	forbidden := []string{"src/big.txt", "node_modules/pkg/index.js", "image.png", "bin/blob", "debug.log"}
	for _, f := range forbidden {
		if got[f] {
			t.Errorf("did not expect entry %q", f)
		}
	}
}

func TestWalkSortsDeterministically(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"z.txt", "a.txt", "m/n.txt"} {
		writeFile(t, filepath.Join(dir, name), "x")
	}
	m, err := ignore.New(ignore.Options{Root: dir})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := Walk(dir, Options{Matcher: m})
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, len(entries))
	for i, e := range entries {
		got[i] = e.RelPath
	}
	want := []string{"a.txt", "m/n.txt", "z.txt"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("entries[%d] = %q, want %q (all=%v)", i, got[i], w, got)
		}
	}
}
