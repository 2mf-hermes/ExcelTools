package files

import "testing"

func TestOffsetFormula(t *testing.T) {
	cases := []struct {
		in          string
		dRow, dCol  int
		want        string
	}{
		{"B2+1", 1, 0, "B3+1"},
		{"B2+1", 0, 1, "C2+1"},
		{"$B$2+A2", 1, 0, "$B$2+A3"},
		{"SUM(A1:B2)", 2, 0, "SUM(A3:B4)"},
		{"LOG10(100)", 1, 0, "LOG10(100)"},
		{"A1", 0, 0, "A1"},
	}
	for _, c := range cases {
		got := offsetFormula(c.in, c.dRow, c.dCol)
		if got != c.want {
			t.Fatalf("offsetFormula(%q,%d,%d)=%q want %q", c.in, c.dRow, c.dCol, got, c.want)
		}
	}
}
