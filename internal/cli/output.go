package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/report"
)

type reportTarget struct {
	format string
	path   string
}

func writeReports(environment Environment, options checkOptions, result unswell.RunResult, inputs []string) error {
	specs := options.reports
	if len(specs) == 0 {
		specs = []string{"text:-"}
	}
	targets, err := reportTargets(environment.Dir, specs, inputs)
	if err != nil {
		return err
	}
	var failures []error
	for _, target := range targets {
		var buffer bytes.Buffer
		err := report.Write(
			&buffer,
			target.format,
			result,
			report.Options{MinSeverity: options.minSeverity, MaxFindings: options.maxFindings},
		)
		if err == nil {
			if target.path == "-" {
				_, err = environment.Out.Write(buffer.Bytes())
			} else {
				err = atomicWrite(target.path, buffer.Bytes())
			}
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("write %s report: %w", target.format, err))
		}
	}
	return errors.Join(failures...)
}

func reportTargets(root string, specs, inputs []string) ([]reportTarget, error) {
	seen := make(map[string]bool)
	result := make([]reportTarget, 0, len(specs))
	for _, spec := range specs {
		format, path, ok := strings.Cut(spec, ":")
		if !ok || path == "" {
			return nil, fmt.Errorf("report must use format:path syntax")
		}
		if path != "-" {
			if !filepath.IsAbs(path) {
				path = filepath.Join(root, path)
			}
			path = filepath.Clean(path)
			if err := protectInputs(path, inputs); err != nil {
				return nil, err
			}
		}
		if seen[path] {
			return nil, fmt.Errorf("duplicate report destination %q", path)
		}
		seen[path] = true
		result = append(result, reportTarget{format: format, path: path})
	}
	return result, nil
}

func protectInputs(output string, inputs []string) error {
	outputInfo, statErr := os.Stat(output)
	if statErr != nil && !os.IsNotExist(statErr) {
		return statErr
	}
	if info, err := os.Lstat(output); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("report destination cannot be a symlink")
	}
	for _, input := range inputs {
		if filepath.Clean(input) == output {
			return fmt.Errorf("report cannot overwrite source %s", input)
		}
		if outputInfo == nil {
			continue
		}
		inputInfo, err := os.Stat(input)
		if err != nil {
			return err
		}
		if os.SameFile(inputInfo, outputInfo) {
			return fmt.Errorf("report aliases source %s", input)
		}
	}
	return nil
}

func atomicWrite(path string, data []byte) (err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".unswell-report-*")
	if err != nil {
		return err
	}
	name := file.Name()
	defer func() {
		if removeErr := os.Remove(name); removeErr != nil && !os.IsNotExist(removeErr) {
			err = errors.Join(err, removeErr)
		}
	}()
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	return os.Rename(name, path)
}
