package provenance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Richonn/sbomforge/internal/config"
)

var execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

type statement struct {
	Type          string    `json:"_type"`
	PredicateType string    `json:"predicateType"`
	Subject       []subject `json:"subject"`
	Predicate     predicate `json:"predicate"`
}

type subject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

type predicate struct {
	Builder    builder    `json:"builder"`
	BuildType  string     `json:"buildType"`
	Invocation invocation `json:"invocation"`
	Metadata   metadata   `json:"metadata"`
	Materials  []material `json:"materials"`
}

type builder struct {
	ID string `json:"id"`
}

type invocation struct {
	ConfigSource configSource      `json:"configSource"`
	Environment  map[string]string `json:"environment"`
}

type configSource struct {
	URI        string            `json:"uri"`
	Digest     map[string]string `json:"digest"`
	EntryPoint string            `json:"entryPoint"`
}

type metadata struct {
	BuildInvocationID string       `json:"buildInvocationId"`
	BuildStartedOn    time.Time    `json:"buildStartedOn"`
	Completeness      completeness `json:"completeness"`
	Reproducible      bool         `json:"reproducible"`
}

type completeness struct {
	Parameters  bool `json:"parameters"`
	Environment bool `json:"environment"`
	Materials   bool `json:"materials"`
}

type material struct {
	URI    string            `json:"uri"`
	Digest map[string]string `json:"digest"`
}

func Generate(ctx context.Context, cfg *config.Config, sbomPath string) (string, error) {
	digest, err := sha256File(sbomPath)
	if err != nil {
		return "", fmt.Errorf("compute sbom digest: %w", err)
	}

	serverURL := os.Getenv("GITHUB_SERVER_URL")
	if serverURL == "" {
		serverURL = "https://github.com"
	}
	repoURI := fmt.Sprintf("git+%s/%s/%s", serverURL, cfg.RepoOwner, cfg.RepoName)
	runID := os.Getenv("GITHUB_RUN_ID")
	workflow := os.Getenv("GITHUB_WORKFLOW")
	sha := os.Getenv("GITHUB_SHA")
	ref := os.Getenv("GITHUB_REF")

	stmt := statement{
		Type:          "https://in-toto.io/Statement/v0.1",
		PredicateType: "https://slsa.dev/provenance/v0.2",
		Subject: []subject{
			{
				Name:   filepath.Base(sbomPath),
				Digest: map[string]string{"sha256": digest},
			},
		},
		Predicate: predicate{
			Builder: builder{
				ID: "https://github.com/Richonn/SBOMForge",
			},
			BuildType: "https://github.com/Richonn/SBOMForge/action@v1",
			Invocation: invocation{
				ConfigSource: configSource{
					URI:        repoURI + "@" + ref,
					Digest:     map[string]string{"sha1": sha},
					EntryPoint: workflow,
				},
				Environment: map[string]string{
					"github_run_id":      runID,
					"github_run_attempt": os.Getenv("GITHUB_RUN_ATTEMPT"),
					"github_sha":         sha,
					"github_ref":         ref,
					"github_repository":  cfg.RepoOwner + "/" + cfg.RepoName,
					"github_event_name":  cfg.EventName,
					"github_workflow":    workflow,
				},
			},
			Metadata: metadata{
				BuildInvocationID: fmt.Sprintf("%s/%s/%s/actions/runs/%s", serverURL, cfg.RepoOwner, cfg.RepoName, runID),
				BuildStartedOn:    time.Now().UTC(),
				Completeness: completeness{
					Parameters:  true,
					Environment: true,
					Materials:   false,
				},
				Reproducible: false,
			},
			Materials: []material{
				{
					URI:    repoURI,
					Digest: map[string]string{"sha1": sha},
				},
			},
		},
	}

	provPath := sbomPath + ".provenance"
	data, err := json.MarshalIndent(stmt, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal provenance: %w", err)
	}
	if err := os.WriteFile(provPath, data, 0644); err != nil {
		return "", fmt.Errorf("write provenance: %w", err)
	}

	cmd := execCommand(ctx, "cosign", "sign-blob",
		"--bundle", provPath+".bundle",
		"--yes",
		provPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cosign sign-blob: %w", err)
	}

	return provPath, nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
