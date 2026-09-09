package repopolicy

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"slices"
	"strings"
)

const internalTestReason = "// White-box tests:"

func checkTestSource(tree fs.FS, name string) error {
	if slices.Contains(strings.Split(name, "/"), "testdata") {
		return nil
	}
	data, err := fs.ReadFile(tree, name)
	if err != nil {
		return err
	}
	positions := token.NewFileSet()
	file, err := parser.ParseFile(positions, name, data, parser.ParseComments)
	if err != nil {
		return err
	}
	internal := strings.HasSuffix(name, "_internal_test.go")
	if strings.HasSuffix(file.Name.Name, "_test") {
		if internal {
			return fmt.Errorf("blackbox test must not use the reserved _internal_test.go suffix: %s", name)
		}
		return nil
	}
	if !internal {
		return fmt.Errorf("whitebox test requires _internal_test.go filename: %s", name)
	}
	if !hasInternalTestReason(positions, file) {
		return fmt.Errorf("whitebox test requires a nonempty %q comment after package and before declarations: %s", internalTestReason, name)
	}
	return nil
}

func hasInternalTestReason(positions *token.FileSet, file *ast.File) bool {
	packageLine := positions.PositionFor(file.Name.End(), false).Line
	limit := file.FileEnd
	if len(file.Decls) != 0 {
		limit = file.Decls[0].Pos()
	}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if comment.Pos() >= limit || positions.PositionFor(comment.Pos(), false).Line <= packageLine {
				continue
			}
			if reason, found := strings.CutPrefix(comment.Text, internalTestReason); found && strings.TrimSpace(reason) != "" {
				return true
			}
		}
	}
	return false
}
