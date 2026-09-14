package repopolicy

import (
	"fmt"
	"io/fs"
	"path"

	"go.yaml.in/yaml/v3"
)

type dependencyUpdates struct {
	Ecosystem    string   `yaml:"package-ecosystem"`
	Directory    string   `yaml:"directory"`
	Directories  []string `yaml:"directories"`
	TargetBranch string   `yaml:"target-branch"`
	Limit        *int     `yaml:"open-pull-requests-limit"`
}

func (update dependencyUpdates) enabledForDefaultBranch() bool {
	return update.Ecosystem == "gomod" && update.TargetBranch == "" && (update.Limit == nil || *update.Limit > 0)
}

func checkDependabot(tree fs.FS, modules []string) error {
	data, err := fs.ReadFile(tree, ".github/dependabot.yml")
	if err != nil {
		return err
	}
	var config struct {
		Version int                 `yaml:"version"`
		Updates []dependencyUpdates `yaml:"updates"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("invalid Dependabot configuration: %w", err)
	}
	if config.Version != 2 {
		return fmt.Errorf("dependabot configuration must use version 2")
	}
	covered := make(map[string]bool, len(modules))
	for _, module := range modules {
		covered[path.Join("/", module)] = false
	}
	for _, update := range config.Updates {
		if !update.enabledForDefaultBranch() {
			continue
		}
		if err := coverDependencyDirectories(update, covered); err != nil {
			return err
		}
	}
	for _, module := range modules {
		if !covered[path.Join("/", module)] {
			return fmt.Errorf("missing default-branch Dependabot version updates for Go module: %s", module)
		}
	}
	return nil
}

func coverDependencyDirectories(update dependencyUpdates, covered map[string]bool) error {
	if (update.Directory != "") == (len(update.Directories) != 0) {
		return fmt.Errorf("dependabot gomod update must set exactly one of directory or directories")
	}
	directories := update.Directories
	if update.Directory != "" {
		directories = []string{update.Directory}
	}
	for _, directory := range directories {
		seen, exists := covered[directory]
		if !exists {
			return fmt.Errorf("dependabot gomod directory must name a classified module explicitly: %s", directory)
		}
		if seen {
			return fmt.Errorf("duplicate Dependabot gomod directory: %s", directory)
		}
		covered[directory] = true
	}
	return nil
}
