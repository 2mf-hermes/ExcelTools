package sheets_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ExcelTools/core/merge/sheets"
	"ExcelTools/core/model"

	"github.com/xuri/excelize/v2"
)

// Integration-style test: real file on disk, progress callbacks, source untouched.
func TestMergeIntegrationProgressAndSourceIntact(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "book.xlsx")

	f := excelize.NewFile()
	_ = f.SetSheetName("Sheet1", "A")
	_, _ = f.NewSheet("B")
	_ = f.SetCellValue("A", "A1", "h")
	_ = f.SetCellValue("A", "A2", "1")
	_ = f.SetCellValue("B", "A1", "h")
	_ = f.SetCellValue("B", "A2", "2")
	if err := f.SaveAs(src); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	before, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	events := 0
	res, err := sheets.Merge(context.Background(), src, sheets.MergeOptions{
		SheetNames:     []string{"A", "B"},
		HeaderMode:     model.HeaderEach,
		AddSourceSheet: true,
	}, func(ev model.ProgressEvent) {
		events++
	})
	if err != nil {
		t.Fatal(err)
	}
	if events == 0 {
		t.Fatal("expected progress events")
	}
	if !res.Success || res.Rows != 3 {
		t.Fatalf("unexpected result %+v", res)
	}

	after, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("source file must be byte-identical")
	}

	out, err := excelize.OpenFile(res.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	sheet := out.GetSheetList()[0]
	v, _ := out.GetCellValue(sheet, "A2")
	if v != "1" {
		t.Fatalf("A2=%q", v)
	}
	v, _ = out.GetCellValue(sheet, "A3")
	if v != "2" {
		t.Fatalf("A3=%q", v)
	}
}
