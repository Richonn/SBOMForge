package oci

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/Richonn/sbomforge/internal/config"
)

var execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

func Attach(ctx context.Context, cfg *config.Config, sbomPath string) error {
	cmd := execCommand(ctx, "cosign", "attach", "sbom", "--sbom", sbomPath, cfg.OCIImage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cosign attach sbom failed: %w", err)
	}

	return nil
}
