package walk

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/jabreeflor/repoblender/internal/detect"
	"github.com/jabreeflor/repoblender/internal/ignore"
)

// FileEntry is a file kept after filtering.
type FileEntry struct {
	// RelPath is forward-slash relative to the repo root.
	RelPath string
	// AbsPath is the resolved absolute path.
	AbsPath string
	// Size in bytes.
	Size int64
}

// Options controls Walk behavior.
type Options struct {
	Matcher     *ignore.Matcher
	MaxFileSize int64 // 0 means no limit
}

// Walk returns every file under root that survives filtering. Results
// are sorted lexicographically by RelPath for deterministic downstream
// rendering.
func Walk(root string, opts Options) ([]FileEntry, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var entries []FileEntry
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == absRoot {
			return nil
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if opts.Matcher != nil && opts.Matcher.Match(rel, d.IsDir()) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if opts.MaxFileSize > 0 && info.Size() > opts.MaxFileSize {
			return nil
		}

		isBin, err := detect.IsBinary(path)
		if err != nil {
			// Unreadable file — skip rather than fail the whole walk.
			if os.IsPermission(err) {
				return nil
			}
			return err
		}
		if isBin {
			return nil
		}

		entries = append(entries, FileEntry{
			RelPath: rel,
			AbsPath: path,
			Size:    info.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RelPath < entries[j].RelPath
	})
	return entries, nil
}
