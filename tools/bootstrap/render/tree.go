package render

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Tree walks templatesDir → outDir; expands path templates; duplicate out path → error.
// Skipped gated files are omitted. Empty directories are not created / are pruned.
func Tree(templatesDir, outDir string, ctx Context) error {
	seen := make(map[string]string) // rel out path → source rel path

	err := filepath.WalkDir(templatesDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(templatesDir, path)
		if err != nil {
			return fmt.Errorf("render tree: %w", err)
		}
		if rel == "." {
			return nil
		}

		outRel, err := expandPath(rel, ctx)
		if err != nil {
			return fmt.Errorf("render tree path %s: %w", rel, err)
		}
		outRel = filepath.Clean(outRel)
		if outRel == "." || strings.HasPrefix(outRel, "..") {
			return fmt.Errorf("render tree: invalid out path %q from %s", outRel, rel)
		}

		if prev, ok := seen[outRel]; ok {
			return fmt.Errorf("render tree: duplicate out path %q from %q and %q", outRel, prev, rel)
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("render tree read %s: %w", rel, err)
		}

		out, skipped, err := File(rel, src, ctx)
		if err != nil {
			return err
		}
		if skipped {
			return nil
		}

		seen[outRel] = rel
		dest := filepath.Join(outDir, outRel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("render tree mkdir %s: %w", outRel, err)
		}
		if err := os.WriteFile(dest, out, 0o644); err != nil {
			return fmt.Errorf("render tree write %s: %w", outRel, err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return pruneEmptyDirs(outDir)
}

func expandPath(rel string, ctx Context) (string, error) {
	if !strings.Contains(rel, "[[") {
		return rel, nil
	}
	tmpl, err := template.New("path").Delims("[[", "]]").Parse(rel)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func pruneEmptyDirs(root string) error {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() && path != root {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	// deepest first
	for i := len(dirs) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(dirs[i])
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if len(entries) == 0 {
			if err := os.Remove(dirs[i]); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}
