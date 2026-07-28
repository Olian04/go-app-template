package main

import (
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// formatGoFiles gofmts every .go file under dir in place.
//
// Mode gates leave behind blank-line runs and misaligned struct tags: a
// `[[ if modeIs ... ]]` on its own line contributes a newline whether or not the
// branch is taken, and field alignment depends on which fields a mode kept.
// Formatting the rendered tree keeps generated projects gofmt-clean in every
// mode without forcing templates into unreadable trim markers.
func formatGoFiles(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("format read %s: %w", path, err)
		}
		out, err := format.Source(src)
		if err != nil {
			return fmt.Errorf("format %s: %w", path, err)
		}
		if len(out) == len(src) && string(out) == string(src) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("format stat %s: %w", path, err)
		}
		if err := os.WriteFile(path, out, info.Mode().Perm()); err != nil {
			return fmt.Errorf("format write %s: %w", path, err)
		}
		return nil
	})
}
