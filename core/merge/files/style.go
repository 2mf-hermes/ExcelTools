package files

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// safeStyleMapper copies conservative style properties into a fresh workbook.
// It avoids raw style IDs and theme-dependent fields that can produce
// repairable-but-invalid OOXML.
type safeStyleMapper struct {
	src   *excelize.File
	dst   *excelize.File
	cache map[int]int
}

func newSafeStyleMapper(src, dst *excelize.File) *safeStyleMapper {
	return &safeStyleMapper{src: src, dst: dst, cache: map[int]int{}}
}

func (m *safeStyleMapper) dstStyleID(srcStyleID int) int {
	if srcStyleID == 0 {
		return 0
	}
	if id, ok := m.cache[srcStyleID]; ok {
		return id
	}
	st, err := m.src.GetStyle(srcStyleID)
	if err != nil || st == nil {
		m.cache[srcStyleID] = 0
		return 0
	}
	safe := sanitizeStyle(st)
	id, err := m.dst.NewStyle(safe)
	if err != nil {
		m.cache[srcStyleID] = 0
		return 0
	}
	m.cache[srcStyleID] = id
	return id
}

// sanitizeStyle keeps only portable RGB-style fields.
func sanitizeStyle(st *excelize.Style) *excelize.Style {
	out := &excelize.Style{}
	if st.NumFmt != 0 {
		out.NumFmt = st.NumFmt
	}
	if st.CustomNumFmt != nil && *st.CustomNumFmt != "" {
		out.CustomNumFmt = st.CustomNumFmt
	}
	if st.Font != nil {
		f := &excelize.Font{
			Bold:      st.Font.Bold,
			Italic:    st.Font.Italic,
			Underline: st.Font.Underline,
			Size:      st.Font.Size,
			Family:    st.Font.Family,
			Strike:    st.Font.Strike,
		}
		if isRGBColor(st.Font.Color) {
			f.Color = ensureHexColor(st.Font.Color)
		}
		out.Font = f
	}
	if st.Fill.Type != "" {
		fill := excelize.Fill{Type: st.Fill.Type, Pattern: st.Fill.Pattern}
		colors := make([]string, 0, len(st.Fill.Color))
		for _, c := range st.Fill.Color {
			if isRGBColor(c) {
				colors = append(colors, ensureHexColor(c))
			}
		}
		if len(colors) > 0 {
			fill.Color = colors
			out.Fill = fill
		}
	}
	if st.Alignment != nil {
		out.Alignment = &excelize.Alignment{
			Horizontal:   st.Alignment.Horizontal,
			Vertical:     st.Alignment.Vertical,
			WrapText:     st.Alignment.WrapText,
			TextRotation: st.Alignment.TextRotation,
			Indent:       st.Alignment.Indent,
		}
	}
	if len(st.Border) > 0 {
		var borders []excelize.Border
		for _, b := range st.Border {
			if b.Style == 0 {
				continue
			}
			nb := excelize.Border{Type: b.Type, Style: b.Style}
			if isRGBColor(b.Color) {
				nb.Color = ensureHexColor(b.Color)
			} else {
				nb.Color = "#FF000000"
			}
			borders = append(borders, nb)
		}
		if len(borders) > 0 {
			out.Border = borders
		}
	}
	return out
}

func isRGBColor(c string) bool {
	c = strings.TrimSpace(strings.TrimPrefix(c, "#"))
	// Accept RRGGBB or AARRGGBB
	if len(c) != 6 && len(c) != 8 {
		return false
	}
	for _, r := range c {
		isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
		if !isHex {
			return false
		}
	}
	return true
}

func ensureHexColor(c string) string {
	c = strings.TrimSpace(c)
	if c == "" {
		return "FF000000"
	}
	if !strings.HasPrefix(c, "#") && isRGBColor(c) {
		if len(c) == 6 {
			return "#" + c
		}
		return "#" + c
	}
	if strings.HasPrefix(c, "#") {
		return c
	}
	return "FF000000"
}

// cellRefPattern matches A1-style refs: $A$1, A1, $A1, A$1
var cellRefPattern = regexp.MustCompile(`(\$?)([A-Za-z]{1,3})(\$?)([0-9]{1,7})`)

// offsetFormula shifts relative A1 references by (dRow, dCol).
func offsetFormula(formula string, dRow, dCol int) string {
	if dRow == 0 && dCol == 0 {
		return formula
	}
	var b strings.Builder
	b.Grow(len(formula))
	last := 0
	for _, loc := range cellRefPattern.FindAllStringIndex(formula, -1) {
		start, end := loc[0], loc[1]
		b.WriteString(formula[last:start])
		last = end

		match := formula[start:end]
		if end < len(formula) && formula[end] == '(' {
			b.WriteString(match)
			continue
		}
		if start > 0 {
			c := formula[start-1]
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_' {
				b.WriteString(match)
				continue
			}
		}

		sub := cellRefPattern.FindStringSubmatch(match)
		if sub == nil {
			b.WriteString(match)
			continue
		}
		dollarCol, col, dollarRow, rowStr := sub[1], sub[2], sub[3], sub[4]
		colNum := colNameToNumber(col)
		if colNum < 1 || colNum > 16384 {
			b.WriteString(match)
			continue
		}
		r, err := strconv.Atoi(rowStr)
		if err != nil || r < 1 || r > 1048576 {
			b.WriteString(match)
			continue
		}
		if dollarCol == "" {
			colNum += dCol
			if colNum < 1 {
				colNum = 1
			}
		}
		if dollarRow == "" {
			r += dRow
			if r < 1 {
				r = 1
			}
		}
		b.WriteString(dollarCol)
		b.WriteString(numberToColName(colNum))
		b.WriteString(dollarRow)
		b.WriteString(strconv.Itoa(r))
	}
	b.WriteString(formula[last:])
	return b.String()
}

func colNameToNumber(name string) int {
	n := 0
	for _, c := range strings.ToUpper(name) {
		if c < 'A' || c > 'Z' {
			return 0
		}
		n = n*26 + int(c-'A'+1)
	}
	return n
}

func numberToColName(n int) string {
	if n < 1 {
		n = 1
	}
	s := ""
	for n > 0 {
		n--
		s = string(rune('A'+n%26)) + s
		n /= 26
	}
	return s
}

// copyCellSafe copies value/formula + sanitized style into a clean workbook.
func copyCellSafe(
	src *excelize.File, srcSheet string, srcRow, srcCol int,
	dst *excelize.File, dstSheet string, dstRow, dstCol int,
	sm *safeStyleMapper,
) {
	from := cellName(srcCol, srcRow)
	to := cellName(dstCol, dstRow)

	if sid, err := src.GetCellStyle(srcSheet, from); err == nil && sid != 0 && sm != nil {
		if did := sm.dstStyleID(sid); did != 0 {
			_ = dst.SetCellStyle(dstSheet, to, to, did)
		}
	}

	formula, err := src.GetCellFormula(srcSheet, from)
	if err == nil && formula != "" {
		if formulaIsSameSheetOnly(formula) {
			shifted := offsetFormula(formula, dstRow-srcRow, dstCol-srcCol)
			if ferr := dst.SetCellFormula(dstSheet, to, shifted); ferr == nil {
				return
			}
		}
		if val, verr := src.GetCellValue(srcSheet, from); verr == nil && val != "" {
			_ = dst.SetCellValue(dstSheet, to, val)
		}
		return
	}

	val, err := src.GetCellValue(srcSheet, from)
	if err != nil || val == "" {
		return
	}
	_ = dst.SetCellValue(dstSheet, to, val)
}

func formulaIsSameSheetOnly(formula string) bool {
	f := strings.TrimSpace(formula)
	if f == "" {
		return false
	}
	if strings.Contains(f, "!") {
		return false
	}
	if strings.Contains(f, "[") && strings.Contains(f, "]") {
		return false
	}
	return true
}

// keep compiler happy if Pattern field naming differs across versions
var _ = excelize.Fill{}
