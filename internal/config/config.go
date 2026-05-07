package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	GitHubToken     string
	Format          string
	ArtifactName    string
	Sign            bool
	AttachToRelease bool
	UploadToSummary bool
	ScanPath        string
	FailOnError     bool

	RepoOwner   string
	RepoName    string
	RefName     string
	EventName   string
	OutputFile  string
	SummaryFile string
}

func Load() (*Config, error) {
	c := &Config{}

	c.GitHubToken = getEnvDefault("INPUT_GITHUB_TOKEN", "")
	if c.GitHubToken == "" {
		return nil, fmt.Errorf("INPUT_GITHUB_TOKEN is required")
	}

	c.Format = getEnvDefault("INPUT_FORMAT", "spdx-json")
	c.ArtifactName = getEnvDefault("INPUT_ARTIFACT_NAME", "sbom")
	c.Sign = parseBool(getEnvDefault("INPUT_SIGN", "true"))
	c.AttachToRelease = parseBool(getEnvDefault("INPUT_ATTACH_TO_RELEASE", "true"))
	c.UploadToSummary = parseBool(getEnvDefault("INPUT_UPLOAD_TO_SUMMARY", "true"))
	c.ScanPath = getEnvDefault("INPUT_SCAN_PATH", ".")
	c.FailOnError = parseBool(getEnvDefault("INPUT_FAIL_ON_ERROR", "true"))

	validFormats := map[string]bool{
		"spdx-json":      true,
		"cyclonedx-json": true,
		"syft-json":      true,
	}
	if !validFormats[c.Format] {
		return nil, fmt.Errorf("invalid format %q: must be spdx-json, cyclonedx-json or syft-json", c.Format)
	}

	repo := os.Getenv("GITHUB_REPOSITORY")
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("GITHUB_REPOSITORY %q has unexpected format", repo)
	}
	c.RepoOwner = parts[0]
	c.RepoName = parts[1]

	c.RefName = os.Getenv("GITHUB_REF_NAME")
	c.EventName = os.Getenv("GITHUB_EVENT_NAME")
	c.OutputFile = os.Getenv("GITHUB_OUTPUT")
	c.SummaryFile = os.Getenv("GITHUB_STEP_SUMMARY")

	return c, nil
}

func (c *Config) WriteOutput(key, value string) error {
	if c.OutputFile == "" {
		return nil
	}
	f, err := os.OpenFile(c.OutputFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open GITHUB_OUTPUT: %w", err)
	}
	defer func() { _ = f.Close() }()
	_, err = fmt.Fprintf(f, "%s=%s\n", key, value)
	return err
}

func (c *Config) WriteSummary(value string) error {
	if c.SummaryFile == "" {
		return nil
	}
	f, err := os.OpenFile(c.SummaryFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open GITHUB_STEP_SUMMARY: %w", err)
	}
	defer func() { _ = f.Close() }()
	_, err = fmt.Fprintf(f, "%s\n", value)
	return err
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}
