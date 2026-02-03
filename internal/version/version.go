// Package version provides a single source of truth for the application version.
package version

const (
	// Version is the semantic version of the application.
	Version = "v5.3.6"

	// Edition describes the deployment edition.
	Edition = "Clinical Edition"

	// FullName combines version and edition for display.
	FullName = Version + " " + Edition
)
