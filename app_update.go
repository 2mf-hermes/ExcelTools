package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"ExcelTools/core/model"
	"ExcelTools/core/update"
)

const (
	// updateEventProgress carries download progress to the UI.
	updateEventProgress = "update:progress"

	// checkTimeout bounds a version check so an unreachable network cannot hang
	// the settings screen.
	checkTimeout = 15 * time.Second

	// installTimeout bounds the whole download-and-install sequence.
	installTimeout = 10 * time.Minute

	// quitDelay lets the UI paint its "restarting" notice before this process
	// exits in favour of the freshly installed one.
	quitDelay = 1500 * time.Millisecond

	// releasesPage is the manual-download fallback.
	releasesPage = "https://github.com/2mf-hermes/ExcelTools/releases/latest"
)

// updater holds update state for this session.
type updater struct {
	mu   sync.Mutex
	last update.Result
	busy bool
}

// runtimeCtx returns a usable context even before Wails startup has run.
func (a *App) runtimeCtx() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

// runningExe returns the absolute path of the running executable, or "" when it
// cannot be resolved. An empty result disables automatic installation only — the
// version check still works.
func runningExe() string {
	p, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return p
}

// CleanupStaleUpdateFiles removes the previous binary's backup and any staged
// download left behind by an interrupted update. Called once at startup.
func CleanupStaleUpdateFiles() {
	update.CleanupStaleBackup(runningExe())
}

// CheckForUpdates queries the release channel and reports the result.
//
// A failure is returned as status "error" carrying a message rather than a Go
// error, because the settings screen renders it inline.
func (a *App) CheckForUpdates() model.UpdateCheckResult {
	return a.runCheck(a.runtimeCtx())
}

// CheckForUpdatesOnStartup runs the automatic launch-time check.
//
// It stays silent when the check is disabled or fails: an offline launch must
// not greet the user with an error before they have asked for anything. The UI
// only reacts to status "available".
func (a *App) CheckForUpdatesOnStartup() model.UpdateCheckResult {
	if !loadSettings().AutoCheckUpdates {
		return model.UpdateCheckResult{
			CurrentVersion: appVersion,
			LatestVersion:  appVersion,
			Status:         model.UpdateSkipped,
			Message:        "auto-check disabled",
		}
	}
	res := a.runCheck(a.runtimeCtx())
	if res.Status == model.UpdateError {
		return model.UpdateCheckResult{
			CurrentVersion: appVersion,
			LatestVersion:  appVersion,
			Status:         model.UpdateSkipped,
			Message:        res.Message,
		}
	}
	return res
}

// runCheck performs a check and records the result for a later install.
func (a *App) runCheck(parent context.Context) model.UpdateCheckResult {
	ctx, cancel := context.WithTimeout(parent, checkTimeout)
	defer cancel()

	res, err := update.NewChecker(appVersion).Check(ctx)
	if err != nil {
		return model.UpdateCheckResult{
			CurrentVersion: appVersion,
			LatestVersion:  appVersion,
			Status:         model.UpdateError,
			Message:        err.Error(),
		}
	}

	a.upd.mu.Lock()
	a.upd.last = res
	a.upd.mu.Unlock()

	out := model.UpdateCheckResult{
		CurrentVersion: appVersion,
		LatestVersion:  res.Latest,
		ReleaseNotes:   res.Release.Body,
		ReleaseURL:     res.Release.HTMLURL,
		AssetName:      res.Asset.Name,
		AssetSize:      res.Asset.Size,
	}
	switch {
	case !res.UpdateAvailable:
		out.Status = model.UpdateUpToDate
		out.LatestVersion = appVersion
		out.Message = "up to date"
	case !update.HasUsableDigest(res.Asset):
		// A newer release exists but publishes no digest we can verify against.
		// Auto-install is refused, so the UI offers the download page instead.
		out.Status = model.UpdateAvailable
		out.CanAutoInstall = false
		out.Message = "no verifiable digest published"
	default:
		out.Status = model.UpdateAvailable
		out.CanAutoInstall = true
	}
	return out
}

// InstallUpdate downloads the pending release, verifies its digest, replaces the
// running executable, and relaunches.
//
// This process exits shortly after a successful install — the replacement has
// already been launched, so the UI's last job is to show the restart notice.
func (a *App) InstallUpdate() (model.UpdateInstallResult, error) {
	a.upd.mu.Lock()
	if a.upd.busy {
		a.upd.mu.Unlock()
		return model.UpdateInstallResult{Status: "error", Message: "an update is already in progress"}, nil
	}
	pending := a.upd.last
	a.upd.busy = true
	a.upd.mu.Unlock()

	defer func() {
		a.upd.mu.Lock()
		a.upd.busy = false
		a.upd.mu.Unlock()
	}()

	if !pending.UpdateAvailable {
		return model.UpdateInstallResult{
			Status:  "error",
			Message: "no update is pending; check for updates first",
		}, nil
	}

	c := update.NewChecker(appVersion)
	c.Progress = a.emitProgress

	ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
	defer cancel()

	if err := update.NewInstaller(c, runningExe()).Install(ctx, pending); err != nil {
		return model.UpdateInstallResult{
			Status:  "error",
			Code:    update.CodeOf(err),
			Message: err.Error(),
		}, nil
	}

	a.upd.mu.Lock()
	a.upd.last = update.Result{}
	a.upd.mu.Unlock()

	// The replacement process is already running; step aside so that it is the
	// only instance left.
	go func() {
		time.Sleep(quitDelay)
		wailsruntime.Quit(a.runtimeCtx())
	}()

	return model.UpdateInstallResult{
		Status:  "installed",
		Message: "updated to " + pending.Latest,
	}, nil
}

// emitProgress forwards download progress to the UI.
func (a *App) emitProgress(done, total int64) {
	pct := 0
	if total > 0 {
		pct = int(done * 100 / total)
	}
	wailsruntime.EventsEmit(a.runtimeCtx(), updateEventProgress, model.UpdateProgress{
		Done:  done,
		Total: total,
		Pct:   pct,
	})
}

// OpenReleasePage opens the project's release page in the default browser. It is
// the fallback whenever automatic installation is unavailable.
func (a *App) OpenReleasePage(url string) error {
	if url == "" {
		url = releasesPage
	}
	// Restrict navigation to the project's own pages so a crafted release payload
	// cannot turn this into an open-anything primitive.
	if !strings.HasPrefix(url, "https://github.com/2mf-hermes/ExcelTools/") {
		return fmt.Errorf("E_INSECURE: refusing to open %q", url)
	}
	wailsruntime.BrowserOpenURL(a.runtimeCtx(), url)
	return nil
}
