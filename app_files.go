package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"ExcelTools/core/model"
	"ExcelTools/core/validate"
	"ExcelTools/platform/fs"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

var excelFilters = []wailsruntime.FileFilter{
	{
		DisplayName: "Excel (.xlsx, .xlsm, .xls)",
		Pattern:     "*.xlsx;*.xlsm;*.xls",
	},
}

// PickExcelFiles opens a multi-select file dialog for Excel workbooks.
func (a *App) PickExcelFiles() ([]model.FileRef, error) {
	if err := a.ensureCtx(); err != nil {
		return nil, err
	}
	paths, err := wailsruntime.OpenMultipleFilesDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:                "選擇 Excel 檔案",
		Filters:              excelFilters,
		CanCreateDirectories: false,
	})
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return []model.FileRef{}, nil
	}
	return a.ClassifyPaths(paths), nil
}

// PickExcelFile opens a single-file dialog.
func (a *App) PickExcelFile() (*model.FileRef, error) {
	if err := a.ensureCtx(); err != nil {
		return nil, err
	}
	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:   "選擇 Excel 檔案",
		Filters: excelFilters,
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}
	refs := a.ClassifyPaths([]string{path})
	if len(refs) == 0 {
		return nil, nil
	}
	return &refs[0], nil
}

// PickFolder opens a directory picker for folder scan.
func (a *App) PickFolder() (string, error) {
	if err := a.ensureCtx(); err != nil {
		return "", err
	}
	return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "選擇資料夾",
	})
}

// ScanFolder lists Excel files directly under dir (non-recursive).
func (a *App) ScanFolder(dir string) ([]model.FileRef, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("E_EMPTY: folder path is empty")
	}
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		return nil, fmt.Errorf("E_FORMAT: folder not found")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("E_PERM: cannot read folder: %w", err)
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".xlsx" || ext == ".xlsm" || ext == ".xls" || ext == ".xltx" || ext == ".xltm" {
			paths = append(paths, filepath.Join(dir, name))
		}
	}
	fs.SortNatural(paths)
	return a.ClassifyPaths(paths), nil
}

// ClassifyPaths validates paths and returns FileRef list (ok/skipped/failed).
func (a *App) ClassifyPaths(paths []string) []model.FileRef {
	const outPrefix = "合併結果_"
	out := make([]model.FileRef, 0, len(paths))
	seen := map[string]bool{}
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		ref := model.FileRef{
			Path: p,
			Name: filepath.Base(p),
			Ext:  validate.ExtOf(p),
		}
		st, err := os.Stat(p)
		if err != nil {
			ref.Status = model.FileFailed
			ref.Reason = "E_FORMAT: file not found"
			out = append(out, ref)
			continue
		}
		if st.IsDir() {
			ref.Status = model.FileFailed
			ref.Reason = "E_FORMAT: path is a folder"
			out = append(out, ref)
			continue
		}
		ref.Size = st.Size()
		base := filepath.Base(p)
		if strings.HasPrefix(base, "~$") {
			ref.Status = model.FileSkipped
			ref.Reason = "E_LOCKED: Excel temp/lock file"
			out = append(out, ref)
			continue
		}
		if strings.HasPrefix(base, outPrefix) {
			ref.Status = model.FileSkipped
			ref.Reason = "skipped: prior tool output"
			out = append(out, ref)
			continue
		}
		if !validate.ExtSupported(p) {
			ref.Status = model.FileFailed
			ref.Reason = "E_FORMAT: unsupported extension"
			out = append(out, ref)
			continue
		}
		if st.Size() == 0 {
			ref.Status = model.FileFailed
			ref.Reason = "E_EMPTY: file is empty"
			out = append(out, ref)
			continue
		}
		ref.Status = model.FileOK
		out = append(out, ref)
	}
	return out
}

// ListSheets returns worksheet metadata for a single workbook (input preview).
// Reads .xlsx/.xlsm via excelize. For .xls, returns a degradation notice without sheet detail.
func (a *App) ListSheets(path string) (*model.SheetListResult, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("E_EMPTY: path is empty")
	}
	if !fs.IsReadableFile(path) {
		return nil, fmt.Errorf("E_FORMAT: file not found")
	}
	ext := validate.ExtOf(path)
	if ext == ".xls" {
		return &model.SheetListResult{
			Sheets: []model.SheetRef{},
			Note:   "E_FORMAT: .xls sheet listing is limited; merge will convert with reduced fidelity",
		}, nil
	}
	if ext != ".xlsx" && ext != ".xlsm" {
		return nil, fmt.Errorf("E_FORMAT: unsupported extension")
	}

	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("E_FORMAT: cannot open workbook: %w", err)
	}
	defer f.Close()

	names := f.GetSheetList()
	refs := make([]model.SheetRef, 0, len(names))
	for _, name := range names {
		sr := model.SheetRef{Name: name}
		if vis, err := f.GetSheetVisible(name); err == nil && !vis {
			sr.Hidden = true
		}
		if dim, err := f.GetSheetDimension(name); err == nil && dim != "" {
			parts := strings.Split(dim, ":")
			end := parts[len(parts)-1]
			if c, r, err := excelize.CellNameToCoordinates(end); err == nil {
				sr.Cols = c
				sr.Rows = r
			}
		}
		if sr.Rows == 0 || (sr.Rows <= 1 && sr.Cols <= 1) {
			if cell, err := f.GetCellValue(name, "A1"); err == nil && strings.TrimSpace(cell) == "" {
				sr.Empty = true
			}
		}
		refs = append(refs, sr)
	}
	return &model.SheetListResult{Sheets: refs}, nil
}

// RevealInExplorer opens the containing folder or file in the OS file manager.
func (a *App) RevealInExplorer(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("E_EMPTY: path is empty")
	}
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", "/select,", filepath.FromSlash(path)).Start()
	case "darwin":
		return exec.Command("open", "-R", path).Start()
	default:
		return exec.Command("xdg-open", filepath.Dir(path)).Start()
	}
}

// SelectFolderAndScan combines picker + scan for the multi-file tool.
func (a *App) SelectFolderAndScan() ([]model.FileRef, error) {
	dir, err := a.PickFolder()
	if err != nil {
		return nil, err
	}
	if dir == "" {
		return []model.FileRef{}, nil
	}
	return a.ScanFolder(dir)
}

// ensureCtx guards dialog APIs that need a live Wails context.
func (a *App) ensureCtx() error {
	if a.ctx == nil {
		return fmt.Errorf("E_INTERNAL: app context not ready")
	}
	return nil
}
