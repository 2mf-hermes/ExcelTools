package update

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// Time bounds for removing leftovers that are still locked by the previous
// process (or an antivirus scan) right after an update.
const (
	cleanupRetryInterval = 500 * time.Millisecond
	cleanupRetryTimeout  = 15 * time.Second
)

// staleFiles returns the artifacts a previous update may have left behind.
func staleFiles(exePath string) []string {
	dir := filepath.Dir(exePath)
	return []string{
		exePath + ".old",
		filepath.Join(dir, "ExcelTools.new.exe"),
		filepath.Join(dir, "ExcelTools.download.zip"),
	}
}

// hasStaleFiles reports whether any leftover artifact still exists.
func hasStaleFiles(exePath string) bool {
	for _, f := range staleFiles(exePath) {
		if _, err := os.Stat(f); err == nil {
			return true
		}
	}
	return false
}

// CleanupStaleBackup removes the .old binary and any staged leftovers from a
// previous update. Call once at startup; failures are intentionally ignored
// because a still-locked file simply gets cleaned up on the next launch.
func CleanupStaleBackup(exePath string) {
	if exePath == "" || !filepath.IsAbs(exePath) {
		return
	}
	for _, f := range staleFiles(exePath) {
		os.Remove(f)
	}
}

// CleanupStaleBackupRetrying behaves like CleanupStaleBackup and, when a file
// is still locked, keeps retrying in the background: right after an update the
// previous process may not have exited yet when the updated app starts, so its
// .old binary can not be deleted on the first attempt. The retry runs every
// 500ms for up to 15s and deletes a file as soon as it is unlocked, so the
// leftover disappears without a manual relaunch.
func CleanupStaleBackupRetrying(exePath string) {
	CleanupStaleBackup(exePath)
	if exePath == "" || !filepath.IsAbs(exePath) || !hasStaleFiles(exePath) {
		return
	}
	go retryRemoveStaleFiles(exePath, cleanupRetryInterval, cleanupRetryTimeout, time.Sleep, os.Remove)
}

// retryRemoveStaleFiles keeps attempting removal until every stale file is
// gone or the timeout has elapsed. sleep and remove are parameters so tests
// can drive the loop without real clocks or file locks.
func retryRemoveStaleFiles(exePath string, interval, timeout time.Duration, sleep func(time.Duration), remove func(string) error) {
	deadline := time.Now().Add(timeout)
	for {
		gone := true
		for _, f := range staleFiles(exePath) {
			if err := remove(f); err != nil && !os.IsNotExist(err) {
				gone = false
			}
		}
		if gone || time.Now().After(deadline) {
			return
		}
		sleep(interval)
	}
}

// removePreviousBackup deletes a leftover backup before the swap renames the
// running binary onto that name. A file held open cannot be removed or
// replaced on Windows, so retry briefly before giving up and blocking the
// update with "cannot rename running executable".
func removePreviousBackup(path string) {
	for i := 0; i < 30; i++ {
		if err := os.Remove(path); err == nil || os.IsNotExist(err) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
