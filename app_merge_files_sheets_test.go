package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

// mkBook writes a workbook with the given sheets, each filled row-wise.
func mkBook(t *testing.T, path string, sheets map[string][][]string, order ...string) {
	t.Helper()
	if len(order) == 0 {
		for n := range sheets {
			order = append(order, n)
		}
	}
	f := excelize.NewFile()
	defer f.Close()
	for i, name := range order {
		if i == 0 {
			if err := f.SetSheetName("Sheet1", name); err != nil {
				t.Fatal(err)
			}
		} else if _, err := f.NewSheet(name); err != nil {
			t.Fatal(err)
		}
		for r, row := range sheets[name] {
			for c, v := range row {
				cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
				if err := f.SetCellValue(name, cell, v); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
}

func TestAppMergeFilesPerFileSheet(t *testing.T) {
	dir := t.TempDir()
	// A: cover sheet first (blank), data on "sales".
	a := filepath.Join(dir, "a.xlsx")
	mkBook(t, a, map[string][][]string{
		"cover": {{"Report"}},
		"sales": {{"Name", "Qty"}, {"apple", "1"}},
	}, "cover", "sales")
	// B: data on the second sheet; user pins it explicitly.
	b := filepath.Join(dir, "b.xlsx")
	mkBook(t, b, map[string][][]string{
		"notes": {{"x"}},
		"data":  {{"Name", "Qty"}, {"banana", "2"}},
	}, "notes", "data")
	// C: uniform "auto" resolves to its only populated sheet "z".
	c := filepath.Join(dir, "c.xlsx")
	mkBook(t, c, map[string][][]string{
		"blank": {},
		"z":     {{"Name", "Qty"}, {"cherry", "3"}},
	}, "blank", "z")

	app := NewApp()
	app.ctx = context.TODO()
	res, err := app.MergeFiles(MergeFilesRequest{
		Paths:          []string{a, b, c},
		HeaderRows:     1,
		SheetMode:      "auto",
		SheetIndex:     1,
		ColumnAlign:    "by-header",
		AddSourceFile:  false,
		SheetOverrides: []SheetOverride{{Path: b, Index: 2}},
		OutputDir:      dir,
	})
	if err != nil {
		t.Fatalf("MergeFiles: %v", err)
	}
	if !res.Success {
		t.Fatalf("not successful: %+v", res)
	}
	if len(res.Sheets) != 3 {
		t.Fatalf("manifest has %d entries, want 3: %+v", len(res.Sheets), res.Sheets)
	}
	want := map[string]string{a: "sales", b: "data", c: "z"}
	for _, s := range res.Sheets {
		if w := want[s.Path]; w != s.Sheet {
			t.Errorf("%s used sheet %q, want %q", s.Name, s.Sheet, w)
		}
	}

	out, err := excelize.OpenFile(res.OutputPath)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer out.Close()
	sh := out.GetSheetList()[0]
	rows, err := out.GetRows(sh)
	if err != nil {
		t.Fatal(err)
	}
	// header + 3 data rows
	if len(rows) != 4 {
		t.Fatalf("output rows = %d, want 4: %v", len(rows), rows)
	}
	for i, wantName := range []string{"apple", "banana", "cherry"} {
		if got := rows[i+1][0]; got != wantName {
			t.Errorf("row %d name = %q, want %q", i+2, got, wantName)
		}
	}

	// Bad mode is rejected rather than silently defaulting.
	if _, err := app.MergeFiles(MergeFilesRequest{
		Paths: []string{a}, SheetMode: "nope", OutputDir: dir,
	}); err == nil {
		t.Error("unknown sheetMode was accepted")
	}
}

