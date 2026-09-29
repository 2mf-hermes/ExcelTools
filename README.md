# ExcelTools

在地優先（local-first）的 Excel 桌面工具集，使用 Wails v2 + React + TypeScript + Go 開發。

**版本：** 0.4.0-m4（Milestone 4）

> English version below ｜ [English](#english)

## 功能特色

| 工具 | 能力 |
|------|------|
| 工作表合併 | 合併單一活頁簿內的多個工作表 |
| 檔案合併 | 合併多個檔案的第 N 個工作表，並對齊欄位 |

- 支援語言：繁體中文、簡體中文、英文、日文、韓文
- 主題：淺色 / 深色 / 跟隨系統
- 設定檔位置：`%AppData%\ExcelTools\settings.json`
- 完全離線、本機處理；**永不修改來源檔案**

## 建置

```powershell
# 執行測試並產出可攜式 zip
powershell -File scripts/package-windows.ps1

# 開發模式
wails dev

# 手動建置可攜版
wails build
```

需求環境：Go 1.25+、Node 20+、Wails v2 CLI。選用：NSIS 以產出安裝檔（`wails build -nsis`）。

## 測試

```powershell
go test ./...
cd frontend; npm run test:i18n
```

## 發行

請參閱 `docs/RELEASE_CHECKLIST.md`。

## 專案結構

```
core/merge/sheets|files   合併引擎
frontend/src              React UI 與 i18n
docs/                     DECISIONS、BEHAVIOR_SPEC、DESIGN_TOKENS、RELEASE_CHECKLIST
scripts/                  package-windows.ps1
```

---

<a name="english"></a>

# ExcelTools (English)

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
