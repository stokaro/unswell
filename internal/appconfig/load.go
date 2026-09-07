// Package appconfig loads local policy resources at the CLI and MCP startup
// boundary. The analysis library receives only the resulting in-memory bundle.
package appconfig

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/stokaro/unswell/config"
)

// Options makes discovery and permission to leave the project root explicit.
// Dir must be absolute. Without Root, discovery uses the nearest Git boundary or
// Dir. With discovery disabled, an explicit Path establishes the starting directory.
type Options struct {
	Dir, Root, Path  string
	Discover         bool
	AllowOutsideRoot bool
	Inline           []byte
}

// Loaded owns resource bytes and records every selected path for output guards.
type Loaded struct {
	Bundle config.Bundle
	Root   string
	Paths  []string
}

type loader struct {
	loaded Loaded
	root   *os.Root
	active map[string]bool
	infos  map[string]os.FileInfo
	total  int
}

// Load reads only explicit or discovered local configuration and its dependencies.
// It never reads a home configuration or downloads a resource.
func Load(ctx context.Context, options Options) (Loaded, error) {
	if err := ctx.Err(); err != nil {
		return Loaded{}, err
	}
	root, selected, err := selectConfig(options)
	if err != nil {
		return Loaded{}, err
	}
	loaded := Loaded{Root: root, Bundle: config.Bundle{
		Root: ".unswell.yaml", Files: make(map[string][]byte), AllowOutsideRoot: options.AllowOutsideRoot,
	}}
	if selected == "" {
		loaded.Bundle.Files[loaded.Bundle.Root] = slices.Clone(options.Inline)
		return loaded, nil
	}
	name, err := filepath.Rel(root, selected)
	if err != nil {
		return Loaded{}, err
	}
	loaded.Bundle.Root, err = config.ResolveReference(".unswell.yaml", filepath.ToSlash(name), options.AllowOutsideRoot)
	if err != nil {
		return Loaded{}, fmt.Errorf("config path: %w; select --project-root or explicitly allow outside-root configuration", err)
	}
	rootHandle, err := os.OpenRoot(root)
	if err != nil {
		return Loaded{}, err
	}
	l := loader{loaded: loaded, root: rootHandle, active: make(map[string]bool), infos: make(map[string]os.FileInfo)}
	err = l.read(ctx, loaded.Bundle.Root, "config", 0)
	err = errors.Join(err, rootHandle.Close())
	if err != nil {
		return Loaded{}, err
	}
	slices.Sort(l.loaded.Paths)
	return l.loaded, nil
}

func (l *loader) read(ctx context.Context, name, kind string, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth > 16 || l.active[name] {
		return fmt.Errorf("configuration cycle or inheritance depth exceeds 16 at %s", name)
	}
	if _, exists := l.loaded.Bundle.Files[name]; exists {
		return nil
	}
	if len(l.loaded.Bundle.Files) >= 64 {
		return fmt.Errorf("configuration exceeds 64 resources")
	}
	data, err := l.resource(name)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	l.total += len(data)
	if l.total > 8<<20 {
		return fmt.Errorf("configuration exceeds 8 MiB total")
	}
	l.loaded.Bundle.Files[name] = data
	l.loaded.Paths = append(l.loaded.Paths, filepath.Join(l.loaded.Root, filepath.FromSlash(name)))
	if kind == "dictionary" {
		return nil
	}
	return l.dependencies(ctx, name, data, depth)
}

func (l *loader) dependencies(ctx context.Context, name string, data []byte, depth int) error {
	refs, err := config.References(data)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	l.active[name] = true
	defer delete(l.active, name)
	for _, ref := range refs {
		resolved, err := config.ResolveReference(name, ref.Path, l.loaded.Bundle.AllowOutsideRoot)
		if err != nil {
			return err
		}
		if err := l.read(ctx, resolved, ref.Kind, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func (l *loader) resource(name string) ([]byte, error) {
	local := filepath.FromSlash(name)
	info, err := l.root.Stat(local)
	if l.loaded.Bundle.AllowOutsideRoot {
		info, err = os.Stat(filepath.Join(l.loaded.Root, local))
	}
	if err != nil {
		return nil, err
	}
	if err := l.regular(name, info); err != nil {
		return nil, err
	}
	file, err := l.open(local)
	if err != nil {
		return nil, err
	}
	data, readErr := readResource(file, info)
	return data, errors.Join(readErr, file.Close())
}

func (l *loader) regular(name string, info os.FileInfo) error {
	if !info.Mode().IsRegular() || info.Size() > 1<<20 {
		return fmt.Errorf("configuration requires regular files of at most 1 MiB")
	}
	for other, previous := range l.infos {
		if os.SameFile(previous, info) {
			return fmt.Errorf("duplicate configuration file through %q and %q", other, name)
		}
	}
	l.infos[name] = info
	return nil
}

func (l *loader) open(name string) (*os.File, error) {
	if l.loaded.Bundle.AllowOutsideRoot {
		// #nosec G304 -- Startup explicitly permits outside-root local configuration; reads remain bounded and regular-file-only.
		return os.Open(filepath.Join(l.loaded.Root, name))
	}
	return l.root.Open(name)
}

func readResource(file *os.File, before os.FileInfo) ([]byte, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || !os.SameFile(before, info) {
		return nil, fmt.Errorf("configuration file changed during loading")
	}
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("configuration exceeds 1 MiB")
	}
	return data, err
}
