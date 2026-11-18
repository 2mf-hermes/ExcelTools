<div align="center">

# 📊 ExcelTools

ExcelTools 在 Windows 上合并 Excel 工作表。合并不需要联网，也不会修改源文件。

[![Version](https://img.shields.io/badge/version-0.5.0--m7-blue.svg)](https://github.com/2mf-hermes/ExcelTools/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25.0-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-61DAFB.svg?logo=react&logoColor=black)](https://react.dev/)
[![Wails](https://img.shields.io/badge/Wails-v2-DF0000.svg)](https://wails.io/)
[![Platform](https://img.shields.io/badge/platform-Windows-0078D6.svg?logo=windows&logoColor=white)](#快速上手)

[繁體中文](README.md) ｜ [简体中文](README.zh-CN.md) ｜ [English](README.en.md)

</div>

## 功能

- **工作表合并**：将同一工作簿中的多个工作表合并为一张表。
- **多文件合并**：从多个工作簿中选择工作表，可按列标题或列位置对齐。
- **工作表选择**：可以为每个文件单独指定工作表，也可以统一选择自动跳过空白工作表，或指定第 N 个工作表。
- **五种语言**：繁體中文、简体中文、English、日本語、한국어。
- **三种外观**：浅色、深色、跟随系统。

例如，每月有一份工作簿，要把各文件里的“销售”工作表汇总成一张表：在“多文件合并”中添加文件。如果“销售”都在第二张，就统一选第 2 个工作表；位置不同的文件，再单独指定。

## 截图

<div align="center">
<img src="docs/screenshots/main-window-cn.png" alt="ExcelTools 主界面" width="720" />
<br/><sub>工具首页</sub>

<br/><br/>

<img src="docs/screenshots/combine-file-cn.png" alt="多文件合并与工作表选择" width="720" />
<br/><sub>多文件合并：统一规则与逐文件选择工作表</sub>
</div>

## 隐私与更新连接

| 项目 | 说明 |
|------|------|
| Excel 文件 | 在本机读写，不会上传，也不会覆盖源文件。 |
| 更新检查 | 启动时向 `api.github.com` 查询最新 Release；可在设置中关闭。 |
| 更新下载 | 只有在用户确认后才下载，并比对 Release 提供的 SHA-256。 |
| 连接安全 | 更新请求仅使用 HTTPS，并逐次检查 URL 和重定向主机。 |
| 遥测 | 不收集分析数据，也没有远程配置或 CDN。 |

离线时，启动检查失败不会弹出错误，也不影响合并。确认安装后，程序会验证下载文件并重启；缺少 SHA-256 或验证不符时不会自动安装。如果重启时旧版本备份（`.old`）仍被占用，程序会在后台每 500 毫秒重试删除，最多 15 秒。

## 快速上手

### 使用已发布版本

从 [Releases](https://github.com/2mf-hermes/ExcelTools/releases) 下载 Windows 便携版，解压后直接运行 `ExcelTools.exe`，无需安装。设置文件位于 `%AppData%\ExcelTools\settings.json`。

### 从源代码运行

需要 Go 1.25.0 或更高版本、Node.js 20.19+（20.x）或 22.12+、Wails CLI v2。若要生成 NSIS 安装包，还需要 NSIS 3.x。

```powershell
git clone https://github.com/2mf-hermes/ExcelTools.git
cd ExcelTools
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails dev
```

### 构建与测试

```powershell
# 执行 Go 测试、i18n 检查并生成便携版 zip
powershell -File scripts/package-windows.ps1

# 单独构建
wails build

# 需要 NSIS 时生成安装包
wails build -nsis

# 测试
go test ./...
cd frontend; npm run test:i18n
```

`wails build` 将可执行文件放在 `build/bin/ExcelTools.exe`。打包脚本将便携版 ZIP 放在 `dist/`，两个输出目录都不纳入版本控制。

## 项目结构

```text
core/merge/       工作表与多文件合并逻辑
core/update/      更新检查、下载、验证与安装
frontend/src/     React 界面与多语言
platform/         文件系统与平台层
docs/             决策、行为规范、设计与发布清单
scripts/          Windows 打包脚本
```

## 参与贡献

欢迎报告问题或提交 Pull Request。提交 PR 前，请运行上面的测试，并说明改了什么、如何验证。发布流程请参考 [`docs/RELEASE_CHECKLIST.md`](docs/RELEASE_CHECKLIST.md)。

## 许可证

本项目采用 [MIT License](LICENSE)。
