package update

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadResult describes a completed, integrity-checked download.
type DownloadResult struct {
	Path   string
	SHA256 string
	Bytes  int64
}

// Download streams url into dest, hashing as it goes, enforcing the size cap,
// and verifying wantDigest ("sha256:<hex>") when non-empty.
//
// On any failure the partial file is removed, so a failed download can never
// leave a truncated file behind for the installer to pick up.
func (c *Checker) Download(ctx context.Context, url, dest, wantDigest string) (res DownloadResult, err error) {
	if err := c.policy.checkURL(url); err != nil {
		return DownloadResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return DownloadResult{}, errCode(CodeInternal, "cannot build request: %v", err)
	}
	req.Header.Set("User-Agent", "ExcelTools-updater")

	resp, err := c.Client.Do(req)
	if err != nil {
		return DownloadResult{}, wrapNetErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return DownloadResult{}, errCode(CodeHTTP, "download returned HTTP %d", resp.StatusCode)
	}
	// Reject up-front when the advertised size already blows the cap.
	if resp.ContentLength > MaxDownloadBytes {
		return DownloadResult{}, errCode(CodeTooLarge,
			"asset is %d bytes, over the %d byte cap", resp.ContentLength, MaxDownloadBytes)
	}

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return DownloadResult{}, errCode(CodePerm, "cannot create %s: %v", dest, err)
	}

	// Clean up on every error path, including a panic mid-copy.
	defer func() {
		if err != nil {
			out.Close()
			os.Remove(dest)
		}
	}()

	h := sha256.New()
	// Read one byte past the cap so an over-long stream is detectable rather
	// than silently truncated.
	body := io.Reader(resp.Body)
	if c.Progress != nil {
		body = &progressReader{
			r:     resp.Body,
			total: resp.ContentLength,
			cb:    c.Progress,
		}
	}
	written, err := io.Copy(io.MultiWriter(out, h), io.LimitReader(body, MaxDownloadBytes+1))
	if err != nil {
		return DownloadResult{}, wrapNetErr(err)
	}
	if written > MaxDownloadBytes {
		return DownloadResult{}, errCode(CodeTooLarge, "download exceeded the %d byte cap", MaxDownloadBytes)
	}
	if err = out.Close(); err != nil {
		return DownloadResult{}, errCode(CodePerm, "cannot flush %s: %v", dest, err)
	}

	got := hex.EncodeToString(h.Sum(nil))
	if wantDigest != "" {
		want := normalizeDigest(wantDigest)
		if got != want {
			err = errCode(CodeDigest,
				"checksum mismatch: expected %s, got %s", want, got)
			return DownloadResult{}, err
		}
	}

	return DownloadResult{Path: dest, SHA256: got, Bytes: written}, nil
}

// normalizeDigest lowercases a digest and strips an optional "sha256:" prefix.
func normalizeDigest(d string) string {
	d = strings.ToLower(strings.TrimSpace(d))
	d = strings.TrimPrefix(d, "sha256:")
	return d
}

// HasUsableDigest reports whether a release asset publishes a sha256 digest we
// can verify against.
func HasUsableDigest(a Asset) bool {
	d := normalizeDigest(a.Digest)
	if len(d) != 64 {
		return false
	}
	for i := 0; i < len(d); i++ {
		c := d[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// ExtractExeFromZip pulls ExcelTools.exe out of the portable archive into dest.
//
// The entry is matched on its base name and written to the destination we
// choose, so an archive containing traversal paths ("../evil.exe") cannot
// escape the target directory.
func ExtractExeFromZip(zipPath, dest string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return errCode(CodeNoAsset, "cannot open archive: %v", err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !strings.EqualFold(filepath.Base(f.Name), ExeName) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return errCode(CodeNoAsset, "cannot read %s from archive: %v", f.Name, err)
		}
		// Scope the copy so a decompression bomb cannot fill the disk.
		n, err := writeFile(dest, io.LimitReader(rc, MaxDownloadBytes+1))
		rc.Close()
		if err != nil {
			return err
		}
		if n > MaxDownloadBytes {
			os.Remove(dest)
			return errCode(CodeTooLarge, "extracted file exceeded the %d byte cap", MaxDownloadBytes)
		}
		return nil
	}
	return errCode(CodeNoAsset, "archive does not contain %s", ExeName)
}

// writeFile writes r to path, returning the byte count.
func writeFile(path string, r io.Reader) (int64, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return 0, errCode(CodePerm, "cannot create %s: %v", path, err)
	}
	n, err := io.Copy(f, r)
	if cerr := f.Close(); err == nil && cerr != nil {
		err = cerr
	}
	if err != nil {
		os.Remove(path)
		return 0, errCode(CodePerm, "cannot write %s: %v", path, err)
	}
	return n, nil
}

// progressReader reports download progress to a callback, throttled to whole
// percentage points so a fast download does not flood the event bridge.
type progressReader struct {
	r     io.Reader
	done  int64
	total int64
	cb    func(done, total int64)
	last  int
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.done += int64(n)
	if p.total > 0 {
		pct := int(p.done * 100 / p.total)
		if pct != p.last {
			p.last = pct
			p.cb(p.done, p.total)
		}
	} else {
		// No Content-Length: throttle on every 256 KiB instead.
		if ki := int(p.done >> 18); ki != p.last {
			p.last = ki
			p.cb(p.done, p.total)
		}
	}
	return n, err
}
