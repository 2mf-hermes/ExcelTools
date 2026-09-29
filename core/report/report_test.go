package report

import "testing"

func TestBuilderCountsAndPartial(t *testing.T) {
	b := New()
	b.Success("a.xlsx", 10)
	b.Skipped("b.xlsx", "empty")
	b.Failed("c.xlsx", "locked")
	ok, skipped, failed := b.Counts()
	if ok != 1 || skipped != 1 || failed != 1 {
		t.Fatalf("counts %d %d %d", ok, skipped, failed)
	}
	if !b.Partial() {
		t.Fatal("expected partial")
	}

	b2 := New()
	b2.Success("a.xlsx", 1)
	if b2.Partial() {
		t.Fatal("single success is not partial")
	}
}
