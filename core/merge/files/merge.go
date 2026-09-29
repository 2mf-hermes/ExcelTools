package files

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"ExcelTools/core/model"
	"ExcelTools/core/report"
	"ExcelTools/platform/fs"

	"github.com/xuri/excelize/v2"
)

const (
	// OutPrefix is the default output filename prefix.
	OutPrefix = "合併結果_"
	// DefaultSourceColumnLabel is the header of the optional source-file column.
	DefaultSourceColumnLabel = "來源檔案"
	maxReasonableCols        = 256
)

// ProgressFunc receives merge progress updates.
type ProgressFunc func(ev model.ProgressEvent)

// Options configures multi-file merge.
type Options struct {
	Paths               []string
	HeaderRows          int
	SheetIndex          int // 1-based
	ColumnAlign         model.ColumnAlign
	AddSourceFile       bool
	KeepBaseOtherSheets bool
	// KeepBlankRows keeps intermediate empty rows from first used data row to last.
	// Default true so output row count matches sources. Source-file column is filled.
	KeepBlankRows     bool
	SourceColumnLabel string
	OutputDir         string
	OutputNameBase    string
}

// Merge appends subsequent workbooks onto the first usable file as base.
// Other sheets of the base file are kept so filters/links stay valid.
func Merge(ctx context.Context, opts Options, progress ProgressFunc) (*model.MergeResult, error) {
	start := time.Now()
	res := &model.MergeResult{}
	rep := report.New()

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

	if opts.SheetIndex < 1 {
		opts.SheetIndex = 1
	}
	if opts.HeaderRows < 0 {
		opts.HeaderRows = 0
	}
	if opts.ColumnAlign == "" {
		opts.ColumnAlign = model.AlignByHeader
	}
	if opts.SourceColumnLabel == "" {
		opts.SourceColumnLabel = DefaultSourceColumnLabel
	}
	if opts.OutputNameBase == "" {
		opts.OutputNameBase = OutPrefix + time.Now().Format("20060102_150405")
	}
	// KeepBlankRows defaults to true when the field is left at zero value by
	// older callers; the UI sends an explicit flag. Use pointer-like sentinel:
	// we treat missing as true by flipping before use if a companion flag is set.
	// For simplicity: always honor the bool; UI default is true.

	seen := map[string]bool{}
	var paths []string
	for _, p := range opts.Paths {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		return res, fmt.Errorf("E_EMPTY: no input files")
	}
	total := len(paths)

	var (
		outWb          *excelize.File
		outSheet       string
		baseExt        = ".xlsx"
		maxCols        int
		nextRow        int
		baseHeader     []string
		firstFileName  string
		firstFileRows  int // last used row of the base file
	)

	defer func() {
		if outWb != nil {
			_ = outWb.Close()
		}
	}()

	for i, path := range paths {
		if err := ctx.Err(); err != nil {
			return res, fmt.Errorf("E_CANCEL: cancelled")
		}
		name := filepath.Base(path)
		say("merge", i, total, name)

		if strings.HasPrefix(name, "~$") {
			rep.Skipped(name, "E_LOCKED: Excel temp/lock file")
			continue
		}
		if strings.HasPrefix(name, OutPrefix) {
			rep.Skipped(name, "skipped: prior tool output")
			continue
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".xlsx" && ext != ".xlsm" {
			rep.Failed(name, "E_FORMAT: unsupported extension")
			continue
		}

		srcWb, err := excelize.OpenFile(path)
		if err != nil {
			rep.Failed(name, "E_FORMAT: cannot open: "+shortErr(err))
			continue
		}

		sheetNames := srcWb.GetSheetList()
		if len(sheetNames) < opts.SheetIndex {
			srcWb.Close()
			rep.Skipped(name, fmt.Sprintf("only %d sheets, need index %d", len(sheetNames), opts.SheetIndex))
			continue
		}
		srcSheet := sheetNames[opts.SheetIndex-1]

		usedRows, usedCols, err := usedExtent(srcWb, srcSheet)
		if err != nil || usedRows == 0 || usedCols == 0 {
			srcWb.Close()
			rep.Skipped(name, "E_EMPTY: target sheet has no data")
			continue
		}
		if usedCols > maxReasonableCols {
			usedCols = maxReasonableCols
		}

		// First data file is the live base (keeps filters, theme, widths, other sheets).
		if outWb == nil {
			outWb = srcWb
			outSheet = srcSheet
			if strings.EqualFold(ext, ".xlsm") {
				baseExt = ".xlsm"
			}
			if idx, err := outWb.GetSheetIndex(outSheet); err == nil && idx > 0 {
				outWb.SetActiveSheet(idx)
			}
			maxCols = usedCols
			if opts.HeaderRows >= 1 {
				baseHeader = readHeaderRow(outWb, outSheet, 1, usedCols)
			}
			rep.Success(name, usedRows)
			firstFileName = name
			firstFileRows = usedRows
			nextRow = usedRows + 1
			continue
		}

		srcStart := opts.HeaderRows + 1
		if srcStart > usedRows {
			srcWb.Close()
			rep.Skipped(name, fmt.Sprintf("no data after %d header rows", opts.HeaderRows))
			continue
		}

		var srcHeader []string
		if opts.HeaderRows >= 1 {
			srcHeader = readHeaderRow(srcWb, srcSheet, 1, usedCols)
		}

		colMap, newMax, alignErr := mapColumns(srcHeader, baseHeader, usedCols, maxCols, opts.ColumnAlign)
		if alignErr != nil {
			srcWb.Close()
			rep.Failed(name, alignErr.Error())
			if opts.ColumnAlign == model.AlignAbort {
				break
			}
			continue
		}
		if colMap != nil && opts.HeaderRows >= 1 {
			for sc, dc := range colMap {
				if dc > maxCols && sc-1 < len(srcHeader) && srcHeader[sc-1] != "" {
					_ = outWb.SetCellValue(outSheet, cellName(dc, 1), srcHeader[sc-1])
				}
			}
			if newMax > len(baseHeader) {
				grown := make([]string, newMax)
				copy(grown, baseHeader)
				for sc, dc := range colMap {
					if dc > maxCols && sc-1 < len(srcHeader) {
						grown[dc-1] = srcHeader[sc-1]
					}
				}
				baseHeader = grown
			}
		}
		if newMax > maxCols {
			maxCols = newMax
		}

		sourceCol := 0
		if opts.AddSourceFile {
			sourceCol = maxCols + 1
		}

		// Carry column width + hidden from source columns into dest (same-index or mapped).
		copyColumnMeta(srcWb, srcSheet, outWb, outSheet, usedCols, colMap)

		sm := newSafeStyleMapper(srcWb, outWb)
		firstDst := nextRow
		for r := srcStart; r <= usedRows; r++ {
			if err := ctx.Err(); err != nil {
				srcWb.Close()
				return res, fmt.Errorf("E_CANCEL: cancelled")
			}
			if !opts.KeepBlankRows && rowIsEmpty(srcWb, srcSheet, r, usedCols) {
				continue
			}
			if h, herr := srcWb.GetRowHeight(srcSheet, r); herr == nil && h > 0 {
				_ = outWb.SetRowHeight(outSheet, nextRow, h)
			}
			for c := 1; c <= usedCols; c++ {
				dstCol := c
				if colMap != nil {
					mapped, ok := colMap[c]
					if !ok || mapped == 0 {
						continue
					}
					dstCol = mapped
				}
				copyCellSafe(srcWb, srcSheet, r, c, outWb, outSheet, nextRow, dstCol, sm)
			}
			if sourceCol > 0 {
				_ = outWb.SetCellValue(outSheet, cellName(sourceCol, nextRow), name)
			}
			// Copy merged-cell regions that start on this source row.
			copyMergesForRow(srcWb, srcSheet, outWb, outSheet, r, nextRow, colMap)
			nextRow++
		}

		if nextRow > firstDst {
			rep.Success(name, nextRow-firstDst)
		} else {
			rep.Skipped(name, "E_EMPTY: no rows appended")
		}
		srcWb.Close()
	}

	if outWb == nil {
		return res, fmt.Errorf("E_EMPTY: no usable data in any file")
	}

	if opts.AddSourceFile {
		srcCol := maxCols + 1
		_ = outWb.SetCellValue(outSheet, cellName(srcCol, 1), opts.SourceColumnLabel)
		_ = outWb.SetColWidth(outSheet, colLetter(srcCol), colLetter(srcCol), 18)
		// Base file rows never went through the append loop — fill their source name.
		if firstFileName != "" && firstFileRows > 0 {
			start := 1
			if opts.HeaderRows >= 1 {
				start = opts.HeaderRows + 1
			}
			for r := start; r <= firstFileRows; r++ {
				if !opts.KeepBlankRows && rowIsEmpty(outWb, outSheet, r, maxCols) {
					continue
				}
				_ = outWb.SetCellValue(outSheet, cellName(srcCol, r), firstFileName)
			}
		}
		maxCols = srcCol
	}

	// Keep base filter if present; extend used range hint via autofilter refresh when possible.
	// Do NOT delete other sheets — Excel filters/links stay valid (B-B06).
	if !opts.KeepBaseOtherSheets {
		// Soft note only; still keep sheets to avoid #REF! and repair issues.
	}

	if idx, err := outWb.GetSheetIndex(outSheet); err == nil && idx > 0 {
		outWb.SetActiveSheet(idx)
		// Move merged sheet to the front (before first other sheet).
		if first := outWb.GetSheetName(0); first != outSheet {
			_ = outWb.MoveSheet(outSheet, first)
		}
	}

	outDir := opts.OutputDir
	if outDir == "" {
		outDir = filepath.Dir(paths[0])
	}
	baseOut := ".xlsx"
	if strings.EqualFold(baseExt, ".xlsm") {
		baseOut = ".xlsm"
	}
	outPath := fs.UniquePath(filepath.Join(outDir, opts.OutputNameBase+baseOut))

	say("write", total, total, filepath.Base(outPath))
	if err := saveWorkbook(ctx, outWb, outPath); err != nil {
		return res, err
	}

	okCount, skippedCount, failedCount := rep.Counts()
	res.Success = okCount > 0
	res.Partial = okCount > 0 && (skippedCount > 0 || failedCount > 0)
	res.OutputPath = outPath
	res.FilesUsed = okCount
	res.Rows = nextRow - 1
	res.Cols = maxCols
	res.DurationMs = time.Since(start).Milliseconds()
	res.Items = rep.Items()
	switch {
	case !res.Success:
		res.Message = "E_EMPTY: no files produced data"
	case res.Partial:
		res.Message = fmt.Sprintf("partial: %d ok, %d skipped, %d failed", okCount, skippedCount, failedCount)
	default:
		res.Message = fmt.Sprintf("ok: %d files, %d rows", okCount, res.Rows)
	}
	say("done", 100, 100, "ok")
	return res, nil
}

func rowIsEmpty(f *excelize.File, sheet string, row, cols int) bool {
	for c := 1; c <= cols; c++ {
		v, err := f.GetCellValue(sheet, cellName(c, row))
		if err == nil && strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

func copyColumnMeta(src *excelize.File, srcSheet string, dst *excelize.File, dstSheet string, srcCols int, colMap map[int]int) {
	for c := 1; c <= srcCols; c++ {
		dstCol := c
		if colMap != nil {
			mapped, ok := colMap[c]
			if !ok || mapped == 0 {
				continue
			}
			dstCol = mapped
		}
		letter := colLetter(c)
		if w, err := src.GetColWidth(srcSheet, letter); err == nil && w > 0 {
			_ = dst.SetColWidth(dstSheet, colLetter(dstCol), colLetter(dstCol), w)
		}
		vis, err := src.GetColVisible(srcSheet, letter)
		if err == nil && !vis {
			_ = dst.SetColVisible(dstSheet, colLetter(dstCol), false)
		}
	}
}

func copyMergesForRow(src *excelize.File, srcSheet string, dst *excelize.File, dstSheet string, srcRow, dstRow int, colMap map[int]int) {
	merges, err := src.GetMergeCells(srcSheet)
	if err != nil {
		return
	}
	for _, m := range merges {
		sc, sr, err1 := excelize.CellNameToCoordinates(m.GetStartAxis())
		ec, er, err2 := excelize.CellNameToCoordinates(m.GetEndAxis())
		if err1 != nil || err2 != nil {
			continue
		}
		if sr != srcRow {
			continue // only start-row merges in this block
		}
		dc1, dc2 := sc, ec
		if colMap != nil {
			a, ok1 := colMap[sc]
			b, ok2 := colMap[ec]
			if !ok1 || a == 0 || !ok2 || b == 0 {
				continue
			}
			dc1, dc2 = a, b
		}
		if dc1 > dc2 {
			dc1, dc2 = dc2, dc1
		}
		_ = dst.MergeCell(dstSheet, cellName(dc1, dstRow), cellName(dc2, dstRow+er-sr))
	}
}

func usedExtent(f *excelize.File, sheet string) (rows, cols int, err error) {
	data, err := f.GetRows(sheet)
	if err != nil {
		return 0, 0, err
	}
	for i, row := range data {
		last := 0
		for c, v := range row {
			if strings.TrimSpace(v) != "" {
				last = c + 1
			}
		}
		if last > 0 {
			rows = i + 1
			if last > cols {
				cols = last
			}
		}
	}
	// Sheet dimension can exceed GetRows when trailing rows only have styles,
	// formulas that evaluate empty, or sparse cells GetRows trims.
	if dimRows, dimCols, derr := dimensionExtent(f, sheet); derr == nil {
		if dimRows > rows && dimRows <= 100000 {
			// Probe up to dimRows for any non-empty or formula cell.
			for r := rows + 1; r <= dimRows; r++ {
				if rowHasContent(f, sheet, r, dimCols) {
					rows = r
				}
			}
		}
		if dimCols > cols && dimCols <= maxReasonableCols {
			for c := cols + 1; c <= dimCols; c++ {
				for r := 1; r <= rows; r++ {
					if v, e := f.GetCellValue(sheet, cellName(c, r)); e == nil && strings.TrimSpace(v) != "" {
						cols = c
						break
					}
					if form, e := f.GetCellFormula(sheet, cellName(c, r)); e == nil && form != "" {
						cols = c
						break
					}
				}
			}
		}
	}
	return rows, cols, nil
}

func dimensionExtent(f *excelize.File, sheet string) (rows, cols int, err error) {
	dim, err := f.GetSheetDimension(sheet)
	if err != nil || dim == "" {
		return 0, 0, nil
	}
	parts := strings.Split(dim, ":")
	end := parts[len(parts)-1]
	c, r, err := excelize.CellNameToCoordinates(end)
	if err != nil {
		return 0, 0, nil
	}
	return r, c, nil
}

func rowHasContent(f *excelize.File, sheet string, row, cols int) bool {
	limit := cols
	if limit > maxReasonableCols {
		limit = maxReasonableCols
	}
	for c := 1; c <= limit; c++ {
		addr := cellName(c, row)
		if v, e := f.GetCellValue(sheet, addr); e == nil && strings.TrimSpace(v) != "" {
			return true
		}
		if form, e := f.GetCellFormula(sheet, addr); e == nil && form != "" {
			return true
		}
	}
	return false
}

func readHeaderRow(f *excelize.File, sheet string, row, cols int) []string {
	if cols < 1 {
		return nil
	}
	if cols > maxReasonableCols {
		cols = maxReasonableCols
	}
	out := make([]string, cols)
	for c := 1; c <= cols; c++ {
		v, err := f.GetCellValue(sheet, cellName(c, row))
		if err != nil {
			continue
		}
		out[c-1] = normalizeHeader(v)
	}
	for len(out) > 1 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}

func normalizeHeader(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "")
	s = strings.ReplaceAll(s, "　", " ")
	s = strings.ReplaceAll(s, " ", " ")
	return strings.TrimSpace(s)
}

func mapColumns(srcHeader, baseHeader []string, srcCols, maxCols int, mode model.ColumnAlign) (map[int]int, int, error) {
	if mode == model.AlignByPosition || mode == "" {
		if srcCols > maxCols {
			return nil, srcCols, nil
		}
		return nil, maxCols, nil
	}
	if len(baseHeader) == 0 {
		if srcCols > maxCols {
			return nil, srcCols, nil
		}
		return nil, maxCols, nil
	}

	baseMap := map[string]int{}
	for i, h := range baseHeader {
		h = normalizeHeader(h)
		if h == "" {
			continue
		}
		if _, ok := baseMap[h]; !ok {
			baseMap[h] = i + 1
		}
	}

	out := make(map[int]int, srcCols)
	nextNew := maxCols + 1
	for c := 1; c <= srcCols; c++ {
		h := ""
		if c-1 < len(srcHeader) {
			h = srcHeader[c-1]
		}
		if h == "" {
			continue
		}
		if col, ok := baseMap[h]; ok {
			out[c] = col
			continue
		}
		if mode == model.AlignAbort {
			return nil, maxCols, fmt.Errorf("E_HEADER_MISMATCH: column %q not in base header", h)
		}
		baseMap[h] = nextNew
		out[c] = nextNew
		nextNew++
	}
	newMax := maxCols
	if nextNew-1 > newMax {
		newMax = nextNew - 1
	}
	return out, newMax, nil
}

func saveWorkbook(ctx context.Context, f *excelize.File, destPath string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("E_CANCEL: cancelled")
	}
	dir := filepath.Dir(destPath)
	tmp, err := fs.TempPathIn(dir, "exceltools-files-*.xlsx")
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

func cellName(col, row int) string {
	s, _ := excelize.CoordinatesToCellName(col, row)
	return s
}

func colLetter(col int) string {
	s, _ := excelize.ColumnNumberToName(col)
	return s
}

func shortErr(err error) string {
	s := err.Error()
	if len(s) > 120 {
		return s[:117] + "..."
	}
	return s
}
