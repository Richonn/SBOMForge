package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("INPUT_GITHUB-TOKEN", "ghp_test")
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")
	t.Setenv("GITHUB_REF_NAME", "v1.0.0")
	t.Setenv("GITHUB_EVENT_NAME", "release")
}

func TestLoad_Defaults(t *testing.T) {
	setBaseEnv(t)

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(c.Formats, []string{"spdx-json"}) {
		t.Errorf("Formats = %q, want [spdx-json]", c.Formats)
	}
	if c.ArtifactName != "sbom" {
		t.Errorf("ArtifactName = %q, want sbom", c.ArtifactName)
	}
	if len(c.ScanPaths) != 1 || c.ScanPaths[0] != "." {
		t.Errorf("ScanPaths = %q, want [.]", c.ScanPaths)
	}
	if !c.Sign {
		t.Error("Sign = false, want true")
	}
	if !c.AttachToRelease {
		t.Error("AttachToRelease = false, want true")
	}
	if !c.UploadToSummary {
		t.Error("UploadToSummary = false, want true")
	}
	if !c.FailOnError {
		t.Error("FailOnError = false, want true")
	}
}

func TestLoad_GitHubRepository(t *testing.T) {
	setBaseEnv(t)

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.RepoOwner != "owner" {
		t.Errorf("RepoOwner = %q, want owner", c.RepoOwner)
	}
	if c.RepoName != "repo" {
		t.Errorf("RepoName = %q, want repo", c.RepoName)
	}
	if c.RefName != "v1.0.0" {
		t.Errorf("RefName = %q, want v1.0.0", c.RefName)
	}
	if c.EventName != "release" {
		t.Errorf("EventName = %q, want release", c.EventName)
	}
}

func TestLoad_MissingToken(t *testing.T) {
	t.Setenv("INPUT_GITHUB-TOKEN", "")
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")

	_, err := Load()
	if err == nil {
		t.Error("expected error for missing token, got nil")
	}
}

func TestLoad_InvalidFormat(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("INPUT_FORMAT", "unknown-format")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid format, got nil")
	}
}

func TestLoad_ValidFormats(t *testing.T) {
	formats := []string{"spdx-json", "cyclonedx-json", "syft-json"}

	for _, f := range formats {
		t.Run(f, func(t *testing.T) {
			setBaseEnv(t)
			t.Setenv("INPUT_FORMAT", f)

			c, err := Load()
			if err != nil {
				t.Fatalf("unexpected error for format %q: %v", f, err)
			}
			if !reflect.DeepEqual(c.Formats, []string{f}) {
				t.Errorf("Formats = %q, want [%q]", c.Formats, f)
			}
		})
	}
}

func TestLoad_MultipleFormats(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("INPUT_FORMAT", "spdx-json,cyclonedx-json")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(c.Formats, []string{"spdx-json", "cyclonedx-json"}) {
		t.Errorf("Formats = %q, want [spdx-json cyclonedx-json]", c.Formats)
	}
}

func TestLoad_MultipleFormatsWithSpaces(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("INPUT_FORMAT", "spdx-json, cyclonedx-json , syft-json")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(c.Formats, []string{"spdx-json", "cyclonedx-json", "syft-json"}) {
		t.Errorf("Formats = %q, want [spdx-json cyclonedx-json syft-json]", c.Formats)
	}
}

func TestLoad_MonorepoScanPaths(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("INPUT_SCAN-PATH", "services/api,services/worker")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.ScanPaths) != 2 || c.ScanPaths[0] != "services/api" || c.ScanPaths[1] != "services/worker" {
		t.Errorf("ScanPaths = %q, want [services/api services/worker]", c.ScanPaths)
	}
}

func TestLoad_MonorepoScanPathsWithSpaces(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("INPUT_SCAN-PATH", "services/api , services/worker , services/frontend")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.ScanPaths) != 3 {
		t.Errorf("ScanPaths = %q, want 3 paths", c.ScanPaths)
	}
}

func TestLoad_OCIImageDefault(t *testing.T) {
	setBaseEnv(t)

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.OCIImage != "" {
		t.Errorf("OCIImage = %q, want empty", c.OCIImage)
	}
}

func TestLoad_OCIImageSet(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("INPUT_OCI-IMAGE", "ghcr.io/owner/app@sha256:abc123")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.OCIImage != "ghcr.io/owner/app@sha256:abc123" {
		t.Errorf("OCIImage = %q, want ghcr.io/owner/app@sha256:abc123", c.OCIImage)
	}
}

func TestLoad_DryRunDefault(t *testing.T) {
	setBaseEnv(t)

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DryRun {
		t.Error("DryRun = true, want false by default")
	}
}

func TestLoad_DryRunEnabled(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("INPUT_DRY-RUN", "true")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.DryRun {
		t.Error("DryRun = false, want true")
	}
}

func TestLoad_InvalidRepository(t *testing.T) {
	t.Setenv("INPUT_GITHUB-TOKEN", "ghp_test")
	t.Setenv("GITHUB_REPOSITORY", "invalid-no-slash")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid GITHUB_REPOSITORY, got nil")
	}
}

func TestLoad_ImageInput(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("INPUT_IMAGE", "alpine:3.21")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Image != "alpine:3.21" {
		t.Errorf("Image = %q, want alpine:3.21", c.Image)
	}
}

func TestLoad_BoolInputs(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("INPUT_SIGN", "false")
	t.Setenv("INPUT_ATTACH-TO-RELEASE", "false")
	t.Setenv("INPUT_UPLOAD-TO-SUMMARY", "false")
	t.Setenv("INPUT_FAIL-ON-ERROR", "false")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Sign {
		t.Error("Sign = true, want false")
	}
	if c.AttachToRelease {
		t.Error("AttachToRelease = true, want false")
	}
	if c.UploadToSummary {
		t.Error("UploadToSummary = true, want false")
	}
	if c.FailOnError {
		t.Error("FailOnError = true, want false")
	}
}

func TestWriteOutput(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "output-*")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	c := &Config{OutputFile: f.Name()}
	if err := c.WriteOutput("sbom-path", "/tmp/sbom.spdx.json"); err != nil {
		t.Fatalf("WriteOutput error: %v", err)
	}

	got, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	want := "sbom-path=/tmp/sbom.spdx.json\n"
	if string(got) != want {
		t.Errorf("WriteOutput wrote %q, want %q", string(got), want)
	}
}

func TestWriteSummary(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "summary-*")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	c := &Config{SummaryFile: f.Name()}
	if err := c.WriteSummary("## SBOM Generated"); err != nil {
		t.Fatalf("WriteSummary error: %v", err)
	}

	got, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	want := "## SBOM Generated\n"
	if string(got) != want {
		t.Errorf("WriteSummary wrote %q, want %q", string(got), want)
	}
}

func TestWriteOutput_NoFile(t *testing.T) {
	c := &Config{OutputFile: ""}
	if err := c.WriteOutput("key", "value"); err != nil {
		t.Errorf("expected nil when OutputFile is empty, got: %v", err)
	}
}

func TestWriteSummary_NoFile(t *testing.T) {
	c := &Config{SummaryFile: ""}
	if err := c.WriteSummary("something"); err != nil {
		t.Errorf("expected nil when SummaryFile is empty, got: %v", err)
	}
}

func TestParseBool(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"1", true},
		{"yes", true},
		{"Yes", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"", false},
		{"random", false},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := parseBool(tc.input)
			if got != tc.want {
				t.Errorf("parseBool(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestGetEnvDefault(t *testing.T) {
	t.Setenv("TEST_EXISTING_KEY", "hello")

	if got := getEnvDefault("TEST_EXISTING_KEY", "fallback"); got != "hello" {
		t.Errorf("got %q, want hello", got)
	}
	if got := getEnvDefault("TEST_MISSING_KEY_XYZ", "fallback"); got != "fallback" {
		t.Errorf("got %q, want fallback", got)
	}
}

func TestWriteOutput_BadPath(t *testing.T) {
	c := &Config{OutputFile: filepath.Join(t.TempDir(), "nonexistent", "output")}
	if err := c.WriteOutput("key", "value"); err == nil {
		t.Error("expected error for bad path, got nil")
	}
}
