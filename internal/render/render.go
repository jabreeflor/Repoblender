// Package render turns a project + file list into llms.txt and
// llms-full.txt output.
package render

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jabreeflor/repoblender/internal/project"
	"github.com/jabreeflor/repoblender/internal/summarize"
	"github.com/jabreeflor/repoblender/internal/walk"
)

// Input is everything needed to render either file.
type Input struct {
	Project project.Info
	Entries []walk.FileEntry
}

// Index writes an llms.txt to w: H1 name, blockquote summary, then H2
// sections grouping files by their top-level directory.
func Index(w io.Writer, in Input) error {
	bw := newBufWriter(w)
	writeHeader(bw, in)
	writeGroups(bw, in.Entries, func(e walk.FileEntry) string {
		return summarize.File(e.AbsPath)
	})
	return bw.err
}

// Full writes llms-full.txt: the same index, then a "## Full File
// Contents" section with each file inlined in a fenced code block.
func Full(w io.Writer, in Input) error {
	bw := newBufWriter(w)
	writeHeader(bw, in)
	writeGroups(bw, in.Entries, func(e walk.FileEntry) string {
		return summarize.File(e.AbsPath)
	})
	bw.writeln("## Full File Contents")
	bw.writeln("")
	for _, e := range in.Entries {
		if bw.err != nil {
			return bw.err
		}
		bw.writeln("### " + e.RelPath)
		bw.writeln("")
		data, err := os.ReadFile(e.AbsPath)
		if err != nil {
			bw.writeln(fmt.Sprintf("_could not read: %s_", err))
			bw.writeln("")
			continue
		}
		lang := languageFor(e.RelPath)
		fence := fenceFor(data)
		if lang != "" {
			bw.writeln(fence + lang)
		} else {
			bw.writeln(fence)
		}
		bw.write(string(data))
		if len(data) > 0 && data[len(data)-1] != '\n' {
			bw.writeln("")
		}
		bw.writeln(fence)
		bw.writeln("")
	}
	return bw.err
}

func writeHeader(bw *bufWriter, in Input) {
	bw.writeln("# " + in.Project.Name)
	bw.writeln("")
	if in.Project.Summary != "" {
		for _, line := range strings.Split(in.Project.Summary, "\n") {
			bw.writeln("> " + line)
		}
		bw.writeln("")
	}
}

func writeGroups(bw *bufWriter, entries []walk.FileEntry, summ func(walk.FileEntry) string) {
	groups := groupByTopDir(entries)
	names := make([]string, 0, len(groups))
	for k := range groups {
		names = append(names, k)
	}
	sort.Slice(names, func(i, j int) bool {
		// "Root" always first, then alphabetical.
		if names[i] == "Root" {
			return true
		}
		if names[j] == "Root" {
			return false
		}
		return names[i] < names[j]
	})

	for _, name := range names {
		bw.writeln("## " + name)
		bw.writeln("")
		for _, e := range groups[name] {
			desc := summ(e)
			line := "- [" + e.RelPath + "](" + e.RelPath + ")"
			if desc != "" {
				line += ": " + desc
			}
			bw.writeln(line)
		}
		bw.writeln("")
	}
}

func groupByTopDir(entries []walk.FileEntry) map[string][]walk.FileEntry {
	out := map[string][]walk.FileEntry{}
	for _, e := range entries {
		parts := strings.SplitN(e.RelPath, "/", 2)
		key := "Root"
		if len(parts) == 2 {
			key = parts[0]
		}
		out[key] = append(out[key], e)
	}
	return out
}

// languageFor returns a conservative language hint for code fences.
func languageFor(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".jsx":
		return "jsx"
	case ".java":
		return "java"
	case ".kt":
		return "kotlin"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	case ".sh", ".bash", ".zsh":
		return "bash"
	case ".c", ".h":
		return "c"
	case ".cc", ".cpp", ".hpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".swift":
		return "swift"
	case ".md", ".markdown":
		return "markdown"
	case ".json", ".jsonc":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".xml":
		return "xml"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".sql":
		return "sql"
	case ".dockerfile":
		return "dockerfile"
	}
	if strings.EqualFold(filepath.Base(path), "Dockerfile") {
		return "dockerfile"
	}
	if strings.EqualFold(filepath.Base(path), "Makefile") {
		return "makefile"
	}
	return ""
}

// fenceFor picks a backtick fence long enough not to collide with
// backticks already inside the content. Default is 3; grows as needed.
func fenceFor(data []byte) string {
	maxRun := 0
	run := 0
	for _, b := range data {
		if b == '`' {
			run++
			if run > maxRun {
				maxRun = run
			}
		} else {
			run = 0
		}
	}
	n := 3
	if maxRun >= 3 {
		n = maxRun + 1
	}
	return strings.Repeat("`", n)
}

// bufWriter is a tiny helper that sticks a single error for the whole
// write sequence so call sites stay clean.
type bufWriter struct {
	w   io.Writer
	err error
}

func newBufWriter(w io.Writer) *bufWriter { return &bufWriter{w: w} }

func (b *bufWriter) write(s string) {
	if b.err != nil {
		return
	}
	_, b.err = io.WriteString(b.w, s)
}

func (b *bufWriter) writeln(s string) {
	b.write(s)
	b.write("\n")
}
