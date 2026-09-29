// Package sheets merges multiple worksheets from one workbook into one sheet.
// Behavior: docs/BEHAVIOR_SPEC.md Tool A (B-A01..B-A12).
package sheets

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"ExcelTools/core/model"
	"ExcelTools/platform/fs"

	"github.com/xuri/excelize/v2"
)

// DefaultSourceColumnLabel is the header of the optional source-sheet column.
const DefaultSourceColumnLabel = "來源工作表"

// ProgressFunc receives merge progress updates.
type ProgressFunc func(ev model.ProgressEvent)

// MergeOptions is the engine-facing option bag.
type MergeOptions struct {
	SheetNames        []string
	HeaderMode        model.HeaderMode
	AddSourceSheet    bool
	SkipEmpty         bool
	OutputSheetName   string
	SourceColumnLabel string
	OutputPath        string // if empty, derived next to source
}

// ListSheets returns sheet metadata for the workbook at path.
func ListSheets(path string) ([]model.SheetRef, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("E_FORMAT: open workbook: %w", err)
	}
	defer f.Close()
	return listFrom(f)
}

func listFrom(f *excelize.File) ([]model.SheetRef, error) {
	names := f.GetSheetList()
	refs := make([]model.SheetRef, 0, len(names))
	for _, name := range names {
		sr := model.SheetRef{Name: name}
		if vis, err := f.GetSheetVisible(name); err == nil && !vis {
			sr.Hidden = true
		}
		rows, _, cols, err := sheetExtent(f, name)
		if err != nil {
			return nil, err
		}
		sr.Rows = len(rows)
		if dim, err := f.GetSheetDimension(name); err == nil && dim != "" {
			parts := strings.Split(dim, ":")
			end := parts[len(parts)-1]
			if c, r, err := excelize.CellNameToCoordinates(end); err == nil {
				if r > sr.Rows {
					sr.Rows = r
				}
				sr.Cols = c
				_ = cols
			}
		}
		sr.Empty = rowsLookEmpty(rows)
		refs = append(refs, sr)
	}
	return refs, nil
}

// Merge copies selected sheets into one output workbook.
// Source path is never modified. Cancel removes any temp output.
func Merge(ctx context.Context, srcPath string, opts MergeOptions, progress ProgressFunc) (*model.MergeResult, error) {
	start := time.Now()
	res := &model.MergeResult{}
	say := func(phase string, cur, total int, label string) {
		if progress == nil {
			return
		}
		pct := 0
		if total > 0 {
			pct = int(float64(cur) / float64(total) * 100)
		}
		progress(model.ProgressEvent{
			Phase: phase, Current: cur, Total: total, Label: label, Percent: pct,
		})
	}
	if err := ctx.Err(); err != nil {
		res.Message = "E_CANCEL: cancelled"
		return res, err
	}

	if opts.HeaderMode == "" {
		opts.HeaderMode = model.HeaderEach
	}
	if opts.OutputSheetName == "" {
		opts.OutputSheetName = "Merged"
	}
	if opts.SourceColumnLabel == "" {
		opts.SourceColumnLabel = DefaultSourceColumnLabel
	}

	srcPath = strings.TrimSpace(srcPath)
	if srcPath == "" {
		return res, fmt.Errorf("E_EMPTY: source path is empty")
	}

	say("read", 0, 1, filepath.Base(srcPath))
	f, err := excelize.OpenFile(srcPath)
	if err != nil {
		return res, fmt.Errorf("E_FORMAT: open source: %w", err)
	}
	defer f.Close()

	if len(opts.SheetNames) == 0 {
		// default: all visible sheets
		for _, n := range f.GetSheetList() {
			vis, _ := f.GetSheetVisible(n)
			if vis {
				opts.SheetNames = append(opts.SheetNames, n)
			}
		}
	}
	if len(opts.SheetNames) == 0 {
		return res, fmt.Errorf("E_EMPTY: no sheets selected")
	}

	existing := map[string]bool{}
	for _, n := range f.GetSheetList() {
		existing[n] = true
	}
	var missing []string
	for _, n := range opts.SheetNames {
		if !existing[n] {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		return res, fmt.Errorf("E_FORMAT: sheet not found: %s", strings.Join(missing, ", "))
	}

	outName := uniqueSheetName(f, opts.OutputSheetName)
	if _, err := f.NewSheet(outName); err != nil {
		return res, fmt.Errorf("E_INTERNAL: create merged sheet: %w", err)
	}

	used := 0
	skipped := 0
	destRow := 1
	headerWritten := false
	maxColOut := 0

	sourceCol := 0
	if opts.AddSourceSheet {
		widest := 0
		for _, name := range opts.SheetNames {
			if name == outName {
				continue
			}
			if opts.SkipEmpty && isSheetEmpty(f, name) {
				continue
			}
			_, _, dimCols, err := sheetExtent(f, name)
			if err != nil {
				continue
			}
			if dimCols > widest {
				widest = dimCols
			}
		}
		if widest == 0 {
			widest = 1
		}
		sourceCol = widest + 1
		maxColOut = sourceCol
	}

	total := len(opts.SheetNames)
	for i, name := range opts.SheetNames {
		if err := ctx.Err(); err != nil {
			return res, fmt.Errorf("E_CANCEL: cancelled")
		}
		say("merge", i, total, name)

		if name == outName {
			continue
		}
		_, dimRows, dimCols, err := sheetExtent(f, name)
		if err != nil {
			return res, err
		}
		if opts.SkipEmpty && isSheetEmpty(f, name) {
			skipped++
			continue
		}
		if dimRows == 0 || dimCols == 0 {
			if opts.SkipEmpty {
				skipped++
			}
			continue
		}

		startSrcRow := 1
		switch opts.HeaderMode {
		case model.HeaderEach:
			if !headerWritten {
				if err := copyRow(f, outName, name, 1, destRow, opts, sourceCol); err != nil {
					return res, err
				}
				headerWritten = true
				destRow++
			}
			startSrcRow = 2
		case model.HeaderFirst:
			if !headerWritten {
				if err := copyRow(f, outName, name, 1, destRow, opts, sourceCol); err != nil {
					return res, err
				}
				headerWritten = true
				destRow++
				startSrcRow = 2
			} else {
				startSrcRow = 2
			}
		case model.HeaderNone:
			// all rows are data
		}

		if dimCols > maxColOut {
			maxColOut = dimCols
		}
		if sourceCol > maxColOut {
			maxColOut = sourceCol
		}

		sheetDestStart := destRow
		dataRows := 0
		for srcRow := startSrcRow; srcRow <= dimRows; srcRow++ {
			if err := ctx.Err(); err != nil {
				return res, fmt.Errorf("E_CANCEL: cancelled")
			}
			if err := copyCellsRow(f, outName, name, srcRow, destRow, 1, dimCols, 0); err != nil {
				return res, err
			}
			if opts.AddSourceSheet && sourceCol > 0 {
				cell := cellName(sourceCol, destRow)
				if err := f.SetCellValue(outName, cell, name); err != nil {
					return res, err
				}
			}
			destRow++
			dataRows++
		}
		if dataRows > 0 {
			if err := copyMergeCells(f, name, outName, startSrcRow, sheetDestStart, 0); err != nil {
				return res, err
			}
			used++
		} else if opts.SkipEmpty {
			skipped++
		}
	}

	if used == 0 {
		return res, fmt.Errorf("E_EMPTY: no data in selected sheets")
	}

	// Column widths from first used sheet.
	for _, name := range opts.SheetNames {
		if name == outName {
			continue
		}
		_, _, dimCols, _ := sheetExtent(f, name)
		if dimCols == 0 {
			continue
		}
		for c := 1; c <= dimCols; c++ {
			w, err := f.GetColWidth(name, colLetter(c))
			if err != nil || w <= 0 {
				continue
			}
			_ = f.SetColWidth(outName, colLetter(c), colLetter(c), w)
		}
		break
	}
	if opts.AddSourceSheet && sourceCol > 0 {
		_ = f.SetColWidth(outName, colLetter(sourceCol), colLetter(sourceCol), 16)
	}

	// Drop original sheets so output is a single combined sheet.
	for _, name := range f.GetSheetList() {
		if name == outName {
			continue
		}
		_ = f.DeleteSheet(name)
	}
	f.SetActiveSheet(0)

	if err := ctx.Err(); err != nil {
		return res, fmt.Errorf("E_CANCEL: cancelled")
	}

	outPath := opts.OutputPath
	if outPath == "" {
		outPath = DefaultOutputPath(srcPath)
	}
	outPath = fs.UniquePath(outPath)

	say("write", total, total, filepath.Base(outPath))
	if err := saveWorkbook(ctx, f, outPath); err != nil {
		return res, err
	}

	res.Success = true
	res.Message = "ok"
	res.OutputPath = outPath
	res.FilesUsed = 1
	res.Rows = destRow - 1
	res.Cols = maxColOut
	res.DurationMs = time.Since(start).Milliseconds()
	res.Items = []model.ItemReport{{
		File:   filepath.Base(srcPath),
		Status: model.ItemSuccess,
		Rows:   res.Rows,
	}}
	if skipped > 0 {
		res.Partial = true
		res.Items = append(res.Items, model.ItemReport{
			File:    filepath.Base(srcPath),
			Status:  model.ItemSkipped,
			Message: fmt.Sprintf("%d empty sheets skipped", skipped),
		})
	}
	say("done", 100, 100, "ok")
	return res, nil
}

// DefaultOutputPath returns <src>-merged.xlsx next to the source.
func DefaultOutputPath(srcPath string) string {
	ext := filepath.Ext(srcPath)
	base := strings.TrimSuffix(srcPath, ext)
	return base + "-merged.xlsx"
}

func saveWorkbook(ctx context.Context, f *excelize.File, destPath string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("E_CANCEL: cancelled")
	}
	dir := filepath.Dir(destPath)
	tmp, err := fs.TempPathIn(dir, "exceltools-merge-*.xlsx")
	if err != nil {
		return fmt.Errorf("E_PERM: create temp: %w", err)
	}
	if err := f.SaveAs(tmp); err != nil {
		_ = fs.RemoveIfExists(tmp)
		return fmt.Errorf("E_PERM: save output: %w", err)
	}
	if err := ctx.Err(); err != nil {
		_ = fs.RemoveIfExists(tmp)
		return fmt.Errorf("E_CANCEL: cancelled")
	}
	if err := fs.RenameOver(tmp, destPath); err != nil {
		_ = fs.RemoveIfExists(tmp)
		return fmt.Errorf("E_PERM: finalize output: %w", err)
	}
	return nil
}

func uniqueSheetName(f *excelize.File, base string) string {
	existing := map[string]bool{}
	for _, n := range f.GetSheetList() {
		existing[n] = true
	}
	if !existing[base] {
		return base
	}
	for i := 2; i < 1000; i++ {
		cand := fmt.Sprintf("%s%d", base, i)
		if !existing[cand] {
			return cand
		}
	}
	return fmt.Sprintf("%s%d", base, time.Now().UnixNano()%100000)
}

func sheetExtent(f *excelize.File, name string) ([][]string, int, int, error) {
	rows, err := f.GetRows(name)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("E_FORMAT: read sheet %q: %w", name, err)
	}
	maxR, maxC := len(rows), 0
	for _, row := range rows {
		if len(row) > maxC {
			maxC = len(row)
		}
	}
	if dim, err := f.GetSheetDimension(name); err == nil && dim != "" {
		parts := strings.Split(dim, ":")
		if len(parts) == 2 {
			if c, r, err := excelize.CellNameToCoordinates(parts[1]); err == nil {
				if r > maxR {
					maxR = r
				}
				if c > maxC {
					maxC = c
				}
			}
		}
	}
	return rows, maxR, maxC, nil
}

func rowsLookEmpty(rows [][]string) bool {
	for _, row := range rows {
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				return false
			}
		}
	}
	return true
}

func isSheetEmpty(f *excelize.File, name string) bool {
	rows, _, _, err := sheetExtent(f, name)
	if err != nil {
		return true
	}
	return rowsLookEmpty(rows)
}

func copyRow(f *excelize.File, outName, srcSheet string, srcRow, destRow int, opts MergeOptions, sourceCol int) error {
	_, _, dimCols, err := sheetExtent(f, srcSheet)
	if err != nil {
		return err
	}
	if err := copyCellsRow(f, outName, srcSheet, srcRow, destRow, 1, dimCols, 0); err != nil {
		return err
	}
	if opts.AddSourceSheet && sourceCol > 0 {
		if err := f.SetCellValue(outName, cellName(sourceCol, destRow), opts.SourceColumnLabel); err != nil {
			return err
		}
	}
	return nil
}

func copyMergeCells(f *excelize.File, srcSheet, outName string, srcStartRow, destStartRow, colOffset int) error {
	merges, err := f.GetMergeCells(srcSheet)
	if err != nil {
		return nil
	}
	for _, m := range merges {
		sc, sr, err1 := excelize.CellNameToCoordinates(m.GetStartAxis())
		ec, er, err2 := excelize.CellNameToCoordinates(m.GetEndAxis())
		if err1 != nil || err2 != nil {
			continue
		}
		if er < srcStartRow {
			continue
		}
		if sr < srcStartRow {
			sr = srcStartRow
		}
		destSR := destStartRow + (sr - srcStartRow)
		destER := destStartRow + (er - srcStartRow)
		_ = f.MergeCell(outName, cellName(sc+colOffset, destSR), cellName(ec+colOffset, destER))
	}
	return nil
}

func copyCellsRow(f *excelize.File, outName, srcSheet string, srcRow, dstRow, colStart, colEnd, colOffset int) error {
	if h, err := f.GetRowHeight(srcSheet, srcRow); err == nil && h > 0 {
		_ = f.SetRowHeight(outName, dstRow, h)
	}
	for c := colStart; c <= colEnd; c++ {
		from := cellName(c, srcRow)
		to := cellName(c+colOffset, dstRow)

		styleID, err := f.GetCellStyle(srcSheet, from)
		if err != nil {
			return err
		}
		if styleID != 0 {
			if err := f.SetCellStyle(outName, to, to, styleID); err != nil {
				return fmt.Errorf("E_INTERNAL: copy style %s->%s: %w", from, to, err)
			}
		}

		formula, err := f.GetCellFormula(srcSheet, from)
		if err != nil {
			return err
		}
		if formula != "" {
			if err := f.SetCellFormula(outName, to, formula); err != nil {
				return fmt.Errorf("E_INTERNAL: copy formula %s->%s: %w", from, to, err)
			}
			continue
		}

		val, err := f.GetCellValue(srcSheet, from)
		if err != nil {
			return err
		}
		if val != "" {
			if err := f.SetCellValue(outName, to, val); err != nil {
				return fmt.Errorf("E_INTERNAL: copy value %s->%s: %w", from, to, err)
			}
		}
	}
	return nil
}

func cellName(col, row int) string {
	s, _ := excelize.CoordinatesToCellName(col, row)
	return s
}

func colLetter(col int) string {
	s, _ := excelize.ColumnNumberToName(col)
	return s
}
