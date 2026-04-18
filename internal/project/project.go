package project

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Info carries the human-facing name and summary for a repo.
type Info struct {
	Name    string
	Summary string
}

// Describe derives a project name and summary from root. If nameOverride
// is non-empty it is used verbatim. Otherwise the name is the base name
// of the absolute root. The summary is the first paragraph after the H1
// of a root-level README (README.md / README.markdown / README), if any.
func Describe(root, nameOverride string) Info {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	info := Info{Name: nameOverride}
	if info.Name == "" {
		info.Name = filepath.Base(abs)
	}
	info.Summary = readmeSummary(abs)
	return info
}

func readmeSummary(root string) string {
	for _, name := range []string{"README.md", "readme.md", "README.markdown", "README"} {
		p := filepath.Join(root, name)
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		sawH1 := false
		var para []string
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if !sawH1 && strings.HasPrefix(line, "# ") {
				sawH1 = true
				continue
			}
			if !sawH1 {
				continue
			}
			if line == "" {
				if len(para) > 0 {
					break
				}
				continue
			}
			if strings.HasPrefix(line, "#") {
				if len(para) > 0 {
					break
				}
				continue
			}
			para = append(para, line)
		}
		return strings.Join(para, " ")
	}
	return ""
}
