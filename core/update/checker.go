package update

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// DefaultAPIURL is the GitHub Releases "latest" endpoint for this project.
	DefaultAPIURL = "https://api.github.com/repos/2mf-hermes/ExcelTools/releases/latest"

	// ExeName is the release asset we install from.
	ExeName = "ExcelTools.exe"

	// ZipName is the portable archive, used when no bare exe is published.
	ZipName = "ExcelTools-portable-win64.zip"

	// MaxDownloadBytes caps a single download (300 MB).
	MaxDownloadBytes = 300 << 20

	// maxRedirects caps the redirect chain length.
	maxRedirects = 5
)

// allowedHosts is the egress allowlist. Every URL the updater contacts — the API
// URL, the asset URL, and each redirect hop — must resolve to one of these.
// GitHub serves release assets via a redirect to its object storage, hence the
// two extra entries.
var allowedHosts = map[string]struct{}{
	"api.github.com":                       {},
	"github.com":                           {},
	"objects.githubusercontent.com":        {},
	"release-assets.githubusercontent.com": {},
}

// Asset is one downloadable file attached to a release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	Digest             string `json:"digest"` // "sha256:<hex>", may be empty
}

// IsZip reports whether this asset is the portable archive rather than a bare exe.
func (a Asset) IsZip() bool {
	return strings.EqualFold(a.Name, ZipName) || strings.HasSuffix(strings.ToLower(a.Name), ".zip")
}

// Release is the subset of the GitHub release payload the updater consumes.
type Release struct {
	TagName    string  `json:"tag_name"`
	Name       string  `json:"name"`
	Body       string  `json:"body"`
	HTMLURL    string  `json:"html_url"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

// Checker performs update checks and downloads for one running version.
type Checker struct {
	// Current is the running version, e.g. "0.5.0-m5".
	Current string

	// APIURL and Client are injectable for tests.
	APIURL string
	Client *http.Client

	// Progress, when set, receives (bytesDone, totalBytes) during a download.
	// totalBytes is 0 when the server sends no Content-Length.
	Progress func(done, total int64)

	// policy is the egress allowlist applied to every request and redirect.
	policy hostPolicy
}

// hostPolicy is the egress allowlist. Each Checker holds one by value so tests
// can substitute a policy pointing at a local TLS test server.
type hostPolicy struct {
	allowed map[string]struct{}
}

// checkURL enforces HTTPS-only transport and the host allowlist.
func (p hostPolicy) checkURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return errCode(CodeInsecure, "unparseable URL")
	}
	if u.Scheme != "https" {
		return errCode(CodeInsecure, "non-HTTPS URL rejected (%q)", u.Scheme)
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return errCode(CodeInsecure, "URL has no host")
	}
	if _, ok := p.allowed[host]; !ok {
		return errCode(CodeInsecure, "host not allowed: %s", host)
	}
	return nil
}

// defaultPolicy is the production egress policy.
var defaultPolicy = hostPolicy{allowed: allowedHosts}

// CheckURL applies the production egress policy to a URL.
func CheckURL(raw string) error { return defaultPolicy.checkURL(raw) }

// NewChecker returns a Checker with the production URL, client, and timeouts.
func NewChecker(current string) *Checker {
	return &Checker{
		Current: current,
		APIURL:  DefaultAPIURL,
		Client:  newClient(defaultPolicy),
		policy:  defaultPolicy,
	}
}

// newClient builds an HTTP client with transport-level timeouts and a redirect
// policy that re-applies the host allowlist on every hop.
func newClient(p hostPolicy) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errCode(CodeInsecure, "too many redirects (%d)", len(via))
			}
			// Validate the hop target, not just the original URL.
			return p.checkURL(req.URL.String())
		},
	}
}

// Result summarises a completed check.
type Result struct {
	Current         string
	Latest          string
	UpdateAvailable bool
	Release         Release
	Asset           Asset // zero when the release has no installable asset
}

// Check asks the release endpoint for the latest version and compares it with
// the running one. A network or protocol failure is returned as an error; being
// up to date is not an error.
func (c *Checker) Check(ctx context.Context) (Result, error) {
	rel, err := c.FetchLatest(ctx)
	if err != nil {
		return Result{}, err
	}
	res := Result{
		Current: c.Current,
		Latest:  rel.TagName,
		Release: rel,
	}
	res.UpdateAvailable = IsNewer(rel.TagName, c.Current)
	if res.UpdateAvailable {
		// A release with no usable asset is still "available" — the UI falls
		// back to the manual download page — so a selection failure is not fatal.
		if a, err := SelectAsset(rel); err == nil {
			res.Asset = a
		}
	}
	return res, nil
}

// FetchLatest retrieves and decodes the latest release payload.
func (c *Checker) FetchLatest(ctx context.Context) (Release, error) {
	if err := c.policy.checkURL(c.APIURL); err != nil {
		return Release{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.APIURL, nil)
	if err != nil {
		return Release{}, errCode(CodeInternal, "cannot build request: %v", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ExcelTools-updater")

	resp, err := c.Client.Do(req)
	if err != nil {
		return Release{}, wrapNetErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, errCode(CodeHTTP, "release endpoint returned HTTP %d", resp.StatusCode)
	}

	var rel Release
	// Cap the metadata read; a release body is text and never legitimately huge.
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&rel); err != nil {
		return Release{}, errCode(CodeHTTP, "cannot parse release metadata: %v", err)
	}
	// The /latest endpoint already excludes these, but never trust it blindly.
	if rel.Draft || rel.Prerelease {
		return Release{}, errCode(CodeNoAsset, "latest release is a draft or pre-release")
	}
	if strings.TrimSpace(rel.TagName) == "" {
		return Release{}, errCode(CodeHTTP, "release metadata has no tag name")
	}
	return rel, nil
}

// SelectAsset picks the installable asset, preferring the bare exe over the
// portable zip. Anything not on the known list is refused so the updater can
// never be pointed at an unexpected file by a compromised release.
func SelectAsset(rel Release) (Asset, error) {
	for _, want := range []string{ExeName, ZipName} {
		for _, a := range rel.Assets {
			if strings.EqualFold(a.Name, want) {
				if a.BrowserDownloadURL == "" {
					return Asset{}, errCode(CodeNoAsset, "asset %s has no download URL", a.Name)
				}
				return a, nil
			}
		}
	}
	return Asset{}, errCode(CodeNoAsset, "release %s has no %s or %s asset", rel.TagName, ExeName, ZipName)
}

// wrapNetErr maps transport failures onto the project's error codes.
func wrapNetErr(err error) error {
	var ce *codedError
	if errors.As(err, &ce) {
		// A policy check already rejected this request (e.g. a redirect to a
		// non-allowlisted host). Keep that code: reporting a generic network
		// error would hide the fact that a security rule fired.
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		return errCode(CodeTimeout, "request timed out: %v", err)
	}
	return errCode(CodeNet, "request failed: %v", err)
}
