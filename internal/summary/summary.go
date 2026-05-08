package summary

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Richonn/sbomforge/internal/config"
)

type sbomDoc struct {
	Packages   []any `json:"packages"`   // spdx-json
	Components []any `json:"components"` // cyclonedx-json
	Artifacts  []any `json:"artifacts"`  // syft-json
}

func Write(cfg *config.Config, sbomPath, sbomURL, bundlePath string) error {
	if !cfg.UploadToSummary {
		return nil
	}

	count := countComponents(sbomPath)

	releaseLink := sbomURL
	if sbomURL != "" {
		releaseLink = fmt.Sprintf("[Download](%s)", sbomURL)
	}

	md := fmt.Sprintf(`## SBOMForge — SBOM Generated

| Field      | Value |
|------------|-------|
| Format     | %s |
| Components | %d |
| Signed     | %v |
| Release    | %s |
`, cfg.Format, count, cfg.Sign, releaseLink)

	return cfg.WriteSummary(md)
}

func countComponents(sbomPath string) int {
	data, err := os.ReadFile(sbomPath)
	if err != nil {
		return 0
	}
	var doc sbomDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return 0
	}
	return len(doc.Packages) + len(doc.Components) + len(doc.Artifacts)
}
