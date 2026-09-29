# ExcelTools

Local-first Excel desktop tool collection (Wails v2 + React + TypeScript + Go).

**Version:** 0.4.0-m4 (Milestone 4)

## Features

| Tool | Capability |
|------|------------|
| Sheet combine | Multi-sheet merge from one workbook |
| File combine | Multi-file merge of Nth sheet, column alignment |

- Languages: zh-TW, zh-CN, en, ja, ko
- Themes: light / dark / system
- Settings: `%AppData%\ExcelTools\settings.json`
- Offline, local-only processing; sources never modified

## Build

```powershell
# Tests + portable zip
powershell -File scripts/package-windows.ps1

# Dev
wails dev

# Manual portable build
wails build
```

Requires: Go 1.25+, Node 20+, Wails v2 CLI. Optional: NSIS for installer (`wails build -nsis`).

## Test

```powershell
go test ./...
cd frontend; npm run test:i18n
```

## Release

See `docs/RELEASE_CHECKLIST.md`.

## Layout

```
core/merge/sheets|files   merge engines
frontend/src              React UI + i18n
docs/                     DECISIONS, BEHAVIOR_SPEC, DESIGN_TOKENS, RELEASE_CHECKLIST
scripts/                  package-windows.ps1
```
