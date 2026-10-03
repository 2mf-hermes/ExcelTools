# ExcelTools Release Checklist (M5)

App version: 0.5.0-m5  
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
- [ ] Settings persist (`autoCheckUpdates`) across restart
- [ ] Update check: launch with a newer release published → banner appears
- [ ] Update check: launch offline → no error shown, app fully usable
- [ ] Update install: confirm → progress bar → app restarts on the new version
- [ ] Update install: tampered asset (digest mismatch) → refuses, does not install
- [ ] `autoCheckUpdates` off → no request is made at launch
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
- [ ] Release publishes a `digest` for the exe asset (required for auto-install)

## Distribution

Distribution is via GitHub Releases; the app's update check reads that same channel.
Publish the exe asset with a SHA-256 `digest` or auto-install will be refused and
users will be sent to the Releases page instead.
Do not host on public CDN without reviewing privacy copy.

## Version bump

Update in:

- `app.go` Health/GetAppInfo version string
- `build/windows/info.json` / icon if needed
- `wails.json` `info.productVersion` — set to `0.5.0.5` (numeric; the app's own
  version lives in `app.go`). **Known issue:** `wails build` on this machine emits
  no `.syso`, so the exe's Windows file-properties tab stays blank despite this
  being configured. The in-app version (from `app.go`) is correct; only the shell
  metadata is missing. Worth investigating before a 1.0.
- This checklist header
