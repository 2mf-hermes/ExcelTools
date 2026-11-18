//go:build windows

package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCleanupRetryingRemovesLockedOld reproduces the post-update condition the
// fix targets: the previous process still holds the .old binary open when the
// updated app starts, so the first removal fails. The background retry must
// delete the file as soon as the handle is released — without a second manual
// launch.
func TestCleanupRetryingRemovesLockedOld(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "ExcelTools.exe")
	old := exe + ".old"
	if err := os.WriteFile(old, []byte("old binary"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Go opens files without FILE_SHARE_DELETE on Windows, so removing while
	// this handle is open fails exactly like it does for the running previous
	// process.
	f, err := os.Open(old)
	if err != nil {
		t.Fatal(err)
	}
	CleanupStaleBackupRetrying(exe)

	// Release the lock a moment later, as the exiting previous process would.
	time.AfterFunc(120*time.Millisecond, func() { f.Close() })

	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(old); os.IsNotExist(err) {
			return // deleted without any further action
		}
		if time.Now().After(deadline) {
			t.Fatal(".old still present after the lock was released")
		}
		time.Sleep(20 * time.Millisecond)
	}
}
