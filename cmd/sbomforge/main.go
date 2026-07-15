package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Richonn/sbomforge/internal/config"
	"github.com/Richonn/sbomforge/internal/oci"
	"github.com/Richonn/sbomforge/internal/release"
	"github.com/Richonn/sbomforge/internal/sbom"
	"github.com/Richonn/sbomforge/internal/sign"
	"github.com/Richonn/sbomforge/internal/summary"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "SBOMForge: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	var sbomPaths, bundlePaths, sbomURLs []string

	client := release.New(cfg)
	for _, format := range cfg.Formats {
		sbomPath, err := sbom.Generate(ctx, cfg, format)
		handleErr(cfg, err, "generate sbom")

		var bundlePath string
		if !cfg.DryRun {
			bundlePath, err = sign.Sign(ctx, cfg, sbomPath)
			handleErr(cfg, err, "sign sbom")
		}

		var sbomURL string
		if !cfg.DryRun {
			sbomURL, err = client.Upload(ctx, sbomPath, bundlePath, format)
			handleErr(cfg, err, "upload to release")
		}

		if !cfg.DryRun && cfg.OCIImage != "" {
			err = oci.Attach(ctx, cfg, sbomPath)
			handleErr(cfg, err, "attach sbom to oci image")
		}

		err = summary.Write(cfg, format, sbomPath, sbomURL, bundlePath)
		handleErr(cfg, err, "write summary")

		sbomPaths = append(sbomPaths, sbomPath)
		bundlePaths = append(bundlePaths, bundlePath)
		sbomURLs = append(sbomURLs, sbomURL)
	}

	_ = cfg.WriteOutput("sbom-path", strings.Join(sbomPaths, ","))
	_ = cfg.WriteOutput("signature-bundle", strings.Join(bundlePaths, ","))
	_ = cfg.WriteOutput("sbom-url", strings.Join(sbomURLs, ","))

	return nil
}

func handleErr(cfg *config.Config, err error, msg string) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "error: %s: %v\n", msg, err)
	if cfg.FailOnError {
		os.Exit(1)
	}
}
