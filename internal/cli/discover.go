package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/stokaro/unswell/config"
)

func discover(ctx context.Context, root string, args []string, files config.Files) ([]string, string, error) {
	if len(args) == 0 {
		args = []string{"."}
	}
	include, err := compileGlobs(files.Include)
	if err != nil {
		return nil, "", err
	}
	exclude, err := compileGlobs(files.Exclude)
	if err != nil {
		return nil, "", err
	}
	tracked, inGit, err := trackedFiles(ctx, root)
	if err != nil {
		return nil, "", err
	}
	mode := "filesystem"
	if inGit {
		mode = "git-tracked-and-explicit"
	}
	selected := make(map[string]bool)
	selection := sourceSelection{root: root, tracked: tracked, inGit: inGit, include: include, exclude: exclude, selected: selected}
	for _, arg := range args {
		if err := selection.argument(ctx, arg); err != nil {
			return nil, mode, err
		}
	}

	paths := make([]string, 0, len(selected))
	for path := range selected {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	return paths, mode, nil
}

func within(root, path string) error {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	if escapesRoot(relative) {
		return fmt.Errorf("source escapes the project root: %s", path)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	relative, err = filepath.Rel(resolvedRoot, resolved)
	if err != nil {
		return err
	}
	if escapesRoot(relative) {
		return fmt.Errorf("source symlink escapes the project root: %s", path)
	}
	return nil
}

func trackedFiles(ctx context.Context, root string) ([]string, bool, error) {
	// #nosec G204 -- A fixed read-only Git subcommand receives the project root as an argv value, without a shell.
	probe := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "--is-inside-work-tree")
	if _, err := probe.Output(); err != nil {
		if ctx.Err() != nil {
			return nil, false, ctx.Err()
		}
		var exit *exec.ExitError
		if errors.As(err, &exit) && strings.Contains(string(exit.Stderr), "not a git repository") {
			return nil, false, nil
		}
		if errors.Is(err, exec.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("inspect Git root: %w", err)
	}
	// #nosec G204 -- Only tracked paths are listed; no shell, hooks, generators, or source programs are executed.
	command := exec.CommandContext(ctx, "git", "-C", root, "ls-files", "-z", "--cached")
	output, err := command.Output()
	if err != nil {
		return nil, true, fmt.Errorf("list tracked files: %w", err)
	}
	return strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00"), true, nil
}

func selectTracked(root, directory string, tracked []string, include, exclude []*regexp.Regexp, selected map[string]bool) error {
	for _, name := range tracked {
		if !eligiblePath(name, include, exclude) {
			continue
		}
		absolute := filepath.Join(root, filepath.FromSlash(name))
		relative, err := filepath.Rel(directory, absolute)
		if err != nil {
			return err
		}
		if escapesRoot(relative) {
			continue
		}
		if err := within(root, absolute); err != nil {
			return err
		}
		info, err := os.Lstat(absolute)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		if info.Mode().IsRegular() {
			selected[absolute] = true
		}
	}
	return nil
}

func walkSources(ctx context.Context, root, directory string, include, exclude []*regexp.Regexp, selected map[string]bool) error {
	visited := 0
	return filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		visited++
		if visited > 100000 {
			return fmt.Errorf("file discovery exceeds 100000 entries")
		}
		return selectWalkEntry(root, path, entry, include, exclude, selected)
	})
}

func selectWalkEntry(root, path string, entry fs.DirEntry, include, exclude []*regexp.Regexp, selected map[string]bool) error {
	name, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	name = filepath.ToSlash(name)
	if entry.IsDir() {
		if entry.Name() == ".git" || matchesGlobs(name+"/", exclude) {
			return filepath.SkipDir
		}
		return nil
	}
	if !entry.Type().IsRegular() {
		return nil
	}
	if eligiblePath(name, include, exclude) {
		selected[path] = true
	}
	return nil
}

func eligiblePath(name string, include, exclude []*regexp.Regexp) bool {
	return name != "" && matchesGlobs(name, include) && !matchesGlobs(name, exclude)
}

func compileGlobs(patterns []string) ([]*regexp.Regexp, error) {
	result := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if len(pattern) > 500 || strings.ContainsAny(pattern, "[]{}\\") {
			return nil, fmt.Errorf("unsupported glob %q; use *, **, and ?", pattern)
		}
		expression := globExpression(pattern)
		compiled, err := regexp.Compile(expression)
		if err != nil {
			return nil, err
		}
		result = append(result, compiled)
	}
	return result, nil
}

func matchesGlobs(name string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(name) {
			return true
		}
	}
	return false
}

type sourceSelection struct {
	root             string
	tracked          []string
	inGit            bool
	include, exclude []*regexp.Regexp
	selected         map[string]bool
}

func (s sourceSelection) argument(ctx context.Context, arg string) error {
	absolute := arg
	if !filepath.IsAbs(absolute) {
		absolute = filepath.Join(s.root, arg)
	}
	absolute = filepath.Clean(absolute)
	if err := within(s.root, absolute); err != nil {
		return err
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("source symlinks are unsupported: %s", arg)
	}
	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("source is not a regular file: %s", arg)
		}
		s.selected[absolute] = true
		return nil
	}
	if s.inGit {
		err = selectTracked(s.root, absolute, s.tracked, s.include, s.exclude, s.selected)
	} else {
		err = walkSources(ctx, s.root, absolute, s.include, s.exclude, s.selected)
	}
	if err != nil {
		return err
	}
	return nil
}

func globExpression(pattern string) string {
	var expression strings.Builder
	expression.WriteByte('^')
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			fragment, end := globStar(pattern, i)
			expression.WriteString(fragment)
			i = end
		case '?':
			expression.WriteString("[^/]")
		default:
			expression.WriteString(regexp.QuoteMeta(pattern[i : i+1]))
		}
	}

	expression.WriteByte('$')
	return expression.String()
}

func globStar(pattern string, index int) (string, int) {
	if index+1 >= len(pattern) || pattern[index+1] != '*' {
		return "[^/]*", index
	}
	index++
	if index+1 < len(pattern) && pattern[index+1] == '/' {
		return "(?:.*/)?", index + 1
	}
	return ".*", index
}

func escapesRoot(relative string) bool {
	return relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
