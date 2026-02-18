package filesafe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteAtomicCreatesAndReplaces(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	if err := WriteAtomic(path, []byte("first"), 0o600); err != nil {
		t.Fatalf("WriteAtomic first: %v", err)
	}
	if got := readFile(t, path); got != "first" {
		t.Fatalf("unexpected content after first write: %q", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perm = %o want %o", info.Mode().Perm(), 0o600)
	}

	if err := WriteAtomic(path, []byte("second"), 0o644); err != nil {
		t.Fatalf("WriteAtomic replace: %v", err)
	}
	if got := readFile(t, path); got != "second" {
		t.Fatalf("unexpected content after replace: %q", got)
	}
}

func TestWriteAtomicFromReader(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "from_reader.txt")

	if err := WriteAtomicFromReader(path, strings.NewReader("hello"), 0); err != nil {
		t.Fatalf("WriteAtomicFromReader: %v", err)
	}
	if got := readFile(t, path); got != "hello" {
		t.Fatalf("unexpected content: %q", got)
	}
}

func TestReplaceAtomic(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	staged := filepath.Join(dir, "staged.txt")

	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	if err := os.WriteFile(staged, []byte("new"), 0o644); err != nil {
		t.Fatalf("seed staged: %v", err)
	}

	if err := ReplaceAtomic(target, staged); err != nil {
		t.Fatalf("ReplaceAtomic: %v", err)
	}
	if got := readFile(t, target); got != "new" {
		t.Fatalf("unexpected content: %q", got)
	}
	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Fatalf("staged file should be moved, got err=%v", err)
	}
}

func TestReplaceAtomicRejectsDifferentDir(t *testing.T) {
	t.Parallel()
	a := t.TempDir()
	b := t.TempDir()
	err := ReplaceAtomic(filepath.Join(a, "target.txt"), filepath.Join(b, "staged.txt"))
	if err == nil {
		t.Fatalf("expected error for different directories")
	}
}

func TestCopyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	if err := os.WriteFile(src, []byte("copied"), 0o644); err != nil {
		t.Fatalf("seed src: %v", err)
	}
	if err := CopyFile(src, dst, 0o640); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}
	if got := readFile(t, dst); got != "copied" {
		t.Fatalf("unexpected copy content: %q", got)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("perm = %o want %o", info.Mode().Perm(), 0o640)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %q: %v", path, err)
	}
	return string(b)
}
