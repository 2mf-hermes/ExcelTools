package files

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"ExcelTools/core/model"

	"github.com/xuri/excelize/v2"
)

// writeMultiSheet builds a workbook whose sheet order is exactly the given
// names; an empty data slice leaves that sheet blank.
func writeMultiSheet(t *testing.T, path string, sheets []string, data map[string][][]string) {
	t.Helper()
	f := excelize.NewFile()
	first := f.GetSheetName(0)
	for i, name := range sheets {
		if i == 0 {
			if name != first {
				if err := f.SetSheetName(first, name); err != nil {
					t.Fatal(err)
				}
			}
		} else if _, err := f.NewSheet(name); err != nil {
			t.Fatal(err)
		}
		for r, row := range data[name] {
			for c, v := range row {
				cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
				_ = f.SetCellValue(name, cell, v)
			}
		}
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
}

func TestResolveSheetAutoSkipsBlank(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "cover.xlsx")
	writeMultiSheet(t, p, []string{"封面", "Data", "Other"}, map[string][][]string{
		"Data":  {{"Name"}, {"apple"}},
		"Other": {{"x"}},
	})
	wb, err := excelize.OpenFile(p)
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()

	name, idx, err := ResolveSheet(wb, p, SheetModeAuto, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if name != "Data" || idx != 2 {
		t.Fatalf("auto picked %q (%d), want Data (2)", name, idx)
	}
}

func TestResolveSheetAutoOnlyLastHasData(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "late.xlsx")
	writeMultiSheet(t, p, []string{"a", "b", "c"}, map[string][][]string{
		"c": {{"Name"}, {"z"}},
	})
	wb, err := excelize.OpenFile(p)
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()

	name, idx, err := ResolveSheet(wb, p, SheetModeAuto, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if name != "c" || idx != 3 {
		t.Fatalf("auto picked %q (%d), want c (3)", name, idx)
	}
}

func TestResolveSheetFirstAndIndex(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.xlsx")
	writeMultiSheet(t, p, []string{"cover", "data"}, map[string][][]string{"data": {{"x"}}})
	wb, err := excelize.OpenFile(p)
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()

	if name, idx, _ := ResolveSheet(wb, p, SheetModeFirst, 2, nil); name != "cover" || idx != 1 {
		t.Fatalf("first picked %q (%d)", name, idx)
	}
	// Empty mode behaves like index, preserving old callers.
	if name, idx, _ := ResolveSheet(wb, p, "", 2, nil); name != "data" || idx != 2 {
		t.Fatalf("index picked %q (%d)", name, idx)
	}
	if _, _, err := ResolveSheet(wb, p, SheetModeIndex, 3, nil); err == nil || !strings.HasPrefix(err.Error(), "E_NO_SHEET") {
		t.Fatalf("expected E_NO_SHEET, got %v", err)
	}
}

func TestResolveSheetOverrideNameAndPathNormalization(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ov.xlsx")
	writeMultiSheet(t, p, []string{"a", "b"}, map[string][][]string{"a": {{"1"}}, "b": {{"2"}}})
	wb, err := excelize.OpenFile(p)
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()

	// Set() stores a canonical key, so lookups never depend on caller spelling.
	ov := NewSheetMap()
	ov.Set(strings.ToUpper(filepath.ToSlash(p)), SheetChoice{Name: "b", Index: 1})
	if len(ov) != 1 {
		t.Fatalf("Set with differently-spelled path created %d keys, want 1", len(ov))
	}
	if name, idx, err := ResolveSheet(wb, p, SheetModeAuto, 1, ov); err != nil || name != "b" || idx != 2 {
		t.Fatalf("override name picked %q (%d) err=%v", name, idx, err)
	}
	// Renamed tab: name no longer exists, fall back to the recorded index.
	ov = NewSheetMap()
	ov.Set(p, SheetChoice{Name: "gone", Index: 1})
	if name, idx, err := ResolveSheet(wb, p, SheetModeAuto, 1, ov); err != nil || name != "a" || idx != 1 {
		t.Fatalf("override fallback picked %q (%d) err=%v", name, idx, err)
	}
	// Index-only override.
	ov = NewSheetMap()
	ov.Set(p, SheetChoice{Index: 2})
	if name, _, err := ResolveSheet(wb, p, SheetModeFirst, 1, ov); err != nil || name != "b" {
		t.Fatalf("override index picked %q err=%v", name, err)
	}
	// Index out of range is reported, not silently ignored.
	ov = NewSheetMap()
	ov.Set(p, SheetChoice{Index: 5})
	if _, _, err := ResolveSheet(wb, p, SheetModeAuto, 1, ov); err == nil {
		t.Fatal("expected error for out-of-range override")
	}
}

func TestMergePerFileSheets(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")
	c := filepath.Join(dir, "file3.xlsx")
	writeMultiSheet(t, a, []string{"cover", "sales"}, map[string][][]string{
		"sales": {{"Name"}, {"apple"}},
	})
	writeMultiSheet(t, b, []string{"report", "data"}, map[string][][]string{
		"data": {{"Name"}, {"banana"}},
	})
	writeMultiSheet(t, c, []string{"x", "y", "z"}, map[string][][]string{
		"z": {{"Name"}, {"cherry"}},
	})

	// Pin file2 by name with a stale index: proves the name wins over position.
	overrides := NewSheetMap()
	overrides.Set(b, SheetChoice{Name: "data", Index: 1})
	if got, ok := overrides[normalizePath(b)]; !ok || got.Name != "data" {
		t.Fatalf("override not stored: %+v (len=%d) key=%q", overrides, len(overrides), normalizePath(b))
	}

	res, err := Merge(context.Background(), Options{
		Paths:          []string{a, b, c},
		HeaderRows:     1,
		SheetMode:      SheetModeAuto,
		ColumnAlign:    model.AlignByHeader,
		SheetOverrides: overrides,
		OutputDir:      dir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("result %+v", res)
	}
	if len(res.Sheets) != 3 {
		t.Fatalf("manifest %+v, want 3 entries", res.Sheets)
	}
	if res.Sheets[0].Sheet != "sales" || res.Sheets[1].Sheet != "data" || res.Sheets[2].Sheet != "z" {
		t.Fatalf("manifest sheets %+v", res.Sheets)
	}
	if got := res.Sheets[1].Index; got != 2 {
		t.Fatalf("resolved index for file2 = %d, want 2", got)
	}

	out := openOut(t, res.OutputPath)
	defer out.Close()
	sh := out.GetSheetList()[0]
	if got := cell(t, out, sh, "A2"); got != "apple" {
		t.Fatalf("A2=%q", got)
	}
	if got := cell(t, out, sh, "A3"); got != "banana" {
		t.Fatalf("A3=%q want banana from pinned sheet data", got)
	}
	if got := cell(t, out, sh, "A4"); got != "cherry" {
		t.Fatalf("A4=%q want cherry from auto-picked sheet z", got)
	}
}
