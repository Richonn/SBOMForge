package oci

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/Richonn/sbomforge/internal/config"
)

func fakeExecCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
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

func TestAttach_Success(t *testing.T) {
	cfg := &config.Config{
		OCIImage: "ghcr.io/owner/app@sha256:abc123",
	}

	sbomFile, err := os.CreateTemp(t.TempDir(), "sbom-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_ = sbomFile.Close()

	orig := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = orig }()

	if err := Attach(context.Background(), cfg, sbomFile.Name()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAttach_Failure(t *testing.T) {
	cfg := &config.Config{
		OCIImage: "ghcr.io/owner/app@sha256:abc123",
	}

	orig := execCommand
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "false")
	}
	defer func() { execCommand = orig }()

	if err := Attach(context.Background(), cfg, "/tmp/sbom.json"); err == nil {
		t.Error("expected error when cosign fails, got nil")
	}
}
