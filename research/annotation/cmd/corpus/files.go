package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

func loadFiles(ctx context.Context, directory string, plan corpus.Plan) (map[string][]byte, error) {
	required, err := corpus.Files(ctx, plan)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	files, readErr := readFiles(ctx, root, required)
	closeErr := root.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return files, nil
}

func readFiles(ctx context.Context, root *os.Root, required []corpus.Notice) (map[string][]byte, error) {
	files := make(map[string][]byte, len(required))
	for _, file := range required {
		data, err := commandio.Await(ctx, func() ([]byte, error) { return readFile(root, file) })
		if err != nil {
			return nil, err
		}
		files[file.Path] = data
	}
	return files, nil
}

// readFile reads one declared file through the root. A symbolic link is
// followed only inside the root, so a notice kept as a link into the
// checkout is read while a link that leaves it is refused.
func readFile(root *os.Root, declared corpus.Notice) ([]byte, error) {
	info, err := root.Stat(declared.Path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() != int64(declared.Bytes) {
		return nil, fmt.Errorf("source must be a regular file with its declared size: %s", declared.Path)
	}
	file, err := root.Open(declared.Path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, int64(declared.Bytes)+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}
