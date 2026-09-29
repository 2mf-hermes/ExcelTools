package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSaveSettingsRoundTrip(t *testing.T) {
	s := loadSettings()
	if s.Language == "" {
		t.Fatal("default language empty")
	}
	s.Theme = "dark"
	s.Language = "en"
	if err := saveSettings(s); err != nil {
		t.Fatal(err)
	}
	got := loadSettings()
	if got.Theme != "dark" || got.Language != "en" {
		t.Fatalf("round trip failed: %+v", got)
	}
	if _, err := resetSettings(); err != nil {
		t.Fatal(err)
	}
	got = loadSettings()
	if got.Theme != "system" {
		t.Fatalf("reset failed: %+v", got)
	}
}

func TestClassifyPaths(t *testing.T) {
	a := &App{settings: loadSettings()}
	dir := t.TempDir()

	write := func(name string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	ok := write("a.xlsx")
	lock := write("~$a.xlsx")
	bad := write("c.txt")
	out := write("合併結果_x.xlsx")

	refs := a.ClassifyPaths([]string{ok, lock, bad, out, ok})
	if len(refs) != 4 {
		t.Fatalf("expected 4 unique refs, got %d", len(refs))
	}
	byName := map[string]string{}
	for _, r := range refs {
		byName[r.Name] = string(r.Status)
	}
	if byName["a.xlsx"] != "ok" {
		t.Fatalf("a.xlsx status %q", byName["a.xlsx"])
	}
	if byName["~$a.xlsx"] != "skipped" {
		t.Fatalf("lock status %q", byName["~$a.xlsx"])
	}
	if byName["c.txt"] != "failed" {
		t.Fatalf("txt status %q", byName["c.txt"])
	}
	if byName["合併結果_x.xlsx"] != "skipped" {
		t.Fatalf("prior output status %q", byName["合併結果_x.xlsx"])
	}
	_ = strings.TrimSpace("")
}
