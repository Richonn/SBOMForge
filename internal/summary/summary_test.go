package summary

import (
	"os"
	"strings"
	"testing"

	"github.com/Richonn/sbomforge/internal/config"
)

func TestWrite_Disabled(t *testing.T) {
	cfg := &config.Config{UploadToSummary: false}

	if err := Write(cfg, "", "", "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWrite_ContainsFields(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "summary-*")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	sbomFile, err := os.CreateTemp(t.TempDir(), "sbom-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = sbomFile.WriteString(`{"artifacts":[{},{}]}`)
	_ = sbomFile.Close()

	cfg := &config.Config{
		UploadToSummary: true,
		Sign:            true,
		SummaryFile:     f.Name(),
	}

	if err := Write(cfg, "spdx-json", sbomFile.Name(), "https://example.com/sbom.json", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	content := string(got)

	for _, want := range []string{"SBOMForge", "spdx-json", "true", "Download"} {
		if !strings.Contains(content, want) {
			t.Errorf("summary missing %q\ngot:\n%s", want, content)
		}
	}
}

func TestWrite_NoURL(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "summary-*")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	cfg := &config.Config{
		UploadToSummary: true,
		Sign:            false,
		SummaryFile:     f.Name(),
	}

	if err := Write(cfg, "syft-json", "/nonexistent/sbom.json", "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	// sbomURL is empty, so releaseLink stays empty string
	if strings.Contains(string(got), "Download") {
		t.Error("should not contain Download link when sbomURL is empty")
	}
}

func TestCountComponents_SPDX(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "sbom-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`{"packages":[{},{},{}]}`)
	_ = f.Close()

	if got := countComponents(f.Name()); got != 3 {
		t.Errorf("countComponents = %d, want 3", got)
	}
}

func TestCountComponents_CycloneDX(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "sbom-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`{"components":[{},{}]}`)
	_ = f.Close()

	if got := countComponents(f.Name()); got != 2 {
		t.Errorf("countComponents = %d, want 2", got)
	}
}

func TestCountComponents_InvalidFile(t *testing.T) {
	if got := countComponents("/nonexistent/file.json"); got != 0 {
		t.Errorf("countComponents = %d, want 0 for missing file", got)
	}
}

func TestCountComponents_InvalidJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "sbom-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`not json`)
	_ = f.Close()

	if got := countComponents(f.Name()); got != 0 {
		t.Errorf("countComponents = %d, want 0 for invalid JSON", got)
	}
}
