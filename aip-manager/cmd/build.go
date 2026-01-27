// cmd/build.go
// Parallel, concise logging build with dry-run, whole-build report, per-job checksums,
// optional Advanced Installer timeout, and tail-on-error.
// - One goroutine per config; each job logs to a buffer.
// - Main goroutine prints each job's result as soon as it's ready (no interleaving).
// - Per-job: header (Config/Note/Build), Generate AIP time, Build MSI time, Total.
// - If --report, writes one JSON report for the whole build at build/msi/<date>/build_report.json
//   with a "jobs" array (one entry per config).
// - If --checksum, writes <packageBase>.sha256 beside each MSI (skipped in --dry-run).
// - If --timeout, bounds Advanced Installer build duration.
// - If --ai-tail, prints last N lines of AI output within the error for easier debugging.

package cmd

import (
	"aip-manager/pkg/config"
	"aip-manager/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type jobResult struct {
	index  int
	text   string
	err    error
	report utils.BuildJobReport
}

var (
	// Populated via flags.
	cfgPaths        []string
	dryRun          bool
	reportEnabled   bool
	checksumEnabled bool
	timeoutDur      time.Duration
	aiTailLines     int
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build MSI(s) from .aip generated per config (parallel, concise logs).",
	Run: func(cmd *cobra.Command, args []string) {
		overall := utils.StartTimer("🚀 Starting Absolute path build ...\n")
		_ = godotenv.Load()

		templatePath, _ := cmd.Flags().GetString("template")
		if len(cfgPaths) == 0 {
			fmt.Fprintln(os.Stderr, "Error: no --config provided (CSV or repeated flag)")
			os.Exit(1)
		}

		// Whole-build report path (date-based folder).
		now := time.Now()
		dateDir := now.Format("2006-01-02")
		buildDirForReport := filepath.Join("build", dateDir)

		n := len(cfgPaths)
		results := make(chan jobResult, n)
		var wg sync.WaitGroup
		wg.Add(n)

		for i, cp := range cfgPaths {
			i := i
			cp := cp
			fmt.Printf("==> [%d/%d] build started\n", i+1, n)

			go func() {
				defer wg.Done()

				start := time.Now()
				var b strings.Builder

				// Load config for header & report fields.
				cfg, err := config.LoadConfig(cp)
				if err != nil {
					results <- jobResult{
						index: i,
						err:   fmt.Errorf("load config (%s): %w", cp, err),
						report: utils.BuildJobReport{
							ConfigPath: cp,
							Status:     "error",
							Error:      err.Error(),
						},
					}
					return
				}
				buildLabel := utils.ComputeBuildTypeLabel(cfg.BuildType, cp)

				// Header (buffered).
				fmt.Fprintf(&b,
					`=============================================
   +  Index:      %d
   +  Config:     %s
   +  Note:       %s
   +  Build:      %s

`, i+1, cp, cfg.NotePath, buildLabel)

				var (
					jobRep   utils.BuildJobReport
					aiLog    string
					genDur   time.Duration
					buildDur time.Duration
				)

				jobRep.ConfigPath = cp
				jobRep.NotePath = cfg.NotePath
				jobRep.BuildTypeLabel = buildLabel

				if dryRun {
					// Plan only (no writes, no AI CLI).
					plan, planDur, err := utils.PlanAIPConcise(templatePath, cp)
					if err != nil {
						results <- jobResult{
							index: i,
							text:  b.String(),
							err:   fmt.Errorf("plan AIP (%s): %w", cp, err),
							report: utils.BuildJobReport{
								ConfigPath: cp,
								NotePath:   cfg.NotePath,
								Status:     "error",
								Error:      err.Error(),
							},
						}
						return
					}
					genDur = planDur
					jobRep.Status = "dry-run"
					jobRep.ProductVersion = plan.Version
					jobRep.PackageBase = plan.PackageBase
					jobRep.AIPPath = plan.AIPPath // intended path
					jobRep.MSIPath = utils.InferMSIPath(plan.AIPPath, plan.PackageBase)
					jobRep.Versions = plan.Versions

					// QoL: show intended artifact locations
					fmt.Fprintf(&b, "   📄 AIP (planned): %s\n", jobRep.AIPPath)
					fmt.Fprintf(&b, "   📦 MSI (planned): %s\n", jobRep.MSIPath)
				} else {
					// Generate .aip and build MSI.
					res, gDur, err := utils.GenerateAIPConcise(templatePath, cp)
					if err != nil {
						results <- jobResult{
							index: i,
							text:  b.String(),
							err:   fmt.Errorf("generate AIP (%s): %w", cp, err),
							report: utils.BuildJobReport{
								ConfigPath: cp,
								NotePath:   cfg.NotePath,
								Status:     "error",
								Error:      err.Error(),
							},
						}
						return
					}
					genDur = gDur

					aiOut, bDur, err := utils.BuildFromAIPWithOpts(res.AIPPath, utils.BuildOptions{
						AdvInstPath: "",
						Timeout:     timeoutDur,
						TailOnError: aiTailLines,
					})
					if err != nil {
						results <- jobResult{
							index: i,
							text:  b.String(),
							err:   fmt.Errorf("build MSI (%s): %w", cp, err),
							report: utils.BuildJobReport{
								ConfigPath: cp,
								NotePath:   cfg.NotePath,
								AIPPath:    res.AIPPath,
								Status:     "error",
								Error:      err.Error(),
							},
						}
						return
					}
					buildDur = bDur
					aiLog = aiOut

					jobRep.Status = "success"
					jobRep.ProductVersion = res.Version
					jobRep.PackageBase = res.PackageBase
					jobRep.AIPPath = res.AIPPath
					jobRep.MSIPath = utils.InferMSIPath(res.AIPPath, res.PackageBase)
					jobRep.Versions = res.Versions

					// QoL: show actual artifact locations
					fmt.Fprintf(&b, "   📄 AIP: %s\n", jobRep.AIPPath)
					fmt.Fprintf(&b, "   📦 MSI: %s\n", jobRep.MSIPath)
				}

				totalDur := time.Since(start)
				jobRep.Timings.GenerateAIP = genDur
				jobRep.Timings.BuildMSI = buildDur
				jobRep.Timings.Total = totalDur

				// Timings summary.
				fmt.Fprintf(&b, "   ⏱  Generate AIP: %s\n", genDur.Truncate(time.Millisecond))
				if dryRun {
					fmt.Fprintf(&b, "   ⏱  Build MSI:    %s (skipped)\n", 0*time.Second)
				} else {
					fmt.Fprintf(&b, "   ⏱  Build MSI:    %s\n", buildDur.Truncate(time.Millisecond))
				}
				fmt.Fprintf(&b, "   ⏱  Total:        %s\n", totalDur.Truncate(time.Millisecond))

				if !dryRun && aiLog != "" {
					fmt.Fprintf(&b, "\n---------- Log ----------\n%s", aiLog)
				}
				fmt.Fprint(&b, "=============================================")

				results <- jobResult{index: i, text: b.String(), report: jobRep}
			}()
		}

		// Close results after workers finish.
		go func() {
			wg.Wait()
			close(results)
		}()

		// Aggregate reports and stream-print job logs.
		var (
			hadErr     bool
			allReports []utils.BuildJobReport
		)
		hostWD, _ := os.Getwd()

		for res := range results {
			if res.text != "" {
				fmt.Println(res.text)
			}
			if res.err != nil {
				hadErr = true
				fmt.Fprintf(os.Stderr, "Error: %v\n", res.err)
			}
			// Collect report and optionally write per-job checksums.
			if res.report.ConfigPath != "" {
				allReports = append(allReports, res.report)

				// Per-job checksums beside MSI (skip in dry-run).
				if checksumEnabled && !dryRun && res.report.Status == "success" {
					dir := filepath.Dir(res.report.AIPPath)
					ckPath := filepath.Join(dir, res.report.PackageBase+".sha256")
					files := []string{res.report.AIPPath, res.report.MSIPath}
					if err := utils.WriteChecksums(ckPath, files); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to write checksums for %s: %v\n", res.report.PackageBase, err)
					} else {
						fmt.Printf("🔐 Checksums written: %s\n", ckPath)
					}
				}
			}
		}

		// Whole-build report (once), if enabled.
		if reportEnabled {
			wholeReportPath := filepath.Join(buildDirForReport, fmt.Sprintf("build_report_%s.json", now.Format("2006-01-02T15")))
			if err := os.MkdirAll(buildDirForReport, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create report dir: %v\n", err)
			} else {
				buildReport := utils.BuildReport{
					GeneratedAt: time.Now(),
					HostWD:      hostWD,
					Jobs:        allReports,
				}
				if err := utils.WriteBuildReport(wholeReportPath, buildReport); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to write build report: %v\n", err)
				} else {
					fmt.Printf("\n📝 Build report written: %s\n", wholeReportPath)
				}
			}
		}

		// Clean Advanced Installer caches in today's build folder (once).
		if !dryRun {
			if removed, paths, err := utils.CleanAdvInstallerCaches(buildDirForReport); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to clean Advanced Installer caches: %v\n", err)
			} else if removed > 0 {
				fmt.Printf("🧹 Removed %d Advanced Installer cache director%s in %s:\n",
					removed, map[bool]string{true: "ies", false: "y"}[removed != 1], buildDirForReport)
				for _, p := range paths {
					fmt.Printf("   • %s\n", p)
				}
			} else {
				fmt.Printf("🧹 No Advanced Installer cache directories found in %s\n", buildDirForReport)
			}
		}

		utils.EndTimer(overall)

		if hadErr {
			os.Exit(1)
		}
		fmt.Println("🎉 All MSI builds completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringSliceVarP(&cfgPaths, "config", "c", nil, "Path(s) to config YAML file(s) (CSV or repeated)")
	buildCmd.Flags().StringP("template", "t", "templates/base.aip", "Path to the template .aip file")
	buildCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Plan builds without writing MSI (skips Advanced Installer)")
	buildCmd.Flags().BoolVar(&reportEnabled, "report", false, "Write one JSON report for the whole build (build/msi/<date>/build_report.json)")
	buildCmd.Flags().BoolVar(&checksumEnabled, "checksum", false, "Write per-job SHA256 checksums beside each MSI (skipped in --dry-run)")
	buildCmd.Flags().DurationVar(&timeoutDur, "timeout", 0, "Timeout for Advanced Installer build (e.g. 30m, 120s). 0 = no timeout")
	buildCmd.Flags().IntVar(&aiTailLines, "ai-tail", 0, "On error, include last N lines of Advanced Installer output")
	_ = buildCmd.MarkFlagRequired("config")

	// Normalize accidental empty entries if someone writes "--config ,a.yaml"
	cobra.OnInitialize(func() {
		if len(cfgPaths) == 0 {
			return
		}
		clean := make([]string, 0, len(cfgPaths))
		for _, v := range cfgPaths {
			if s := strings.TrimSpace(v); s != "" {
				clean = append(clean, s)
			}
		}
		cfgPaths = clean
	})
}
