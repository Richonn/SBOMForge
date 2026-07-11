# SBOMForge

> GitHub Action that automatically generates a signed Software Bill of Materials (SBOM) for your project using Syft and Cosign, and attaches it to your GitHub releases.

[![Marketplace](https://img.shields.io/badge/GitHub%20Marketplace-SBOMForge-blue?logo=github)](https://github.com/marketplace/actions/sbomforge)
![CI](https://github.com/Richonn/SBOMForge/actions/workflows/ci.yml/badge.svg)
![License](https://img.shields.io/github/license/Richonn/SBOMForge)
![Go version](https://img.shields.io/github/go-mod/go-version/Richonn/SBOMForge)
![Latest release](https://img.shields.io/github/v/release/Richonn/SBOMForge)

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
      - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2

      - uses: Richonn/SBOMForge@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

---

## Inputs

| Input | Required | Default | Description |
|---|---|---|---|
| `github-token` | yes | — | GitHub token to upload the SBOM as a release asset |
| `format` | no | `spdx-json` | SBOM format(s): `spdx-json`, `cyclonedx-json`, `syft-json`. Comma-separated for multiple (e.g. `spdx-json,cyclonedx-json`) |
| `artifact-name` | no | `sbom` | Output filename prefix |
| `sign` | no | `true` | Sign the SBOM with Cosign keyless |
| `attach-to-release` | no | `true` | Attach the SBOM to the GitHub Release |
| `upload-to-summary` | no | `true` | Show a summary in the GitHub Actions Job Summary |
| `scan-path` | no | `.` | Directory to scan (useful for monorepos) |
| `image` | no | — | Docker image to scan (e.g. `alpine:3.21`, `ghcr.io/org/app:latest`). If set, `scan-path` is ignored |
| `fail-on-error` | no | `true` | Fail the job if SBOM generation fails |
| `dry-run` | no | `false` | Generate the SBOM without signing or uploading. Useful for testing |

## Outputs

| Output | Description |
|---|---|
| `sbom-path` | Local path(s) of the generated SBOM file(s). Comma-separated when multiple formats are used |
| `sbom-url` | Download URL(s) of the SBOM(s) on the GitHub Release. Comma-separated when multiple formats are used |
| `signature-bundle` | Path(s) to the Cosign signature bundle(s). Comma-separated when multiple formats are used |

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

## Multiple formats

To generate SBOMs in multiple formats in a single run:

```yaml
- uses: Richonn/SBOMForge@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    format: spdx-json,cyclonedx-json
```

All formats are generated, signed, and uploaded to the release in one pass. Outputs (`sbom-path`, `sbom-url`, `signature-bundle`) are comma-separated when multiple formats are used.

---

## Dry-run mode

To generate the SBOM without signing or uploading (useful for PRs or testing):

```yaml
- uses: Richonn/SBOMForge@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    dry-run: "true"
```

The SBOM is still generated and visible in the Job Summary, but no signature is created and nothing is uploaded to the release.

---

## Docker image scanning

To scan a Docker image instead of source code, pass the `image` input:

```yaml
- uses: Richonn/SBOMForge@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    image: alpine:3.21
```

When `image` is set, `scan-path` is ignored.

---

## Roadmap

- [x] Docker image SBOM support
- [x] Multiple formats in a single run
- [ ] Monorepo support
- [ ] SLSA attestation level 2
- [x] Dry-run mode
- [ ] OCI registry upload (ghcr.io)

---

## License

MIT — Copyright 2026 Léandre Cacarié
