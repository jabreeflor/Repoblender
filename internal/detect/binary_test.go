package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasBinaryExt(t *testing.T) {
	cases := map[string]bool{
		"foo.png":       true,
		"foo.PNG":       true,
		"archive.tar":   true,
		"lib.so":        true,
		"notes.txt":     false,
		"script.go":     false,
		"README":        false,
		"dir/pic.jpeg":  true,
		"dir/data.json": false,
	}
	for path, want := range cases {
		if got := HasBinaryExt(path); got != want {
			t.Errorf("HasBinaryExt(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestIsBinaryNullByte(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "plain.txt")
	binPath := filepath.Join(dir, "weird")

	if err := os.WriteFile(textPath, []byte("hello world\nmore text"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte("abc\x00def"), 0o644); err != nil {
		t.Fatal(err)
	}

	if ok, err := IsBinary(textPath); err != nil || ok {
		t.Errorf("IsBinary(text) = %v, %v; want false, nil", ok, err)
	}
	if ok, err := IsBinary(binPath); err != nil || !ok {
		t.Errorf("IsBinary(bin) = %v, %v; want true, nil", ok, err)
	}
}

func TestIsBinaryExtensionShortCircuit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "image.png")
	if err := os.WriteFile(path, []byte("plain ascii"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := IsBinary(path)
	if err != nil || !ok {
		t.Errorf("IsBinary(%q) = %v, %v; want true via extension", path, ok, err)
	}
}
