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

// Update status values reported to the UI.
const (
	// UpdateUpToDate means the running version is the newest published one.
	UpdateUpToDate = "up-to-date"
	// UpdateAvailable means a newer release exists.
	UpdateAvailable = "available"
	// UpdateError means the check itself failed (offline, HTTP error, …).
	UpdateError = "error"
	// UpdateSkipped means no check was attempted or a failure was suppressed so
	// that an offline launch stays silent.
	UpdateSkipped = "skipped"
)

// UpdateCheckResult is returned by the settings "check updates" action.
type UpdateCheckResult struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	Status         string `json:"status"` // up-to-date | available | error | skipped
	Message        string `json:"message"`
	ReleaseNotes   string `json:"releaseNotes,omitempty"`
	ReleaseURL     string `json:"releaseUrl,omitempty"`
	AssetName      string `json:"assetName,omitempty"`
	AssetSize      int64  `json:"assetSize,omitempty"`
	// CanAutoInstall is false when the release publishes no verifiable digest,
	// in which case the UI offers the download page instead.
	CanAutoInstall bool `json:"canAutoInstall"`
}

// UpdateInstallResult reports the outcome of an automatic install attempt.
type UpdateInstallResult struct {
	Status  string `json:"status"` // installed | error
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// UpdateProgress is the payload of the "update:progress" event.
type UpdateProgress struct {
	Done  int64 `json:"done"`
	Total int64 `json:"total"`
	Pct   int   `json:"pct"`
}

// AppInfo is static metadata for the home/about surface.
type AppInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
	// LocalOnly means user documents are never uploaded: all processing happens
	// on this machine. The update check is the single outbound request and is
	// disableable; see UpdateCheck.
	LocalOnly bool `json:"localOnly"`
	// UpdateCheck reports whether this build can check for updates.
	UpdateCheck bool `json:"updateCheck"`
}

// Settings is the persisted local preference file shape.
type Settings struct {
	Language string `json:"language"`
	Theme    string `json:"theme"`
	// AutoCheckUpdates controls the version check at launch. It is the only
	// outbound request the app makes, so it is user-controlled.
	AutoCheckUpdates bool              `json:"autoCheckUpdates"`
	Defaults         map[string]string `json:"defaults"`
}

// DefaultSettings returns factory preferences.
func DefaultSettings() Settings {
	return Settings{
		Language:         "system",
		Theme:            "system",
		AutoCheckUpdates: true,
		Defaults: map[string]string{
			"outputDirMode":       "same-as-source",
			"headerMode":          string(HeaderEach),
			"columnAlign":         string(AlignByHeader),
			"skipEmptySheets":     "true",
			"addSourceColumn":     "true",
			"keepBaseOtherSheets": "true",
		},
	}
}
