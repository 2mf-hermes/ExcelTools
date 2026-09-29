package files

import (
	"context"
	"path/filepath"
	"testing"

	"ExcelTools/core/model"

	"github.com/xuri/excelize/v2"
)

func TestKeepBlankRowsAndNoMissingRows(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")

	// file1: header + data + blank + data
	f1 := excelize.NewFile()
	sh1 := f1.GetSheetName(0)
	_ = f1.SetCellValue(sh1, "A1", "Name")
	_ = f1.SetCellValue(sh1, "B1", "Qty")
	_ = f1.SetCellValue(sh1, "A2", "apple")
	_ = f1.SetCellValue(sh1, "B2", "1")
	// row 3 empty
	_ = f1.SetCellValue(sh1, "A4", "cherry")
	_ = f1.SetCellValue(sh1, "B4", "3")
	if err := f1.SaveAs(a); err != nil {
		t.Fatal(err)
	}
	_ = f1.Close()

	// file2: header + data + blank + data (same structure)
	f2 := excelize.NewFile()
	sh2 := f2.GetSheetName(0)
	_ = f2.SetCellValue(sh2, "A1", "Name")
	_ = f2.SetCellValue(sh2, "B1", "Qty")
	_ = f2.SetCellValue(sh2, "A2", "banana")
	_ = f2.SetCellValue(sh2, "B2", "2")
	_ = f2.SetCellValue(sh2, "A4", "date")
	_ = f2.SetCellValue(sh2, "B4", "4")
	if err := f2.SaveAs(b); err != nil {
		t.Fatal(err)
	}
	_ = f2.Close()

	res, err := Merge(context.Background(), Options{
		Paths:         []string{a, b},
		HeaderRows:    1,
		ColumnAlign:   model.AlignByHeader,
		AddSourceFile: true,
		KeepBlankRows: true,
		OutputDir:     dir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	out, err := excelize.OpenFile(res.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	sheet := out.GetSheetList()[0]

	// file1: rows 1-4 (header, apple, blank, cherry)
	// file2 data starts at row 5: banana, blank, date → rows 5-7
	if v, _ := out.GetCellValue(sheet, "A2"); v != "apple" {
		t.Fatalf("A2=%q", v)
	}
	// blank row preserved with source tag
	if v, _ := out.GetCellValue(sheet, "A3"); v != "" {
		t.Fatalf("A3 should be blank, got %q", v)
	}
	if v, _ := out.GetCellValue(sheet, "C3"); v != "file1.xlsx" {
		t.Fatalf("C3 source=%q want file1.xlsx on blank row", v)
	}
	if v, _ := out.GetCellValue(sheet, "A4"); v != "cherry" {
		t.Fatalf("A4=%q", v)
	}
	if v, _ := out.GetCellValue(sheet, "A5"); v != "banana" {
		t.Fatalf("A5=%q want banana (second file, no dropped rows)", v)
	}
	if v, _ := out.GetCellValue(sheet, "A7"); v != "date" {
		t.Fatalf("A7=%q want date", v)
	}
	if v, _ := out.GetCellValue(sheet, "C5"); v != "file2.xlsx" {
		t.Fatalf("C5=%q", v)
	}
}

func TestSkipBlankRowsWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")

	f1 := excelize.NewFile()
	sh1 := f1.GetSheetName(0)
	_ = f1.SetCellValue(sh1, "A1", "Name")
	_ = f1.SetCellValue(sh1, "A2", "apple")
	_ = f1.SetCellValue(sh1, "A4", "cherry")
	_ = f1.SaveAs(a)
	_ = f1.Close()

	f2 := excelize.NewFile()
	sh2 := f2.GetSheetName(0)
	_ = f2.SetCellValue(sh2, "A1", "Name")
	_ = f2.SetCellValue(sh2, "A2", "banana")
	_ = f2.SaveAs(b)
	_ = f2.Close()

	res, err := Merge(context.Background(), Options{
		Paths:         []string{a, b},
		HeaderRows:    1,
		ColumnAlign:   model.AlignByHeader,
		AddSourceFile: true,
		KeepBlankRows: false,
		OutputDir:     dir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := excelize.OpenFile(res.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	sheet := out.GetSheetList()[0]
	// Base keeps its blank row; subsequent file skips blanks.
	if v, _ := out.GetCellValue(sheet, "A4"); v != "cherry" {
		t.Fatalf("A4=%q want cherry (base unchanged)", v)
	}
	if v, _ := out.GetCellValue(sheet, "A5"); v != "banana" {
		t.Fatalf("A5=%q want banana (second file blank skipped)", v)
	}
}
