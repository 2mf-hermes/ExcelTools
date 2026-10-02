# Auto-update — Design Spec

**Date:** 2026-10-03
**Status:** Implemented (0.5.0-m5)
**Target version:** 0.5.0-m5

## Problem

ExcelTools ships as a portable Windows `.exe`. Users have no way to learn that a
newer build exists, and no way to replace their binary short of manually
re-downloading from GitHub Releases and overwriting the file by hand.

## Goal

In Settings, detect a newer published release. When the user confirms, download
it, verify it, install it by replacing the running executable, and restart.

## Non-goals

- Delta / incremental patching.
- macOS and Linux install paths (check-only for now; install returns
  `E_UNSUPPORTED` off Windows).
- Rollback beyond keeping the previous binary as `<exe>.old`.
- Silent unattended install. A user confirmation is always required before
  anything is written over the running binary.

## Constraint conflict (recorded deliberately)

`docs/DECISIONS.md` previously listed auto-update under "Out of scope for MVP"
and **S04** locked "No telemetry, no runtime CDN, offline core". This feature
amends that posture: the app now performs a **network request on startup**.

Mitigations adopted, because the offline promise was a product-level commitment:

| Control | Effect |
|---|---|
| Startup check is a persisted user preference (`autoCheckUpdates`) | User can restore fully-offline behaviour |
| Startup check fails silently (errors discarded) | Offline users see no error, no modal, no delay to first paint |
| No telemetry, no identifiers, no request body | Request is a plain GET of a public endpoint |
| Only user-initiated install contacts the network again | Checking never downloads |
| S04 and the out-of-scope list amended, README privacy line updated | Documentation does not misrepresent behaviour |

## Architecture

New package `core/update` — pure Go, no Wails, no Excel, no filesystem side
effects in the check/download path, so it is unit-testable with `httptest`.

```
core/update/
  version.go    Version parsing + ordering
  checker.go    Release metadata fetch/parse, asset selection, host policy
  download.go   Streaming download, in-flight SHA-256, size cap
  install.go    Platform-neutral install orchestration
  install_windows.go  Self-replace + relaunch (Windows)
  install_other.go    E_UNSUPPORTED off Windows
```

The Wails-facing layer is `app_update.go` in package `main`, which binds
`CheckForUpdates` (existing name, now real) and `DownloadAndInstallUpdate`.

### Why this split

`core/` already holds logic with no UI dependency, and `docs/DECISIONS.md` D12
states core logic lives in Go and the UI never receives bulk payloads. Keeping
network and file-swap logic out of package `main` lets the security-critical
paths (host policy, digest check) be tested without launching Wails.

## Version comparison

Tags look like `v0.5.0-m5`, `0.4.0-m4`, `0.6.0`. Rules, in order:

1. Strip a leading `v`.
2. Parse `major.minor.patch` numerically. Unparseable → treated as `0`.
3. Compare the numeric triple.
4. If equal, a version with **no** suffix sorts **higher** than one with a
   suffix (plain release beats its own pre-release).
5. If both have suffixes, compare case-insensitively, digit runs numerically so
   `m10 > m9` (a plain string compare would get this wrong).

`Compare(a, b) int` returns -1 / 0 / 1. Ties mean "up to date".

## Update check

`GET https://api.github.com/repos/2mf-hermes/ExcelTools/releases/latest`

Response fields consumed: `tag_name`, `name`, `body`, `html_url`, `prerelease`,
`draft`, `assets[]` where each asset has `name`, `browser_download_url`, `size`,
`digest`.

`digest` is returned as `sha256:<hex>`. It is absent on assets uploaded before
GitHub added the field.

Draft and prerelease entries are ignored for the "latest" decision as a
defensive measure even though the `latest` endpoint already excludes them.

### Asset selection

Prefer, in order: `ExcelTools.exe` → the portable `.zip` (extract
`ExcelTools.exe` from it). Any other asset name is refused — a release that
contains neither yields `E_NO_ASSET` rather than downloading something
unexpected.

## Security policy

This is the part that matters. Controls, all enforced in `core/update`:

1. **HTTPS only.** Any non-`https` URL is rejected before a request is made.
2. **Host allowlist.** `api.github.com`, `github.com`,
   `objects.githubusercontent.com`, `release-assets.githubusercontent.com`.
   Applies to the API URL *and* the asset URL *and* every redirect hop.
3. **Redirect validation.** `http.Client.CheckRedirect` re-applies the host
   policy on each hop and caps the chain length. This blocks the classic
   "trusted URL 302s to attacker host" path.
4. **SHA-256 verification.** Hash computed in-flight while streaming, compared
   against the release asset digest. **Mismatch aborts and the temp file is
   deleted**; nothing is installed.
5. **Missing digest = no auto-install.** If the release publishes no digest we
   cannot verify integrity, so the app refuses to self-replace and offers the
   Releases page instead. Fail closed, not open.
6. **Size cap.** Rejected if `Content-Length` or the observed stream exceeds the
   cap (300 MB — generous for a ~14 MB binary, tight enough to stop a disk-fill).
7. **Timeouts.** Dial/TLS/response-header timeouts plus an overall download
   deadline, so a stalled server cannot hang the UI.
8. **No elevation, no shell.** The installer never requests admin and never
   executes a downloaded script. Only the EXE we verified is placed on disk.
9. **Release-build guard.** Self-replace is refused when the running binary is
   not an `ExcelTools.exe` release build (protects `go run` and `wails dev`).

## Install sequence (Windows)

A running executable cannot be overwritten, but it *can* be renamed.

1. Download to `<dir>/ExcelTools.new.exe` (same volume as the target, so the
   subsequent renames are atomic and never cross a filesystem boundary).
2. Verify SHA-256 (step 4 above).
3. `Rename(ExcelTools.exe, ExcelTools.exe.old)` — permitted while running.
4. `Rename(ExcelTools.new.exe, ExcelTools.exe)`.
   - On failure: rename `.old` back and abort with `E_PERM`. The user is left
     with a working install.
5. Spawn the new `ExcelTools.exe` detached.
6. `os.Exit(0)` the current process.
7. On next startup, delete a leftover `ExcelTools.exe.old`.

Chosen over the common "write a .bat helper and shell out" approach because no
script file is created and no interpreter runs — less attack surface and less
antivirus suspicion.

## UI

Settings gains an update card, replacing the current version panel:

- Current version, and latest version when known.
- Release name + notes (truncated, scrollable).
- States: idle → checking → up-to-date | available → downloading (percent) →
  verifying → installing → error.
- 「下載並安裝」requires an explicit click; the download does not start on check.
- 「啟動時自動檢查更新」toggle, persisted via the existing settings `Defaults`
  map (`autoCheckUpdates`) so the `Settings` struct shape is unchanged and no
  migration is needed.
- When no digest is published, the card shows the manual-download notice instead
  of the install button.

All new strings go into all five locales (zh-TW, zh-CN, en, ja, ko); the
`check-i18n.cjs` parity gate enforces this.

## Error codes

Extends the existing table in `DECISIONS.md`.

| Code | Meaning |
|---|---|
| `E_NET` | Network failure / unreachable |
| `E_HTTP` | Non-200 response |
| `E_NO_ASSET` | No acceptable asset in the release |
| `E_DIGEST` | SHA-256 mismatch or digest unavailable for auto-install |
| `E_INSECURE` | URL failed scheme or host policy |
| `E_TOO_LARGE` | Exceeded the download size cap |
| `E_TIMEOUT` | Exceeded a deadline |
| `E_UNSUPPORTED` | Install not supported on this platform |
| `E_NOT_RELEASE` | Running binary is not a release build |

## Testing

- `version_test.go` — ordering table incl. `m9` vs `m10`, `v` prefix, release vs
  pre-release, malformed input.
- `checker_test.go` — JSON parse, asset preference order, `E_NO_ASSET`,
  allowlist rejection, `http:` rejection, redirect-to-foreign-host rejection.
- `download_test.go` — happy path, digest mismatch deletes temp file, size cap,
  non-200.
- Existing `go test ./...` stays green; `npm run test:i18n` parity gate passes.
- `wails build` compiles and regenerates bindings.

**Stated limit:** the swap-and-relaunch path itself cannot be exercised without
a real published release and a packaged binary. Logic around it is tested; the
end-to-end replace is verified on first real update, not claimed here.
