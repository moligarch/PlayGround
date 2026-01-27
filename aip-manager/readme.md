# AIP Manager (Go) – Standalone CLI

[![pipeline status](http://gitlab-agent.tblco.local/verkiani_m/aip-manager/badges/main/pipeline.svg)](http://gitlab-agent.tblco.local/verkiani_m/aip-manager/-/pipelines)

AIP Manager is a Windows‑focused Go CLI for generating Advanced Installer projects (.aip), building MSI packages, managing product versions, and (optionally) running Enigma pack profiles. This README documents only the **aip-manager** project so it can live independently of any installer repository.

---

## Features

* **Prepare**: Render a .aip from a template and a YAML config, encrypting sensitive INI content and injecting a structured ARPCOMMENTS note.
* **Build**: Compile MSI via Advanced Installer CLI with clear, non‑interleaved logs, per‑job timings, and optional timeouts.
* **Parallel multi‑config**: Provide multiple `--config` paths and each is processed concurrently.
* **Deterministic artifact naming**: `installer_<version>_<buildTypeLabel>_<HHmm>`.
* **Dry‑run**: Plan the AIP/MSI artifacts without writing files.
* **Reports & checksums**: Whole‑build JSON report and per‑job SHA256 files (opt‑in).
* **Versioning**: Show, set, or bump the version in `version.yaml` and persist it; AIP `ProductVersion` is written via XML (no AI scripting).
* **Enigma integration** (optional): Rewrite `.enigma64` profile paths to your workspace and invoke Enigma.

---

## Quick start

```powershell
# Build the CLI
PS> go build -o .\Manager.exe ./aip-manager

# Show current product version
PS> .\Manager.exe version show

# Prepare AIP from template + config
PS> .\Manager.exe prepare -t templates\base.aip -c configs\config.basic.yaml

# Build MSI (one or many configs, parallel)
PS> .\Manager.exe build -t templates\base.aip -c configs\config.basic.packed.yaml,configs\config.ai.packed.yaml --report --checksum --timeout 45m --ai-tail 120

# Dry-run planning (no writes)
PS> .\Manager.exe build -t templates\base.aip -c configs\config.basic.yaml --dry-run

# Run Enigma using a profile (optional)
PS> .\Manager.exe enigma --profile configs\enigma\profile.enigma64 --in installer_src\files\Unpacked --out installer_src\files\Packed --timeout 10m --tail 200
```

> The binary name is standardized to **`Manager.exe`** to avoid versioned filenames in scripts.

---

## Commands

### `build`

Generate .aip **and** build MSI for one or multiple configs in parallel.

**Flags:**

* `-t, --template <path>`: Path to base `.aip` template (required).
* `-c, --config <path[,path2,...]>`: YAML config(s). CSV or repeated flags (required).
* `--report`: Write a single `build_report_<HHmm>.json` in the date folder (whole‑build report with per‑job entries).
* `--checksum`: Write `<packageBase>.sha256` listing checksums for the `.aip` and `.msi` artifacts.
* `--timeout <dur>`: Bound Advanced Installer time (e.g., `30m`, `120s`).
* `--ai-tail <N>`: On AI error, include the last `N` lines of the AI console output in the error.
* `--dry-run`: Plan only—compute intended names/paths and version but write nothing.

**Output layout**

```
build/<YYYY-MM-DD>/
  installer_<version>_<buildTypeLabel>_<HHmm>.aip
  installer_<version>_<buildTypeLabel>_<HHmm>.msi
  installer_<version>_<buildTypeLabel>_<HHmm>.sha256   # if --checksum
  build_report_<HHmm>.json                              # if --report
```

**Build type label logic**

* Base: if the **config file name** contains `"ai"` → `ai`; otherwise `basic`.
* Suffix: if `build_type: packed` → `_packed`; otherwise no suffix.

Examples: `ai_packed`, `basic`.

---

### `prepare`

Generate only the `.aip`, with the same transformation pipeline as `build`.

**Flags:** same `-t/--template` and `-c/--config` as `build`.

Output: the `.aip` in `build/<YYYY-MM-DD>/`.

---

### `msi`

Build an MSI from an existing `.aip` file.

**Usage:**

```powershell
PS> .\Manager.exe msi path\to\file.aip --timeout 20m --ai-tail 120
```

---

### `version`

Manage `version.yaml` and the product version set into the AIP.

Subcommands:

* `version show` — display current version and the file path.
* `version set [a.b.c.d]` — set explicitly (or via `--major/--minor/--patch/--build`).
* `version bump --major|--minor|--patch|--build` — bump exactly one part.

> The AIP `ProductVersion` is updated via XML using `SetProductVersion` (no external AI command).

---

### `enigma` (optional)

Rewrite absolute paths inside an `.enigma64` profile to your workspace and run Enigma Protector.

**Flags:**

* `-p, --profile <file.enigma64>` — profile to rewrite and run.
* `-i, --in <dir>` — directory containing source (Unpacked) files.
* `-o, --out <dir>` — directory to place packed outputs.
* `--timeout <dur>` — max run time for Enigma.
* `--tail <N>` — tail the last `N` lines on error for quick triage.

**What is rewritten?**

* `<Input><FileName>` and `<Output><FileName>` (top‑level).
* `<AdvanceInput><Files><File><Input|Output>>` entries (advanced list).

---

## Configuration & Notes

* **YAML config** is parsed by `pkg/config`; plaintext INI blocks are generated in a stable order and **encrypted** via `pkg/crypto` before being written to the AIP registry.
* **ARPCOMMENTS**: We inject a JSON note (from `note_path`) into the AIP. A placeholder is used and spliced post‑serialization to preserve formatting for Windows Installer UI.
* **Source paths**: File `SourcePath` entries inside the AIP are resolved to absolute paths under your working tree according to the config `build_type` (e.g., prefer `Packed` vs `Unpacked`).
* **Advanced Installer path**: Discovered automatically or overridden with `ADVANCED_INSTALLER_PATH`.
* **Enigma path**: Discovered (typical install) or set via `ENIGMA_EXE`.

### `version_includes.yaml` (optional)

Use this file to inject **runtime file versions** into the MSI’s `ARPCOMMENTS` note JSON automatically during `prepare`/`build` (and in `--dry-run` planning). It lets you display component versions (e.g., `CyberCore.exe`, `CyberRes.dll`) in the installer metadata without hardcoding them.

**Location:**

* The manager looks for `version_includes.yaml` (relative to the working directory). If not found, it ignores quietly.

**Format:**

```yaml
includes:
  - key: cybercore
    path: installer_src/files/Common/CyberCore.exe
  - key: ui_version
    path: installer_src/files/Common/AgentNotification/AgentNotification.exe
  - key: cyberres
    path: installer_src/files/Common/CyberRes.dll
  - key: cybercomrules
    path: installer_src/files/Common/CyberCommonRules.dll
```

* `key`: The name that will appear under the `version` object in the note JSON.
* `path`: File to read the Windows file version from. Absolute paths are used as-is; relative paths are resolved against the project root. If the file can’t be found or has no version info, the value becomes `"n/a"` (build continues).

**How it is used:**

* On `prepare`/`build`, the manager reads each `path`, extracts its 4-part file version (when available), and merges into the note JSON under a `version` object:

```json
{
  "build_date": "2025-10-14T09:18:27Z",
  "version": {
    "cybercore": "1.2.3.4",
    "ui_version": "2.0.0.0",
    "cyberres": "1.0.130.0",
    "cybercomrules": "n/a"
  }
}
```

**CLI interaction:**

* No extra flags are needed. If `version_includes.yaml` is present, it’s applied automatically by `prepare`, `build`, and `--dry-run` planning.

**Notes & caveats:**

* Version extraction relies on Windows version resources. Plain files without version metadata will yield `"n/a"`.
* This enrichment complements the MSI **ProductVersion** managed by `aip-manager version` commands; it does not change the MSI version.
* The merged note is written into `ARPCOMMENTS` (escaped) so it shows up in **Control Panel → Programs → Details → Comments** and in the MSI summary.

---

## Development

* Run all tests: `go test ./...`
* Packages:

  * `pkg/aip` — AIP XML model & processors (`SetBuildOutput`, `ResolveSourcePaths`, `SetProductVersion`, …)
  * `pkg/config` — YAML loader and INI generators (order‑preserving)
  * `pkg/crypto` — encryption for registry values
  * `pkg/utils` — timers, AI runners, reporting, checksums
  * `pkg/versioning` — `version.yaml` parse/save, bump helpers
  * `pkg/enigma` — profile rewrite & runner

---

## Environment variables

* `ADVANCED_INSTALLER_PATH` — full path to `AdvancedInstaller.com` (optional).
* `ENIGMA_EXE` — full path to `enigma64.exe` (optional).

---

## Troubleshooting

* **AI not found** → set `ADVANCED_INSTALLER_PATH`.
* **Enigma not found** → set `ENIGMA_EXE`.
* **Interleaved logs** → `build` uses buffered per‑job logging; if you wrap in other tools, avoid mixing stdout/stderr asynchronously.

---

## License

Internal use. Adjust licensing per your organization’s policies.
