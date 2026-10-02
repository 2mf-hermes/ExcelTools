// Package update implements update checking and self-installation.
//
// It is deliberately free of Wails, Excel, and global-state dependencies so the
// security-critical paths (URL policy, digest verification, version ordering)
// can be unit-tested in isolation.
//
// Design: docs/superpowers/specs/2026-10-03-auto-update-design.md
package update

import (
	"errors"
	"fmt"
	"strings"
)

// Error codes. These extend the table in docs/DECISIONS.md.
const (
	CodeNet         = "E_NET"
	CodeHTTP        = "E_HTTP"
	CodeNoAsset     = "E_NO_ASSET"
	CodeDigest      = "E_DIGEST"
	CodeInsecure    = "E_INSECURE"
	CodeTooLarge    = "E_TOO_LARGE"
	CodeTimeout     = "E_TIMEOUT"
	CodeUnsupported = "E_UNSUPPORTED"
	CodeNotRelease  = "E_NOT_RELEASE"
	CodePerm        = "E_PERM"
	CodeInternal    = "E_INTERNAL"
)

// codedError carries a machine-readable code alongside the human-readable
// detail, so the code survives wrapping by net/http.
type codedError struct {
	code string
	msg  string
}

func (e *codedError) Error() string { return e.code + ": " + e.msg }

// Code returns the error's machine-readable code.
func (e *codedError) Code() string { return e.code }

// errCode builds a coded error using the project's "E_CODE: detail" convention.
func errCode(code, format string, args ...any) error {
	return &codedError{code: code, msg: fmt.Sprintf(format, args...)}
}

// CodeOf extracts the "E_*" code from an error, or "" if absent.
//
// It unwraps, so a coded error stays identifiable after net/http has wrapped it
// in a *url.Error — the case that matters when a redirect is refused.
func CodeOf(err error) string {
	if err == nil {
		return ""
	}
	var ce *codedError
	if errors.As(err, &ce) {
		return ce.code
	}
	// Errors raised outside this package follow the same textual convention.
	msg := err.Error()
	if !strings.HasPrefix(msg, "E_") {
		return ""
	}
	if i := strings.IndexByte(msg, ':'); i > 0 {
		return msg[:i]
	}
	return ""
}
