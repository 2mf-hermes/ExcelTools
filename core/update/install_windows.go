//go:build windows

package update

import (
	"os"
	"os/exec"
	"path/filepath"
)

// SwapAndRestart replaces the running executable with staged and, when restart
// is true, launches the replacement.
//
// Windows will not let a running executable be overwritten, but it does allow it
// to be renamed. Renaming the current binary out of the way and moving the new
// one into its place therefore achieves a self-update with no helper script and
// no shell interpreter involved — smaller attack surface than the usual
// write-a-batch-file-and-exec-it approach.
func SwapAndRestart(exePath, staged string, restart bool) error {
	if _, err := os.Stat(staged); err != nil {
		return errCode(CodePerm, "staged binary is missing: %v", err)
	}

	backup := exePath + ".old"
	os.Remove(backup)

	if err := os.Rename(exePath, backup); err != nil {
		return errCode(CodePerm, "cannot rename running executable: %v", err)
	}

	// Past this point exePath holds no executable, so every failure must restore
	// the backup. The user must never be left without a working install.
	if err := os.Rename(staged, exePath); err != nil {
		if rbErr := os.Rename(backup, exePath); rbErr != nil {
			return errCode(CodePerm,
				"cannot place new binary (%v) and rollback failed (%v); rename %s back to ExcelTools.exe manually",
				err, rbErr, backup)
		}
		return errCode(CodePerm, "cannot place new binary, previous version restored: %v", err)
	}

	if !restart {
		return nil
	}

	cmd := exec.Command(exePath)
	cmd.Dir = filepath.Dir(exePath)
	if err := cmd.Start(); err != nil {
		// The new binary is installed and verified; only the relaunch failed, so
		// tell the user to start it themselves rather than report a failed update.
		return errCode(CodeInternal, "update installed, but relaunch failed — start ExcelTools manually: %v", err)
	}
	// Detach: the child is an independent process and outlives us.
	_ = cmd.Process.Release()
	return nil
}
