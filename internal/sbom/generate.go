package sbom

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Richonn/sbomforge/internal/config"
)

func Generate(ctx context.Context, cfg *config.Config) (string, error) {
	outputPath := filepath.Join(os.TempDir(), cfg.ArtifactName+"."+cfg.Format+".json")

	cmd := exec.CommandContext(ctx, "syft", "scan", cfg.ScanPath, "-o", cfg.Format+"="+outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("syft scan failed: %w", err)
	}

	return outputPath, nil
}
