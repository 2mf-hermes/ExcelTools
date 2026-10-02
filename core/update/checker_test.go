package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testHost is the loopback name httptest serves on; the test policy allows it so
// the production allowlist can stay strict.
const testHost = "127.0.0.1"

func testPolicy() hostPolicy {
	return hostPolicy{allowed: map[string]struct{}{testHost: {}}}
}

// newTestChecker wires a Checker to a local TLS test server. srv.Client()
// supplies the transport that trusts the server's self-signed certificate.
func newTestChecker(t *testing.T, srv *httptest.Server, current string) *Checker {
	t.Helper()
	p := testPolicy()
	c := newClient(p)
	c.Transport = srv.Client().Transport
	return &Checker{
		Current: current,
		APIURL:  srv.URL + "/releases/latest",
		Client:  c,
		policy:  p,
	}
}

func TestCheckURLPolicy(t *testing.T) {
	cases := []struct {
		url  string
		ok   bool
		code string
	}{
		{"https://api.github.com/repos/x/y/releases/latest", true, ""},
		{"https://github.com/2mf-hermes/ExcelTools/releases", true, ""},
		{"https://objects.githubusercontent.com/foo", true, ""},
		{"https://release-assets.githubusercontent.com/foo", true, ""},

		// Bare HTTP is refused even for an allowlisted host.
		{"http://api.github.com/repos/x/y", false, CodeInsecure},
		// Off-allowlist hosts are refused even over HTTPS.
		{"https://evil.example.com/ExcelTools.exe", false, CodeInsecure},
		// Lookalike hosts must not slip through.
		{"https://api.github.com.evil.example.com/x", false, CodeInsecure},
		{"https://notgithub.com/x", false, CodeInsecure},
		{"https://github.com@evil.example.com/x", false, CodeInsecure},
		// Other schemes.
		{"ftp://github.com/x", false, CodeInsecure},
		{"file:///C:/Windows/System32/calc.exe", false, CodeInsecure},
		{"", false, CodeInsecure},
	}
	for _, c := range cases {
		err := CheckURL(c.url)
		if c.ok {
			if err != nil {
				t.Errorf("CheckURL(%q) = %v, want nil", c.url, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("CheckURL(%q) = nil, want rejection", c.url)
			continue
		}
		if got := CodeOf(err); got != c.code {
			t.Errorf("CheckURL(%q) code = %q, want %q", c.url, got, c.code)
		}
	}
}

func TestCheckRejectsNonAllowlistedAPIHost(t *testing.T) {
	// A Checker whose API URL points off-allowlist must refuse before dialling.
	p := testPolicy()
	c := &Checker{Current: "0.4.0-m4", APIURL: "https://evil.example.com/latest", Client: newClient(p), policy: p}
	if _, err := c.Check(context.Background()); CodeOf(err) != CodeInsecure {
		t.Fatalf("Check() error = %v, want %s", err, CodeInsecure)
	}
}

const sampleRelease = `{
  "tag_name": "v0.5.0-m5",
  "name": "ExcelTools v0.5.0-m5",
  "body": "更新功能",
  "html_url": "https://github.com/2mf-hermes/ExcelTools/releases/tag/v0.5.0-m5",
  "draft": false,
  "prerelease": false,
  "assets": [
    {"name":"ExcelTools-portable-win64.zip","browser_download_url":"https://github.com/dl/a.zip","size":5510000,"digest":"sha256:aa"},
    {"name":"ExcelTools.exe","browser_download_url":"https://github.com/dl/ExcelTools.exe","size":14060000,"digest":"sha256:bb"}
  ]
}`

func TestFetchLatestParses(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Errorf("Accept header = %q", got)
		}
		fmt.Fprint(w, sampleRelease)
	}))
	defer srv.Close()

	rel, err := newTestChecker(t, srv, "0.4.0-m4").FetchLatest(context.Background())
	if err != nil {
		t.Fatalf("FetchLatest: %v", err)
	}
	if rel.TagName != "v0.5.0-m5" || rel.Name != "ExcelTools v0.5.0-m5" || rel.Body != "更新功能" {
		t.Errorf("parsed release = %+v", rel)
	}
	if len(rel.Assets) != 2 {
		t.Fatalf("assets = %d, want 2", len(rel.Assets))
	}
}

func TestCheckDetectsUpdateAndPrefersExe(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sampleRelease)
	}))
	defer srv.Close()

	res, err := newTestChecker(t, srv, "0.4.0-m4").Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !res.UpdateAvailable || res.Latest != "v0.5.0-m5" {
		t.Errorf("result = %+v, want update available to v0.5.0-m5", res)
	}
	if res.Asset.Name != ExeName {
		t.Errorf("asset = %q, want %q (exe preferred over zip)", res.Asset.Name, ExeName)
	}
}

func TestCheckNoUpdateWhenCurrent(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sampleRelease)
	}))
	defer srv.Close()

	res, err := newTestChecker(t, srv, "0.5.0-m5").Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if res.UpdateAvailable {
		t.Errorf("UpdateAvailable = true for identical versions")
	}
}

func TestFetchLatestErrors(t *testing.T) {
	cases := []struct {
		name string
		code int
		body string
		want string
	}{
		{"not found", http.StatusNotFound, `{}`, CodeHTTP},
		{"rate limited", http.StatusForbidden, `{}`, CodeHTTP},
		{"server error", http.StatusInternalServerError, `{}`, CodeHTTP},
		{"malformed json", http.StatusOK, `{not json`, CodeHTTP},
		{"empty tag", http.StatusOK, `{"tag_name":"","assets":[]}`, CodeHTTP},
		{"draft", http.StatusOK, `{"tag_name":"v9.9.9","draft":true}`, CodeNoAsset},
		{"prerelease", http.StatusOK, `{"tag_name":"v9.9.9","prerelease":true}`, CodeNoAsset},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.code)
				fmt.Fprint(w, c.body)
			}))
			defer srv.Close()

			_, err := newTestChecker(t, srv, "0.4.0-m4").FetchLatest(context.Background())
			if got := CodeOf(err); got != c.want {
				t.Errorf("error code = %q (%v), want %q", got, err, c.want)
			}
		})
	}
}

func TestSelectAsset(t *testing.T) {
	exe := Asset{Name: ExeName, BrowserDownloadURL: "https://github.com/x.exe"}
	zip := Asset{Name: ZipName, BrowserDownloadURL: "https://github.com/x.zip"}
	other := Asset{Name: "notes.txt", BrowserDownloadURL: "https://github.com/x.txt"}

	cases := []struct {
		name string
		rel  Release
		want string
		code string
	}{
		{"prefers exe", Release{TagName: "v1", Assets: []Asset{zip, exe}}, ExeName, ""},
		{"falls back to zip", Release{TagName: "v1", Assets: []Asset{zip}}, ZipName, ""},
		{"exe in any position", Release{TagName: "v1", Assets: []Asset{exe, zip}}, ExeName, ""},
		{"case insensitive", Release{TagName: "v1", Assets: []Asset{{Name: "exceltools.EXE", BrowserDownloadURL: "u"}}}, "exceltools.EXE", ""},
		{"unknown assets refused", Release{TagName: "v1", Assets: []Asset{other}}, "", CodeNoAsset},
		{"no assets", Release{TagName: "v1"}, "", CodeNoAsset},
		{"missing download url", Release{TagName: "v1", Assets: []Asset{{Name: ExeName}}}, "", CodeNoAsset},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := SelectAsset(c.rel)
			if c.code != "" {
				if CodeOf(err) != c.code {
					t.Fatalf("error = %v, want code %s", err, c.code)
				}
				return
			}
			if err != nil {
				t.Fatalf("SelectAsset: %v", err)
			}
			if got.Name != c.want {
				t.Errorf("asset = %q, want %q", got.Name, c.want)
			}
		})
	}
}

func TestHasUsableDigest(t *testing.T) {
	valid := strings.Repeat("ab", 32) // 64 hex chars
	cases := []struct {
		in   string
		want bool
	}{
		{"sha256:" + valid, true},
		{valid, true},
		{"SHA256:" + strings.ToUpper(valid), true},
		{"", false},
		{"sha256:", false},
		{strings.Repeat("a", 63), false},
		{strings.Repeat("a", 65), false},
		{strings.Repeat("z", 64), false}, // non-hex
		{"md5:" + valid, false},
	}
	for _, c := range cases {
		if got := HasUsableDigest(Asset{Digest: c.in}); got != c.want {
			t.Errorf("HasUsableDigest(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestDownloadVerifiesDigest(t *testing.T) {
	payload := []byte("fake installer bytes")
	sum := sha256.Sum256(payload)
	digest := "sha256:" + hex.EncodeToString(sum[:])

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	}))
	defer srv.Close()

	c := newTestChecker(t, srv, "0.4.0-m4")
	dest := filepath.Join(t.TempDir(), "out.exe")

	t.Run("matching digest", func(t *testing.T) {
		res, err := c.Download(context.Background(), srv.URL+"/a", dest, digest)
		if err != nil {
			t.Fatalf("Download: %v", err)
		}
		if res.Bytes != int64(len(payload)) {
			t.Errorf("bytes = %d, want %d", res.Bytes, len(payload))
		}
		got, _ := os.ReadFile(dest)
		if string(got) != string(payload) {
			t.Errorf("file contents differ from payload")
		}
	})

	t.Run("mismatched digest aborts and removes file", func(t *testing.T) {
		wrong := "sha256:" + strings.Repeat("00", 32)
		if _, err := c.Download(context.Background(), srv.URL+"/a", dest, wrong); CodeOf(err) != CodeDigest {
			t.Fatalf("error = %v, want %s", err, CodeDigest)
		}
		if _, err := os.Stat(dest); !os.IsNotExist(err) {
			t.Errorf("partial file was left behind after failed verification")
		}
	})
}

func TestDownloadRejectsOffAllowlistURL(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	c := newTestChecker(t, srv, "0.4.0-m4")
	dest := filepath.Join(t.TempDir(), "out.exe")
	if _, err := c.Download(context.Background(), "https://evil.example.com/x.exe", dest, ""); CodeOf(err) != CodeInsecure {
		t.Errorf("error = %v, want %s", err, CodeInsecure)
	}
}

func TestDownloadRejectsOversizedContentLength(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprint(MaxDownloadBytes+1))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestChecker(t, srv, "0.4.0-m4")
	dest := filepath.Join(t.TempDir(), "out.exe")
	_, err := c.Download(context.Background(), srv.URL+"/a", dest, "")
	if CodeOf(err) != CodeTooLarge && CodeOf(err) != CodeNet {
		t.Errorf("error = %v, want %s", err, CodeTooLarge)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Errorf("oversized download left a file behind")
	}
}

func TestDownloadHTTPError(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestChecker(t, srv, "0.4.0-m4")
	dest := filepath.Join(t.TempDir(), "out.exe")
	if _, err := c.Download(context.Background(), srv.URL+"/a", dest, ""); CodeOf(err) != CodeHTTP {
		t.Errorf("error = %v, want %s", err, CodeHTTP)
	}
}

func TestDownloadReportsProgress(t *testing.T) {
	payload := make([]byte, 512*1024)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	}))
	defer srv.Close()

	c := newTestChecker(t, srv, "0.4.0-m4")
	var calls int
	var lastDone int64
	c.Progress = func(done, total int64) { calls++; lastDone = done }

	if _, err := c.Download(context.Background(), srv.URL+"/a", filepath.Join(t.TempDir(), "o"), ""); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if calls == 0 {
		t.Errorf("progress callback never fired")
	}
	if lastDone != int64(len(payload)) {
		t.Errorf("final progress = %d, want %d", lastDone, len(payload))
	}
}

func TestDownloadRedirectToDisallowedHostIsBlocked(t *testing.T) {
	// The initial request is on the allowed test host, but the server tries to
	// redirect the download to an attacker-controlled host. The per-hop allowlist
	// check must stop it.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://evil.example.com/payload.exe", http.StatusFound)
	}))
	defer srv.Close()

	c := newTestChecker(t, srv, "0.4.0-m4")
	dest := filepath.Join(t.TempDir(), "out.exe")
	_, err := c.Download(context.Background(), srv.URL+"/a", dest, "")
	if err == nil {
		t.Fatal("redirect to a disallowed host was followed")
	}
	if CodeOf(err) != CodeInsecure {
		t.Errorf("error = %v, want %s", err, CodeInsecure)
	}
}
