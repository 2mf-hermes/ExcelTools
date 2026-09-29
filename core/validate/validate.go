// Package validate checks user input before merge runs.
package validate

import (
	"fmt"
	"path/filepath"
	"strings"

	"ExcelTools/core/model"
)

// SupportedExts lists readable workbook extensions.
var SupportedExts = []string{".xlsx", ".xlsm", ".xls"}

// ExtSupported reports whether ext is a supported workbook extension.
func ExtSupported(ext string) bool {
	e := strings.ToLower(filepath.Ext(ext))
	for _, s := range SupportedExts {
		if e == s {
			return true
		}
	}
	return false
}

// ExtOf returns the lowercase extension including the leading dot.
func ExtOf(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

// ValidateSheetOptions checks single-file merge options.
func ValidateSheetOptions(sheets []model.SheetRef, opts model.SheetMergeOptions) error {
	if len(sheets) == 0 {
		return fmt.Errorf("E_EMPTY: no sheets listed")
	}
	selected := 0
	for _, s := range sheets {
		if !s.Empty {
			selected++
		}
	}
	if selected == 0 && !opts.SkipEmpty {
		return fmt.Errorf("E_EMPTY: all sheets look empty")
	}
	switch opts.HeaderMode {
	case model.HeaderEach, model.HeaderFirst, model.HeaderNone, "":
	default:
		return fmt.Errorf("E_INTERNAL: unknown header mode %q", opts.HeaderMode)
	}
	return nil
}

// ValidateFileOptions checks multi-file merge options.
func ValidateFileOptions(files []model.FileRef, opts model.FileMergeOptions) error {
	ok := 0
	for _, f := range files {
		if f.Status == model.FileOK || f.Status == "" {
			ok++
		}
	}
	if ok == 0 {
		return fmt.Errorf("E_EMPTY: no usable files")
	}
	if opts.SheetIndex < 1 {
		return fmt.Errorf("E_INTERNAL: sheet index must be >= 1")
	}
	if opts.HeaderRows < 0 {
		return fmt.Errorf("E_INTERNAL: header rows must be >= 0")
	}
	switch opts.ColumnAlign {
	case model.AlignByPosition, model.AlignByHeader, model.AlignAbort, "":
	default:
		return fmt.Errorf("E_INTERNAL: unknown column align %q", opts.ColumnAlign)
	}
	return nil
}

// IsTempOrToolOutput reports Excel lock files and our own prior outputs.
func IsTempOrToolOutput(name, outPrefix string) bool {
	base := filepath.Base(name)
	if strings.HasPrefix(base, "~$") {
		return true
	}
	if outPrefix != "" && strings.HasPrefix(base, outPrefix) {
		return true
	}
	return false
}
