// Package repopolicy checks repository contracts against the actual source tree.
package repopolicy

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const modulePath = "github.com/stokaro/unswell"

// Check validates module coverage, library boundaries, the API ledger and CI pins.
func Check(tree fs.FS) error {
	if _, err := fs.Stat(tree, ".github/workflows/ci.yml"); err != nil {
		return err
	}
	modules, err := inventory(tree)
	if err != nil {
		return err
	}
	ledger, err := fs.ReadFile(tree, "docs/public_api.md")
	if err != nil {
		return err
	}
	err = fs.WalkDir(tree, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		return checkEntry(tree, modules, string(ledger), name, entry, walkErr)
	})
	return err
}

func checkEntry(tree fs.FS, modules []string, ledger, name string, entry fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		return walkErr
	}
	if entry.IsDir() {
		return skipDirectory(entry.Name())
	}
	if path.Base(name) == "go.mod" && !slices.Contains(modules, path.Dir(name)) {
		return fmt.Errorf("unclassified Go module: %s", name)
	}
	if strings.HasSuffix(name, ".go") {
		return checkGoEntry(tree, modules, name, ledger)
	}
	if strings.HasPrefix(name, ".github/workflows/") {
		return checkWorkflow(tree, name)
	}
	return nil
}

func checkGoEntry(tree fs.FS, modules []string, name, ledger string) error {
	if strings.HasSuffix(name, "_test.go") {
		if err := checkTestSource(tree, name); err != nil {
			return err
		}
	}
	if strings.HasPrefix(name, "goanalysis/") {
		return checkAdapterSource(tree, name, ledger)
	}
	if inRootModule(name, modules) {
		return checkSource(tree, name, ledger)
	}
	return nil
}

func checkAdapterSource(tree fs.FS, name, ledger string) error {
	if strings.Contains(name, "/testdata/") {
		return nil
	}
	if !strings.HasPrefix(name, "goanalysis/cmd/") {
		if err := checkSource(tree, name, ledger); err != nil {
			return err
		}
	}
	data, err := fs.ReadFile(tree, name)
	if err != nil {
		return err
	}
	file, err := parser.ParseFile(token.NewFileSet(), name, data, parser.ImportsOnly)
	if err != nil {
		return err
	}
	for _, item := range file.Imports {
		importPath, err := strconv.Unquote(item.Path.Value)
		if err != nil {
			return err
		}
		if strings.HasPrefix(importPath, modulePath+"/internal/") || strings.HasPrefix(importPath, modulePath+"/cmd/") {
			return fmt.Errorf("go/analysis adapter must use public Unswell packages: %s", name)
		}
	}
	return nil
}

func inventory(tree fs.FS) ([]string, error) {
	data, err := fs.ReadFile(tree, ".gomodules")
	if err != nil {
		return nil, err
	}
	var modules []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || !slices.Contains([]string{"runtime", "consumer", "tools"}, fields[1]) {
			return nil, fmt.Errorf("invalid module inventory row: %q", line)
		}
		if !fs.ValidPath(fields[0]) || slices.Contains(modules, fields[0]) {
			return nil, fmt.Errorf("invalid or duplicate module: %q", fields[0])
		}
		if err := moduleRole(fields[0], fields[1]); err != nil {
			return nil, err
		}
		if _, err := fs.Stat(tree, path.Join(fields[0], "go.mod")); err != nil {
			return nil, err
		}
		modules = append(modules, fields[0])
	}
	if !slices.Contains(modules, ".") {
		return nil, fmt.Errorf("root module is missing")
	}
	return modules, nil
}

func moduleRole(directory, role string) error {
	if directory == "." && role != "runtime" {
		return fmt.Errorf("root module must be runtime")
	}
	if (directory == "tools") != (role == "tools") {
		return fmt.Errorf("only the tools directory can have the tools role")
	}
	return nil
}

func skipDirectory(name string) error {
	if slices.Contains([]string{".git", "vendor", "artifacts", "bin", "dist"}, name) {
		return fs.SkipDir
	}
	return nil
}

func inRootModule(name string, modules []string) bool {
	for _, module := range modules {
		if module != "." && strings.HasPrefix(name, module+"/") {
			return false
		}
	}
	return true
}

func checkSource(tree fs.FS, name, ledger string) error {
	if strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "cmd/") {
		return nil
	}
	data, err := fs.ReadFile(tree, name)
	if err != nil {
		return err
	}
	file, err := parser.ParseFile(token.NewFileSet(), name, data, parser.ImportsOnly)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(name, "internal/") {
		if err := checkLedger(name, ledger); err != nil {
			return err
		}
	}
	if strings.HasPrefix(name, "internal/cli/") || strings.HasPrefix(name, "internal/repopolicy/") ||
		strings.HasPrefix(name, "internal/appconfig/") {
		return nil
	}
	return checkImports(name, file.Imports)
}

func checkImports(name string, imports []*ast.ImportSpec) error {
	for _, item := range imports {
		importPath, err := strconv.Unquote(item.Path.Value)
		if err != nil {
			return err
		}
		if forbiddenImport(importPath) {
			return fmt.Errorf("library boundary: %s imports %s", name, importPath)
		}
	}
	return nil
}

func checkLedger(name, ledger string) error {
	pkg := strings.TrimSuffix(modulePath+"/"+path.Dir(name), "/.")
	if !strings.Contains(ledger, "`"+pkg+"`") {
		return fmt.Errorf("public package absent from API ledger: %s", pkg)
	}
	return nil
}

func forbiddenImport(name string) bool {
	// URL parsing is a pure transformation required for SARIF locations.
	if name == "net/url" {
		return false
	}
	for _, prefix := range []string{
		"os", "net", "syscall", modulePath + "/internal/cli", modulePath + "/internal/repopolicy", modulePath + "/internal/appconfig",
	} {
		if name == prefix || strings.HasPrefix(name, prefix+"/") {
			return true
		}
	}
	return strings.HasPrefix(name, modulePath+"/cmd/") || strings.HasPrefix(name, "github.com/spf13/")
}

func checkWorkflow(tree fs.FS, name string) error {
	data, err := fs.ReadFile(tree, name)
	if err != nil {
		return err
	}
	pinned := regexp.MustCompile(`^[^@\s]+@[a-f0-9]{40}$`)
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- "))
		if !strings.HasPrefix(line, "uses:") {
			continue
		}
		value := strings.TrimSpace(strings.SplitN(strings.TrimPrefix(line, "uses:"), "#", 2)[0])
		if !strings.HasPrefix(value, "./") && !pinned.MatchString(value) {
			return fmt.Errorf("unpinned action in %s: %s", name, value)
		}
	}
	if name == ".github/workflows/ci.yml" {
		return checkCICoverage(string(data))
	}
	return nil
}

func checkCICoverage(workflow string) error {
	for _, required := range []string{
		"ubuntu-latest", "macos-latest", "windows-latest", "GOTOOLCHAIN: local",
		"go-version-file: go.mod", "scripts/modules.sh test", "make check",
	} {
		if !strings.Contains(workflow, required) {
			return fmt.Errorf("missing CI coverage: %s", required)
		}
	}
	return nil
}
