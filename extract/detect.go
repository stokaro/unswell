package extract

import (
	"bytes"
	"path"
	"strings"

	"github.com/stokaro/unswell/document"
)

// Detect selects a supported format from a filename or an explicit shebang.
// Shebangs refine generic .sh files and identify extensionless inputs.
// A shebang is inspected as text; its interpreter and arguments are never executed.
func Detect(name string, prefix []byte) (document.Format, bool) {
	name = strings.ReplaceAll(name, "\\", "/")
	if format := namedFormat(path.Base(name)); format != "" {
		return format, true
	}
	format := extensionFormat(path.Ext(name))
	if format != "" && format != document.Shell {
		return format, true
	}
	if interpreter := shebangFormat(prefix); interpreter != "" {
		return interpreter, true
	}
	return format, format != ""
}

func shebangFormat(prefix []byte) document.Format {
	line, _, _ := bytes.Cut(prefix[:min(len(prefix), 512)], []byte("\n"))
	if !bytes.HasPrefix(line, []byte("#!")) {
		return ""
	}
	fields := strings.Fields(string(line[2:]))
	if len(fields) == 0 {
		return ""
	}
	interpreter := path.Base(fields[0])
	if interpreter == "env" {
		interpreter = envInterpreter(fields[1:])
	}
	return interpreterFormat(interpreter)
}

func namedFormat(name string) document.Format {
	switch name {
	case ".bashrc", ".bash_profile", ".bash_login", ".bash_logout":
		return document.Bash
	case ".zshrc", ".zprofile", ".zshenv", ".zlogin", ".zlogout":
		return document.Zsh
	case ".profile":
		return document.Shell
	}
	return ""
}

func extensionFormat(extension string) document.Format {
	if extension == ".C" {
		return document.CPP
	}
	formats := map[string]document.Format{
		".txt": document.Plain, ".md": document.Markdown, ".markdown": document.Markdown,
		".go": document.Go, ".js": document.JavaScript, ".jsx": document.JavaScript,
		".mjs": document.JavaScript, ".cjs": document.JavaScript,
		".ts": document.TypeScript, ".mts": document.TypeScript, ".cts": document.TypeScript, ".tsx": document.TSX,
		".py": document.Python, ".pyi": document.Python, ".rs": document.Rust, ".java": document.Java,
		".c": document.C, ".h": document.C, ".cc": document.CPP, ".cpp": document.CPP,
		".cxx": document.CPP, ".hpp": document.CPP, ".hh": document.CPP, ".hxx": document.CPP,
		".cs": document.CSharp, ".csx": document.CSharp, ".yaml": document.YAML, ".yml": document.YAML,
		".sh": document.Shell, ".bash": document.Bash, ".zsh": document.Zsh, ".fish": document.Fish,
		".ps1": document.PowerShell, ".psm1": document.PowerShell, ".psd1": document.PowerShell,
	}
	return formats[strings.ToLower(extension)]
}

func envInterpreter(fields []string) string {
	for _, field := range fields {
		if field == "-S" || strings.Contains(field, "=") {
			continue
		}
		if strings.HasPrefix(field, "-") {
			return ""
		}
		return path.Base(field)
	}
	return ""
}

func interpreterFormat(name string) document.Format {
	formats := map[string]document.Format{
		"bash": document.Bash, "sh": document.Shell, "dash": document.Shell, "ash": document.Shell,
		"zsh": document.Zsh, "fish": document.Fish, "pwsh": document.PowerShell, "powershell": document.PowerShell,
		"python": document.Python, "python3": document.Python, "node": document.JavaScript,
	}
	return formats[name]
}
