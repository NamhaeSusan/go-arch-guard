package core

import (
	"path/filepath"
	"strings"
)

// NormalizeMatchPath canonicalizes both exclude patterns and the file paths
// rules emit so they match consistently. The supported equivalences are:
//
//   - Backslashes are converted to forward slashes (Windows-shell paste).
//   - Leading "/" is trimmed (absolute-path fallback in analysisutil).
//   - Leading "./" is trimmed (rule-emitted relative paths).
//   - Trailing "/" is trimmed for non-recursive patterns; a recursive
//     pattern keeps its "..." suffix intact since matchExcludePattern
//     reads it.
//
// As a result "internal/foo", "/internal/foo", "./internal/foo",
// "internal\\foo", and "internal/foo/" are all the same key.
func NormalizeMatchPath(path string) string {
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
