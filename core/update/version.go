package update

import (
	"strconv"
	"strings"
)

// Version is a parsed semantic version with an optional pre-release suffix.
//
// The project tags releases as "v0.5.0-m5" (milestone 5 of 0.5.0), so the
// suffix carries real ordering meaning and cannot be discarded.
type Version struct {
	Major  int
	Minor  int
	Patch  int
	Suffix string // e.g. "m5"; empty for a plain release.
}

// ParseVersion parses "v0.5.0-m5", "0.4.0-m4", "0.6.0", "1.2.3-rc1" and
// similar. A leading "v"/"V" is ignored. Unparseable numeric segments become 0
// rather than erroring, so a malformed tag can never panic the caller.
func ParseVersion(s string) Version {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")

	var suffix string
	if i := strings.IndexByte(s, '-'); i >= 0 {
		suffix = s[i+1:]
		s = s[:i]
	}

	parts := strings.Split(s, ".")
	var nums [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		nums[i] = leadingInt(parts[i])
	}
	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2], Suffix: suffix}
}

// Compare reports whether a sorts below (-1), equal to (0), or above (1) b.
func Compare(a, b string) int {
	return ParseVersion(a).Compare(ParseVersion(b))
}

// Compare orders two parsed versions.
//
// Ordering rules, applied in sequence:
//  1. numeric major.minor.patch
//  2. a version with no suffix outranks one with a suffix (0.5.0 > 0.5.0-m1)
//  3. otherwise compare suffixes with digit runs compared numerically (m10 > m9)
func (a Version) Compare(b Version) int {
	if c := cmpInt(a.Major, b.Major); c != 0 {
		return c
	}
	if c := cmpInt(a.Minor, b.Minor); c != 0 {
		return c
	}
	if c := cmpInt(a.Patch, b.Patch); c != 0 {
		return c
	}
	switch {
	case a.Suffix == "" && b.Suffix == "":
		return 0
	case a.Suffix == "":
		return 1
	case b.Suffix == "":
		return -1
	}
	return compareSuffix(a.Suffix, b.Suffix)
}

// compareSuffix compares pre-release suffixes case-insensitively, treating
// embedded digit runs as numbers so "m10" sorts after "m9". A plain string
// comparison would place "m10" before "m9".
func compareSuffix(a, b string) int {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		ca, cb := a[i], b[j]
		if isDigit(ca) && isDigit(cb) {
			runA, nextA := digitRun(a, i)
			runB, nextB := digitRun(b, j)
			va, _ := strconv.Atoi(runA)
			vb, _ := strconv.Atoi(runB)
			if c := cmpInt(va, vb); c != 0 {
				return c
			}
			i, j = nextA, nextB
			continue
		}
		la, lb := lowerByte(ca), lowerByte(cb)
		if la != lb {
			if la < lb {
				return -1
			}
			return 1
		}
		i++
		j++
	}
	switch {
	case i < len(a):
		return 1
	case j < len(b):
		return -1
	}
	return 0
}

// IsNewer reports whether candidate is strictly newer than current.
func IsNewer(candidate, current string) bool {
	return Compare(candidate, current) > 0
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// leadingInt parses the leading run of digits in s, ignoring any trailing
// non-digit characters ("12abc" -> 12). Returns 0 when there is no digit.
func leadingInt(s string) int {
	end := 0
	for end < len(s) && isDigit(s[end]) {
		end++
	}
	if end == 0 {
		return 0
	}
	n, err := strconv.Atoi(s[:end])
	if err != nil {
		return 0
	}
	return n
}

// digitRun returns the maximal digit substring starting at i and the index just
// past it.
func digitRun(s string, i int) (string, int) {
	j := i
	for j < len(s) && isDigit(s[j]) {
		j++
	}
	return s[i:j], j
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func lowerByte(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}
