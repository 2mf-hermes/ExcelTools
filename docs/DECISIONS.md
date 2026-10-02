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
| D15 | In-app update: check **GitHub Releases API**, automatic install after explicit user confirmation | DECISION (M5) |

## Stack

| ID | Decision | Status |
|----|----------|--------|
| S01 | UI: React + Vite + TypeScript | DECISION |
| S02 | Backend: Go; Excel primary engine: **excelize** (planned M2) | DECISION |
| S03 | Settings file: AppData JSON, deletable, no sheet content | DECISION |
| S04 | No telemetry, no runtime CDN; offline core. The update check is the **single** outbound request, user-disableable | DECISION (amended in M5) |
| S05 | No paid licenses; new deps need license + activity review | DECISION |
| S06 | Update artifacts must be **HTTPS from the project's own GitHub hosts**, digest-verified (SHA-256) before install | DECISION (M5) |
| S07 | Update check runs at launch, silently; installer never elevates, never shells out, never runs in the background | DECISION (M5) |

### S04 amendment (M5)

The original wording was "no network at all". Shipping an update feature makes that
literally false, so the decision is restated rather than quietly contradicted:

- User documents are **never** uploaded — merging stays purely local.
- The **only** outbound request is the update check (`api.github.com`) plus the
  update download when the user clicks install.
- The launch-time check has a Settings toggle (`autoCheckUpdates`, default on) and
  fails silently when offline.
- No telemetry, no analytics, no runtime CDN, no remote fonts/config.

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
- Update install is refused when the release publishes no digest we can verify
  against; the UI then opens the Releases page instead of installing.
- The exe swap-and-restart path (`core/update/install_windows.go`) is covered by
  tests against a staged temp dir, but a true end-to-end swap needs a published
  release to exercise in the wild.

## Out of scope for MVP

- Accounts, cloud sync
- Delta/incremental updates and background auto-download (install is always user-confirmed)
- Plugin system
- Guaranteed round-trip of images/charts/pivot/VBA
- Product-grade ja/ko (architecture only)
- Background folder watching
