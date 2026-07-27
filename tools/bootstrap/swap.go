package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// Paths kept at repo root while staging is swapped in (bootstrap machinery until wipe).
var preserveDuringSwap = map[string]struct{}{
	".git":           {},
	".bootstrap-out": {},
	"templates":      {},
	"tools":          {},
	"bootstrap.sh":   {},
}

func swapStaging(root, staging string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("swap read root: %w", err)
	}
	for _, e := range entries {
		name := e.Name()
		if _, ok := preserveDuringSwap[name]; ok {
			continue
		}
		path := filepath.Join(root, name)
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("swap remove %s: %w", name, err)
		}
	}

	if err := copyTree(staging, root); err != nil {
		return fmt.Errorf("swap copy: %w", err)
	}
	if err := os.RemoveAll(staging); err != nil {
		return fmt.Errorf("swap remove staging: %w", err)
	}
	return nil
}

func tidyModule(root string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = root
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	return nil
}

func wipeBootstrap(root string) error {
	targets := []string{
		filepath.Join(root, "templates"),
		filepath.Join(root, "tools", "bootstrap"),
		filepath.Join(root, "bootstrap.sh"),
		filepath.Join(root, ".bootstrap-out"),
	}
	for _, path := range targets {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("wipe %s: %w", path, err)
		}
	}
	toolsDir := filepath.Join(root, "tools")
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("wipe read tools: %w", err)
	}
	if len(entries) == 0 {
		if err := os.Remove(toolsDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("wipe empty tools: %w", err)
		}
	}
	return nil
}

func copyTree(srcRoot, dstRoot string) error {
	return filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dst := filepath.Join(dstRoot, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return copyFile(path, dst)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
