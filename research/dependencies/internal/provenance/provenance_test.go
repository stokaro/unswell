package provenance_test

import (
	"encoding/base64"
	"encoding/json"
	"runtime/debug"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/dependencies/internal/provenance"
)

const parserPath = "github.com/bioshock/gospacy/v3"

func checksum(text string) string {
	return "h1:" + base64.StdEncoding.EncodeToString([]byte(strings.Repeat(text, 32)))
}

func TestProviderTracksSelectedBuild(t *testing.T) {
	for _, version := range []string{"v3.8.14-port.2", "v3.8.14-port.99"} {
		t.Run(version, func(t *testing.T) {
			c := qt.New(t)
			info := &debug.BuildInfo{
				Deps:     []*debug.Module{{Path: parserPath, Version: version, Sum: checksum("a")}},
				Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "main-module-revision"}},
			}
			provider, err := provenance.FromBuildInfo(info)
			c.Assert(err, qt.IsNil)
			c.Assert(provider.Path, qt.Equals, parserPath)
			c.Assert(provider.Version, qt.Equals, version)
			c.Assert(provider.Sum, qt.Equals, checksum("a"))
			c.Assert(provider.Revision, qt.IsNil)
			c.Assert(provider.Requested, qt.IsNil)
			encoded, err := json.Marshal(provider)
			c.Assert(err, qt.IsNil)
			c.Assert(string(encoded), qt.Contains, `"revision":null`)
		})
	}
}

func TestVersionedReplacementUsesActualCode(t *testing.T) {
	c := qt.New(t)
	info := &debug.BuildInfo{Deps: []*debug.Module{{
		Path: parserPath, Version: "v3.8.14-port.2", Sum: checksum("a"),
		Replace: &debug.Module{Path: "example.org/parser/v3", Version: "v3.1.0", Sum: checksum("b")},
	}}}
	provider, err := provenance.FromBuildInfo(info)
	c.Assert(err, qt.IsNil)
	c.Assert(provider.Path, qt.Equals, "example.org/parser/v3")
	c.Assert(provider.Version, qt.Equals, "v3.1.0")
	c.Assert(provider.Sum, qt.Equals, checksum("b"))
	c.Assert(provider.Requested, qt.DeepEquals, &provenance.Module{Path: parserPath, Version: "v3.8.14-port.2"})
	c.Assert(provider.Revision, qt.IsNil)
}

func TestUnverifiableProviderIsAnError(t *testing.T) {
	for _, row := range []struct {
		name string
		info *debug.BuildInfo
	}{
		{"no build information", nil},
		{"provider absent", &debug.BuildInfo{Deps: []*debug.Module{nil, {Path: "example.org/other"}}}},
		{"no version", moduleInfo("", checksum("a"))},
		{"development version", moduleInfo("(devel)", checksum("a"))},
		{"no checksum", moduleInfo("v3.0.0", "")},
		{"invalid checksum", moduleInfo("v3.0.0", "h1:not-base64")},
		{"wrong checksum length", moduleInfo("v3.0.0", "h1:YQ==")},
		{"wrong checksum algorithm", moduleInfo("v3.0.0", "h2:"+strings.TrimPrefix(checksum("a"), "h1:"))},
		{"local replacement", replacementInfo(&debug.Module{Path: "../parser"})},
		{"replacement without checksum", replacementInfo(&debug.Module{Path: parserPath, Version: "v3.0.0"})},
		{"nested replacement", replacementInfo(&debug.Module{Path: parserPath, Version: "v3.0.0", Replace: &debug.Module{}})},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			provider, err := provenance.FromBuildInfo(row.info)
			c.Assert(err, qt.ErrorMatches, "dependency provenance: .*")
			c.Assert(provider, qt.DeepEquals, provenance.Provider{})
		})
	}
}

func moduleInfo(version, sum string) *debug.BuildInfo {
	return &debug.BuildInfo{Deps: []*debug.Module{{Path: parserPath, Version: version, Sum: sum}}}
}

func replacementInfo(replacement *debug.Module) *debug.BuildInfo {
	info := moduleInfo("v3.8.14-port.2", checksum("a"))
	info.Deps[0].Replace = replacement
	return info
}
