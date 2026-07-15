package provenance

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/Richonn/sbomforge/internal/config"
)

func fakeExecCommandSuccess(ctx context.Context, name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(0)
}

func cfg() *config.Config {
	return &config.Config{
		RepoOwner: "owner",
		RepoName:  "repo",
		EventName: "release",
	}
}

func TestGenerate_CreatesProvenanceFile(t *testing.T) {
	sbomFile, err := os.CreateTemp(t.TempDir(), "sbom-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = sbomFile.WriteString(`{"artifacts":[]}`)
	_ = sbomFile.Close()

	orig := execCommand
	execCommand = fakeExecCommandSuccess
	defer func() { execCommand = orig }()

	provPath, err := Generate(context.Background(), cfg(), sbomFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(provPath); os.IsNotExist(err) {
		t.Errorf("provenance file not created at %s", provPath)
	}
}

func TestGenerate_ProvenanceIsValidJSON(t *testing.T) {
	sbomFile, err := os.CreateTemp(t.TempDir(), "sbom-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = sbomFile.WriteString(`{"artifacts":[]}`)
	_ = sbomFile.Close()

	orig := execCommand
	execCommand = fakeExecCommandSuccess
	defer func() { execCommand = orig }()

	provPath, err := Generate(context.Background(), cfg(), sbomFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(provPath)
	if err != nil {
		t.Fatalf("read provenance: %v", err)
	}
	var stmt statement
	if err := json.Unmarshal(data, &stmt); err != nil {
		t.Fatalf("provenance is not valid JSON: %v", err)
	}
	if stmt.PredicateType != "https://slsa.dev/provenance/v0.2" {
		t.Errorf("predicateType = %q, want https://slsa.dev/provenance/v0.2", stmt.PredicateType)
	}
	if len(stmt.Subject) != 1 {
		t.Errorf("expected 1 subject, got %d", len(stmt.Subject))
	}
}

func TestGenerate_SubjectDigestMatchesSBOM(t *testing.T) {
	sbomFile, err := os.CreateTemp(t.TempDir(), "sbom-*.json")
	if err != nil {
		t.Fatal(err)
	}
	content := `{"artifacts":[]}`
	_, _ = sbomFile.WriteString(content)
	_ = sbomFile.Close()

	orig := execCommand
	execCommand = fakeExecCommandSuccess
	defer func() { execCommand = orig }()

	provPath, err := Generate(context.Background(), cfg(), sbomFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(provPath)
	if err != nil {
		t.Fatal(err)
	}
	var stmt statement
	_ = json.Unmarshal(data, &stmt)

	digest := stmt.Subject[0].Digest["sha256"]
	if digest == "" {
		t.Error("subject sha256 digest is empty")
	}
}

func TestGenerate_CosignFailure(t *testing.T) {
	sbomFile, err := os.CreateTemp(t.TempDir(), "sbom-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = sbomFile.WriteString(`{"artifacts":[]}`)
	_ = sbomFile.Close()

	orig := execCommand
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "false")
	}
	defer func() { execCommand = orig }()

	_, err = Generate(context.Background(), cfg(), sbomFile.Name())
	if err == nil {
		t.Error("expected error when cosign fails, got nil")
	}
	if !strings.Contains(err.Error(), "cosign sign-blob") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSha256File(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "test-*")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("hello")
	_ = f.Close()

	digest, err := sha256File(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// sha256("hello") = 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if digest != want {
		t.Errorf("sha256 = %q, want %q", digest, want)
	}
}

func TestSha256File_Missing(t *testing.T) {
	_, err := sha256File("/nonexistent/file")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
