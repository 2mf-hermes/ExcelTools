package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUniquePath(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "out.xlsx")
	got := UniquePath(p)
	if got != p {
		t.Fatalf("expected %s, got %s", p, got)
	}
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got = UniquePath(p)
	if got != filepath.Join(dir, "out_2.xlsx") {
		t.Fatalf("expected out_2.xlsx, got %s", got)
	}
}

func TestNaturalLess(t *testing.T) {
	if !NaturalLess("file2.xlsx", "file10.xlsx") {
		t.Fatal("file2 should sort before file10")
	}
	if NaturalLess("file10.xlsx", "file2.xlsx") {
		t.Fatal("file10 should not sort before file2")
	}
}

func TestDirWritable(t *testing.T) {
	dir := t.TempDir()
	if !DirWritable(dir) {
		t.Fatal("temp dir should be writable")
	}
	if DirWritable(filepath.Join(dir, "missing")) {
		t.Fatal("missing dir should not be writable")
	}
}

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.bin")
	if err := WriteFileAtomic(p, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil || string(b) != "hello" {
		t.Fatalf("read back failed: %v %q", err, b)
	}
}
