package appconfig

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/stokaro/unswell/config"
)

// LoadResources resolves a configuration graph through an explicit bounded reader.
// Root identifies the logical project for returned paths; no path is opened here.
// The reader must honor the caller's context and enforce its own read byte limit.
// An empty selected name chooses the builtin default without calling the reader.
func LoadResources(ctx context.Context, root, selected string, read func(string) ([]byte, error)) (Loaded, error) {
	if err := ctx.Err(); err != nil {
		return Loaded{}, err
	}
	if !filepath.IsAbs(root) {
		return Loaded{}, fmt.Errorf("resource project root must be absolute")
	}
	loaded := Loaded{Root: filepath.Clean(root), Bundle: config.Bundle{
		Root: ".unswell.yaml", Files: make(map[string][]byte),
	}}
	if selected == "" {
		loaded.Bundle.Files[loaded.Bundle.Root] = nil
		return loaded, nil
	}
	if read == nil {
		return Loaded{}, fmt.Errorf("configuration resource reader is required")
	}
	name, err := config.ResolveReference(loaded.Bundle.Root, selected, false)
	if err != nil {
		return Loaded{}, err
	}
	loaded.Bundle.Root = name
	l := loader{loaded: loaded, reader: read, active: make(map[string]bool)}
	if err := l.read(ctx, name, "config", 0); err != nil {
		return Loaded{}, err
	}
	if err := ctx.Err(); err != nil {
		return Loaded{}, err
	}
	slices.Sort(l.loaded.Paths)
	return l.loaded, nil
}
