# DECISIONS

Status legend: [DECISION] locked · [ASSUMPTION] working default · [OPEN] still open

## Product / Architecture

| ID | Decision | Status |
|----|----------|--------|
| D01 | Desktop shell: **Wails v2** | DECISION |
| D02 | IPC: **Wails native bindings** (no Gin HTTP) | DECISION |
| D03 | Formats: **.xlsx / .xlsm / .xls** | DECISION |
| D04 | Input: multi-select + drag-drop primary; folder scan secondary | DECISION |
| D05 | Column mismatch: user-selectable; default **by-header** | DECISION + ASSUMPTION (default) |
| D06 | Fidelity required: values, primary styles, col/row sizes, merges, formulas. Images/charts/pivot/data validation: best effort | DECISION |
| D07 | Partial success: emit output when any data succeeded; include success/skip/fail list | DECISION |
| D08 | Output: same directory as source, timestamped name; auto `_2`/`_3` on conflict | DECISION |
| D09 | Product languages MVP: **zh-TW / zh-CN / en** (keys reserved for ja/ko) | DECISION |
| D10 | Windows delivery: **portable exe + installer** | DECISION |
| D11 | **.xls** path: read-only convert → always write **.xlsx**; UI must show fidelity degradation | DECISION |
| D12 | Core merge logic lives in Go `core/`; UI never receives full sheet payloads | DECISION |
| D13 | Input files are opened read-only; never deleted/overwritten by the app | DECISION |
| D14 | MVP target OS for acceptance: **Windows 10 1809+**; macOS/Linux smoke in M4 | DECISION |

## Stack

| ID | Decision | Status |
|----|----------|--------|
| S01 | UI: React + Vite + TypeScript | DECISION |
| S02 | Backend: Go; Excel primary engine: **excelize** (planned M2) | DECISION |
| S03 | Settings file: AppData JSON, deletable, no sheet content | DECISION |
| S04 | No telemetry, no runtime CDN, offline core | DECISION |
| S05 | No paid licenses; new deps need license + activity review | DECISION |

## Error codes

| Code | Meaning |
|------|---------|
| E_FORMAT | Unsupported or unreadable format |
| E_EMPTY | No data to merge |
| E_LOCKED | Source locked by another process |
| E_PERM | Permission / write failure |
| E_HEADER_MISMATCH | Column alignment failure |
| E_CANCEL | User cancelled |
| E_PARTIAL | Completed with partial success |
| E_MEM | Size / memory guard |
| E_INTERNAL | Unexpected internal error |

## Residual assumptions

- Column-align default `by-header` is safer for non-technical users; switchable to `by-position` or `abort`.
- .xls style/formula fidelity is lower than xlsx; never claim full fidelity in UI copy.
- M0 settings are in-memory; persistence lands with M1 settings screen work.

## Out of scope for MVP

- Auto-update, accounts, cloud sync
- Plugin system
- Guaranteed round-trip of images/charts/pivot/VBA
- Product-grade ja/ko (architecture only)
- Background folder watching
