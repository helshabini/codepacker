# codepacker

A simple tool to pack your code into a single file. Typically used to pack a code project for use with an LLM.

## Features

- Concatenates all code files from a directory into a single file
- Automatically detects and uses appropriate comment syntax for different programming languages
- Preserves relative path information in file headers
- Respects `.gitignore` rules and common ignore patterns
- Supports 50+ programming languages and file types
- Maintains project structure in the output file

## Installation

```bash
go install github.com/helshabini/codepacker@latest
```

## Usage

Basic usage:
```bash
codepacker -indir ./myproject
```

This will create a file named `myproject.txt` in your current directory containing all the code files.

### Command Line Options

```bash
codepacker [flags]

Flags:
  -indir string
        Input directory to process (default ".")
  -outfile string
        Output file path (in current directory)
  -verbose
        Enable verbose output
  -force
        Force overwrite of existing output file
  -help
        Show help message
```

### Examples

Process the current directory:
```bash
codepacker
```

Process a specific project directory:
```bash
codepacker -indir ~/projects/myapp
```

Specify custom output file:
```bash
codepacker -indir ./src -outfile code_review.txt
```

Enable verbose output:
```bash
codepacker -indir ./project -verbose
```

Force overwrite existing output:
```bash
codepacker -indir ./project -force
```

## File Type Support

The tool supports many common programming languages and file types, including:

- C/C++ (.c, .h, .cpp, .hpp)
- Web (.js, .ts, .jsx, .tsx, .html, .css)
- System (.sh, .bash, .ps1)
- Modern Languages (Go, Rust, Swift, Kotlin)
- Traditional Languages (Java, Python, Ruby, Perl)
- Configuration (.yaml, .toml, .ini)
- And many more...

Each file type is processed with its appropriate comment syntax.

## Ignored Paths

The tool automatically skips:

- Files and directories specified in `.gitignore`
- Common dependency directories (node_modules, vendor)
- Build directories (dist, build, target)
- VCS directories (.git)
- IDE directories (.vscode, .idea)
- Cache directories (__pycache__, .mypy_cache)

## Output Format

The output file contains each source file preceded by a comment header showing its path relative to the project root:

```
// project/src/main.go
package main
...

// project/src/utils/helper.go
package utils
...
```

## Use with LLMs

The output file is formatted to be easily readable by Large Language Models. Each file is clearly delimited with comments and maintains its original structure, making it ideal for:

- Code review requests
- Architecture analysis
- Documentation generation
- Pattern recognition tasks

## Contributing

Contributions are welcome! Areas that could use improvements:

- Additional language support
- Enhanced gitignore pattern matching
- Output format options
- Test coverage
- Documentation

## License

MIT License

Copyright (c) 2025 Hazem Elshabini

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

