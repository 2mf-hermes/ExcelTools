package files

import (
	"context"
	"path/filepath"
	"testing"

	"ExcelTools/core/model"

	"github.com/xuri/excelize/v2"
)

func TestFirstFileGetsSourceColumn(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")

	f1 := excelize.NewFile()
	sh1 := f1.GetSheetName(0)
	_ = f1.SetCellValue(sh1, "A1", "Name")
	_ = f1.SetCellValue(sh1, "B1", "Qty")
	_ = f1.SetCellValue(sh1, "A2", "apple")
	_ = f1.SetCellValue(sh1, "B2", "1")
	if err := f1.SaveAs(a); err != nil {
		t.Fatal(err)
	}
	_ = f1.Close()

	f2 := excelize.NewFile()
	sh2 := f2.GetSheetName(0)
	_ = f2.SetCellValue(sh2, "A1", "Name")
	_ = f2.SetCellValue(sh2, "B1", "Qty")
	_ = f2.SetCellValue(sh2, "A2", "banana")
	_ = f2.SetCellValue(sh2, "B2", "2")
	if err := f2.SaveAs(b); err != nil {
		t.Fatal(err)
	}
	_ = f2.Close()

	res, err := Merge(context.Background(), Options{
		Paths:         []string{a, b},
		HeaderRows:    1,
		ColumnAlign:   model.AlignByHeader,
		AddSourceFile: true,
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

	// Header label on C1
	if v, _ := out.GetCellValue(sheet, "C1"); v != DefaultSourceColumnLabel {
		t.Fatalf("C1=%q want label", v)
	}
	// First file data row
	if v, _ := out.GetCellValue(sheet, "C2"); v != "file1.xlsx" {
		t.Fatalf("C2=%q want file1.xlsx", v)
	}
	// Second file data row
	if v, _ := out.GetCellValue(sheet, "C3"); v != "file2.xlsx" {
		t.Fatalf("C3=%q want file2.xlsx", v)
	}
}
