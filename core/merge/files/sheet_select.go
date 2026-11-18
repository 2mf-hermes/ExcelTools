package files

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/xuri/excelize/v2"
)

// SheetMode selects how the source worksheet is chosen for each file.
type SheetMode string

const (
	// SheetModeAuto picks the first worksheet that has data, falling back to the
	// last worksheet when every earlier one is blank. Cover sheets, notes pages
	// and other empty tabs are skipped automatically.
	SheetModeAuto SheetMode = "auto"
	// SheetModeFirst always uses the first worksheet by position.
	SheetModeFirst SheetMode = "first"
	// SheetModeIndex uses the SheetIndex-th worksheet (1-based); SheetModeIndex
	// is also the default when the option is left blank.
	SheetModeIndex SheetMode = "index"
	// SheetModeName uses the worksheet with a matching name.
	SheetModeName SheetMode = "name"
)

// IsValidSheetMode reports whether m is a known mode. An empty mode is allowed
// and is treated as SheetModeIndex.
func IsValidSheetMode(m SheetMode) bool {
	switch m {
	case "", SheetModeAuto, SheetModeFirst, SheetModeIndex, SheetModeName:
		return true
	}
	return false
}

// SheetChoice overrides the global sheet mode for one file.
//
// Name takes precedence over Index when both are set. Name survives a sheet
// reorder; Index is the position-based fallback for files whose tab names
// differ from the base file.
type SheetChoice struct {
	Name  string
	Index int // 1-based
}

// SheetSheet maps a file path to its sheet override. Depending on user
// settings, paths may differ in case, slash direction or be relative, so the
// key is matched through normalizePath.
type SheetMap map[string]SheetChoice

// NewSheetMap returns an empty override set.
func NewSheetMap() SheetMap { return SheetMap{} }

// Set pins one file to a worksheet. The key is canonicalised, so a
// plain assignment to the map would not be found by ResolveSheet.
func (m SheetMap) Set(path string, choice SheetChoice) {
	k := normalizePath(path)
	if k == "" {
		return
	}
	m[k] = choice
}

// Lookup returns the override pinned for path, if any.
func (m SheetMap) Lookup(path string) (SheetChoice, bool) {
	ov, ok := m[normalizePath(path)]
	return ov, ok
}

// normalizePath canonicalises a file path for override lookup. Windows paths
// are ASCII, so simple lower-casing is enough and stays valid UTF-8.
func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	p = filepath.Clean(p)
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	return strings.ToLower(strings.ReplaceAll(p, "\\", "/"))
}

func sameName(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == b {
		return true
	}
	equalFold := strings.EqualFold(a, b)
	if !equalFold {
		return false
	}
	// Only ASCII letter case is treated as the same name; non-ASCII folding is
	// too broad (e.g. different localised letters) to merge silently.
	for _, s := range []string{a, b} {
		for _, r := range s {
			if r > unicode.MaxASCII {
				return false
			}
		}
	}
	return true
}

// ManifestEntry describes which worksheet one file contributed to the merge.
type ManifestEntry struct {
	Path  string
	Name  string // file name
	Sheet string // worksheet used
	Index int    // 1-based position of that worksheet
	Error string // why no sheet was used
}

// ResolveSheet decides which worksheet of wb a file contributes.
//
//   - A per-file override with a Name wins when that name exists; otherwise its
//     Index is used, so a renamed tab still resolves.
//   - A per-file override with only an Index uses that position.
//   - Otherwise mode applies: auto skips blank sheets, first uses position 1,
//     index uses idx.
//
// Only visible sheets are considered by auto and first; choosing an index or
// name directly still works for hidden sheets, because the user asked for them.
func ResolveSheet(wb *excelize.File, path string, mode SheetMode, idx int, overrides SheetMap) (string, int, error) {
	names := wb.GetSheetList()
	if len(names) == 0 {
		return "", 0, fmt.Errorf("E_EMPTY: workbook has no worksheets")
	}

	var ov SheetChoice
	hasOv := false
	if len(overrides) > 0 {
		ov, hasOv = overrides.Lookup(path)
		hasOv = hasOv && (strings.TrimSpace(ov.Name) != "" || ov.Index > 0)
	}
	if hasOv && strings.TrimSpace(ov.Name) != "" {
		want := strings.TrimSpace(ov.Name)
		for i, n := range names {
			if n == want {
				return n, i + 1, nil
			}
		}
		for i, n := range names {
			if sameName(n, want) {
				return n, i + 1, nil
			}
		}
		if ov.Index > 0 && ov.Index <= len(names) {
			// The file was re-ordered or renamed; fall back to the position.
			return names[ov.Index-1], ov.Index, nil
		}
		return "", 0, fmt.Errorf("E_NO_SHEET: sheet %q not found", want)
	}
	if hasOv {
		if ov.Index > len(names) {
			return "", 0, fmt.Errorf("E_NO_SHEET: only %d sheets, need index %d", len(names), ov.Index)
		}
		return names[ov.Index-1], ov.Index, nil
	}

	switch mode {
	case SheetModeAuto:
		for i, n := range names {
			if sheetIsBlank(wb, n) {
				continue
			}
			if vis, err := wb.GetSheetVisible(n); err == nil && !vis {
				continue
			}
			return n, i + 1, nil
		}
		// Every visible sheet is blank: keep the old behaviour and use the first.
		return names[0], 1, nil
	case SheetModeFirst:
		return names[0], 1, nil
	default:
		if idx < 1 {
			idx = 1
		}
		if idx > len(names) {
			return "", 0, fmt.Errorf("E_NO_SHEET: only %d sheets, need index %d", len(names), idx)
		}
		return names[idx-1], idx, nil
	}
}

// sheetIsBlank reports whether a worksheet holds no value or formula.
func sheetIsBlank(wb *excelize.File, sheet string) bool {
	rows, cols, err := usedExtent(wb, sheet)
	if err != nil {
		return false
	}
	return rows == 0 || cols == 0
}
