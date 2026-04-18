package summarize

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MaxLen bounds the length of a generated summary.
const MaxLen = 120

// kind classifies files for summary extraction.
type kind int

const (
	kindGeneric kind = iota
	kindSlashSlash
	kindHash
	kindPython
	kindMarkdown
	kindHTML
	kindData
)

var extKinds = map[string]kind{
	".go":    kindSlashSlash,
	".js":    kindSlashSlash,
	".jsx":   kindSlashSlash,
	".ts":    kindSlashSlash,
	".tsx":   kindSlashSlash,
	".java":  kindSlashSlash,
	".c":     kindSlashSlash,
	".h":     kindSlashSlash,
	".cc":    kindSlashSlash,
	".cpp":   kindSlashSlash,
	".hpp":   kindSlashSlash,
	".cs":    kindSlashSlash,
	".rs":    kindSlashSlash,
	".swift": kindSlashSlash,
	".kt":    kindSlashSlash,
	".scala": kindSlashSlash,
	".php":   kindSlashSlash,
	".m":     kindSlashSlash,

	".py":   kindPython,
	".sh":   kindHash,
	".bash": kindHash,
	".zsh":  kindHash,
	".rb":   kindHash,
	".pl":   kindHash,
	".yaml": kindHash,
	".yml":  kindHash,
	".toml": kindHash,
	".r":    kindHash,
	".ex":   kindHash,
	".exs":  kindHash,

	".md":       kindMarkdown,
	".markdown": kindMarkdown,

	".html": kindHTML,
	".htm":  kindHTML,
	".xml":  kindHTML,
	".svg":  kindHTML,

	".json":  kindData,
	".jsonc": kindData,
	".ini":   kindData,
	".cfg":   kindData,
	".conf":  kindData,
	".env":   kindData,
	".csv":   kindData,
	".tsv":   kindData,
	".txt":   kindData,
	".sql":   kindData,
}

// File returns a one-line summary of the file at path, or "" if none
// can be extracted. It never returns an error; unreadable files map to
// the empty string.
func File(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return Content(path, data)
}

// Content returns a summary for content that came from path. Split out
// for testability.
func Content(path string, data []byte) string {
	k := extKinds[strings.ToLower(filepath.Ext(path))]
	base := strings.ToLower(filepath.Base(path))
	if base == "makefile" {
		k = kindHash
	}
	if base == "dockerfile" {
		k = kindHash
	}

	var s string
	switch k {
	case kindSlashSlash:
		s = slashSlash(data)
	case kindHash:
		s = hashComment(data)
	case kindPython:
		s = pythonDoc(data)
	case kindMarkdown:
		s = markdownFirstPara(data)
	case kindHTML:
		s = htmlFirstTextOrComment(data)
	default:
		s = firstMeaningfulLine(data)
	}
	if s == "" {
		s = firstMeaningfulLine(data)
	}
	return clip(s)
}

func clip(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > MaxLen {
		// Cut at a word boundary when possible.
		cut := MaxLen
		if idx := strings.LastIndexByte(s[:MaxLen], ' '); idx > 40 {
			cut = idx
		}
		s = strings.TrimRight(s[:cut], " .,;:") + "…"
	}
	return s
}

func firstMeaningfulLine(data []byte) string {
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		return line
	}
	return ""
}

// slashSlash captures the leading // comment block or /* ... */ block.
func slashSlash(data []byte) string {
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	var out []string
	inBlock := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !inBlock && len(out) == 0 && line == "" {
			continue
		}
		if !inBlock && strings.HasPrefix(line, "/*") {
			inBlock = true
			line = strings.TrimPrefix(line, "/*")
			if strings.Contains(line, "*/") {
				line = line[:strings.Index(line, "*/")]
				out = append(out, strings.TrimLeft(strings.TrimSpace(line), "* "))
				break
			}
			line = strings.TrimLeft(strings.TrimSpace(line), "* ")
			if line != "" {
				out = append(out, line)
			}
			continue
		}
		if inBlock {
			if strings.Contains(line, "*/") {
				line = line[:strings.Index(line, "*/")]
				line = strings.TrimLeft(strings.TrimSpace(line), "* ")
				if line != "" {
					out = append(out, line)
				}
				break
			}
			out = append(out, strings.TrimLeft(line, "* "))
			continue
		}
		if strings.HasPrefix(line, "//") {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(line, "//")))
			continue
		}
		break
	}
	return strings.TrimSpace(strings.Join(out, " "))
}

func hashComment(data []byte) string {
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	var out []string
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if len(out) == 0 && (line == "" || strings.HasPrefix(line, "#!")) {
			continue
		}
		if strings.HasPrefix(line, "#") {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(line, "#")))
			continue
		}
		break
	}
	return strings.TrimSpace(strings.Join(out, " "))
}

var pyDocRe = regexp.MustCompile(`(?s)^\s*(?:#!.*\n)?(?:(?:#[^\n]*\n)|\s)*(?:"""(.*?)"""|'''(.*?)''')`)

func pythonDoc(data []byte) string {
	m := pyDocRe.FindSubmatch(data)
	if m == nil {
		return hashComment(data)
	}
	doc := m[1]
	if len(doc) == 0 {
		doc = m[2]
	}
	return strings.TrimSpace(string(doc))
}

func markdownFirstPara(data []byte) string {
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	var para []string
	sawH1 := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !sawH1 && strings.HasPrefix(line, "# ") {
			sawH1 = true
			continue
		}
		if line == "" {
			if len(para) > 0 {
				break
			}
			continue
		}
		if strings.HasPrefix(line, "#") {
			// subsequent heading, stop
			if len(para) > 0 {
				break
			}
			continue
		}
		para = append(para, line)
	}
	return strings.Join(para, " ")
}

var htmlCommentRe = regexp.MustCompile(`(?s)<!--(.*?)-->`)
var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func htmlFirstTextOrComment(data []byte) string {
	if m := htmlCommentRe.FindSubmatch(data); m != nil {
		return strings.TrimSpace(string(m[1]))
	}
	stripped := htmlTagRe.ReplaceAllString(string(data), " ")
	return firstMeaningfulLine([]byte(stripped))
}
