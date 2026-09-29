package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"ExcelTools/core/merge/files"
	"ExcelTools/core/model"
	"ExcelTools/core/validate"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const fileProgressEvent = "filemerge:progress"

type fileJob struct {
	cancel context.CancelFunc
	mu     sync.Mutex
	active bool
}

func (a *App) beginFileJob() (context.Context, error) {
	a.fileJob.mu.Lock()
	defer a.fileJob.mu.Unlock()
	if a.fileJob.active {
		return nil, fmt.Errorf("E_INTERNAL: a file merge is already running")
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.fileJob.cancel = cancel
	a.fileJob.active = true
	return ctx, nil
}

func (a *App) endFileJob() {
	a.fileJob.mu.Lock()
	defer a.fileJob.mu.Unlock()
	if a.fileJob.cancel != nil {
		a.fileJob.cancel()
		a.fileJob.cancel = nil
	}
	a.fileJob.active = false
}

// CancelFileMerge cancels the running multi-file merge, if any.
func (a *App) CancelFileMerge() bool {
	a.fileJob.mu.Lock()
	defer a.fileJob.mu.Unlock()
	if a.fileJob.active && a.fileJob.cancel != nil {
		a.fileJob.cancel()
		return true
	}
	return false
}

// MergeFilesRequest is the UI payload for multi-file merge.
type MergeFilesRequest struct {
	Paths               []string `json:"paths"`
	HeaderRows          int      `json:"headerRows"`
	SheetIndex          int      `json:"sheetIndex"`
	ColumnAlign         string   `json:"columnAlign"`
	AddSourceFile       bool     `json:"addSourceFile"`
	KeepBaseOtherSheets bool     `json:"keepBaseOtherSheets"`
	KeepBlankRows       bool     `json:"keepBlankRows"`
	OutputDir           string   `json:"outputDir"`
}

// MergeFiles runs multi-workbook merge.
func (a *App) MergeFiles(req MergeFilesRequest) (*model.MergeResult, error) {
	if len(req.Paths) == 0 {
		return nil, fmt.Errorf("E_EMPTY: no input files")
	}
	for _, p := range req.Paths {
		if !validate.ExtSupported(p) {
			return nil, fmt.Errorf("E_FORMAT: unsupported extension: %s", p)
		}
	}
	align := model.ColumnAlign(req.ColumnAlign)
	switch align {
	case model.AlignByPosition, model.AlignByHeader, model.AlignAbort, "":
	default:
		return nil, fmt.Errorf("E_INTERNAL: unknown column align")
	}

	ctx, err := a.beginFileJob()
	if err != nil {
		return nil, err
	}
	defer a.endFileJob()

	progress := func(ev model.ProgressEvent) {
		wailsruntime.EventsEmit(a.ctx, fileProgressEvent, ev)
	}

	res, merr := files.Merge(ctx, files.Options{
		Paths:               req.Paths,
		HeaderRows:          req.HeaderRows,
		SheetIndex:          req.SheetIndex,
		ColumnAlign:         align,
		AddSourceFile:       req.AddSourceFile,
		KeepBaseOtherSheets: req.KeepBaseOtherSheets,
		KeepBlankRows:       req.KeepBlankRows,
		OutputDir:           req.OutputDir,
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
