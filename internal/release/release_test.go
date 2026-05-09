package release

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Richonn/sbomforge/internal/config"
	"github.com/google/go-github/v71/github"
)

func newTestServer(t *testing.T, mux *http.ServeMux) (*httptest.Server, *github.Client) {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	url := srv.URL + "/"
	client := github.NewClient(nil).WithAuthToken("test-token")
	client.BaseURL, _ = client.BaseURL.Parse(url)
	client.UploadURL, _ = client.UploadURL.Parse(url)
	return srv, client
}

func baseCfg() *config.Config {
	return &config.Config{
		GitHubToken:     "test-token",
		RepoOwner:       "owner",
		RepoName:        "repo",
		RefName:         "v1.0.0",
		Format:          "spdx-json",
		Sign:            false,
		AttachToRelease: true,
	}
}

func TestNew(t *testing.T) {
	cfg := baseCfg()
	c := New(cfg)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.cfg != cfg {
		t.Error("expected cfg to be set")
	}
	if c.gh == nil {
		t.Error("expected gh client to be set")
	}
}

func TestUpload_AttachDisabled(t *testing.T) {
	cfg := baseCfg()
	cfg.AttachToRelease = false

	c := &Client{cfg: cfg, gh: github.NewClient(nil)}
	url, err := c.Upload(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "" {
		t.Errorf("expected empty url, got %q", url)
	}
}

func TestUpload_ReleaseNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/releases/tags/v1.0.0", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, ghClient := newTestServer(t, mux)
	c := &Client{cfg: baseCfg(), gh: ghClient}

	_, err := c.Upload(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for missing release, got nil")
	}
	if !strings.Contains(err.Error(), "no release found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpload_Success(t *testing.T) {
	releaseID := int64(42)
	mux := http.NewServeMux()

	mux.HandleFunc("/repos/owner/repo/releases/tags/v1.0.0", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&github.RepositoryRelease{ID: &releaseID})
	})

	mux.HandleFunc("/repos/owner/repo/releases/42/assets", func(w http.ResponseWriter, r *http.Request) {
		downloadURL := "https://example.com/sbom.json"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&github.ReleaseAsset{BrowserDownloadURL: &downloadURL})
	})

	_, ghClient := newTestServer(t, mux)
	c := &Client{cfg: baseCfg(), gh: ghClient}

	sbomFile, err := os.CreateTemp(t.TempDir(), "sbom-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = sbomFile.WriteString(`{}`)
	_ = sbomFile.Close()

	sbomURL, err := c.Upload(context.Background(), sbomFile.Name(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sbomURL != "https://example.com/sbom.json" {
		t.Errorf("unexpected sbomURL: %s", sbomURL)
	}
}

func TestAssetLabel(t *testing.T) {
	cases := map[string]string{
		"spdx-json":      "SBOM (SPDX JSON)",
		"cyclonedx-json": "SBOM (CycloneDX JSON)",
		"syft-json":      "SBOM (Syft JSON)",
		"unknown":        "SBOM",
	}
	for format, want := range cases {
		if got := assetLabel(format); got != want {
			t.Errorf("assetLabel(%q) = %q, want %q", format, got, want)
		}
	}
}

func TestAssetName(t *testing.T) {
	cases := map[string]string{
		"/tmp/sbom.spdx-json.json": "sbom.spdx-json.json",
		"sbom.json":                "sbom.json",
		"/a/b/c/file.json":         "file.json",
	}
	for path, want := range cases {
		if got := assetName(path); got != want {
			t.Errorf("assetName(%q) = %q, want %q", path, got, want)
		}
	}
}
