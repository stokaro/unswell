package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

// Acquisition walk limits bound what one checkout can make the command read.
const (
	maxAcquireFiles = 50000
	maxAcquireBytes = 512 << 20
)

type acquireOptions struct {
	root, record string
}

// runAcquire applies an acquisition record to a local checkout and writes the
// shard manifest with every exclusion. The checkout was obtained by the
// research script beforehand; this command reads only regular files beneath
// the root and never follows a symbolic link or leaves the root.
func runAcquire(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	options, err := acquireFlags(args[1:])
	if err != nil {
		return err
	}
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return readLocalArtifact(options.record, corpus.MaxAcquisitionBytes, "acquisition record")
	})
	if err != nil {
		return err
	}
	record, err := corpus.LoadAcquisition(ctx, data)
	if err != nil {
		return err
	}
	files, skipped, err := walkCheckout(ctx, options.root, record.Selection)
	if err != nil {
		return err
	}
	result, err := corpus.Acquire(ctx, record, files)
	if err != nil {
		return err
	}
	// Directories the walk never entered are exclusions too; the library only
	// sees files, so the command adds them and keeps the list sorted.
	for _, directory := range skipped {
		result.Excluded = append(result.Excluded, corpus.AcquisitionExclusion{Path: directory, Reason: "excluded_segment_directory"})
	}
	sort.Slice(result.Excluded, func(a, b int) bool { return result.Excluded[a].Path < result.Excluded[b].Path })
	return writeResult(ctx, "acquire", result, output)
}

func acquireFlags(args []string) (acquireOptions, error) {
	var options acquireOptions
	flags := flag.NewFlagSet("corpus acquire", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.root, "root", "", "Local checkout of the pinned repository snapshot")
	flags.StringVar(&options.record, "record", "", "Acquisition record naming the snapshot, rights, and selection rules")
	if err := flags.Parse(args); err != nil {
		return acquireOptions{}, err
	}
	if flags.NArg() != 0 || options.root == "" || options.record == "" {
		return acquireOptions{}, fmt.Errorf("acquire requires --root and --record")
	}
	return options, nil
}

// walkCheckout reads every regular file beneath the root in slash-path form
// and returns the directories it did not enter. Files above the record's byte
// cap are read to the cap plus one byte so the library can exclude them by
// rule; version-control metadata and excluded segments are never entered.
func walkCheckout(ctx context.Context, directory string, rules corpus.SelectionRules) (map[string][]byte, []string, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, nil, err
	}
	walk, walkErr := walkRoot(ctx, root, rules)
	closeErr := root.Close()
	if walkErr != nil {
		return nil, nil, walkErr
	}
	return walk.files, walk.skipped, closeErr
}

type checkoutWalk struct {
	root     *os.Root
	limit    int
	total    int
	excluded []string
	files    map[string][]byte
	skipped  []string
}

func walkRoot(ctx context.Context, root *os.Root, rules corpus.SelectionRules) (*checkoutWalk, error) {
	walk := &checkoutWalk{root: root, limit: max(rules.MaxSourceBytes, corpus.MaxSourceBytes) + 1,
		excluded: rules.ExcludedSegments, files: map[string][]byte{}}
	err := fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		return walk.visit(name, entry)
	})
	if err != nil {
		return nil, err
	}
	return walk, nil
}

func (walk *checkoutWalk) visit(name string, entry fs.DirEntry) error {
	if entry.IsDir() {
		base := path.Base(name)
		if name != "." && strings.HasPrefix(base, ".git") {
			return fs.SkipDir
		}
		if name != "." && slices.Contains(walk.excluded, base) {
			walk.skipped = append(walk.skipped, name)
			return fs.SkipDir
		}
		return nil
	}
	if !entry.Type().IsRegular() {
		return nil
	}
	if len(walk.files) >= maxAcquireFiles {
		return fmt.Errorf("checkout exceeds %d files", maxAcquireFiles)
	}
	data, err := readCheckoutFile(walk.root, name, walk.limit)
	if err != nil {
		return err
	}
	walk.total += len(data)
	if walk.total > maxAcquireBytes {
		return fmt.Errorf("checkout exceeds %d bytes", maxAcquireBytes)
	}
	walk.files[name] = data
	return nil
}

func readCheckoutFile(root *os.Root, name string, limit int) ([]byte, error) {
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, int64(limit)))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	return data, closeErr
}
