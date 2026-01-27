// pkg/utils/aip_ops.go
// Shared AIP/MSI helpers used by build/prepare/msi commands.
// This module exposes concise, non-printing operations so callers control logging.
//
// Exposed functions:
//   - PlanAIPConcise(templatePath, configPath) -> (plan, duration, err)  // in-memory AIP (no writes)
//   - GenerateAIPConcise(templatePath, configPath) -> (res, duration, err) // writes .aip per plan
//   - BuildFromAIPConcise(aipPath, advinstPathOpt) -> (aiOutput, duration, err)
//   - BuildFromAIPWithOpts(aipPath, opts) -> (aiOutput, duration, err) // timeout + tail-on-error
//   - ComputeBuildTypeLabel(cfgBuildType, configPath) -> "basic" | "ai" (+ optional "_packed")
//   - TailLines(s, n) -> string
//
// Artifact naming convention:
//   installer_<version>_<buildTypeLabel>
//   (version comes from version.yaml; ProductCode remains fixed and is not mutated here)

package utils

import (
	"aip-manager/pkg/aip"
	"aip-manager/pkg/config"
	"aip-manager/pkg/crypto"
	"aip-manager/pkg/versioning"
	"context"
	"encoding/json"
	"encoding/xml"

	// "encoding/json"
	// "encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ArpCommentsPlaceholder is injected into the AIP before XML serialization,
// then replaced in the serialized string with the escaped JSON payload.
const ArpCommentsPlaceholder = "___AIP_MANAGER_ARPCOMMENTS_PLACEHOLDER___"

// GenerateAIPResult captures key outputs from GenerateAIPConcise.
type GenerateAIPResult struct {
	// AIPPath is the fully-qualified path to the written .aip file.
	AIPPath string
	// PackageBase is the MSI base name (no extension), e.g., "installer_1.0.130.0_ai_packed".
	PackageBase string
	// Version is the ProductVersion that was written into the AIP (from version.yaml).
	Version string
	// BuildTypeLabel is the label used in names: basic|ai plus optional _packed suffix.
	BuildTypeLabel string
	// Versions is the merged version map injected into the note (from version_includes.yaml).
	Versions map[string]string
}

// PlannedAIP describes a fully prepared AIP (in-memory) without writing it.
type PlannedAIP struct {
	// Intended output locations.
	BuildDir     string // "build/msi/YYYY-MM-DD"
	AIPPath      string // BuildDir + "/" + PackageBase + ".aip"
	PackageBase  string // "installer_<version>_<buildTypeLabel>"
	Version      string // MSI ProductVersion (from version.yaml)
	BuildTypeLbl string // "basic" | "ai" (+ "_packed")
	// Rendered AIP XML with ARPCOMMENTS spliced; safe to write to disk as-is.
	AIPXML string
	// Versions merged from version_includes.yaml (key → "a.b.c.d" or "n/a").
	Versions map[string]string
}

// ComputeBuildTypeLabel builds the label used in artifact names.
func ComputeBuildTypeLabel(cfgBuildType, configPath string) string {
	name := strings.ToLower(strings.TrimSuffix(filepath.Base(configPath), filepath.Ext(configPath)))
	base := "basic"
	if strings.Contains(name, "ai") {
		base = "ai"
	}
	if strings.EqualFold(cfgBuildType, "packed") {
		return base + "_packed"
	}
	return base
}

// PlanAIPConcise prepares an AIP entirely in-memory:
// - Reads version.yaml and sets ProductVersion (keeps ProductCode fixed).
// - Applies encrypted registry values.
// - Resolves source paths (validates inputs).
// - Builds the ARPCOMMENTS payload (note + version_includes) and splices it.
// - Produces the final AIP XML string and planned output paths.
// No files are written. Callers can write plan.AIPXML to plan.AIPPath if desired.
func PlanAIPConcise(templatePath, configPath string) (*PlannedAIP, time.Duration, error) {
	start := time.Now()

	rootDir, err := os.Getwd()
	if err != nil {
		return nil, 0, fmt.Errorf("getwd: %w", err)
	}
	filesRoot := filepath.Join(rootDir, "installer_src", "files")

	// Load config and AIP template.
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, 0, fmt.Errorf("load config: %w", err)
	}
	doc, err := aip.LoadAIP(templatePath)
	if err != nil {
		return nil, 0, fmt.Errorf("load AIP template: %w", err)
	}

	// Versioning from version.yaml.
	v, _, err := versioning.FindAndLoad()
	if err != nil {
		return nil, 0, fmt.Errorf("version.yaml not found or invalid: %w", err)
	}
	verStr := v.String()

	// Set ProductVersion in the document (keeps ProductCode fixed by not touching it).
	if err := doc.SetProductVersion(verStr); err != nil {
		return nil, 0, fmt.Errorf("set ProductVersion: %w", err)
	}

	// Naming: installer_<version>_<buildTypeLabel>
	btLabel := ComputeBuildTypeLabel(cfg.BuildType, configPath)
	packageBase := fmt.Sprintf("installer_%s_%s", verStr, btLabel)

	// Set MSI output base name (Advanced Installer places MSI next to AIP).
	if err := doc.SetBuildOutput(".", packageBase); err != nil {
		return nil, 0, fmt.Errorf("set package output: %w", err)
	}

	// Registry values (plaintext → encrypt).
	modPlain := cfg.GenerateModulesPlaintext()
	encMod, err := crypto.Encrypt(modPlain)
	if err != nil {
		return nil, 0, fmt.Errorf("encrypt modules: %w", err)
	}
	if err := doc.SetRegistryValue("modules", encMod); err != nil {
		return nil, 0, fmt.Errorf("set modules reg: %w", err)
	}
	stsPlain := cfg.GenerateSTSPlaintext()
	encSTS, err := crypto.Encrypt(stsPlain)
	if err != nil {
		return nil, 0, fmt.Errorf("encrypt sts: %w", err)
	}
	if err := doc.SetRegistryValue("sts", encSTS); err != nil {
		return nil, 0, fmt.Errorf("set sts reg: %w", err)
	}

	// Resolve sources for Advanced Installer (validate inputs).
	if _, err = doc.ResolveSourcePaths(filesRoot, cfg.BuildType); err != nil {
		return nil, 0, fmt.Errorf("resolve source paths: %w", err)
	}

	// ARPCOMMENTS payload (note + optional version_includes).
	now := time.Now()
	if cfg.NotePath == "" {
		return nil, 0, fmt.Errorf("config %s missing required field `note_path`", configPath)
	}
	if err := doc.SetProperty("ARPCOMMENTS", ArpCommentsPlaceholder); err != nil {
		return nil, 0, fmt.Errorf("set ARPCOMMENTS placeholder: %w", err)
	}
	noteRaw, err := os.ReadFile(cfg.NotePath)
	if err != nil {
		return nil, 0, fmt.Errorf("read note JSON at %s: %w", cfg.NotePath, err)
	}
	var note map[string]any
	_ = json.Unmarshal(noteRaw, &note)
	note["build_date"] = now.Format(time.RFC3339)

	// Optional version_includes.yaml at aip-manager/.
	versMap := make(map[string]string)
	if vi, viErr := config.LoadVersionIncludes(filepath.Join(rootDir, "version_includes.yaml")); viErr == nil && vi != nil {
		for _, inc := range vi.Includes {
			if inc.Key == "" || inc.Path == "" {
				continue
			}
			abs := inc.Path
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(rootDir, inc.Path)
			}
			if _, statErr := os.Stat(abs); statErr != nil {
				alt := filepath.Join(filesRoot, inc.Path)
				if _, statAlt := os.Stat(alt); statAlt == nil {
					abs = alt
				}
			}
			if vstr, rErr := ReadFileVersion(abs); rErr == nil && vstr != "" {
				versMap[inc.Key] = vstr
			} else {
				versMap[inc.Key] = "n/a"
			}
		}
	}
	if len(versMap) > 0 {
		note["version"] = versMap
		note["core_version"] = versMap["CyberCore"]
	}

	// Serialize ARPCOMMENTS with XML escaping and newline substitution so it renders in MSI UI.
	pretty, _ := json.MarshalIndent(note, "", "    ")
	var esc strings.Builder
	_ = xml.EscapeText(&esc, pretty)
	finalComment := "###&#xA;" + strings.ReplaceAll(esc.String(), "\n", "&#xA;") + "&#xA;###"

	// Final AIP XML string (splice ARPCOMMENTS).
	xmlBytes, err := aip.GenerateAIPBytes(doc)
	if err != nil {
		return nil, 0, fmt.Errorf("generate AIP XML: %w", err)
	}
	xmlContent := strings.Replace(string(xmlBytes), ArpCommentsPlaceholder, finalComment, 1)
	// xmlContent := string(xmlBytes)

	// Planned output paths.
	buildDir := filepath.Join("build", now.Format("2006-01-02"))
	aipPath := filepath.Join(buildDir, packageBase+".aip")

	return &PlannedAIP{
		BuildDir:     buildDir,
		AIPPath:      aipPath,
		PackageBase:  packageBase,
		Version:      verStr,
		BuildTypeLbl: btLabel,
		AIPXML:       xmlContent,
		Versions:     versMap,
	}, time.Since(start), nil
}

// GenerateAIPConcise writes the planned AIP to disk.
// It reuses PlanAIPConcise for deterministic behavior.
func GenerateAIPConcise(templatePath, configPath string) (*GenerateAIPResult, time.Duration, error) {
	plan, dur, err := PlanAIPConcise(templatePath, configPath)
	if err != nil {
		return nil, 0, err
	}
	if err := os.MkdirAll(plan.BuildDir, 0o755); err != nil {
		return nil, 0, fmt.Errorf("create build dir: %w", err)
	}
	if err := os.WriteFile(plan.AIPPath, []byte(plan.AIPXML), 0o644); err != nil {
		return nil, 0, fmt.Errorf("save final AIP: %w", err)
	}
	return &GenerateAIPResult{
		AIPPath:        plan.AIPPath,
		PackageBase:    plan.PackageBase,
		Version:        plan.Version,
		BuildTypeLabel: plan.BuildTypeLbl,
		Versions:       plan.Versions,
	}, dur, nil
}

// BuildOptions controls the behavior of Advanced Installer CLI invocation.
type BuildOptions struct {
	AdvInstPath string        // override path; if empty, auto-detect/env
	Timeout     time.Duration // 0 = no timeout
	TailOnError int           // if >0, attach last N lines of AI output to the error
}

// TailLines returns the last n logical lines of s.
// It is Windows-safe: trims CRLFs and drops trailing blank lines so
// "last 2 lines" really means the last two non-empty lines of output.
func TailLines(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	// Split on LF, then normalize CRLF and drop trailing blanks.
	lines := strings.Split(s, "\n")

	// Trim trailing blanks (including lines that are just "\r").
	for len(lines) > 0 && strings.TrimRight(lines[len(lines)-1], "\r") == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return ""
	}

	// Strip trailing '\r' from each line.
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}

	if len(lines) <= n {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

// BuildFromAIPWithOpts runs Advanced Installer with timeout/tail-on-error.
func BuildFromAIPWithOpts(aipPath string, opts BuildOptions) (string, time.Duration, error) {
	start := time.Now()

	adv := opts.AdvInstPath
	var err error
	if adv == "" {
		adv = os.Getenv("ADVANCED_INSTALLER_PATH")
		if adv == "" {
			adv, err = FindAdvancedInstallerPath()
			if err != nil {
				return "", 0, fmt.Errorf("find Advanced Installer: %w", err)
			}
		}
	}

	// Always use a context (deadline if Timeout > 0).
	ctx := context.Background()
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, adv, "/build", aipPath)
	out, runErr := cmd.CombinedOutput()
	dur := time.Since(start)

	if runErr != nil {
		// Distinguish timeout vs generic failure.
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			if opts.TailOnError > 0 {
				return string(out), dur, fmt.Errorf("msi build timed out after %s\n--- last %d lines ---\n%s",
					opts.Timeout, opts.TailOnError, TailLines(string(out), opts.TailOnError))
			}
			return string(out), dur, fmt.Errorf("msi build timed out after %s", opts.Timeout)
		}
		if opts.TailOnError > 0 {
			return string(out), dur, fmt.Errorf("msi build failed: %v\n--- last %d lines ---\n%s",
				runErr, opts.TailOnError, TailLines(string(out), opts.TailOnError))
		}
		return string(out), dur, fmt.Errorf("msi build failed: %v", runErr)
	}
	return string(out), dur, nil
}

// BuildFromAIPConcise keeps backward-compatible behavior (no timeout, no tail).
func BuildFromAIPConcise(aipPath, advinstPathOpt string) (string, time.Duration, error) {
	return BuildFromAIPWithOpts(aipPath, BuildOptions{
		AdvInstPath: advinstPathOpt,
		Timeout:     0,
		TailOnError: 0,
	})
}
