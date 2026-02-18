package filesafe

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const defaultFilePerm os.FileMode = 0o644

// WriteAtomic writes content to path using same-directory temp file + rename.
func WriteAtomic(path string, content []byte, perm os.FileMode) error {
	if path == "" {
		return errors.New("path is required")
	}
	if perm == 0 {
		perm = defaultFilePerm
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".pdfmeta-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	cleanup = false

	if err := syncDir(dir); err != nil {
		return fmt.Errorf("sync parent dir: %w", err)
	}
	return nil
}

// ReplaceAtomic promotes an existing staged file into target path atomically.
// stagedPath must live in the same directory as targetPath.
func ReplaceAtomic(targetPath, stagedPath string) error {
	if targetPath == "" || stagedPath == "" {
		return errors.New("target and staged paths are required")
	}
	targetDir := filepath.Clean(filepath.Dir(targetPath))
	stagedDir := filepath.Clean(filepath.Dir(stagedPath))
	if targetDir != stagedDir {
		return errors.New("staged file must be in the same directory as target")
	}
	if err := os.Rename(stagedPath, targetPath); err != nil {
		return fmt.Errorf("rename staged file: %w", err)
	}
	if err := syncDir(targetDir); err != nil {
		return fmt.Errorf("sync parent dir: %w", err)
	}
	return nil
}

// CopyFile copies src to dst, replacing dst if it exists.
func CopyFile(srcPath, dstPath string, perm os.FileMode) error {
	if srcPath == "" || dstPath == "" {
		return errors.New("source and destination paths are required")
	}
	if perm == 0 {
		perm = defaultFilePerm
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer src.Close()

	if err := WriteAtomicFromReader(dstPath, src, perm); err != nil {
		return err
	}
	return nil
}

// WriteAtomicFromReader writes reader data atomically into path.
func WriteAtomicFromReader(path string, reader io.Reader, perm os.FileMode) error {
	if path == "" {
		return errors.New("path is required")
	}
	if reader == nil {
		return errors.New("reader is required")
	}
	if perm == 0 {
		perm = defaultFilePerm
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".pdfmeta-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if _, err := io.Copy(tmp, reader); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	cleanup = false

	if err := syncDir(dir); err != nil {
		return fmt.Errorf("sync parent dir: %w", err)
	}
	return nil
}

func syncDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
