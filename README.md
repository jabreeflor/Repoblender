# Repoblender

A Go CLI that converts an entire repository into the
[llms.txt](https://llmstxt.org) format. It produces two files:

- `llms.txt` — a structured index of the repo grouped by top-level
  directory, with a one-line summary per file.
- `llms-full.txt` — the same index followed by the full contents of
  every kept file, inlined in fenced Markdown code blocks.

## Install

```
go install github.com/jabreeflor/repoblender/cmd/repoblender@latest
```

Or build from a checkout:

```
go build -o repoblender ./cmd/repoblender
```

## Usage

```
repoblender [flags] [path]
```

Path defaults to the current directory. By default the tool respects
`.gitignore`, skips common build/vendor directories, drops binary files,
and skips files larger than 256 KiB.

Useful flags:

- `-o, --output DIR` — where to write `llms.txt` / `llms-full.txt`
- `--name NAME` — project name override
- `--index-only` — skip `llms-full.txt`
- `--include PATTERN` / `--exclude PATTERN` — gitignore-style filters
  (repeatable)
- `--max-file-size BYTES` — skip files over this size (0 = no limit)
- `--respect-gitignore=false` — disable gitignore handling

Example:

```
repoblender --name MyApp --exclude 'testdata/**' -o docs .
```

## How summaries are chosen

File descriptions come from simple per-language heuristics, no LLM calls:

- `.go/.js/.ts/.java/.c/...` → leading `//` or `/* */` comment block
- `.py` → module-level docstring, then `#` comments as fallback
- `.sh/.yaml/.toml/.rb/...` → leading `#` comment block (skipping shebang)
- `.md` → first paragraph after the `# H1`
- Other text files → first non-blank line
- README summary is reused as the top-level blockquote

Summaries are trimmed to roughly 120 characters.

## Testing

```
go test ./...
```

Golden-file integration tests live in `internal/render/render_test.go`
and compare output against files in `testdata/golden/`. When intentional
output changes happen, regenerate goldens with:

```
go test ./internal/render -update
```

Fixture repos under `testdata/fixtures/` ship with a `dot.gitignore`
file that the test harness renames to `.gitignore` inside a temp
checkout, so fixture ignore rules don't leak into the outer repo.
