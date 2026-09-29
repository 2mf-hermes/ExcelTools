// Package report builds success/skip/fail summaries for merge results.
package report

import "ExcelTools/core/model"

// Builder accumulates per-file outcomes.
type Builder struct {
	items []model.ItemReport
}

// New creates an empty builder.
func New() *Builder {
	return &Builder{items: make([]model.ItemReport, 0, 8)}
}

// Success records a successful file.
func (b *Builder) Success(file string, rows int) {
	b.items = append(b.items, model.ItemReport{
		File: file, Status: model.ItemSuccess, Rows: rows,
	})
}

// Skipped records a skipped file with a short reason.
func (b *Builder) Skipped(file, reason string) {
	b.items = append(b.items, model.ItemReport{
		File: file, Status: model.ItemSkipped, Message: reason,
	})
}

// Failed records a failed file with a short reason.
func (b *Builder) Failed(file, reason string) {
	b.items = append(b.items, model.ItemReport{
		File: file, Status: model.ItemFailed, Message: reason,
	})
}

// Items returns accumulated reports.
func (b *Builder) Items() []model.ItemReport {
	return b.items
}

// Counts returns success, skipped, failed totals.
func (b *Builder) Counts() (ok, skipped, failed int) {
	for _, it := range b.items {
		switch it.Status {
		case model.ItemSuccess:
			ok++
		case model.ItemSkipped:
			skipped++
		case model.ItemFailed:
			failed++
		}
	}
	return
}

// Partial reports whether any file succeeded while others did not.
func (b *Builder) Partial() bool {
	ok, skipped, failed := b.Counts()
	return ok > 0 && (skipped > 0 || failed > 0)
}
