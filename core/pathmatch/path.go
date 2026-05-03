package pathmatch

import (
	"path/filepath"
	"strings"
)

// Normalize canonicalizes paths used by exclude patterns and rule-level
// path matching so callers agree on slash, prefix, and trailing-slash shape.
func Normalize(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "/")
	for strings.HasPrefix(path, "./") {
		path = strings.TrimPrefix(path, "./")
	}
	if strings.HasSuffix(path, "/") && !strings.HasSuffix(path, "...") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}
