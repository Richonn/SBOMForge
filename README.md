# SBOMForge

> GitHub Action that automatically generates a signed Software Bill of Materials (SBOM) for your project using Syft and Cosign, and attaches it to your GitHub releases.


---

## Why SBOMForge?

Software supply chain security is no longer optional. The **EU Cyber Resilience Act** and frameworks like SLSA require software producers to document the components they ship. SBOMForge automates this with zero configuration: generate, sign, and publish your SBOM as part of every release.

---

## Quick start

```yaml
on:
  release:
    types: [published]

jobs:
  sbom:
    runs-on: ubuntu-latest
    permissions:
      contents: write      # upload release asset
      id-token: write      # cosign keyless signing

    steps:
      - uses: actions/checkout@v4

      - uses: Richonn/sbomforge@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

---

## Inputs

| Input | Required | Default | Description |
|---|---|---|---|
| `github-token` | yes | — | GitHub token to upload the SBOM as a release asset |
| `format` | no | `spdx-json` | SBOM format: `spdx-json`, `cyclonedx-json`, `syft-json` |
| `artifact-name` | no | `sbom` | Output filename prefix |
| `sign` | no | `true` | Sign the SBOM with Cosign keyless |
| `attach-to-release` | no | `true` | Attach the SBOM to the GitHub Release |
| `upload-to-summary` | no | `true` | Show a summary in the GitHub Actions Job Summary |
| `scan-path` | no | `.` | Directory to scan (useful for monorepos) |
| `fail-on-error` | no | `true` | Fail the job if SBOM generation fails |

## Outputs

| Output | Description |
|---|---|
| `sbom-path` | Local path of the generated SBOM file |
| `sbom-url` | Download URL of the SBOM on the GitHub Release |
| `signature-bundle` | Path to the Cosign signature bundle |

---

## Verify the signature

Once the action runs, you can verify the SBOM signature locally:

```bash
cosign verify-blob \
  --bundle=sbom.spdx-json.json.bundle \
  sbom.spdx-json.json
```

---

## Supported formats

| Format | Flag | Output file |
|---|---|---|
| SPDX JSON | `spdx-json` | `sbom.spdx-json.json` |
| CycloneDX JSON | `cyclonedx-json` | `sbom.cyclonedx-json.json` |
| Syft JSON | `syft-json` | `sbom.syft-json.json` |

---

## Roadmap

- [ ] Docker image SBOM support
- [ ] Multiple formats in a single run
- [ ] Monorepo support
- [ ] SLSA attestation level 2
- [ ] Dry-run mode
- [ ] OCI registry upload (ghcr.io)

---

## License

MIT — Copyright 2026 Léandre Cacarié
