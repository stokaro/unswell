package goanalysis

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/stokaro/unswell/document"
)

type sourceInputs struct {
	sources []document.Source
	files   map[string]*token.File
}

func (o Options) readSources(pass *analysis.Pass) (sourceInputs, error) {
	input := sourceInputs{files: make(map[string]*token.File)}
	if err := validatePass(pass); err != nil {
		return input, err
	}
	total := 0
	for _, syntax := range pass.Files {
		if err := o.Context.Err(); err != nil {
			return input, err
		}
		file, err := syntaxFile(pass.Fset, syntax)
		if err != nil {
			return input, err
		}
		source, err := o.readSource(pass, file, total)
		if err != nil {
			return input, err
		}
		if input.files[source.Name] != nil {
			return input, fmt.Errorf("duplicate logical Go source: %s; supply a shared Root for distinct directories", source.Name)
		}
		total += len(source.Bytes)
		input.sources = append(input.sources, source)
		input.files[source.Name] = file
	}
	slices.SortFunc(input.sources, func(a, b document.Source) int { return strings.Compare(a.Name, b.Name) })
	return input, nil
}

func validatePass(pass *analysis.Pass) error {
	if pass == nil || pass.Fset == nil || pass.ReadFile == nil || pass.Report == nil || len(pass.Files) == 0 {
		return fmt.Errorf("analysis requires Go files, a FileSet, ReadFile, and Report")
	}
	return nil
}

func syntaxFile(fset *token.FileSet, syntax *ast.File) (*token.File, error) {
	if syntax == nil {
		return nil, fmt.Errorf("analysis contains a nil syntax tree")
	}
	file := fset.File(syntax.FileStart)
	if file == nil || syntax.FileStart != file.Pos(0) || syntax.FileEnd != file.Pos(file.Size()) {
		return nil, fmt.Errorf("analysis syntax tree does not match its FileSet")
	}
	return file, nil
}

func (o Options) readSource(pass *analysis.Pass, file *token.File, total int) (document.Source, error) {
	name, err := o.logicalName(file.Name())
	if err != nil {
		return document.Source{}, err
	}
	policy, err := o.Engine.PolicyForFile(name)
	if err != nil {
		return document.Source{}, err
	}
	if file.Size() > policy.Analysis.MaxFileBytes || file.Size() > policy.Analysis.MaxTotalBytes-total {
		return document.Source{}, fmt.Errorf("source exceeds analysis byte limits: %s", name)
	}
	data, err := pass.ReadFile(file.Name())
	if err != nil {
		return document.Source{}, fmt.Errorf("read Go source %s: %w", name, err)
	}
	if len(data) != file.Size() {
		return document.Source{}, fmt.Errorf("source size does not match FileSet: %s", name)
	}
	return document.Source{Name: name, Format: document.Go, Bytes: slices.Clone(data)}, nil
}

func (o Options) logicalName(filename string) (string, error) {
	name := filepath.Base(filename)
	if o.Root != "" {
		relative, err := filepath.Rel(o.Root, filename)
		if err != nil || !filepath.IsLocal(relative) {
			return "", fmt.Errorf("source must be inside analysis root: %s", filename)
		}
		name = relative
	}
	if !filepath.IsLocal(name) || name == "." {
		return "", fmt.Errorf("invalid logical Go source: %s", name)
	}
	return filepath.ToSlash(name), nil
}
