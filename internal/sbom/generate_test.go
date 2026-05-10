package sbom

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/Richonn/sbomforge/internal/config"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	for _, a := range args {
		if strings.Contains(a, "=") && !strings.HasPrefix(a, "--") {
			parts := strings.SplitN(a, "=", 2)
			if len(parts) == 2 {
				_ = os.WriteFile(parts[1], []byte(`{"artifacts":[]}`), 0644)
			}
		}
	}
	os.Exit(0)
}

func fakeExecCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestGenerate_Success(t *testing.T) {
	cfg := &config.Config{
		ArtifactName: "sbom",
		Format:       "spdx-json",
		ScanPath:     ".",
	}

	origExecCommand := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = origExecCommand }()

	path, err := Generate(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("output file not created at %s", path)
	}
}

func TestGenerate_OutputPath(t *testing.T) {
	cfg := &config.Config{
		ArtifactName: "mysbom",
		Format:       "cyclonedx-json",
		ScanPath:     ".",
	}

	origExecCommand := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = origExecCommand }()

	path, err := Generate(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(path, "mysbom.cyclonedx-json.json") {
		t.Errorf("unexpected path: %s", path)
	}
}

func TestGenerate_DockerImage(t *testing.T) {
	cfg := &config.Config{
		ArtifactName: "sbom",
		Format:       "spdx-json",
		Image:        "alpine:3.21",
	}

	origExecCommand := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = origExecCommand }()

	path, err := Generate(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestGenerate_DockerImageSource(t *testing.T) {
	cfg := &config.Config{
		ArtifactName: "sbom",
		Format:       "spdx-json",
		Image:        "alpine:3.21",
	}

	var capturedArgs []string
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		capturedArgs = args
		return fakeExecCommand(ctx, name, args...)
	}
	defer func() { execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, name, args...)
	}}()

	_, _ = Generate(context.Background(), cfg)

	if len(capturedArgs) == 0 || capturedArgs[1] != "docker:alpine:3.21" {
		t.Errorf("expected source docker:alpine:3.21, got args: %v", capturedArgs)
	}
}

func TestGenerate_SyftFailure(t *testing.T) {
	cfg := &config.Config{
		ArtifactName: "sbom",
		Format:       "spdx-json",
		ScanPath:     ".",
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, "false") // always exits 1
		return cmd
	}
	defer func() {
		execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		}
	}()

	_, err := Generate(context.Background(), cfg)
	if err == nil {
		t.Error("expected error when syft fails, got nil")
	}
}
