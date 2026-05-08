package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Richonn/sbomforge/internal/config"
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
		return fmt.Errorf("error: %v\n", err)
	}

	sbomPath, err := sbom.Generate(ctx, cfg)
	handleErr(cfg, err, "generate sbom")
	_ = cfg.WriteOutput("sbom-path", sbomPath)

	bundlePath, err := sign.Sign(ctx, cfg, sbomPath)
	handleErr(cfg, err, "sign sbom")
	_ = cfg.WriteOutput("signature-bundle", bundlePath)

	client := release.New(cfg)
	sbomURL, err := client.Upload(ctx, sbomPath, bundlePath)
	handleErr(cfg, err, "upload to release")
	_ = cfg.WriteOutput("sbom-url", sbomURL)

	err = summary.Write(cfg, sbomPath, sbomURL, bundlePath)
	handleErr(cfg, err, "write summary")

	return err
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
