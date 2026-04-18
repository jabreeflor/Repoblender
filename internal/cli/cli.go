// Package cli implements the repoblender command-line interface.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	"github.com/jabreeflor/repoblender/internal/ignore"
	"github.com/jabreeflor/repoblender/internal/project"
	"github.com/jabreeflor/repoblender/internal/render"
	"github.com/jabreeflor/repoblender/internal/walk"
)

const defaultMaxSize = 256 * 1024 // 256 KiB

// Run parses args and executes the command, writing status to stdout/stderr.
// It returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := pflag.NewFlagSet("repoblender", pflag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		outputDir    = fs.StringP("output", "o", ".", "directory to write output files")
		nameOverride = fs.String("name", "", "project name (default: repo directory name)")
		full         = fs.Bool("full", true, "also emit llms-full.txt")
		indexOnly    = fs.Bool("index-only", false, "only emit llms.txt (overrides --full)")
		includeGlobs = fs.StringArray("include", nil, "include only files matching gitignore-style pattern (repeatable)")
		excludeGlobs = fs.StringArray("exclude", nil, "exclude files matching gitignore-style pattern (repeatable)")
		maxFileSize  = fs.Int64("max-file-size", defaultMaxSize, "skip files larger than N bytes (0 = no limit)")
		respectGI    = fs.Bool("respect-gitignore", true, "honor .gitignore at the repo root")
		showHelp     = fs.BoolP("help", "h", false, "show help")
	)

	fs.Usage = func() {
		fmt.Fprintf(stderr, "usage: repoblender [flags] [path]\n\n")
		fmt.Fprintf(stderr, "Convert a repository into llms.txt format.\n\n")
		fmt.Fprintf(stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if err == pflag.ErrHelp {
			return 0
		}
		return 2
	}
	if *showHelp {
		fs.Usage()
		return 0
	}

	root := "."
	if fs.NArg() > 0 {
		root = fs.Arg(0)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(stderr, "repoblender: %v\n", err)
		return 1
	}
	fi, err := os.Stat(absRoot)
	if err != nil {
		fmt.Fprintf(stderr, "repoblender: %v\n", err)
		return 1
	}
	if !fi.IsDir() {
		fmt.Fprintf(stderr, "repoblender: %s is not a directory\n", absRoot)
		return 1
	}

	matcher, err := ignore.New(ignore.Options{
		Root:             absRoot,
		RespectGitignore: *respectGI,
		ExtraPatterns:    ignore.NormalizePatterns(*excludeGlobs),
		IncludePatterns:  ignore.NormalizePatterns(*includeGlobs),
	})
	if err != nil {
		fmt.Fprintf(stderr, "repoblender: %v\n", err)
		return 1
	}

	entries, err := walk.Walk(absRoot, walk.Options{
		Matcher:     matcher,
		MaxFileSize: *maxFileSize,
	})
	if err != nil {
		fmt.Fprintf(stderr, "repoblender: %v\n", err)
		return 1
	}

	info := project.Describe(absRoot, *nameOverride)
	in := render.Input{Project: info, Entries: entries}

	outDir, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Fprintf(stderr, "repoblender: %v\n", err)
		return 1
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(stderr, "repoblender: %v\n", err)
		return 1
	}

	indexPath := filepath.Join(outDir, "llms.txt")
	if err := writeFile(indexPath, func(w io.Writer) error { return render.Index(w, in) }); err != nil {
		fmt.Fprintf(stderr, "repoblender: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "wrote %s (%d files)\n", indexPath, len(entries))

	if *full && !*indexOnly {
		fullPath := filepath.Join(outDir, "llms-full.txt")
		if err := writeFile(fullPath, func(w io.Writer) error { return render.Full(w, in) }); err != nil {
			fmt.Fprintf(stderr, "repoblender: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "wrote %s\n", fullPath)
	}
	return 0
}

func writeFile(path string, render func(io.Writer) error) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := render(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
