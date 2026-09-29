# BEHAVIOR_SPEC

Rules are derived from the two legacy tools and the locked decisions in `DECISIONS.md`.
Each rule is testable. Implementations must match this table, not ad-hoc UI wording.

## Shared

| ID | Rule |
|----|------|
| B-S01 | Open all inputs read-only. Never modify or delete user source files. |
| B-S02 | Write output to a new path. Default folder = folder of the primary source. |
| B-S03 | Default output name uses timestamp; if path exists, append `_2`, `_3`, … |
| B-S04 | Write via temp file in the same directory, then rename. On cancel/failure remove temp. |
| B-S05 | Skip Excel lock files named `~$…`. |
| B-S06 | Skip files whose name starts with tool output prefix `合併結果_`. |
| B-S07 | Natural-sort multi-file lists (`file2` before `file10`). Users may reorder manually. |
| B-S08 | Partial success: if at least one file/sheet produced data, write output and set Partial when any item failed/skipped. |
| B-S09 | If zero usable data, do not write output; return `E_EMPTY`. |
| B-S10 | `.xls` is read with reduced fidelity and always saved as `.xlsx`; UI shows a degradation notice. |

## Tool A — Sheet Combine (single workbook)

Target source: Excel Sheet Combiner (Go/excelize).

| ID | Rule |
|----|------|
| B-A01 | Accept one file: `.xlsx`, `.xlsm`, `.xls`. |
| B-A02 | List sheets with name, rows, cols, empty flag, hidden flag. |
| B-A03 | User selects and reorders sheets to merge. |
| B-A04 | Header modes: `each` (default) / `first` / `none`. |
| B-A05 | `each`: first selected non-empty sheet writes its first row as header once; later sheets skip their first row. |
| B-A06 | `first`: only the first selected sheet contributes the header row; other sheets’ first rows are data only if mode says so — when mode is `first`, subsequent sheets skip first row as well (same skip as each for data, header written once). |
| B-A07 | `none`: all rows are data; no special header write. |
| B-A08 | Optional source-sheet column: header label default `來源工作表` (localized later); placed after widest data column. |
| B-A09 | Optional skip-empty: sheets with no non-empty cells are skipped and counted. |
| B-A10 | Copy values, styles, row heights, column widths (from first used sheet), merged cells (clipped to copied range). |
| B-A11 | Preserve formulas when safe in the same workbook; otherwise write cached value and note. |
| B-A12 | Output is a new workbook containing the merged sheet (and only that sheet for pure combine). |

## Tool B — File Combine (many workbooks)

Target source: ExcelCombiner (ClosedXML).

| ID | Rule |
|----|------|
| B-B01 | Accept many files via multi-select, drag-drop, or folder scan. |
| B-B02 | On scan/collect: classify ok / skipped / failed with reasons shown in list. |
| B-B03 | Header rows: integer 0..N (default 1). First file keeps its header rows; subsequent files skip N header rows when appending. |
| B-B04 | Sheet index: 1-based. Each file contributes its Nth sheet. Files with fewer sheets are skipped with reason. |
| B-B05 | First file that has data on the target sheet becomes the output base (preserves its theme, widths, other sheets if `keepBaseOtherSheets`). |
| B-B06 | Other sheets of the base file are kept so cross-sheet formulas do not break. Merged result lives in the target sheet moved to position 1. |
| B-B07 | Column align strategies: `by-position` (legacy), `by-header` (default), `abort` (stop on mismatch). |
| B-B08 | `by-position`: map column i → column i (legacy ExcelCombiner). |
| B-B09 | `by-header`: map by header cell text of the first header row among files; unmatched columns append as new columns and are reported. |
| B-B10 | `abort`: if any subsequent file’s header row differs from the base, fail that file (or the run) with `E_HEADER_MISMATCH` and do not silently mis-align. |
| B-B11 | Optional source-file column after max data column, marking each row’s origin file. |
| B-B12 | Copy primary styles as a whole-style assignment when possible; convert theme/index colors against the source workbook before write. |
| B-B13 | Move formulas using R1C1 when possible; cross-file relative refs that cannot be safely moved become cached values + item note. |
| B-B14 | Natural sort collect order unless user reordered. |
| B-B15 | Single-file failures do not abort the run; they become Failed items. |

## Cancel

| ID | Rule |
|----|------|
| B-C01 | Cancel stops further file/sheet work promptly. |
| B-C02 | Cancel deletes temp output; no final file is left. |
| B-C03 | UI returns to input with the file list preserved for retry. |

## UI copy obligations

| ID | Rule |
|----|------|
| B-U01 | Errors show user message + one primary recovery action. Technical detail is collapsed. |
| B-U02 | Never show raw stack traces in the primary UI. |
| B-U03 | Empty states always name the next action. |
| B-U04 | .xls degradation and “images/charts not guaranteed” appear where relevant, not only in About. |
