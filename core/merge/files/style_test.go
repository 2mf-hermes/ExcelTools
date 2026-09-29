package files

import (
	"context"
	"path/filepath"
	"testing"

	"ExcelTools/core/model"

	"github.com/xuri/excelize/v2"
)

func TestMergeCopiesStylesAndFormulas(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "file1.xlsx")
	b := filepath.Join(dir, "file2.xlsx")

	// File 1: bold header, yellow fill on data
	f1 := excelize.NewFile()
	sh1 := f1.GetSheetName(0)
	_ = f1.SetCellValue(sh1, "A1", "Name")
	_ = f1.SetCellValue(sh1, "B1", "Qty")
	styleBold, _ := f1.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	_ = f1.SetCellStyle(sh1, "A1", "B1", styleBold)
	_ = f1.SetCellValue(sh1, "A2", "apple")
	_ = f1.SetCellValue(sh1, "B2", "1")
	if err := f1.SaveAs(a); err != nil {
		t.Fatal(err)
	}
	_ = f1.Close()

	// File 2: green fill on name cell, formula =B2+1 in C
	f2 := excelize.NewFile()
	sh2 := f2.GetSheetName(0)
	_ = f2.SetCellValue(sh2, "A1", "Name")
	_ = f2.SetCellValue(sh2, "B1", "Qty")
	_ = f2.SetCellValue(sh2, "C1", "Total")
	_ = f2.SetCellValue(sh2, "A2", "banana")
	_ = f2.SetCellValue(sh2, "B2", "2")
	fillStyle, _ := f2.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#C6EFCE"}, Pattern: 1},
	})
	_ = f2.SetCellStyle(sh2, "A2", "A2", fillStyle)
	_ = f2.SetCellFormula(sh2, "C2", "B2+1")
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

	out, err := excelize.OpenFile(res.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	sheet := out.GetSheetList()[0]

	// Values aligned
	if v, _ := out.GetCellValue(sheet, "A3"); v != "banana" {
		t.Fatalf("A3=%q", v)
	}

	// Style from file2 should exist on A3 (fill), not style ID 0 blindly
	sid, err := out.GetCellStyle(sheet, "A3")
	if err != nil {
		t.Fatal(err)
	}
	if sid == 0 {
		t.Fatal("A3 has no style; expected fill copied from source")
	}
	st, err := out.GetStyle(sid)
	if err != nil || st == nil {
		t.Fatal(err)
	}
	if st.Fill.Type != "pattern" || len(st.Fill.Color) == 0 {
		t.Fatalf("fill not preserved: %+v", st.Fill)
	}

	// Formula preserved
	formula, _ := out.GetCellFormula(sheet, "C3")
	if formula != "B3+1" {
		// Either original text or adjusted; must not be empty if source had formula
		// B2+1 as-is is acceptable (Excel recalculates); also accept B3+1
		t.Fatalf("C3 formula=%q", formula)
	}
}
