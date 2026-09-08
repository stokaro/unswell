package appconfig

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectRoot resolves only the project boundary, without discovering or loading
// configuration. Immutable-resource adapters select policy from their own source.
func ProjectRoot(options Options) (string, error) {
	if !filepath.IsAbs(options.Dir) {
		return "", fmt.Errorf("configuration working directory must be absolute")
	}
	return projectRoot(options, options.Dir)
}

func selectConfig(options Options) (string, string, error) {
	if !filepath.IsAbs(options.Dir) {
		return "", "", fmt.Errorf("configuration working directory must be absolute")
	}
	selected := absolute(options.Dir, options.Path)
	start := options.Dir
	if !options.Discover && selected != "" {
		start = filepath.Dir(selected)
	}
	root, err := projectRoot(options, start)
	if err != nil {
		return "", "", err
	}
	if selected != "" || !options.Discover || options.Inline != nil {
		return root, selected, nil
	}
	if _, err := relativeWithin(root, options.Dir); err != nil {
		return "", "", err
	}
	selected, err = discoverConfig(root, options.Dir)
	return root, selected, err
}

func discoverConfig(root, dir string) (string, error) {
	for ; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, ".unswell.yaml")
		_, err := os.Lstat(candidate)
		if err == nil {
			return candidate, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		if dir == root {
			return "", nil
		}
	}
}

func projectRoot(options Options, start string) (string, error) {
	if options.Root != "" {
		return absolute(options.Dir, options.Root), nil
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		_, err := os.Lstat(filepath.Join(dir, ".git"))
		if err == nil {
			return dir, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		if filepath.Dir(dir) == dir {
			return start, nil
		}
	}
}

func absolute(dir, name string) string {
	if name == "" {
		return ""
	}
	if filepath.IsAbs(name) {
		return filepath.Clean(name)
	}
	return filepath.Join(dir, name)
}

func relativeWithin(root, name string) (string, error) {
	relative, err := filepath.Rel(root, name)
	if err != nil {
		return "", err
	}
	if relative != "." && !filepath.IsLocal(relative) {
		return "", fmt.Errorf("working directory must be inside project root")
	}
	return relative, nil
}
