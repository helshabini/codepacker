package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const helpText = `Code Packer - concatenates source code files with appropriate comment markers

Usage:
  program [flags]

Flags:
  -indir string
        Input directory to process (default ".")
  -outfile string
        Output file path. If not specified, uses input directory name + ".txt"
  -verbose
        Enable verbose output
  -force
        Force overwrite of existing output file
  -help
        Show this help message

Example:
  codepacker -indir ./myproject -outfile output.txt -verbose
  codepacker -indir /path/to/code/project -force

The program will:
1. Walk through all files in the input directory
2. Identify code files by their extensions
3. Add appropriate comment markers for each language
4. Concatenate all code files into a single output file`

// Maximum path length varies by OS, adding extra bytes for comment markers and newlines
// Windows MAX_PATH is 260, Unix typically 4096
const maxBufferSize = 4096 + 100 // path length + extra space for comments and formatting

// CommentStyle defines the structure for comment syntax
type CommentStyle struct {
	Prepend string // Opening/starting comment symbol
	Append  string // Closing comment symbol (if needed)
}

// GitIgnore holds the ignore patterns and their base directory
type GitIgnore struct {
	patterns []string
	baseDir  string
}

// LoadGitIgnore loads .gitignore files from the given directory and its parents
func LoadGitIgnore(dir string) (*GitIgnore, error) {
	patterns := make([]string, 0)

	// Start from the given directory and move up until we find a .git folder or reach root
	currentDir := dir
	for {
		gitignorePath := filepath.Join(currentDir, ".gitignore")
		if _, err := os.Stat(gitignorePath); err == nil {
			file, err := os.Open(gitignorePath)
			if err != nil {
				return nil, fmt.Errorf("error opening .gitignore: %v", err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				// Skip empty lines and comments
				if line != "" && !strings.HasPrefix(line, "#") {
					patterns = append(patterns, line)
				}
			}

			if scanner.Err() != nil {
				return nil, fmt.Errorf("error reading .gitignore: %v", scanner.Err())
			}
		}

		// Check if we're in a git repository
		if _, err := os.Stat(filepath.Join(currentDir, ".git")); err == nil {
			// Found the repository root, stop here
			return &GitIgnore{
				patterns: patterns,
				baseDir:  currentDir,
			}, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// We've reached the root directory
			break
		}
		currentDir = parentDir
	}

	// If we didn't find a .git directory, just use the patterns we found (if any)
	return &GitIgnore{
		patterns: patterns,
		baseDir:  dir,
	}, nil
}

// ShouldIgnore checks if a path should be ignored based on gitignore patterns
func (gi *GitIgnore) ShouldIgnore(path string) bool {
	// Convert path to be relative to the base directory
	relPath, err := filepath.Rel(gi.baseDir, path)
	if err != nil {
		return false
	}

	// Common directories to ignore even if not in .gitignore
	commonIgnores := []string{
		"node_modules",
		"vendor",
		"build",
		"dist",
		"target",
		"bin",
		"obj",
		".git",
		".idea",
		".vscode",
		"__pycache__",
		".pytest_cache",
		".mypy_cache",
	}

	// Check common ignores first
	pathParts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range pathParts {
		for _, ignore := range commonIgnores {
			if part == ignore {
				return true
			}
		}
	}

	// Check each gitignore pattern
	for _, pattern := range gi.patterns {
		matched, err := filepath.Match(pattern, relPath)
		if err == nil && matched {
			return true
		}

		// Handle directory wildcards (e.g., **/node_modules)
		if strings.Contains(pattern, "**") {
			pattern = strings.ReplaceAll(pattern, "**", "*")
			for _, part := range pathParts {
				matched, err := filepath.Match(pattern, part)
				if err == nil && matched {
					return true
				}
			}
		}
	}

	return false
}

func main() {
	// Get cmd line arguments using flags
	indir := flag.String("indir", ".", "Input directory")
	outfile := flag.String("outfile", "", "Output file")
	verbose := flag.Bool("verbose", false, "Verbose output")
	force := flag.Bool("force", false, "Force overwrite output file")
	help := flag.Bool("help", false, "Show help message")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", helpText)
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Clean and resolve the input directory path
	cleanInDir := filepath.Clean(*indir)
	absdir, err := filepath.Abs(cleanInDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving input directory path: %v\n", err)
		os.Exit(1)
	}

	// Get current working directory for output file
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Handle output file path - always in current working directory
	var outfilepath string
	if *outfile == "" {
		// Use the base name of input directory for the output file
		dirname := filepath.Base(absdir)
		outfilepath = filepath.Join(cwd, dirname+".txt")
	} else {
		// Put the specified output file in current directory
		outfilepath = filepath.Join(cwd, *outfile)
	}

	// Check if output file exists
	if !*force {
		if _, err := os.Stat(outfilepath); err == nil {
			fmt.Fprintf(os.Stderr, "Output file already exists. Use -force to overwrite.\n")
			os.Exit(1)
		}
	}

	if *verbose {
		println("Input directory:", absdir)
		println("Output file:", outfilepath)
	}

	// Load gitignore patterns
	gitignore, err := LoadGitIgnore(absdir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error loading .gitignore: %v\n", err)
		// Continue without gitignore if there's an error
	}

	f, err := os.Create(outfilepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	sb := strings.Builder{}
	sb.Grow(maxBufferSize)

	err = filepath.Walk(absdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

    // Check if path should be ignored based on gitignore rules
		if gitignore != nil && gitignore.ShouldIgnore(path) {
			if *verbose {
				fmt.Println("Skipping (ignored by gitignore):", path)
			}
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		ext := filepath.Ext(path)
		commentStyle, ok := FileExtToComment[ext]
		if !ok {
			if *verbose {
				fmt.Println("Skipping (not a code file):", path)
			}
			return nil
		}

		code := readCodeFile(path)
		if code == nil {
			if *verbose {
				fmt.Println("Skipping (empty file):", path)
			}
			return nil
		}

		if *verbose {
			fmt.Println("Processing:", path)
		}

		// Calculate path relative to indir, keep the directory structure
		relPath, err := filepath.Rel(absdir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %v", err)
		}

		sb.Reset()
		sb.WriteString(commentStyle.Prepend)
		sb.WriteString(" ")

		// Add the input directory name as prefix to maintain context
		dirName := filepath.Base(absdir)
		sb.WriteString(filepath.Join(dirName, relPath))

		sb.WriteString(" ")
		sb.WriteString(commentStyle.Append)
		sb.WriteString("\n")

		if _, err := f.WriteString(sb.String()); err != nil {
			return fmt.Errorf("error writing to output file: %v", err)
		}
		if _, err := f.Write(code); err != nil {
			return fmt.Errorf("error writing code to output file: %v", err)
		}
		if _, err := f.WriteString("\n\n"); err != nil {
			return fmt.Errorf("error writing newlines to output file: %v", err)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}
}

func readCodeFile(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil
	}

	return b
}

var FileExtToComment = map[string]CommentStyle{
	// C and C-like languages
	".c":   {Prepend: "//", Append: ""},
	".h":   {Prepend: "//", Append: ""},
	".cpp": {Prepend: "//", Append: ""},
	".hpp": {Prepend: "//", Append: ""},
	".cc":  {Prepend: "//", Append: ""},
	".hh":  {Prepend: "//", Append: ""},
	".cxx": {Prepend: "//", Append: ""},
	".cs":  {Prepend: "//", Append: ""}, // C#

	// Web development
	".js":   {Prepend: "//", Append: ""},   // JavaScript
	".jsx":  {Prepend: "//", Append: ""},   // React JSX
	".ts":   {Prepend: "//", Append: ""},   // TypeScript
	".tsx":  {Prepend: "//", Append: ""},   // TypeScript React
	".php":  {Prepend: "//", Append: ""},   // PHP (also supports #)
	".css":  {Prepend: "/*", Append: "*/"}, // CSS
	".scss": {Prepend: "//", Append: ""},   // SASS
	".less": {Prepend: "//", Append: ""},   // LESS

	// System/Shell scripting
	".sh":   {Prepend: "#", Append: ""}, // Shell script
	".bash": {Prepend: "#", Append: ""}, // Bash script
	".zsh":  {Prepend: "#", Append: ""}, // Zsh script
	".fish": {Prepend: "#", Append: ""}, // Fish script
	".ksh":  {Prepend: "#", Append: ""}, // Korn shell
	".ps1":  {Prepend: "#", Append: ""}, // PowerShell
	".psm1": {Prepend: "#", Append: ""}, // PowerShell module

	// Modern languages
	".go":    {Prepend: "//", Append: ""}, // Go
	".rs":    {Prepend: "//", Append: ""}, // Rust
	".dart":  {Prepend: "//", Append: ""}, // Dart
	".swift": {Prepend: "//", Append: ""}, // Swift
	".kt":    {Prepend: "//", Append: ""}, // Kotlin
	".scala": {Prepend: "//", Append: ""}, // Scala

	// Traditional languages
	".java":   {Prepend: "//", Append: ""}, // Java
	".groovy": {Prepend: "//", Append: ""}, // Groovy
	".rb":     {Prepend: "#", Append: ""},  // Ruby
	".py":     {Prepend: "#", Append: ""},  // Python
	".pl":     {Prepend: "#", Append: ""},  // Perl
	".pm":     {Prepend: "#", Append: ""},  // Perl module
	".lua":    {Prepend: "--", Append: ""}, // Lua
	".tcl":    {Prepend: "#", Append: ""},  // Tcl

	// Configuration and markup
	".yaml": {Prepend: "#", Append: ""},       // YAML
	".yml":  {Prepend: "#", Append: ""},       // YAML
	".toml": {Prepend: "#", Append: ""},       // TOML
	".ini":  {Prepend: ";", Append: ""},       // INI
	".conf": {Prepend: "#", Append: ""},       // Config files
	".xml":  {Prepend: "<!--", Append: "-->"}, // XML
	".html": {Prepend: "<!--", Append: "-->"}, // HTML

	// Database
	".sql":   {Prepend: "--", Append: ""}, // SQL
	".psql":  {Prepend: "--", Append: ""}, // PostgreSQL
	".mysql": {Prepend: "--", Append: ""}, // MySQL

	// Other
	".r":   {Prepend: "#", Append: ""},    // R
	".jl":  {Prepend: "#", Append: ""},    // Julia
	".fs":  {Prepend: "//", Append: ""},   // F#
	".fsx": {Prepend: "//", Append: ""},   // F# script
	".f90": {Prepend: "!", Append: ""},    // Fortran
	".f95": {Prepend: "!", Append: ""},    // Fortran
	".f":   {Prepend: "!", Append: ""},    // Fortran
	".elm": {Prepend: "--", Append: ""},   // Elm
	".ex":  {Prepend: "#", Append: ""},    // Elixir
	".exs": {Prepend: "#", Append: ""},    // Elixir script
	".erl": {Prepend: "%", Append: ""},    // Erlang
	".hrl": {Prepend: "%", Append: ""},    // Erlang header
	".hs":  {Prepend: "--", Append: ""},   // Haskell
	".lhs": {Prepend: "--", Append: ""},   // Literate Haskell
	".ml":  {Prepend: "(*", Append: "*)"}, // OCaml
	".mli": {Prepend: "(*", Append: "*)"}, // OCaml interface
	".v":   {Prepend: "//", Append: ""},   // Verilog
	".vh":  {Prepend: "//", Append: ""},   // Verilog header
	".vhd": {Prepend: "--", Append: ""},   // VHDL
}
