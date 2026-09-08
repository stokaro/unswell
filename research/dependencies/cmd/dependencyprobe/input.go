package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/stokaro/unswell/document"
)

type sourceCase struct {
	ID     string          `json:"id"`
	Format document.Format `json:"format"`
	Text   string          `json:"text"`
}

func loadCases(name string) ([]sourceCase, string, error) {
	data, err := readLimited(name, 1<<20)
	if err != nil {
		return nil, "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var cases []sourceCase
	if err := decoder.Decode(&cases); err != nil {
		return nil, "", err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, "", fmt.Errorf("case file must contain one JSON array")
	}
	if len(cases) == 0 || len(cases) > 1000 {
		return nil, "", fmt.Errorf("require 1 through 1000 source cases")
	}
	seen := make(map[string]bool)
	for _, item := range cases {
		if item.ID == "" || seen[item.ID] || len(item.Text) > 8192 {
			return nil, "", fmt.Errorf("case IDs must be unique and nonempty; source limit is 8192 bytes")
		}
		seen[item.ID] = true
	}
	return cases, fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func readLimited(name string, limit int64) ([]byte, error) {
	// #nosec G304 -- This developer command reads an explicitly supplied local path.
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	err = errors.Join(err, file.Close())
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("file exceeds %d bytes", limit)
	}
	return data, nil
}

func inventoryModel(path string) (map[string]string, int64, error) {
	files := make(map[string]string)
	var size int64
	err := filepath.WalkDir(path, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		if !entry.Type().IsRegular() || len(files) >= 100 {
			return fmt.Errorf("model must contain at most 100 files and only regular files")
		}
		data, err := readLimited(name, 32<<20-size)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(path, name)
		if err != nil {
			return err
		}
		size += int64(len(data))
		files[filepath.ToSlash(relative)] = fmt.Sprintf("%x", sha256.Sum256(data))
		return nil
	})
	return files, size, err
}

func checkModelMetadata(path string) error {
	data, err := readLimited(filepath.Join(path, "meta.json"), 1<<20)
	if err != nil {
		return err
	}
	var meta struct {
		Lang    string `json:"lang"`
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return err
	}
	if meta.Lang != "en" || meta.Name != "core_web_sm" || meta.Version != "3.8.0" {
		return fmt.Errorf("experiment requires en_core_web_sm 3.8.0")
	}
	return nil
}
