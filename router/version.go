package router

import "strings"

// Version creates a version-prefixed sub-router (e.g. "v1" -> "/api/v1" or "/v1").
func Version(r Router, version string, fn func(Router)) {
	cleanVersion := strings.Trim(version, "/")
	if !strings.HasPrefix(cleanVersion, "v") {
		cleanVersion = "v" + cleanVersion
	}
	sub := NewGroup(r, "/api/"+cleanVersion)
	fn(sub)
}
