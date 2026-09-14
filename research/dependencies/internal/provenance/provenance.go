// Package provenance identifies the dependency parser selected for a probe build.
package provenance

import (
	"encoding/base64"
	"fmt"
	"runtime/debug"
	"strings"
)

const modulePath = "github.com/bioshock/gospacy/v3"

// Module identifies a selected module requirement.
type Module struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

// Provider records the parser's actual code identity, including any replacement.
type Provider struct {
	Module
	Sum       string  `json:"sum"`
	Revision  *string `json:"revision"`
	Requested *Module `json:"requested,omitempty"`
}

// FromBuildInfo reads the selected parser module without consulting the working
// directory or the network. A local replacement or missing checksum is an error.
// Build settings describe the main module, so they cannot establish a dependency
// revision; Revision remains nil even when vcs.revision is present.
func FromBuildInfo(info *debug.BuildInfo) (Provider, error) {
	if info == nil {
		return Provider{}, fmt.Errorf("dependency provenance: build information unavailable")
	}
	for _, dep := range info.Deps {
		if dep != nil && dep.Path == modulePath {
			return selectedProvider(dep)
		}
	}
	return Provider{}, fmt.Errorf("dependency provenance: %s missing from build information", modulePath)
}

func selectedProvider(dep *debug.Module) (Provider, error) {
	var requested *Module
	if dep.Replace != nil {
		requested = &Module{Path: dep.Path, Version: dep.Version}
		dep = dep.Replace
	}
	if dep.Replace != nil || dep.Path == "" || dep.Version == "" || dep.Version == "(devel)" {
		return Provider{}, fmt.Errorf("dependency provenance: require a versioned module; local or nested replacements are unsupported")
	}
	hash, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dep.Sum, "h1:"))
	if err != nil || !strings.HasPrefix(dep.Sum, "h1:") || len(hash) != 32 {
		return Provider{}, fmt.Errorf("dependency provenance: %s@%s has no valid module checksum", dep.Path, dep.Version)
	}
	return Provider{
		Module: Module{Path: dep.Path, Version: dep.Version}, Sum: dep.Sum, Requested: requested,
	}, nil
}
