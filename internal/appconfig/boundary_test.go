package appconfig_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/internal/appconfig"
)

// The startup boundary is the only place that touches the filesystem, so it
// resolves paths itself instead of trusting a caller's relative directory.
func TestStartupRequiresAnAbsoluteWorkingDirectory(t *testing.T) {
	c := qt.New(t)
	_, err := appconfig.ProjectRoot(appconfig.Options{Dir: "relative"})
	c.Assert(err, qt.ErrorMatches, "configuration working directory must be absolute")
	_, err = appconfig.Load(t.Context(), appconfig.Options{Dir: "relative", Discover: true})
	c.Assert(err, qt.ErrorMatches, "configuration working directory must be absolute")
}

// An explicit project root bounds discovery. A working directory outside it is a
// mistake, not a reason to walk upward past the declared boundary.
func TestDiscoveryRejectsAWorkingDirectoryOutsideTheRoot(t *testing.T) {
	c := qt.New(t)
	base := t.TempDir()
	root := filepath.Join(base, "project")
	outside := filepath.Join(base, "elsewhere")
	c.Assert(os.MkdirAll(root, 0o700), qt.IsNil)
	c.Assert(os.MkdirAll(outside, 0o700), qt.IsNil)
	_, err := appconfig.Load(t.Context(), appconfig.Options{Dir: outside, Root: root, Discover: true})
	c.Assert(err, qt.ErrorMatches, "working directory must be inside project root")
}

// Reading a configuration outside the project root needs explicit permission,
// and the message says how to grant it.
func TestOutsideRootConfigurationNeedsPermission(t *testing.T) {
	c := qt.New(t)
	base := t.TempDir()
	root := filepath.Join(base, "project")
	c.Assert(os.MkdirAll(root, 0o700), qt.IsNil)
	writeConfig(c, base, "shared.yaml", "version: 1\nextends: [builtin:strict]\n")
	_, err := appconfig.Load(t.Context(),
		appconfig.Options{Dir: root, Root: root, Path: filepath.Join(base, "shared.yaml")})
	c.Assert(err, qt.ErrorMatches, "config path: .*escapes project root.*select --project-root.*")
	loaded, err := appconfig.Load(t.Context(),
		appconfig.Options{Dir: root, Root: root, Path: filepath.Join(base, "shared.yaml"), AllowOutsideRoot: true})
	c.Assert(err, qt.IsNil)
	c.Assert(loaded.Paths, qt.HasLen, 1)
}

// A project root that does not exist fails at the boundary instead of producing
// an empty configuration that looks like a valid default.
func TestMissingProjectRootFailsToOpen(t *testing.T) {
	c := qt.New(t)
	base := t.TempDir()
	root := filepath.Join(base, "absent")
	writeConfig(c, base, "policy.yaml", "version: 1\nextends: [builtin:strict]\n")
	_, err := appconfig.Load(t.Context(),
		appconfig.Options{Dir: base, Root: root, Path: filepath.Join(base, "policy.yaml"), AllowOutsideRoot: true})
	c.Assert(err, qt.IsNotNil)
}

// The graph is bounded by resource count and total bytes so a generated
// configuration tree cannot exhaust memory before analysis starts.
func TestConfigurationGraphLimitsAreEnforced(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	var top, nested strings.Builder
	top.WriteString("version: 1\nextends: [")
	nested.WriteString("version: 1\nextends: [")
	for i := range 60 {
		if i > 0 {
			top.WriteString(", ")
		}
		fmt.Fprintf(&top, "layer%02d.yaml", i)
		writeConfig(c, root, fmt.Sprintf("layer%02d.yaml", i), "version: 1\n")
	}
	for i := range 10 {
		if i > 0 {
			nested.WriteString(", ")
		}
		fmt.Fprintf(&nested, "extra%02d.yaml", i)
		writeConfig(c, root, fmt.Sprintf("extra%02d.yaml", i), "version: 1\n")
	}
	top.WriteString("]\n")
	nested.WriteString("]\n")
	writeConfig(c, root, "layer00.yaml", nested.String())
	writeConfig(c, root, ".unswell.yaml", top.String())
	_, err := appconfig.Load(t.Context(), appconfig.Options{Dir: root, Root: root, Discover: true})
	c.Assert(err, qt.ErrorMatches, "configuration exceeds 64 resources")
}

// A directory or a device is not a configuration file, and the loader says so
// rather than reading whatever the operating system returns.
func TestConfigurationResourcesMustBeRegularFiles(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	c.Assert(os.MkdirAll(filepath.Join(root, ".unswell.yaml"), 0o700), qt.IsNil)
	_, err := appconfig.Load(t.Context(), appconfig.Options{Dir: root, Root: root, Path: ".unswell.yaml"})
	c.Assert(err, qt.ErrorMatches, ".*configuration requires regular files of at most 1 MiB")
}

// LoadResources never opens a path; it still refuses an unusable project root,
// a canceled context, and a reference that leaves the project.
func TestResourceLoadingChecksItsArgumentsBeforeReading(t *testing.T) {
	c := qt.New(t)
	_, err := appconfig.LoadResources(t.Context(), "relative", "policy.yaml", resourceMap(nil))
	c.Assert(err, qt.ErrorMatches, "resource project root must be absolute")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = appconfig.LoadResources(ctx, t.TempDir(), "policy.yaml", resourceMap(nil))
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = appconfig.LoadResources(t.Context(), t.TempDir(), "../outside.yaml", resourceMap(nil))
	c.Assert(err, qt.ErrorMatches, ".*escapes project root.*")
}
