package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"ExcelTools/core/merge/sheets"
	"ExcelTools/core/model"
	"ExcelTools/core/validate"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const progressEvent = "merge:progress"

// sheetJob tracks the in-flight single-file merge for cancel support.
type sheetJob struct {
	cancel context.CancelFunc
	mu     sync.Mutex
	active bool
}

func (a *App) beginSheetJob() (context.Context, error) {
	a.sheet.mu.Lock()
	defer a.sheet.mu.Unlock()
	if a.sheet.active {
		return nil, fmt.Errorf("E_INTERNAL: a merge is already running")
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.sheet.cancel = cancel
	a.sheet.active = true
	return ctx, nil
}

func (a *App) endSheetJob() {
	a.sheet.mu.Lock()
	defer a.sheet.mu.Unlock()
	if a.sheet.cancel != nil {
		a.sheet.cancel()
		a.sheet.cancel = nil
	}
	a.sheet.active = false
}

// CancelSheetMerge cancels the running single-file merge, if any.
func (a *App) CancelSheetMerge() bool {
	a.sheet.mu.Lock()
	defer a.sheet.mu.Unlock()
	if a.sheet.active && a.sheet.cancel != nil {
		a.sheet.cancel()
		return true
	}
	return false
}

// MergeSheetsRequest is the UI payload for single-file multi-sheet merge.
type MergeSheetsRequest struct {
	SourcePath        string   `json:"sourcePath"`
	SheetNames        []string `json:"sheetNames"`
	HeaderMode        string   `json:"headerMode"`
	AddSourceSheet    bool     `json:"addSourceSheet"`
	SkipEmpty         bool     `json:"skipEmpty"`
	OutputSheetName   string   `json:"outputSheetName"`
	SourceColumnLabel string   `json:"sourceColumnLabel"`
	OutputPath        string   `json:"outputPath"`
}

// MergeSheets runs the single-workbook multi-sheet merge.
func (a *App) MergeSheets(req MergeSheetsRequest) (*model.MergeResult, error) {
	src := strings.TrimSpace(req.SourcePath)
	if src == "" {
		return nil, fmt.Errorf("E_EMPTY: source path is empty")
	}
	if !validate.ExtSupported(src) {
		return nil, fmt.Errorf("E_FORMAT: unsupported extension")
	}
	mode := model.HeaderMode(req.HeaderMode)
	switch mode {
	case model.HeaderEach, model.HeaderFirst, model.HeaderNone, "":
	default:
		return nil, fmt.Errorf("E_INTERNAL: unknown header mode")
	}

	ctx, err := a.beginSheetJob()
	if err != nil {
		return nil, err
	}
	defer a.endSheetJob()

	progress := func(ev model.ProgressEvent) {
		wailsruntime.EventsEmit(a.ctx, progressEvent, ev)
	}

	res, merr := sheets.Merge(ctx, src, sheets.MergeOptions{
		SheetNames:        req.SheetNames,
		HeaderMode:        mode,
		AddSourceSheet:    req.AddSourceSheet,
		SkipEmpty:         req.SkipEmpty,
		OutputSheetName:   req.OutputSheetName,
		SourceColumnLabel: req.SourceColumnLabel,
		OutputPath:        req.OutputPath,
	}, progress)

	if merr != nil {
		code := "E_INTERNAL"
		msg := merr.Error()
		if strings.HasPrefix(msg, "E_") {
			if i := strings.Index(msg, ":"); i > 0 {
				code = msg[:i]
			} else {
				code = msg
			}
		}
		if res == nil {
			res = &model.MergeResult{}
		}
		res.Success = false
		res.Message = msg
		return res, fmt.Errorf("%s: %s", code, msg)
	}
	return res, nil
}
