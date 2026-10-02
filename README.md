<div align="center">

# 📊 ExcelTools

**在地優先的 Excel 桌面工具集，讓多檔案、多工作表合併一鍵完成，來源檔絕不被修改。**

[![Version](https://img.shields.io/badge/version-0.4.0--m4-blue.svg)](https://github.com/2mf-hermes/ExcelTools/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB.svg?logo=react&logoColor=black)](https://react.dev/)
[![Wails](https://img.shields.io/badge/Wails-v2-DF0000.svg)](https://wails.io/)
[![Platform](https://img.shields.io/badge/platform-Windows-0078D6.svg?logo=windows&logoColor=white)](#快速上手)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](#貢獻指南)

繁體中文 ｜ [English](#english)

</div>

---

## 📖 這是什麼

ExcelTools 是一套**本機優先、離線可用**的 Excel 桌面工具，用 Wails v2 打包 Go 引擎與 React 介面。它把日常最惱人的「把一堆表格拼在一起」變成幾次點擊——而且**永遠不會改動你的原始檔案**。

## 🎬 展示畫面

> 📌 截圖佔位：請將實際畫面放到 `docs/screenshots/` 後，把下方路徑換成真實檔名。

<div align="center">

<!-- 範例：![ExcelTools 主畫面](docs/screenshots/main.png) -->
<img src="https://placehold.co/820x480/1e293b/e2e8f0?text=ExcelTools+Demo" alt="ExcelTools 展示畫面（佔位）" width="720" />

</div>

## ✨ 主要特色

- 🧩 **工作表合併** — 把單一活頁簿裡的多個工作表合併成一張。
- 🗂️ **檔案合併** — 合併多個檔案的第 N 個工作表，並自動對齊欄位。
- 🔒 **來源零修改** — 所有處理都在本機完成，原始檔案原封不動。
- 🌐 **五種語言** — 繁體中文、簡體中文、English、日本語、한국어。
- 🎨 **三種主題** — 淺色 / 深色 / 跟隨系統。
- 📴 **離線可用** — 檔案不上傳，全部在本機處理。
- 🔄 **自動更新** — 啟動時檢查 GitHub Releases；經你確認後才下載安裝，並以 SHA-256 驗證檔案。

## 🔒 隱私與連線

| 行為 | 說明 |
|------|------|
| 你的檔案 | **永不離開本機**。合併、讀寫全部在電腦上完成，沒有任何上傳。 |
| 唯一的對外連線 | 啟動時的**更新檢查**（`api.github.com`），以及你按下「下載並安裝」後的更新下載。可在 **設定 → 啟動時自動檢查更新** 關閉。 |
| 離線時 | 更新檢查會**靜默失敗**，不會跳錯誤、不影響功能。 |
| 更新安全 | 僅允許 HTTPS 與本專案 GitHub 網域；重新導向逐跳驗證；下載後比對官方 SHA-256，**不符即中止不安裝**；不要求系統管理員權限、不呼叫外部指令、不常駐背景程序。 |
| 遙測 | **無**。沒有分析、沒有 CDN、沒有遠端設定。 |

> 早期版本 README 寫「完全不連網」。新增更新功能後這句話不再成立，因此在此明確改寫，而非默默略過。

## 🚀 快速上手

### 前置需求

| 工具 | 版本 |
|------|------|
| [Go](https://go.dev/) | 1.25+ |
| [Node.js](https://nodejs.org/) | 20+ |
| [Wails CLI](https://wails.io/docs/gettingstarted/installation) | v2 |
| NSIS（選用，用於產生安裝檔） | 最新版 |

### 安裝與建置

```powershell
# 1. 取得原始碼
git clone https://github.com/2mf-hermes/ExcelTools.git
cd ExcelTools

# 2. 安裝 Wails CLI（若尚未安裝）
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 3. 開發模式（熱重載）
wails dev
```

## 💡 使用範例

```powershell
# 產生可攜版執行檔（含測試）
powershell -File scripts/package-windows.ps1

# 手動建置可攜版
wails build

# 產生 Windows 安裝檔（需 NSIS）
wails build -nsis
```

建置完成後，可攜版執行檔位於 `dist/`，直接雙擊 `ExcelTools.exe` 即可使用，無需安裝。設定檔會存放於 `%AppData%\ExcelTools\settings.json`。

### 執行測試

```powershell
go test ./...
cd frontend; npm run test:i18n
```

## 🗂️ 專案結構

```
core/merge/sheets|files   合併引擎（工作表 / 檔案）
core/report | validate    報表與驗證邏輯
frontend/src              React UI 與 i18n
platform/                 檔案系統與 Excel 平台層
docs/                     DECISIONS、BEHAVIOR_SPEC、DESIGN_TOKENS、RELEASE_CHECKLIST
scripts/                  package-windows.ps1、圖示產生器
```

## 🤝 貢獻指南

歡迎參與！流程如下：

1. Fork 本專案並建立分支：`git checkout -b feature/your-feature`
2. 提交變更前先跑過測試：`go test ./...`
3. 送出 Pull Request，並在描述中說明變更內容與測試方式。

發行流程細節請見 [`docs/RELEASE_CHECKLIST.md`](docs/RELEASE_CHECKLIST.md)。

## 📄 授權條款

本專案採用 [MIT License](LICENSE) 授權，可自由使用、修改與散布。

---

<a name="english"></a>

## 🌍 English

**ExcelTools** — a local-first Excel desktop toolkit that merges multiple files and sheets in a few clicks, and never modifies your source files.

Built with Wails v2 + React + TypeScript + Go.

### Features

- 🧩 **Sheet combine** — merge multiple sheets from one workbook into one.
- 🗂️ **File combine** — merge the Nth sheet across multiple files with automatic column alignment.
- 🔒 **Sources never modified** — everything runs locally; originals stay untouched.
- 🌐 **5 languages** — zh-TW, zh-CN, English, 日本語, 한국어.
- 🎨 **3 themes** — light / dark / system.
- 📴 **Offline-capable** — your files are never uploaded; everything runs locally.
- 🔄 **Auto-update** — checks GitHub Releases at launch and installs only after you confirm, with SHA-256 verification.

### Privacy & Network

Your documents never leave your machine. The **only** outbound request is the
update check (`api.github.com`) plus the update download you explicitly confirm;
you can disable the check in **Settings → Check for updates at startup**. An offline
launch fails silently. Updates are HTTPS-only, restricted to this project's GitHub
hosts, validated on every redirect hop, and installed only if the SHA-256 digest
matches. No telemetry, no CDN, no elevation, no background processes.

### Getting Started

**Requirements:** Go 1.25+, Node.js 20+, Wails v2 CLI. Optional: NSIS for the installer.

```powershell
git clone https://github.com/2mf-hermes/ExcelTools.git
cd ExcelTools
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails dev
```

### Usage

```powershell
# Tests + portable zip
powershell -File scripts/package-windows.ps1

# Manual portable build
wails build

# Windows installer (requires NSIS)
wails build -nsis
```

### Test

```powershell
go test ./...
cd frontend; npm run test:i18n
```

### Contributing & License

PRs are welcome — fork, branch, run `go test ./...`, and open a PR. Licensed under the [MIT License](LICENSE).
