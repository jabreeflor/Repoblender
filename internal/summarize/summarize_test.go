package summarize

import "testing"

func TestContent(t *testing.T) {
	cases := []struct {
		name string
		path string
		body string
		want string
	}{
		{
			name: "go line comments",
			path: "foo.go",
			body: "// Package foo does things.\n// It is useful.\npackage foo\n",
			want: "Package foo does things. It is useful.",
		},
		{
			name: "go block comment",
			path: "foo.go",
			body: "/* Package foo\n * manages widgets.\n */\npackage foo\n",
			want: "Package foo manages widgets.",
		},
		{
			name: "python docstring",
			path: "mod.py",
			body: "#!/usr/bin/env python\n\"\"\"A tidy module.\"\"\"\nimport os\n",
			want: "A tidy module.",
		},
		{
			name: "python no docstring falls back to hash",
			path: "mod.py",
			body: "# helper script\nimport os\n",
			want: "helper script",
		},
		{
			name: "shell hash comment skips shebang",
			path: "run.sh",
			body: "#!/bin/bash\n# Build the thing.\nmake\n",
			want: "Build the thing.",
		},
		{
			name: "markdown first paragraph after h1",
			path: "README.md",
			body: "# Title\n\nFirst paragraph line one.\nLine two.\n\nSecond para.\n",
			want: "First paragraph line one. Line two.",
		},
		{
			name: "yaml hash",
			path: "conf.yaml",
			body: "# Server config\nport: 8080\n",
			want: "Server config",
		},
		{
			name: "json falls back to first line",
			path: "pkg.json",
			body: "{\n  \"name\": \"x\"\n}\n",
			want: "{",
		},
		{
			name: "empty file yields empty",
			path: "x.go",
			body: "",
			want: "",
		},
		{
			name: "truncates very long summary",
			path: "a.md",
			body: "# t\n\n" + longString(400),
			want: "", // checked separately
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Content(c.path, []byte(c.body))
			if c.name == "truncates very long summary" {
				if len(got) > MaxLen+3 { // allow ellipsis
					t.Errorf("expected truncation, got len=%d", len(got))
				}
				return
			}
			if got != c.want {
				t.Errorf("Content(%q)\n got=%q\nwant=%q", c.path, got, c.want)
			}
		})
	}
}

func longString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a' + byte(i%26)
	}
	return string(b)
}
