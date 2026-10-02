package main

import (
	"context"
	"fmt"
	"runtime"

	"ExcelTools/core/model"
	"ExcelTools/core/report"
	"ExcelTools/platform/fs"
)

// App is the Wails-bound application façade.
type App struct {
	ctx      context.Context
	settings model.Settings
	sheet    sheetJob
	fileJob  fileJob
	upd      updater
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{settings: loadSettings()}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.settings = loadSettings()
}

// appVersion is the single source of truth for displayed version.
const appVersion = "0.5.0-m5"

// Health returns a simple liveness payload for the UI shell.
func (a *App) Health() map[string]string {
	return map[string]string{
		"status":  "ok",
		"version": appVersion,
	}
}

// GetAppInfo returns static app metadata.
func (a *App) GetAppInfo() model.AppInfo {
	return model.AppInfo{
		Name:        "ExcelTools",
		Version:     appVersion,
		Platform:    runtime.GOOS,
		LocalOnly:   true,
		UpdateCheck: true,
	}
}

// GetSettings returns current local preferences (disk-backed).
func (a *App) GetSettings() model.Settings {
	a.settings = loadSettings()
	return a.settings
}

// SetSettings replaces and persists local preferences.
func (a *App) SetSettings(s model.Settings) (model.Settings, error) {
	if s.Language == "" {
		s.Language = "system"
	}
	if s.Theme == "" {
		s.Theme = "system"
	}
	if s.Defaults == nil {
		s.Defaults = model.DefaultSettings().Defaults
	}
	if err := saveSettings(s); err != nil {
		return a.settings, fmt.Errorf("E_PERM: cannot save settings: %w", err)
	}
	a.settings = s
	return a.settings, nil
}

// ResetSettings restores factory defaults.
func (a *App) ResetSettings() (model.Settings, error) {
	s, err := resetSettings()
	if err != nil {
		return a.settings, err
	}
	a.settings = s
	return s, nil
}

// Ping is a trivial IPC check used by the home skeleton.
func (a *App) Ping() string {
	return "pong"
}

// DirWritable is a thin platform probe for output-path UI.
func (a *App) DirWritable(dir string) bool {
	return fs.DirWritable(dir)
}

// ReportDemo returns an empty report shape for Result view typing.
func (a *App) ReportDemo() model.MergeResult {
	b := report.New()
	ok, _, _ := b.Counts()
	return model.MergeResult{
		Success:   false,
		Partial:   b.Partial(),
		Message:   "merge engines not implemented in M0/M1",
		FilesUsed: ok,
		Items:     b.Items(),
	}
}
