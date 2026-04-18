package render_test

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jabreeflor/repoblender/internal/ignore"
	"github.com/jabreeflor/repoblender/internal/project"
	"github.com/jabreeflor/repoblender/internal/render"
	"github.com/jabreeflor/repoblender/internal/walk"
)

var update = flag.Bool("update", false, "regenerate golden files")

// stageFixture copies a fixture tree into a temp dir, renaming any
// `dot.gitignore` files to `.gitignore` so the real name isn't carried
// inside the outer repo's checkout.
func stageFixture(t *testing.T, name string) string {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "fixtures", name)
	dst := t.TempDir()
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if filepath.Base(rel) == "dot.gitignore" {
			target = filepath.Join(filepath.Dir(target), ".gitignore")
		}
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("stageFixture: %v", err)
	}
	return dst
}

func runRender(t *testing.T, root string, renderFn func(io.Writer, render.Input) error) []byte {
	t.Helper()
	m, err := ignore.New(ignore.Options{Root: root, RespectGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := walk.Walk(root, walk.Options{Matcher: m, MaxFileSize: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	info := project.Describe(root, "Simple Fixture")
	var buf bytes.Buffer
	if err := renderFn(&buf, render.Input{Project: info, Entries: entries}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// scrubPaths replaces the staged temp path with a stable placeholder
// so goldens don't depend on the temp dir. The staged root is passed in.
func scrubPaths(out []byte, root string) []byte {
	return []byte(strings.ReplaceAll(string(out), root, "<ROOT>"))
}

func checkGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	goldenPath := filepath.Join("..", "..", "testdata", "golden", name)
	if *update {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to create)", goldenPath, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}

func TestIndexGolden(t *testing.T) {
	root := stageFixture(t, "simple")
	got := scrubPaths(runRender(t, root, render.Index), root)
	checkGolden(t, "simple.llms.txt", got)
}

func TestFullGolden(t *testing.T) {
	root := stageFixture(t, "simple")
	got := scrubPaths(runRender(t, root, render.Full), root)
	checkGolden(t, "simple.llms-full.txt", got)
}
