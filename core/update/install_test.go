package update

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeExe writes a file named ExcelTools.exe in a fresh temp dir.
func makeExe(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ExeName)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return path
}

func TestIsReleaseBinary(t *testing.T) {
	exe := makeExe(t, "old")

	t.Run("accepts packaged exe", func(t *testing.T) {
		if !IsReleaseBinary(exe) {
			t.Errorf("IsReleaseBinary(%q) = false, want true", exe)
		}
	})

	t.Run("rejects empty and relative paths", func(t *testing.T) {
		// A relative path cannot be trusted: after a directory change it may
		// resolve somewhere else entirely.
		for _, p := range []string{"", ExeName, `.\ExcelTools.exe`, "sub/ExcelTools.exe"} {
			if IsReleaseBinary(p) {
				t.Errorf("IsReleaseBinary(%q) = true, want false", p)
			}
		}
	})

	t.Run("rejects other file names", func(t *testing.T) {
		other := filepath.Join(filepath.Dir(exe), "go-build123.exe")
		os.WriteFile(other, []byte("tmp"), 0o600)
		if IsReleaseBinary(other) {
			t.Errorf("IsReleaseBinary(%q) = true, want false", other)
		}
	})

	t.Run("rejects missing file", func(t *testing.T) {
		missing := filepath.Join(filepath.Dir(exe), "sub", ExeName)
		if IsReleaseBinary(missing) {
			t.Errorf("IsReleaseBinary(%q) = true for a nonexistent file", missing)
		}
	})
}

func TestSwapAndRestartReplacesBinary(t *testing.T) {
	exe := makeExe(t, "OLD")

	staged := filepath.Join(filepath.Dir(exe), "ExcelTools.new.exe")
	if err := os.WriteFile(staged, []byte("NEW"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// restart=false: swapping is what we are testing, not process spawning.
	if err := SwapAndRestart(exe, staged, false); err != nil {
		t.Fatalf("SwapAndRestart: %v", err)
	}

	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatalf("reading installed binary: %v", err)
	}
	if string(got) != "NEW" {
		t.Errorf("installed binary = %q, want %q", got, "NEW")
	}

	backup, err := os.ReadFile(exe + ".old")
	if err != nil {
		t.Fatalf("backup not kept: %v", err)
	}
	if string(backup) != "OLD" {
		t.Errorf("backup = %q, want %q", backup, "OLD")
	}

	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Errorf("staged file still present after swap")
	}
}

func TestSwapAndRestartRefusesMissingStaged(t *testing.T) {
	exe := makeExe(t, "OLD")
	missing := filepath.Join(filepath.Dir(exe), "ExcelTools.new.exe")

	err := SwapAndRestart(exe, missing, false)
	if CodeOf(err) != CodePerm {
		t.Fatalf("error = %v, want %s", err, CodePerm)
	}
	// The running binary must be untouched.
	got, _ := os.ReadFile(exe)
	if string(got) != "OLD" {
		t.Errorf("binary was modified despite refusal: %q", got)
	}
}

func TestSwapAndRestartOverwritesStaleBackup(t *testing.T) {
	exe := makeExe(t, "OLD")
	// Simulate a leftover backup from an earlier update.
	os.WriteFile(exe+".old", []byte("STALE"), 0o600)

	staged := filepath.Join(filepath.Dir(exe), "ExcelTools.new.exe")
	os.WriteFile(staged, []byte("NEW"), 0o600)

	if err := SwapAndRestart(exe, staged, false); err != nil {
		t.Fatalf("SwapAndRestart: %v", err)
	}
	backup, _ := os.ReadFile(exe + ".old")
	if string(backup) != "OLD" {
		t.Errorf("backup = %q, want the previous binary %q", backup, "OLD")
	}
}

func TestCleanupStaleBackup(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, ExeName)
	os.WriteFile(exe, []byte("cur"), 0o600)
	os.WriteFile(exe+".old", []byte("old"), 0o600)
	os.WriteFile(filepath.Join(dir, "ExcelTools.new.exe"), []byte("staged"), 0o600)
	os.WriteFile(filepath.Join(dir, "ExcelTools.download.zip"), []byte("zip"), 0o600)

	CleanupStaleBackup(exe)

	for _, name := range []string{exe + ".old", "ExcelTools.new.exe", "ExcelTools.download.zip"} {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			t.Errorf("%s was not cleaned up", filepath.Base(name))
		}
	}
	// The running binary itself must survive.
	if _, err := os.Stat(exe); err != nil {
		t.Errorf("CleanupStaleBackup removed the running binary: %v", err)
	}
}

func TestCleanupStaleBackupIgnoresBadPath(t *testing.T) {
	// Must never panic or touch anything for a relative/empty path.
	CleanupStaleBackup("")
	CleanupStaleBackup("ExcelTools.exe")
}

// writeZip builds a zip archive at path from name/content pairs.
func writeZip(t *testing.T, path string, entries [][2]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for _, e := range entries {
		w, err := zw.Create(e[0])
		if err != nil {
			t.Fatalf("zip entry: %v", err)
		}
		if _, err := w.Write([]byte(e[1])); err != nil {
			t.Fatalf("zip write: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
}

func TestExtractExeFromZip(t *testing.T) {
	t.Run("finds exe at archive root", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "a.zip")
		writeZip(t, zipPath, [][2]string{
			{"README.txt", "docs"},
			{ExeName, "BINARY"},
		})
		dest := filepath.Join(dir, "out.exe")
		if err := ExtractExeFromZip(zipPath, dest); err != nil {
			t.Fatalf("ExtractExeFromZip: %v", err)
		}
		got, _ := os.ReadFile(dest)
		if string(got) != "BINARY" {
			t.Errorf("extracted = %q, want %q", got, "BINARY")
		}
	})

	t.Run("finds exe in a subdirectory", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "a.zip")
		writeZip(t, zipPath, [][2]string{{"ExcelTools/" + ExeName, "NESTED"}})
		dest := filepath.Join(dir, "out.exe")
		if err := ExtractExeFromZip(zipPath, dest); err != nil {
			t.Fatalf("ExtractExeFromZip: %v", err)
		}
		got, _ := os.ReadFile(dest)
		if string(got) != "NESTED" {
			t.Errorf("extracted = %q, want %q", got, "NESTED")
		}
	})

	t.Run("traversal path cannot escape the destination", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "sub")
		os.Mkdir(sub, 0o700)
		zipPath := filepath.Join(sub, "a.zip")
		// The entry claims to live above its directory. We match on base name and
		// write to the caller's path, so the declared path is never used.
		writeZip(t, zipPath, [][2]string{{"../../escaped.exe", "PAYLOAD"}, {ExeName, "SAFE"}})

		dest := filepath.Join(sub, "out.exe")
		if err := ExtractExeFromZip(zipPath, dest); err != nil {
			t.Fatalf("ExtractExeFromZip: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dir, "escaped.exe")); !os.IsNotExist(err) {
			t.Errorf("archive entry escaped the destination directory")
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "escaped.exe")); !os.IsNotExist(err) {
			t.Errorf("archive entry escaped above the destination directory")
		}
	})

	t.Run("archive without the exe is refused", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "a.zip")
		writeZip(t, zipPath, [][2]string{{"README.txt", "docs"}})
		err := ExtractExeFromZip(zipPath, filepath.Join(dir, "out.exe"))
		if CodeOf(err) != CodeNoAsset {
			t.Errorf("error = %v, want %s", err, CodeNoAsset)
		}
	})

	t.Run("nonexistent archive is refused", func(t *testing.T) {
		dir := t.TempDir()
		err := ExtractExeFromZip(filepath.Join(dir, "missing.zip"), filepath.Join(dir, "out.exe"))
		if CodeOf(err) != CodeNoAsset {
			t.Errorf("error = %v, want %s", err, CodeNoAsset)
		}
	})
}

func TestInstallRefusesWithoutDigest(t *testing.T) {
	// Fail closed: no digest means no proof the bytes are the maintainer's, so
	// auto-install must be refused rather than silently trusting the download.
	c := NewChecker("0.4.0-m4")
	in := NewInstaller(c, makeExe(t, "OLD"))
	res := Result{
		Current:         "0.4.0-m4",
		Latest:          "v0.5.0-m5",
		UpdateAvailable: true,
		Asset:           Asset{Name: ExeName, BrowserDownloadURL: "https://github.com/x.exe"},
	}
	err := in.Install(t.Context(), res)
	if CodeOf(err) != CodeDigest {
		t.Fatalf("error = %v, want %s", err, CodeDigest)
	}
}

func TestInstallRefusesNonReleaseBuild(t *testing.T) {
	c := NewChecker("0.4.0-m4")
	in := NewInstaller(c, filepath.Join(t.TempDir(), "go-build1234.exe"))
	res := Result{
		UpdateAvailable: true,
		Latest:          "v0.5.0-m5",
		Asset: Asset{
			Name:               ExeName,
			BrowserDownloadURL: "https://github.com/x.exe",
			Digest:             "sha256:" + strings.Repeat("ab", 32),
		},
	}
	err := in.Install(t.Context(), res)
	if CodeOf(err) != CodeNotRelease {
		t.Fatalf("error = %v, want %s", err, CodeNotRelease)
	}
}

func TestInstallRefusesWhenNoUpdate(t *testing.T) {
	c := NewChecker("0.5.0-m5")
	in := NewInstaller(c, makeExe(t, "OLD"))
	if err := in.Install(t.Context(), Result{UpdateAvailable: false}); CodeOf(err) != CodeInternal {
		t.Fatalf("error = %v, want %s", err, CodeInternal)
	}
}

func TestInstallRefusesMissingAsset(t *testing.T) {
	c := NewChecker("0.4.0-m4")
	in := NewInstaller(c, makeExe(t, "OLD"))
	err := in.Install(t.Context(), Result{UpdateAvailable: true, Latest: "v0.5.0-m5"})
	if CodeOf(err) != CodeNoAsset {
		t.Fatalf("error = %v, want %s", err, CodeNoAsset)
	}
}
