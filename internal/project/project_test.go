package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDescribeNameFallback(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "MyProj")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	info := Describe(sub, "")
	if info.Name != "MyProj" {
		t.Errorf("Name = %q, want MyProj", info.Name)
	}
	if info.Summary != "" {
		t.Errorf("Summary = %q, want empty", info.Summary)
	}
}

func TestDescribeReadme(t *testing.T) {
	dir := t.TempDir()
	readme := "# Cool Project\n\nA tiny tool that blends things.\nIt is nice.\n\n## Usage\n\nstuff\n"
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644); err != nil {
		t.Fatal(err)
	}
	info := Describe(dir, "Override")
	if info.Name != "Override" {
		t.Errorf("Name = %q, want Override", info.Name)
	}
	want := "A tiny tool that blends things. It is nice."
	if info.Summary != want {
		t.Errorf("Summary = %q, want %q", info.Summary, want)
	}
}
