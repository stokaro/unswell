package evaluation

import (
	"runtime"
	"runtime/debug"
)

// Execution records the build producing the comparison, separately from each
// prediction producer. Unknown VCS metadata remains explicit.
type Execution struct {
	GoVersion string `json:"go_version"`
	GOOS      string `json:"goos"`
	GOARCH    string `json:"goarch"`
	Revision  string `json:"revision"`
	Modified  bool   `json:"modified"`
}

func executionIdentity() Execution {
	result := Execution{GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Revision: "unknown"}
	info, exists := debug.ReadBuildInfo()
	if !exists {
		return result
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			result.Revision = setting.Value
		case "vcs.modified":
			result.Modified = setting.Value == "true"
		}
	}
	return result
}
