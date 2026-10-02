package update

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Installer runs the download → verify → replace sequence.
type Installer struct {
	Checker *Checker

	// ExePath is the absolute path of the running executable.
	ExePath string

	// Restart relaunches the app once the swap succeeds.
	Restart bool
}

// NewInstaller returns an installer for the given executable, restarting by
// default.
func NewInstaller(c *Checker, exePath string) *Installer {
	return &Installer{Checker: c, ExePath: exePath, Restart: true}
}

// Install downloads, verifies, and swaps in the new binary.
//
// On success the caller must terminate the process so the replacement can take
// over. Guards run before any network or filesystem work so an ineligible
// install is refused cheaply and without side effects.
func (in *Installer) Install(ctx context.Context, res Result) error {
	if !res.UpdateAvailable {
		return errCode(CodeInternal, "no update to install")
	}
	if res.Asset.Name == "" {
		return errCode(CodeNoAsset, "release %s has no installable asset", res.Latest)
	}
	// Fail closed: without a published digest there is no way to prove the bytes
	// match what the maintainer built, so we refuse to overwrite the binary and
	// the UI offers the release page instead.
	if !HasUsableDigest(res.Asset) {
		return errCode(CodeDigest,
			"release %s publishes no sha256 digest; automatic install is disabled", res.Latest)
	}
	if !IsReleaseBinary(in.ExePath) {
		return errCode(CodeNotRelease,
			"running binary %q is not a packaged release build", filepath.Base(in.ExePath))
	}

	dir := filepath.Dir(in.ExePath)
	staged := filepath.Join(dir, "ExcelTools.new.exe")
	os.Remove(staged) // clear a previous attempt

	if res.Asset.IsZip() {
		archive := filepath.Join(dir, "ExcelTools.download.zip")
		os.Remove(archive)
		defer os.Remove(archive)

		if _, err := in.Checker.Download(ctx, res.Asset.BrowserDownloadURL, archive, res.Asset.Digest); err != nil {
			return err
		}
		if err := ExtractExeFromZip(archive, staged); err != nil {
			os.Remove(staged)
			return err
		}
	} else {
		if _, err := in.Checker.Download(ctx, res.Asset.BrowserDownloadURL, staged, res.Asset.Digest); err != nil {
			return err
		}
	}

	if err := SwapAndRestart(in.ExePath, staged, in.Restart); err != nil {
		os.Remove(staged)
		return err
	}
	return nil
}

// IsReleaseBinary reports whether path is a packaged ExcelTools build rather
// than the temporary binary produced by `go run` or `wails dev`. Self-replace is
// meaningless for those and would clobber a developer's toolchain output.
func IsReleaseBinary(path string) bool {
	if path == "" || !filepath.IsAbs(path) {
		return false
	}
	if !strings.EqualFold(filepath.Base(path), ExeName) {
		return false
	}
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// CleanupStaleBackup removes the .old binary and any staged leftovers from a
// previous update. Call once at startup; failures are intentionally ignored
// because a still-locked file simply gets cleaned up on the next launch.
func CleanupStaleBackup(exePath string) {
	if exePath == "" || !filepath.IsAbs(exePath) {
		return
	}
	os.Remove(exePath + ".old")
	dir := filepath.Dir(exePath)
	os.Remove(filepath.Join(dir, "ExcelTools.new.exe"))
	os.Remove(filepath.Join(dir, "ExcelTools.download.zip"))
}
