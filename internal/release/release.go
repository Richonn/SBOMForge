package release

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Richonn/sbomforge/internal/config"
	"github.com/google/go-github/v71/github"
	"golang.org/x/oauth2"
)

type Client struct {
	gh  *github.Client
	cfg *config.Config
}

func New(cfg *config.Config) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.GitHubToken})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		gh:  github.NewClient(tc),
		cfg: cfg,
	}
}

func (c *Client) Upload(ctx context.Context, sbomPath, bundlePath string) (string, error) {
	if !c.cfg.AttachToRelease {
		return "", nil
	}

	release, err := c.getRelease(ctx)
	if err != nil {
		return "", err
	}

	sbomURL, err := c.uploadAsset(ctx, release, sbomPath, assetLabel(c.cfg.Format))
	if err != nil {
		return "", fmt.Errorf("upload sbom: %w", err)
	}

	if c.cfg.Sign && bundlePath != "" {
		if _, err := c.uploadAsset(ctx, release, bundlePath, "Cosign bundle"); err != nil {
			return "", fmt.Errorf("upload bundle: %w", err)
		}
	}

	return sbomURL, nil
}

func (c *Client) getRelease(ctx context.Context) (*github.RepositoryRelease, error) {
	release, resp, err := c.gh.Repositories.GetReleaseByTag(ctx, c.cfg.RepoOwner, c.cfg.RepoName, c.cfg.RefName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("no release found for tag %q - create the release before running SBOMForge", c.cfg.RefName)
		}
		return nil, fmt.Errorf("get release: %w", err)
	}
	return release, nil
}

func (c *Client) uploadAsset(ctx context.Context, release *github.RepositoryRelease, path, label string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	opts := &github.UploadOptions{
		Name:      assetName(path),
		Label:     label,
		MediaType: "application/json",
	}

	assets, _, err := c.gh.Repositories.UploadReleaseAsset(ctx, c.cfg.RepoOwner, c.cfg.RepoName, release.GetID(), opts, f)
	if err != nil {
		return "", fmt.Errorf("upload asset %s: %w", opts.Name, err)
	}

	return assets.GetBrowserDownloadURL(), nil
}

func assetName(path string) string {
	return path[lastSlash(path)+1:]
}

func lastSlash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}

func assetLabel(format string) string {
	switch format {
	case "spdx-json":
		return "SBOM (SPDX JSON)"
	case "cyclonedx-json":
		return "SBOM (CycloneDX JSON)"
	case "syft-json":
		return "SBOM (Syft JSON)"
	default:
		return "SBOM"
	}
}
