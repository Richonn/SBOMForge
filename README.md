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
| `scan-path` | no | `.` | Directory or directories to scan. Comma-separated for monorepos (e.g. `services/api,services/worker`). If `image` is set, this is ignored |
| `image` | no | — | Docker image to scan (e.g. `alpine:3.21`, `ghcr.io/org/app:latest`). If set, `scan-path` is ignored |
| `fail-on-error` | no | `true` | Fail the job if SBOM generation fails |
| `dry-run` | no | `false` | Generate the SBOM without signing or uploading. Useful for testing |
| `attest` | no | `false` | Generate a signed SLSA provenance attestation for each SBOM. Requires `id-token: write` permission |
| `oci-image` | no | — | OCI image reference to attach the SBOM to (e.g. `ghcr.io/org/app@sha256:...`). Requires prior login with `docker/login-action` |

## Outputs

| Output | Description |
|---|---|
| `sbom-path` | Local path(s) of the generated SBOM file(s). Comma-separated when multiple formats or paths are used |
| `sbom-url` | Download URL(s) of the SBOM(s) on the GitHub Release. Comma-separated when multiple formats or paths are used |
| `signature-bundle` | Path(s) to the Cosign signature bundle(s). Comma-separated when multiple formats or paths are used |

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

When scanning multiple paths, the directory basename is included in the filename: `sbom.api.spdx-json.json`, `sbom.worker.spdx-json.json`.

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

## OCI registry upload

To attach the SBOM directly to an image in a registry, use the `oci-image` input. This makes the SBOM discoverable by tools like `cosign`, `grype`, or `syft` without needing to find it in the release assets.

Authenticate first with `docker/login-action`, then pass the image reference:

```yaml
- uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}

- uses: Richonn/SBOMForge@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    oci-image: ghcr.io/owner/app@sha256:abc123
```

The SBOM can then be retrieved with:

```bash
cosign download sbom ghcr.io/owner/app@sha256:abc123
```

---

## SLSA provenance attestation

To generate a signed SLSA provenance attestation alongside each SBOM, enable the `attest` input:

```yaml
jobs:
  sbom:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      id-token: write      # required for cosign keyless signing

    steps:
      - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2

      - uses: Richonn/SBOMForge@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          attest: "true"
```

For each SBOM file generated, SBOMForge produces a `.provenance` file — an [in-toto](https://in-toto.io) statement with a [SLSA 0.2](https://slsa.dev/provenance/v0.2) predicate — signed with Cosign keyless and uploaded to the release. It records:

- The SHA256 digest of the SBOM
- The source commit (`GITHUB_SHA`), ref, and workflow
- The build invocation ID (link to the Actions run)

Verify the provenance signature locally:

```bash
cosign verify-blob \
  --bundle=sbom.spdx-json.json.provenance.bundle \
  sbom.spdx-json.json.provenance
```

---

## Monorepo support

To scan multiple directories in a single run, pass a comma-separated list to `scan-path`:

```yaml
- uses: Richonn/SBOMForge@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    scan-path: "services/api,services/worker,services/frontend"
```

Each path generates its own SBOM per format. Output files are named using the directory basename: `sbom.api.spdx-json.json`, `sbom.worker.spdx-json.json`, etc. All outputs (`sbom-path`, `sbom-url`, `signature-bundle`) are comma-separated.

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
- [x] Monorepo support
- [x] SLSA attestation level 2
- [x] Dry-run mode
- [x] OCI registry upload (ghcr.io)

---

## License

MIT — Copyright 2026 Léandre Cacarié
