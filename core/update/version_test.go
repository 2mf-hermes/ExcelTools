package update

import "testing"

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in                  string
		major, minor, patch int
		suffix              string
	}{
		{"v0.5.0-m5", 0, 5, 0, "m5"},
		{"0.4.0-m4", 0, 4, 0, "m4"},
		{"0.6.0", 0, 6, 0, ""},
		{"V1.2.3-rc1", 1, 2, 3, "rc1"},
		{"  0.5.0-m5  ", 0, 5, 0, "m5"},
		{"1.2", 1, 2, 0, ""},
		{"", 0, 0, 0, ""},
		{"garbage", 0, 0, 0, ""},
		{"1.2.3.4", 1, 2, 3, ""},
	}
	for _, c := range cases {
		got := ParseVersion(c.in)
		if got.Major != c.major || got.Minor != c.minor || got.Patch != c.patch || got.Suffix != c.suffix {
			t.Errorf("ParseVersion(%q) = {%d %d %d %q}, want {%d %d %d %q}",
				c.in, got.Major, got.Minor, got.Patch, got.Suffix,
				c.major, c.minor, c.patch, c.suffix)
		}
	}
}

func TestCompareOrdering(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		// Numeric triples dominate.
		{"0.4.0-m4", "0.5.0-m5", -1},
		{"0.5.0-m5", "0.4.0-m4", 1},
		{"0.5.1", "0.5.0", 1},
		{"1.0.0", "0.9.9", 1},
		{"0.6.0", "0.5.0-m5", 1},

		// Equal, including the optional "v" prefix and suffix case.
		{"0.4.0-m4", "0.4.0-m4", 0},
		{"v0.4.0-m4", "0.4.0-m4", 0},
		{"0.5.0-M5", "0.5.0-m5", 0},

		// A plain release outranks its own pre-releases.
		{"0.5.0", "0.5.0-m5", 1},
		{"0.5.0-m5", "0.5.0", -1},

		// Digit runs in the suffix compare numerically, not lexically.
		{"0.5.0-m9", "0.5.0-m10", -1},
		{"0.5.0-m10", "0.5.0-m9", 1},
		{"0.5.0-m2", "0.5.0-m10", -1},

		// Unparseable input degrades to 0.0.0 instead of panicking.
		{"garbage", "0.0.0", 0},
		{"", "0.0.0", 0},
	}
	for _, c := range cases {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
		if got := Compare(c.b, c.a); got != -c.want {
			t.Errorf("Compare(%q, %q) = %d, want %d (antisymmetry)", c.b, c.a, got, -c.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		candidate, current string
		want               bool
	}{
		{"0.5.0-m5", "0.4.0-m4", true},
		{"0.4.0-m4", "0.4.0-m4", false},
		{"0.4.0-m4", "0.5.0-m5", false},
		{"", "0.5.0-m5", false},
		{"0.5.0-m5", "", true},
	}
	for _, c := range cases {
		if got := IsNewer(c.candidate, c.current); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.candidate, c.current, got, c.want)
		}
	}
}
