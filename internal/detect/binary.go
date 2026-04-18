package detect

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const sniffLen = 8192

var binaryExts = map[string]struct{}{
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".bmp": {}, ".tiff": {},
	".ico": {}, ".webp": {}, ".pdf": {}, ".zip": {}, ".gz": {}, ".tgz": {},
	".bz2": {}, ".xz": {}, ".7z": {}, ".rar": {}, ".tar": {}, ".exe": {},
	".dll": {}, ".so": {}, ".dylib": {}, ".class": {}, ".jar": {}, ".war": {},
	".wasm": {}, ".o": {}, ".a": {}, ".obj": {}, ".lib": {}, ".woff": {},
	".woff2": {}, ".ttf": {}, ".otf": {}, ".eot": {}, ".mp3": {}, ".mp4": {},
	".wav": {}, ".ogg": {}, ".flac": {}, ".avi": {}, ".mov": {}, ".webm": {},
	".mkv": {}, ".bin": {}, ".dat": {}, ".db": {}, ".sqlite": {}, ".sqlite3": {},
	".pyc": {}, ".pyo": {},
}

// HasBinaryExt reports whether the extension is on the known-binary denylist.
func HasBinaryExt(path string) bool {
	_, ok := binaryExts[strings.ToLower(filepath.Ext(path))]
	return ok
}

// IsBinary returns true if the file looks binary: either its extension is
// denylisted, or a null byte appears within the first sniffLen bytes.
func IsBinary(path string) (bool, error) {
	if HasBinaryExt(path) {
		return true, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, sniffLen)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, err
	}
	return bytes.IndexByte(buf[:n], 0) != -1, nil
}
