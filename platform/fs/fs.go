// Package fs wraps filesystem helpers used by merge (paths, locks, atomic write).
package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// UniquePath returns destPath, or destPath with _2/_3… if it already exists.
func UniquePath(destPath string) string {
	if _, err := os.Stat(destPath); err != nil {
		return destPath
	}
	ext := filepath.Ext(destPath)
	base := strings.TrimSuffix(destPath, ext)
	for i := 2; i < 1000; i++ {
		cand := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(cand); err != nil {
			return cand
		}
	}
	return fmt.Sprintf("%s_%d%s", base, os.Getpid(), ext)
}

// TempPathIn creates an empty temp file path in dir with the given pattern.
// Caller owns cleanup via RemoveIfExists or successful RenameOver.
func TempPathIn(dir, pattern string) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return "", err
	}
	_ = os.Remove(name)
	return name, nil
}

// RemoveIfExists deletes path, ignoring not-exist errors.
func RemoveIfExists(path string) error {
	err := os.Remove(path)
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

// RenameOver moves src to dest, replacing dest if needed (Windows-friendly).
func RenameOver(src, dest string) error {
	if err := os.Rename(src, dest); err == nil {
		return nil
	}
	// Windows may refuse rename onto existing file.
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(src, dest)
}

// WriteFileAtomic writes data to destPath via a same-directory temp file then rename.
func WriteFileAtomic(destPath string, data []byte) error {
	dir := filepath.Dir(destPath)
	tmp, err := os.CreateTemp(dir, ".tmp-exceltools-*")
	if err != nil {
		return fmt.Errorf("E_PERM: create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("E_PERM: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("E_PERM: close temp: %w", err)
	}
	if err := os.Rename(tmpName, destPath); err != nil {
		return fmt.Errorf("E_PERM: rename output: %w", err)
	}
	return nil
}

// NaturalLess compares names so "file2" sorts before "file10".
func NaturalLess(a, b string) bool {
	na, nb := filepath.Base(a), filepath.Base(b)
	i, j := 0, 0
	for i < len(na) && j < len(nb) {
		ca, cb := na[i], nb[j]
		if isDigit(ca) && isDigit(cb) {
			si, sj := i, j
			for i < len(na) && isDigit(na[i]) {
				i++
			}
			for j < len(nb) && isDigit(nb[j]) {
				j++
			}
			va, _ := strconv.Atoi(na[si:i])
			vb, _ := strconv.Atoi(nb[sj:j])
			if va != vb {
				return va < vb
			}
			continue
		}
		la, lb := lower(ca), lower(cb)
		if la != lb {
			return la < lb
		}
		i++
		j++
	}
	return len(na) < len(nb)
}

// SortNatural sorts paths by natural filename order.
func SortNatural(paths []string) {
	sort.SliceStable(paths, func(i, j int) bool {
		return NaturalLess(paths[i], paths[j])
	})
}

// IsReadableFile reports whether path exists and is a regular file.
func IsReadableFile(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Mode().IsRegular()
}

// DirWritable reports whether dir can accept a new file.
func DirWritable(dir string) bool {
	if dir == "" {
		return false
	}
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		return false
	}
	f, err := os.CreateTemp(dir, ".exceltools-write-test-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func lower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}
