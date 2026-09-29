// Package model holds domain types shared by merge tools and the UI bridge.
// No OS, Wails, or Excel library dependencies.
package model

// ToolID identifies a home-screen tool.
type ToolID string

const (
	ToolSheetCombine ToolID = "sheet-combine"
	ToolFileCombine  ToolID = "file-combine"
)

// FileStatus classifies an input file before or during merge.
type FileStatus string

const (
	FileOK      FileStatus = "ok"
	FileSkipped FileStatus = "skipped"
	FileFailed  FileStatus = "failed"
)

// HeaderMode controls how leading header rows are treated in sheet merge.
type HeaderMode string

const (
	HeaderEach  HeaderMode = "each"
	HeaderFirst HeaderMode = "first"
	HeaderNone  HeaderMode = "none"
)

// ColumnAlign controls multi-file column mapping when headers differ.
type ColumnAlign string

const (
	AlignByPosition ColumnAlign = "by-position"
	AlignByHeader   ColumnAlign = "by-header"
	AlignAbort      ColumnAlign = "abort"
)

// ItemStatus is the per-file outcome in a merge report.
type ItemStatus string

const (
	ItemSuccess ItemStatus = "success"
	ItemSkipped ItemStatus = "skipped"
	ItemFailed  ItemStatus = "failed"
)

// FileRef is a user-selected input file.
type FileRef struct {
	Path   string     `json:"path"`
	Name   string     `json:"name"`
	Size   int64      `json:"size"`
	Ext    string     `json:"ext"`
	Status FileStatus `json:"status"`
	Reason string     `json:"reason,omitempty"`
}

// SheetRef describes one worksheet in a workbook.
type SheetRef struct {
	Name   string `json:"name"`
	Rows   int    `json:"rows"`
	Cols   int    `json:"cols"`
	Empty  bool   `json:"empty"`
	Hidden bool   `json:"hidden"`
}

// SheetMergeOptions configures single-workbook multi-sheet merge.
type SheetMergeOptions struct {
	HeaderMode      HeaderMode `json:"headerMode"`
	AddSourceSheet  bool       `json:"addSourceSheet"`
	SkipEmpty       bool       `json:"skipEmpty"`
	OutputSheetName string     `json:"outputSheetName"`
}

// FileMergeOptions configures multi-file merge.
type FileMergeOptions struct {
	HeaderRows          int         `json:"headerRows"`
	SheetIndex          int         `json:"sheetIndex"`
	ColumnAlign         ColumnAlign `json:"columnAlign"`
	AddSourceFile       bool        `json:"addSourceFile"`
	KeepBaseOtherSheets bool        `json:"keepBaseOtherSheets"`
}

// ProgressEvent is emitted to the UI during long operations.
type ProgressEvent struct {
	Phase   string `json:"phase"`
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Label   string `json:"label"`
	Percent int    `json:"percent"`
}

// ItemReport is one file's outcome in MergeResult.
type ItemReport struct {
	File    string     `json:"file"`
	Status  ItemStatus `json:"status"`
	Rows    int        `json:"rows"`
	Message string     `json:"message,omitempty"`
}

// MergeResult is the outcome of a merge run.
type MergeResult struct {
	Success    bool         `json:"success"`
	Partial    bool         `json:"partial"`
	Message    string       `json:"message"`
	OutputPath string       `json:"outputPath"`
	FilesUsed  int          `json:"filesUsed"`
	Rows       int          `json:"rows"`
	Cols       int          `json:"cols"`
	DurationMs int64        `json:"durationMs"`
	Items      []ItemReport `json:"items"`
}

// SheetListResult is the UI payload for workbook sheet preview.
type SheetListResult struct {
	Sheets []SheetRef `json:"sheets"`
	Note   string     `json:"note,omitempty"`
}

// UpdateCheckResult is returned by the settings "check updates" action.
type UpdateCheckResult struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	Status         string `json:"status"` // up-to-date | no-source | error
	Message        string `json:"message"`
}

// AppInfo is static metadata for the home/about surface.
type AppInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Platform  string `json:"platform"`
	LocalOnly bool   `json:"localOnly"`
}

// Settings is the persisted local preference file shape.
type Settings struct {
	Language string            `json:"language"`
	Theme    string            `json:"theme"`
	Defaults map[string]string `json:"defaults"`
}

// DefaultSettings returns factory preferences.
func DefaultSettings() Settings {
	return Settings{
		Language: "system",
		Theme:    "system",
		Defaults: map[string]string{
			"outputDirMode":     "same-as-source",
			"headerMode":        string(HeaderEach),
			"columnAlign":       string(AlignByHeader),
			"skipEmptySheets":   "true",
			"addSourceColumn":   "true",
			"keepBaseOtherSheets": "true",
		},
	}
}
