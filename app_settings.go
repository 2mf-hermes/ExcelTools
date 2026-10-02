package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"ExcelTools/core/model"
)

// settingsPath returns the AppData settings file path.
func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(dir, "ExcelTools")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(root, "settings.json"), nil
}

// loadSettings reads settings from disk, falling back to defaults.
func loadSettings() model.Settings {
	def := model.DefaultSettings()
	path, err := settingsPath()
	if err != nil {
		return def
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return def
	}
	var s model.Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return def
	}
	// A missing key is not the same as an explicit false: a settings file written
	// before the update feature existed must keep the check enabled.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		s.AutoCheckUpdates = def.AutoCheckUpdates
	} else if _, ok := raw["autoCheckUpdates"]; !ok {
		s.AutoCheckUpdates = def.AutoCheckUpdates
	}
	if s.Language == "" {
		s.Language = def.Language
	}
	if s.Theme == "" {
		s.Theme = def.Theme
	}
	if s.Defaults == nil {
		s.Defaults = def.Defaults
	} else {
		for k, v := range def.Defaults {
			if _, ok := s.Defaults[k]; !ok {
				s.Defaults[k] = v
			}
		}
	}
	return s
}

// saveSettings writes settings to disk.
func saveSettings(s model.Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// resetSettings restores factory defaults on disk and returns them.
func resetSettings() (model.Settings, error) {
	s := model.DefaultSettings()
	if err := saveSettings(s); err != nil {
		return s, err
	}
	return s, nil
}
