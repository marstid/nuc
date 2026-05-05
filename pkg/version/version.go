// Package version provides build-time version information for the nuc CLI.
// Variables are set via ldflags at build time.
package version

// Build-time variables, set via:
//
//	-ldflags "-X github.com/marstid/nuc/pkg/version.Version=v1.0.0
//	          -X github.com/marstid/nuc/pkg/version.Commit=abc1234
//	          -X github.com/marstid/nuc/pkg/version.Date=2024-03-20T14:30:00Z"
var (
	// Version is the semantic version of the binary.
	Version = "dev"

	// Commit is the git commit SHA that produced the binary.
	Commit = "none"

	// Date is the build timestamp in ISO 8601 format.
	Date = "unknown"
)

// String returns a formatted version string suitable for display.
func String() string {
	return Version + " (commit: " + Commit + ", built: " + Date + ")"
}
