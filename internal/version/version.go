// Package version holds build-time identity for the application.
// Values are injected via -ldflags during production builds.
package version

// These defaults are used for local `go run` / `wails3 dev` when ldflags are absent.
var (
	// Version is the semantic version (e.g. "0.1.0" or "0.1.0-rc.1").
	Version = "dev"
	// Commit is the short git SHA used for the build.
	Commit = "unknown"
	// BuildTime is the UTC build timestamp in RFC3339 form.
	BuildTime = "unknown"
)

// Summary returns a human-readable version string.
func Summary() string {
	if Commit == "" || Commit == "unknown" {
		return Version
	}
	return Version + "+" + Commit
}
