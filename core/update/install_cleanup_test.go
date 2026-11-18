package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The .old file succeeds only on the third attempt, as if a lock was released
// while the retry loop was already running.
func TestRetryRemoveStaleFilesRetriesUntilGone(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "ExcelTools.exe")
	if err := os.WriteFile(exe+".old", []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	attempts := 0
	remove := func(path string) error {
		if filepath.Ext(path) == ".old" {
			attempts++
			if attempts < 3 {
				return os.ErrPermission
			}
		}
		return os.Remove(path)
	}
	var sleeps []time.Duration
	retryRemoveStaleFiles(exe, 500*time.Millisecond, time.Minute,
		func(d time.Duration) { sleeps = append(sleeps, d) }, remove)

	if _, err := os.Stat(exe + ".old"); !os.IsNotExist(err) {
		t.Errorf(".old still exists after retry loop: %v", err)
	}
	if len(sleeps) != 2 {
		t.Fatalf("sleeps = %v, want 2 (one between each retry)", sleeps)
	}
	for _, d := range sleeps {
		if d != 500*time.Millisecond {
			t.Errorf("sleep = %v, want the retry interval", d)
		}
	}
}

// A file that stays locked must survive: the loop gives up at the timeout
// instead of spinning forever or deleting something it could not.
func TestRetryRemoveStaleFilesTimesOut(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "ExcelTools.exe")
	old := exe + ".old"
	if err := os.WriteFile(old, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	stuck := func(string) error { return os.ErrPermission }

	start := time.Now()
	retryRemoveStaleFiles(exe, 10*time.Millisecond, 80*time.Millisecond, time.Sleep, stuck)

	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("loop ran %v, want about the 80ms timeout", elapsed)
	}
	if _, err := os.Stat(old); err != nil {
		t.Errorf("stale file vanished despite failing removals: %v", err)
	}
}

// CleanupStaleBackupRetrying on an unlocked leftover removes it at once and
// does not spawn needless work.
func TestCleanupStaleBackupRetryingRemovesPromptly(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "ExcelTools.exe")
	old := exe + ".old"
	if err := os.WriteFile(old, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	CleanupStaleBackupRetrying(exe)

	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(old); os.IsNotExist(err) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal(".old not removed within 3s")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Bad paths stay no-ops, including in the retrying variant.
func TestCleanupStaleBackupRetryingIgnoresBadPath(t *testing.T) {
	CleanupStaleBackupRetrying("")
	CleanupStaleBackupRetrying("ExcelTools.exe")
}
