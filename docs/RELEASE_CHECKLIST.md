# ExcelTools Release Checklist (M4)

App version: 0.4.0-m4  
Target: Windows 10 1809+ (primary), macOS/Linux smoke (secondary)

## Build outputs

| Artifact | How | Path |
|----------|-----|------|
| Portable exe | `wails build` | `build/bin/ExcelTools.exe` |
| NSIS installer | `wails build -nsis` (requires makensis) | `build/bin/ExcelTools-*.exe` installer |

### Portable package

```powershell
wails build
# Ship: build/bin/ExcelTools.exe
# Optional: zip with README
Compress-Archive -Path build/bin/ExcelTools.exe -DestinationPath dist/ExcelTools-portable-win64.zip
```

### Installer (NSIS)

Prerequisites: NSIS 3.x on PATH (`makensis`).

```powershell
wails build -nsis
```

If `makensis` is missing, ship portable zip only and note installer as optional.

## Pre-release verification

- [ ] `go test ./...` all green
- [ ] `npm run test:i18n` — 5 locales × key parity
- [ ] `wails build` succeeds
- [ ] App launches offline (disconnect network, start exe)
- [ ] Sheet combine: merge 3-sheet workbook; source unchanged
- [ ] File combine: merge 2 workbooks; by-header align; partial skip list
- [ ] Cancel mid-merge leaves no output
- [ ] Settings persist (language/theme) across restart
- [ ] Light/Dark switch
- [ ] Keyboard: Tab through home tools, Enter activates
- [ ] Focus rings visible
- [ ] No admin required to run
- [ ] Input files not modified (spot-check timestamps/bytes)

## Localization QA

Languages: zh-TW, zh-CN, en, ja, ko

- [ ] Switch each language in Settings; home + both tools readable
- [ ] No hardcoded Chinese left in EN UI (spot-check buttons)
- [ ] Error strings use i18n keys
- [ ] `{n}` placeholders substitute correctly

## Known limitations (document for users)

- Multi-file merge does not accept `.xls` (single-file tool does, read-only)
- Cross-file formulas written as cached values
- Images/charts/pivot not guaranteed
- macOS/Linux builds not product-signed in this milestone

## Code signing (optional, pre-public)

- [ ] Authenticode sign `ExcelTools.exe` if cert available
- [ ] SmartScreen note for unsigned builds

## Distribution

Offline USB / internal share is the default. No auto-update in MVP.
Do not host on public CDN without reviewing privacy copy.

## Version bump

Update in:

- `app.go` Health/GetAppInfo version string
- `build/windows/info.json` / icon if needed
- This checklist header
