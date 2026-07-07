package sbom

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Richonn/sbomforge/internal/config"
)

var execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

func Generate(ctx context.Context, cfg *config.Config, format string) (string, error) {
	outputPath := filepath.Join(os.TempDir(), cfg.ArtifactName+"."+format+".json")

	source := cfg.ScanPath
	if cfg.Image != "" {
		source = "docker:" + cfg.Image
	}
	cmd := execCommand(ctx, "syft", "scan", source, "-o", format+"="+outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("syft scan failed: %w", err)
	}

	return outputPath, nil
}
