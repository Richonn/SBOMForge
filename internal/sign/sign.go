package sign

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/Richonn/sbomforge/internal/config"
)

func Sign(ctx context.Context, cfg *config.Config, sbomPath string) (string, error) {
	if !cfg.Sign {
		return "", nil
	}
	bundlePath := sbomPath + ".bundle"

	cmd := exec.CommandContext(ctx, "cosign", "sign-blob", "--bundle="+bundlePath, "--yes", sbomPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	return bundlePath, nil
}
