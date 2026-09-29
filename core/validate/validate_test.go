package validate

import (
	"testing"

	"ExcelTools/core/model"
)

func TestExtSupported(t *testing.T) {
	for _, p := range []string{"a.xlsx", "b.XLSM", "c.xls"} {
		if !ExtSupported(p) {
			t.Fatalf("expected support for %s", p)
		}
	}
	if ExtSupported("d.csv") {
		t.Fatal("csv must not be supported")
	}
}

func TestValidateSheetOptions(t *testing.T) {
	sheets := []model.SheetRef{{Name: "S1", Rows: 1}}
	if err := ValidateSheetOptions(sheets, model.SheetMergeOptions{HeaderMode: model.HeaderEach}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSheetOptions(nil, model.SheetMergeOptions{}); err == nil {
		t.Fatal("empty sheet list should fail")
	}
}

func TestValidateFileOptions(t *testing.T) {
	files := []model.FileRef{{Path: "a.xlsx", Status: model.FileOK}}
	opts := model.FileMergeOptions{SheetIndex: 1, ColumnAlign: model.AlignByHeader}
	if err := ValidateFileOptions(files, opts); err != nil {
		t.Fatal(err)
	}
	bad := opts
	bad.SheetIndex = 0
	if err := ValidateFileOptions(files, bad); err == nil {
		t.Fatal("sheet index 0 should fail")
	}
}

func TestIsTempOrToolOutput(t *testing.T) {
	if !IsTempOrToolOutput("~$book.xlsx", "合併結果_") {
		t.Fatal("lock file should be filtered")
	}
	if !IsTempOrToolOutput("合併結果_20240101.xlsx", "合併結果_") {
		t.Fatal("prior output should be filtered")
	}
	if IsTempOrToolOutput("report.xlsx", "合併結果_") {
		t.Fatal("normal file should not be filtered")
	}
}
