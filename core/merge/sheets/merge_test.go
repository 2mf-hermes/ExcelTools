package sheets

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ExcelTools/core/model"

	"github.com/xuri/excelize/v2"
)

func writeFixture(t *testing.T, path string) {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()

	_ = f.SetSheetName("Sheet1", "Data1")
	_, _ = f.NewSheet("Data2")
	_, _ = f.NewSheet("Empty")

	for i, h := range []string{"Name", "Qty"} {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue("Data1", cell, h)
		_ = f.SetCellValue("Data2", cell, h)
	}
	_ = f.SetCellValue("Data1", "A2", "apple")
	_ = f.SetCellValue("Data1", "B2", "1")
	_ = f.SetCellValue("Data2", "A2", "banana")
	_ = f.SetCellValue("Data2", "B2", "2")
	_ = f.SetCellValue("Data2", "A3", "cherry")
	_ = f.SetCellValue("Data2", "B3", "3")

	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
}

func TestMergeHeaderEach(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.xlsx")
	writeFixture(t, src)

	out := filepath.Join(dir, "out.xlsx")
	res, err := Merge(context.Background(), src, MergeOptions{
		SheetNames:     []string{"Data1", "Data2", "Empty"},
		HeaderMode:     model.HeaderEach,
		SkipEmpty:      true,
		AddSourceSheet: true,
		OutputPath:     out,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.Rows != 4 {
		t.Fatalf("expected 1 header + 3 data rows, got %+v", res)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatal("output missing")
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatal("source must remain")
	}

	f, err := excelize.OpenFile(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) != 1 {
		t.Fatalf("expected single output sheet, got %v", sheets)
	}
	h1, _ := f.GetCellValue(sheets[0], "A1")
	if h1 != "Name" {
		t.Fatalf("header A1=%q", h1)
	}
	// source col after 2 data cols => C1
	hc, _ := f.GetCellValue(sheets[0], "C1")
	if hc != DefaultSourceColumnLabel {
		t.Fatalf("source header C1=%q", hc)
	}
	a2, _ := f.GetCellValue(sheets[0], "A2")
	if a2 != "apple" {
		t.Fatalf("A2=%q", a2)
	}
	c2, _ := f.GetCellValue(sheets[0], "C2")
	if c2 != "Data1" {
		t.Fatalf("C2 source=%q", c2)
	}
	a3, _ := f.GetCellValue(sheets[0], "A3")
	if a3 != "banana" {
		t.Fatalf("A3=%q", a3)
	}
}

func TestMergeHeaderNone(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.xlsx")
	writeFixture(t, src)
	out := filepath.Join(dir, "out.xlsx")
	res, err := Merge(context.Background(), src, MergeOptions{
		SheetNames: []string{"Data1", "Data2"},
		HeaderMode: model.HeaderNone,
		OutputPath: out,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Data1 has 2 rows, Data2 has 3 rows
	if res.Rows != 5 {
		t.Fatalf("expected 5 rows, got %d", res.Rows)
	}
}

func TestMergeCancel(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.xlsx")
	writeFixture(t, src)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out := filepath.Join(dir, "out.xlsx")
	_, err := Merge(ctx, src, MergeOptions{
		SheetNames: []string{"Data1"},
		OutputPath: out,
	}, nil)
	if err == nil {
		t.Fatal("expected cancel error")
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Fatal("cancel must not leave output")
	}
}

func TestDefaultOutputPathUnique(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "book.xlsx")
	writeFixture(t, src)
	if _, err := Merge(context.Background(), src, MergeOptions{
		SheetNames: []string{"Data1"},
	}, nil); err != nil {
		t.Fatal(err)
	}
	// second merge auto-uniques
	res, err := Merge(context.Background(), src, MergeOptions{
		SheetNames: []string{"Data1"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.OutputPath == "" {
		t.Fatal("no output path")
	}
	if _, err := os.Stat(res.OutputPath); err != nil {
		t.Fatal(err)
	}
}
