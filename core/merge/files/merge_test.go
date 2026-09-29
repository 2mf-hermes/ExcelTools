package files

import (
	"context"
	"path/filepath"
	"testing"

	"ExcelTools/core/model"

	"github.com/xuri/excelize/v2"
)

func writeFile(t *testing.T, path string, headers []string, data [][]string) {
	t.Helper()
	f := excelize.NewFile()
	sh := f.GetSheetName(0)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sh, cell, h)
	}
	for r, row := range data {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = f.SetCellValue(sh, cell, v)
		}
	}
	if _, err := f.NewSheet("Other"); err == nil {
		_ = f.SetCellValue("Other", "A1", "x")
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
}

func openOut(t *testing.T, path string) *excelize.File {
	t.Helper()
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func cell(t *testing.T, f *excelize.File, sheet, addr string) string {
	t.Helper()
	v, err := f.GetCellValue(sheet, addr)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestMergeByPosition(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")
	writeFile(t, a, []string{"Name", "Qty"}, [][]string{{"apple", "1"}})
	writeFile(t, b, []string{"Name", "Qty"}, [][]string{{"banana", "2"}, {"cherry", "3"}})

	res, err := Merge(context.Background(), Options{
		Paths:       []string{a, b}, // caller order preserved
		HeaderRows:  1,
		SheetIndex:  1,
		ColumnAlign: model.AlignByPosition,
		OutputDir:   dir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.FilesUsed != 2 || res.Rows != 4 {
		t.Fatalf("result %+v", res)
	}
	out := openOut(t, res.OutputPath)
	defer out.Close()
	sh := out.GetSheetList()[0]
	if got := cell(t, out, sh, "A2"); got != "apple" {
		t.Fatalf("A2=%q", got)
	}
	if got := cell(t, out, sh, "A3"); got != "banana" {
		t.Fatalf("A3=%q", got)
	}
	if got := cell(t, out, sh, "B3"); got != "2" {
		t.Fatalf("B3=%q", got)
	}
}

func TestMergeByHeaderReorder(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")
	writeFile(t, a, []string{"Name", "Qty"}, [][]string{{"apple", "1"}})
	writeFile(t, b, []string{"Qty", "Name"}, [][]string{{"9", "banana"}})

	res, err := Merge(context.Background(), Options{
		Paths:       []string{a, b},
		HeaderRows:  1,
		SheetIndex:  1,
		ColumnAlign: model.AlignByHeader,
		OutputDir:   dir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := openOut(t, res.OutputPath)
	defer out.Close()
	sh := out.GetSheetList()[0]
	// Header row stays Name, Qty. Row 3: banana under Name, 9 under Qty.
	if got := cell(t, out, sh, "A1"); got != "Name" {
		t.Fatalf("A1 header=%q", got)
	}
	if got := cell(t, out, sh, "A3"); got != "banana" {
		t.Fatalf("A3 name=%q", got)
	}
	if got := cell(t, out, sh, "B3"); got != "9" {
		t.Fatalf("B3 qty=%q", got)
	}
}

func TestMergeByHeaderExtraColumns(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")
	writeFile(t, a, []string{"Name", "Qty"}, [][]string{{"apple", "1"}})
	writeFile(t, b, []string{"Name", "Qty", "Note"}, [][]string{{"banana", "2", "hi"}})

	res, err := Merge(context.Background(), Options{
		Paths:       []string{a, b},
		HeaderRows:  1,
		ColumnAlign: model.AlignByHeader,
		OutputDir:   dir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := openOut(t, res.OutputPath)
	defer out.Close()
	sh := out.GetSheetList()[0]
	if got := cell(t, out, sh, "A3"); got != "banana" {
		t.Fatalf("A3=%q", got)
	}
	if got := cell(t, out, sh, "B3"); got != "2" {
		t.Fatalf("B3=%q", got)
	}
	// New column Note at C
	if got := cell(t, out, sh, "C1"); got != "Note" {
		t.Fatalf("C1 header=%q", got)
	}
	if got := cell(t, out, sh, "C3"); got != "hi" {
		t.Fatalf("C3=%q", got)
	}
}

func TestMergeByHeaderMissingColumn(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")
	writeFile(t, a, []string{"Name", "Qty", "City"}, [][]string{{"apple", "1", "TPE"}})
	writeFile(t, b, []string{"Name", "Qty"}, [][]string{{"banana", "2"}})

	res, err := Merge(context.Background(), Options{
		Paths:       []string{a, b},
		HeaderRows:  1,
		ColumnAlign: model.AlignByHeader,
		OutputDir:   dir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := openOut(t, res.OutputPath)
	defer out.Close()
	sh := out.GetSheetList()[0]
	if got := cell(t, out, sh, "A3"); got != "banana" {
		t.Fatalf("A3=%q", got)
	}
	if got := cell(t, out, sh, "B3"); got != "2" {
		t.Fatalf("B3=%q", got)
	}
	// City column stays empty for second file (no scramble into City)
	if got := cell(t, out, sh, "C3"); got != "" {
		t.Fatalf("C3 should be empty, got %q", got)
	}
}

func TestMergeAbortOnMismatch(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")
	writeFile(t, a, []string{"Name", "Qty"}, [][]string{{"apple", "1"}})
	writeFile(t, b, []string{"Foo", "Bar"}, [][]string{{"x", "y"}})

	res, err := Merge(context.Background(), Options{
		Paths:       []string{a, b},
		HeaderRows:  1,
		ColumnAlign: model.AlignAbort,
		OutputDir:   dir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.FilesUsed != 1 {
		t.Fatalf("expected base-only success: %+v items=%+v", res, res.Items)
	}
	found := false
	for _, it := range res.Items {
		if it.Status == model.ItemFailed {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected failed item: %+v", res.Items)
	}
}

func TestMergeSheetIndex2(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")
	writeFile(t, a, []string{"Name"}, [][]string{{"apple"}})
	writeFile(t, b, []string{"Name"}, [][]string{{"banana"}})

	res, err := Merge(context.Background(), Options{
		Paths:      []string{a, b},
		HeaderRows: 0,
		SheetIndex: 2,
		OutputDir:  dir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := openOut(t, res.OutputPath)
	defer out.Close()
	sh := out.GetSheetList()[0]
	if got := cell(t, out, sh, "A1"); got != "x" {
		t.Fatalf("A1=%q want x from Other", got)
	}
}

func TestMergeHeaderTrim(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")
	// trailing spaces on headers should still match
	f1 := excelize.NewFile()
	sh1 := f1.GetSheetName(0)
	_ = f1.SetCellValue(sh1, "A1", "Name ")
	_ = f1.SetCellValue(sh1, "B1", " Qty")
	_ = f1.SetCellValue(sh1, "A2", "apple")
	_ = f1.SetCellValue(sh1, "B2", "1")
	if err := f1.SaveAs(a); err != nil {
		t.Fatal(err)
	}
	_ = f1.Close()

	f2 := excelize.NewFile()
	sh2 := f2.GetSheetName(0)
	_ = f2.SetCellValue(sh2, "A1", "Qty")
	_ = f2.SetCellValue(sh2, "B1", "Name")
	_ = f2.SetCellValue(sh2, "A2", "9")
	_ = f2.SetCellValue(sh2, "B2", "banana")
	if err := f2.SaveAs(b); err != nil {
		t.Fatal(err)
	}
	_ = f2.Close()

	res, err := Merge(context.Background(), Options{
		Paths:       []string{a, b},
		HeaderRows:  1,
		ColumnAlign: model.AlignByHeader,
		OutputDir:   dir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := openOut(t, res.OutputPath)
	defer out.Close()
	sheet := out.GetSheetList()[0]
	if got := cell(t, out, sheet, "A3"); got != "banana" {
		t.Fatalf("A3=%q want banana under Name", got)
	}
	if got := cell(t, out, sheet, "B3"); got != "9" {
		t.Fatalf("B3=%q want 9 under Qty", got)
	}
}
