package lrcrules

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/HexmosTech/git-lrc/storage"
)

// CollectZipExtras walks .lrc/ under repoRoot (if present) and returns its
// files as a map suitable for reviewapi.CreateZipArchiveWithExtras, keyed
// by repo-relative path (e.g. ".lrc/rules/security.md") with "/"
// separators. Returns a nil map (no error) when .lrc/ does not exist.
//
// A file or subdirectory that can't be read (e.g. a permission error) is
// skipped and reported via the returned warnings slice rather than
// aborting the whole walk, so a single unreadable entry doesn't drop all
// Repository Rules from the review bundle.
func CollectZipExtras(repoRoot string) (map[string][]byte, []string, error) {
	lrcDir, ok, err := Load(repoRoot)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, nil
	}

	extras := map[string][]byte{}
	var warnings []string
	walkErr := filepath.WalkDir(lrcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skipping %s: %v", path, err))
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skipping %s: failed to compute relative path: %v", path, err))
			return nil
		}
		content, err := storage.ReadFile(path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skipping %s: failed to read file: %v", relPath, err))
			return nil
		}
		extras[filepath.ToSlash(relPath)] = content
		return nil
	})
	if walkErr != nil {
		return extras, warnings, walkErr
	}

	return extras, warnings, nil
}
