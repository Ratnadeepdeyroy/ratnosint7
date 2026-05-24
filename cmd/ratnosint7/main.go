// Copyright 2026 Ratnadeep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/config"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/engine"
	apperrors "github.com/Ratnadeepdeyroy/ratnosint7/pkg/errors"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/installer"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/parser"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/plugins"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/runner"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/ui"
	"github.com/Ratnadeepdeyroy/ratnosint7/pkg/ui/tui"
	"github.com/charmbracelet/lipgloss"

	"github.com/spf13/cobra"
)

var (
	verbose         bool
	quiet           bool
	debug           bool
	silent          bool
	noDashboardScan bool
	passive         bool
	active          bool
	both            bool
	timeout         string
	outputFormat    string
	overwrite       bool
	pprofOn         bool
	useCache        bool
	configDirScan   string
	configDirTools  string
)

func main() {
	orange := "\033[38;5;208m"
	reset := "\033[0m"

	root := &cobra.Command{
		Use:   "ratnosint7",
		Short: "High-performance OSINT subdomain enumeration engine",
		Long: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true).Render("RATNOSINT7") +
			"\n\nHigh-performance Go subdomain reconnaissance orchestration engine.\n" +
			"Runs 13 enumeration tools concurrently, deduplicates results, and streams to disk.\n\n" +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Render("Commands:") + "\n" +
			"  ratnosint7 scan              — Start interactive subdomain scan\n" +
			"  ratnosint7 update-tools      — Install or update enumeration tools\n\n" +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Render("Global Flags:") + "\n" +
			"  --verbose                    — Verbose output\n" +
			"  --debug                      — Debug logging\n" +
			"  --silent                     — No CLI output (file still written)\n" +
			"  --quiet                      — Minimal output\n",
	}

	root.PersistentFlags().BoolVar(&verbose, "verbose", false, orange+"verbose slog output"+reset)
	root.PersistentFlags().BoolVar(&quiet, "quiet", false, orange+"minimal slog output"+reset)
	root.PersistentFlags().BoolVar(&silent, "silent", false, orange+"no CLI UI output (still writes results)"+reset)
	root.PersistentFlags().BoolVar(&debug, "debug", false, orange+"debug slog output and subprocess logs"+reset)

	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan a domain for subdomains (interactive mode)",
		Args:  cobra.NoArgs,
		Long: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true).Render("SCAN") +
			"\n\nStart an interactive subdomain enumeration scan.\n" +
			"Prompts for: scan mode (passive/active/both) → cache option → target domain.\n\n" +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Render("Examples:") + "\n" +
			"  ratnosint7 scan                      — Interactive mode (prompts all options)\n" +
			"  ratnosint7 scan --passive --cache    — Passive scan with cache enabled\n" +
			"  ratnosint7 scan --active             — Active scan (DNS brute force)\n" +
			"  ratnosint7 scan --both               — Passive + active scan (merged results)\n" +
			"  ratnosint7 scan --format json        — JSON output\n",
		RunE:  runScan,
	}
	scanCmd.Flags().BoolVar(&passive, "passive", false, orange+"run only passive plugins"+reset)
	scanCmd.Flags().BoolVar(&active, "active", false, orange+"run only active plugins"+reset)
	scanCmd.Flags().BoolVar(&both, "both", false, orange+"run passive + active plugins (merged results)"+reset)
	scanCmd.Flags().StringVar(&timeout, "timeout", "30m", orange+"scan timeout"+reset)
	scanCmd.Flags().StringVar(&outputFormat, "format", "txt", orange+"output format (txt|json|csv)"+reset)
	scanCmd.Flags().BoolVar(&overwrite, "overwrite", false, orange+"overwrite output file"+reset)
	scanCmd.Flags().BoolVar(&pprofOn, "pprof", false, orange+"start pprof server on :6060"+reset)
	scanCmd.Flags().BoolVar(&useCache, "cache", false, orange+"use cached results (24h TTL)"+reset)
	scanCmd.Flags().BoolVar(&noDashboardScan, "no-dashboard", true, orange+"disable fullscreen scan dashboard (TTY only)"+reset)
	scanCmd.Flags().StringVar(&configDirScan, "config", "configs", orange+"config directory"+reset)

	updateToolsCmd := &cobra.Command{
		Use:   "update-tools",
		Short: "Install or update enumeration tools",
		Long: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true).Render("UPDATE-TOOLS") +
			"\n\nInstall or update all enumeration tools.\n" +
			"Idempotent — safe to run multiple times.\n\n" +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Render("Examples:") + "\n" +
			"  ratnosint7 update-tools              — Install all tools\n" +
			"  ratnosint7 update-tools --verbose    — Show detailed install logs\n\n" +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Render("Tools installed:") + "\n" +
			"  Passive:  subfinder, amass, findomain, assetfinder, sublist3r,\n" +
			"            subscraper, dome, as3nt, substr3am, tugarecon (needs python3.12)\n" +
			"  Active:   amass, findomain, dnsx, tugarecon (needs python3.12)\n\n" +
			"  Prerequisites: " + installer.SuggestedInstallFix("tugarecon") + "\n",
		RunE:  runUpdateTools,
	}
	updateToolsCmd.Flags().StringVar(&configDirTools, "config", "configs", orange+"config directory"+reset)

	root.AddCommand(scanCmd)
	root.AddCommand(updateToolsCmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func canRunPlugin(tc *config.ToolConfig, mode string) bool {
	switch mode {
	case "active", "both":
		return tc.ActiveRun != "" || tc.PassiveRun != ""
	default:
		return tc.PassiveRun != ""
	}
}

func promptScanOptions() (mode string, cacheEnable bool, domain string, err error) {
	st := ui.NewStyles(ui.ThemeDefault, ui.IsInteractive(os.Stdout))

	fmt.Println()
	fmt.Println(st.Title.Render("ratnosint7 recon engine"))
	fmt.Println(st.Dim.Render("======================"))
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Scan mode with validation loop
	for {
		fmt.Println(st.Title.Render("Scan mode:"))
		fmt.Println(st.Accent.Render("  [1]") + " passive  — OSINT only, no DNS interaction")
		fmt.Println(st.Accent.Render("  [2]") + " active   — DNS brute force + validation")
		fmt.Println(st.Accent.Render("  [3]") + " both     — OSINT + DNS brute force (merged results)")
		fmt.Print(st.Running.Render("Select [1/2/3]: "))
		modeInput, _ := reader.ReadString('\n')
		modeInput = strings.TrimSpace(strings.ToLower(modeInput))
		if modeInput == "active" || modeInput == "2" {
			mode = "active"
			break
		} else if modeInput == "passive" || modeInput == "1" {
			mode = "passive"
			break
		} else if modeInput == "both" || modeInput == "3" {
			mode = "both"
			break
		}
		fmt.Println(st.ErrorStyled.Render("⚠️  Please select 1, 2, or 3") + "\n")
	}

	// Cache option with validation loop
	fmt.Println()
	for {
		fmt.Println(st.Title.Render("Cache (24h TTL):"))
		fmt.Println(st.Accent.Render("  [1]") + " yes — use cached results if available")
		fmt.Println(st.Accent.Render("  [2]") + " no  — run fresh scan")
		fmt.Print(st.Running.Render("Select [1/2]: "))
		cacheInput, _ := reader.ReadString('\n')
		cacheInput = strings.TrimSpace(strings.ToLower(cacheInput))
		if cacheInput == "yes" || cacheInput == "y" || cacheInput == "1" {
			cacheEnable = true
			break
		} else if cacheInput == "no" || cacheInput == "n" || cacheInput == "2" {
			cacheEnable = false
			break
		}
		fmt.Println(st.ErrorStyled.Render("⚠️  Please select 1 or 2") + "\n")
	}

	// Target domain with validation loop
	fmt.Println()
	for {
		fmt.Print(st.Running.Render("Target domain: "))
		domain, _ = reader.ReadString('\n')
		domain = strings.TrimSpace(domain)
		if parser.IsValidDomain(domain) {
			break
		}
		fmt.Printf(st.ErrorStyled.Render("⚠️  Invalid domain: %s. Try again.")+"\n", domain)
	}

	fmt.Println()
	fmt.Println(st.Success.Render("[✓] Domain validated"))
	fmt.Println()
	fmt.Printf(st.Box.Render(fmt.Sprintf("Mode: %s | Cache: %v | Domain: %s", st.Accent.Render(mode), cacheEnable, st.Accent.Render(domain))+"\n") + "\n")

	return mode, cacheEnable, domain, nil
}

func setupLogging() {
	var level slog.Level
	switch {
	case debug:
		level = slog.LevelDebug
	case silent || quiet:
		level = slog.LevelWarn
	case verbose:
		level = slog.LevelInfo
	default:
		level = slog.LevelError
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
}

func writeErrorLog(domain string, errors []*apperrors.PluginError) error {
	if len(errors) == 0 {
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	logDir := filepath.Join(home, ".ratnosint7", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		slog.Warn("failed to create log directory", "dir", logDir, "err", err)
		return err
	}
	logPath := filepath.Join(logDir, fmt.Sprintf("errors_%d.log", time.Now().Unix()))
	f, err := os.Create(logPath)
	if err != nil {
		slog.Warn("failed to create error log", "path", logPath, "err", err)
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "ratnosint7 error log\n")
	fmt.Fprintf(f, "====================\n")
	fmt.Fprintf(f, "Domain: %s\n", domain)
	fmt.Fprintf(f, "Time: %s\n\n", time.Now().Format(time.RFC3339))

	for i, e := range errors {
		if e == nil {
			continue
		}
		fmt.Fprintf(f, "[%d] %s\n", i+1, e.Plugin)
		fmt.Fprintf(f, "    Code: %s\n", e.Code)
		if apperrors.IsFatal(e.Code) {
			fmt.Fprintf(f, "    Severity: FATAL\n")
		} else if apperrors.IsRetryable(e.Code) {
			fmt.Fprintf(f, "    Severity: RETRYABLE\n")
		}
		if e.Reason != "" {
			fmt.Fprintf(f, "    Reason: %s\n", e.Reason)
		}
		if e.Fix != "" {
			fmt.Fprintf(f, "    Fix: %s\n", e.Fix)
		}
		if e.Err != nil {
			fmt.Fprintf(f, "    Error: %v\n", e.Err)
		}
		fmt.Fprintf(f, "\n")
	}

	fmt.Printf("Error log saved to: %s\n", logPath)
	return nil
}

func runScan(cmd *cobra.Command, args []string) error {
	setupLogging()
	runner.SetDebugLogs(debug)

	th := ui.ThemeDefault

	var err error
	var scanMode string
	var cacheEnabled bool
	var domain string

	// Always interactive
	scanMode, cacheEnabled, domain, err = promptScanOptions()
	if err != nil {
		return err
	}

	// Command-line flags override prompts
	if both {
		scanMode = "both"
	} else if passive && !active {
		scanMode = "passive"
	} else if active && !passive {
		scanMode = "active"
	}
	if cmd.Flags().Changed("cache") {
		cacheEnabled = useCache
	}

	if !parser.IsValidDomain(domain) {
		return fmt.Errorf("invalid domain: %s", domain)
	}

	colorStdout := ui.IsInteractive(os.Stdout)

	if pprofOn {
		go func() {
			_ = http.ListenAndServe(":6060", nil)
		}()
		slog.Debug("pprof server on :6060")
	}

	cfg, err := config.Load(configDirScan)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if len(cfg.Tools) == 0 {
		cfg, _ = config.Load(".")
	}
	if len(cfg.Tools) == 0 {
		return fmt.Errorf("no tools configured in %s/tools.yaml", configDirScan)
	}

	timeoutDur, err := time.ParseDuration(timeout)
	if err != nil {
		timeoutDur = 30 * time.Minute
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if timeoutDur > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeoutDur)
		defer cancel()
	}

	var uiBus *ui.Bus
	if !silent {
		uiBus = ui.NewBus()
		defer uiBus.CloseSubscribers()
	}

	var pluginList []engine.Plugin
	var pluginNames []string
	for i := range cfg.Tools {
		tc := &cfg.Tools[i]
		pl := plugins.ConfigPlugin{Config: tc, Mode: scanMode}
		if !canRunPlugin(tc, scanMode) {
			continue
		}

		if !installer.ToolExists(tc.Name) && !installer.ToolExistsWithWorkDir(tc.Name, tc.WorkDir) {
			spec := pl.InstallSpec()
			var installErr error
			switch spec.Name {
			case "findomain":
				if debug {
					slog.Debug("installing tool", "tool", spec.Name)
				}
				installErr = installer.Install(ctx, spec.Name, spec.Cmd, spec.Args)
				if installErr != nil {
					if debug {
						slog.Debug("trying brew", "tool", spec.Name)
					}
					installErr = installer.Install(ctx, spec.Name, "brew", []string{"install", "findomain"})
				}
				if installErr != nil {
					if debug {
						slog.Debug("downloading from GitHub", "tool", spec.Name)
					}
					installErr = installer.InstallFindomainFromGitHub(ctx)
				}
			default:
				if debug {
					slog.Debug("installing tool", "tool", spec.Name)
				}
				installErr = installer.InstallFromConfig(ctx, *tc)
			}
			if installErr != nil {
				fix := installer.SuggestedInstallFix(spec.Name)
				pe := apperrors.DetailedPluginError(spec.Name, apperrors.InstallFailed, installErr.Error(), fix, installErr)
				if uiBus != nil {
					uiBus.Publish(ui.PluginSkipped{Name: spec.Name, Err: pe})
				} else {
					slog.Warn("tool unavailable, skipping plugin", "tool", spec.Name, "err", installErr)
				}
				continue
			}
		}

		if err := pl.Preflight(ctx); err != nil {
			slog.Warn("preflight failed", "plugin", pl.Name(), "err", err)
		}

		pluginList = append(pluginList, pl)
		pluginNames = append(pluginNames, tc.Name)
	}
	if len(pluginList) == 0 {
		return fmt.Errorf("no plugins available (install failed or --passive/--active/--both filter excludes all)")
	}

	resolversHash := ""
	if len(cfg.Resolvers) > 0 {
		resolversHash = fmt.Sprintf("%d", len(cfg.Resolvers))
	}

	dashboardOn := colorStdout && !silent && !noDashboardScan

	var wg sync.WaitGroup
	var result *engine.ScanResult
	var scanErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		passiveMode := scanMode == "passive"
		activeMode := scanMode == "active"
		bothMode := scanMode == "both"
		result, scanErr = engine.Scan(ctx, domain, engine.ScanConfig{
			Config:       cfg,
			Plugins:      pluginList,
			PluginNames:  pluginNames,
			Passive:      passiveMode,
			Active:       activeMode,
			Both:         bothMode,
			Resolvers:    resolversHash,
			OutputFormat: outputFormat,
			Overwrite:    overwrite,
			NoCache:      !cacheEnabled,
			Events:       uiBus,
			DebugLogs:    debug,
		})
	}()

	if dashboardOn && uiBus != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := tui.RunScanDashboard(ctx, domain, uiBus, th, colorStdout); err != nil {
				slog.Debug("dashboard stopped", "err", err)
			}
		}()
	}

	if !(dashboardOn || silent) && uiBus != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ui.ConsumeScan(ctx, os.Stdout, uiBus, th, colorStdout); err != nil &&
				!errors.Is(err, context.Canceled) {
				slog.Warn("static scan UI stopped", "err", err)
			}
		}()
	}

	wg.Wait()

	if scanErr != nil {
		return scanErr
	}
	if silent {
		return nil
	}

	// Show scan summary and first 50 lines of output
	if result != nil && !silent {
		fmt.Println()
		fmt.Println("=====================================")
		fmt.Printf("Scan Settings: Mode=%s | Cache=%v\n", scanMode, cacheEnabled)
		fmt.Printf("Domain: %s\n", domain)
		fmt.Printf("Total Subdomains: %d\n", result.TotalDomains)
		fmt.Printf("Output File: %s\n", result.OutputPath)
		fmt.Println("=====================================")
		fmt.Println()

		// Show first 50 lines
		if result.TotalDomains > 0 {
			file, err := os.Open(result.OutputPath)
			if err == nil {
				defer file.Close()
				scanner := bufio.NewScanner(file)
				lineCount := 0
				fmt.Println("First subdomains:")
				for scanner.Scan() && lineCount < 50 {
					fmt.Println(scanner.Text())
					lineCount++
				}
				if result.TotalDomains > 50 {
					fmt.Printf("\n... and %d more (see full output in %s)\n", result.TotalDomains-50, result.OutputPath)
				}
			}
		}
		fmt.Println()
	}

	if result != nil {
		allErrs := result.Errors
		for pluginName, stat := range result.PluginStats {
			if stat.Err != nil {
				plugErr := apperrors.NewPluginError(pluginName, apperrors.Unknown, stat.Err)
				allErrs = append(allErrs, plugErr)
			}
		}
		if len(allErrs) > 0 {
			_ = writeErrorLog(domain, allErrs)
		}
	}

	return nil
}

func runUpdateTools(cmd *cobra.Command, args []string) error {
	setupLogging()
	runner.SetDebugLogs(debug)

	color := ui.IsInteractive(os.Stdout)
	st := ui.NewStyles(ui.ThemeDefault, color)

	ctx := context.Background()
	tools, err := config.LoadTools(filepath.Join(configDirTools, "tools.yaml"))
	if err != nil {
		tools, _ = config.LoadTools("tools.yaml")
	}
	if len(tools) == 0 {
		return fmt.Errorf("no tools configured")
	}

	if !silent {
		printUpdateToolsHeader(st, color, len(tools))
	}

	started := time.Now()
	ok, fail := 0, 0
	total := len(tools)
	for i, tc := range tools {
		idx := i + 1
		if !silent {
			printToolInstallStart(os.Stdout, st, color, idx, total, tc.Name)
			flushStdout()
		}

		installErr := installer.InstallFromConfig(ctx, tc)
		if tc.Name == "findomain" && installErr != nil {
			installErr = retryFindomainInstall(ctx, tc.Name, installErr, st, color, !silent)
		}

		if installErr != nil {
			fail++
			if !silent {
				printToolInstallFailure(os.Stdout, st, color, tc.Name, installErr, verbose || debug)
				flushStdout()
			}
			slog.Warn("install failed", "tool", tc.Name, "err", installErr)
		} else {
			ok++
			if !silent {
				printToolInstallSuccess(os.Stdout, st, color, tc.Name)
				flushStdout()
			}
		}
	}

	if silent {
		return nil
	}
	lines := []string{
		st.Title.Render("Update tools summary"),
		fmt.Sprintf("Success:  %d / %d", ok, total),
		fmt.Sprintf("Failed:   %d / %d", fail, total),
		fmt.Sprintf("Duration: %s", time.Since(started).Round(time.Second)),
	}
	if fail > 0 {
		lines = append(lines, st.Dim.Render("Re-run after fixing failed tools, or install them manually."))
	}
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	fmt.Fprint(os.Stdout, "\n"+st.Box.Render(body)+"\n")

	return nil
}

func printUpdateToolsHeader(st ui.Styles, color bool, total int) {
	fmt.Fprintln(os.Stdout)
	if color {
		fmt.Fprintln(os.Stdout, st.Title.Render("ratnosint7 — install enumeration tools"))
		fmt.Fprintf(os.Stdout, "%s\n\n", st.Dim.Render(fmt.Sprintf("%d tools configured", total)))
	} else {
		fmt.Println("ratnosint7 — install enumeration tools")
		fmt.Printf("%d tools configured\n\n", total)
	}
}

func printToolInstallStart(w *os.File, st ui.Styles, color bool, idx, total int, name string) {
	label := fmt.Sprintf("[%d/%d] installing %s …", idx, total, name)
	if color {
		fmt.Fprintln(w, st.Running.Render(label))
	} else {
		fmt.Fprintln(w, label)
	}
}

func printToolInstallSuccess(w *os.File, st ui.Styles, color bool, name string) {
	if color {
		fmt.Fprintf(w, "         %s  %s installed successfully\n", st.Success.Render("✓"), name)
	} else {
		fmt.Fprintf(w, "         OK   %s installed successfully\n", name)
	}
}

func printToolInstallFailure(w *os.File, st ui.Styles, color bool, name string, installErr error, fullDetail bool) {
	if color {
		fmt.Fprintf(w, "         %s  %s installation failed\n", st.ErrorStyled.Render("✗"), name)
	} else {
		fmt.Fprintf(w, "         FAIL %s installation failed\n", name)
	}
	detail := installErr.Error()
	if !fullDetail {
		detail = firstInstallErrorLine(detail)
	}
	for _, line := range strings.Split(detail, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if color {
			fmt.Fprintf(w, "         %s\n", st.Dim.Render(line))
		} else {
			fmt.Fprintf(w, "         %s\n", line)
		}
	}
	if fix := installer.SuggestedInstallFix(name); fix != "" {
		if color {
			fmt.Fprintf(w, "         %s\n", st.Accent.Render("Fix: "+fix))
		} else {
			fmt.Fprintf(w, "         Fix: %s\n", fix)
		}
	}
}

func retryFindomainInstall(ctx context.Context, name string, firstErr error, st ui.Styles, color, printStatus bool) error {
	if runtime.GOOS == "darwin" {
		if printStatus {
			printInstallRetryNote(st, color, name, "primary install failed, trying Homebrew…")
		}
		brewErr := installer.Install(ctx, name, "brew", []string{"install", "findomain"})
		if brewErr == nil {
			if printStatus {
				printInstallRetryNote(st, color, name, "installed via Homebrew")
			}
			return nil
		}
		if printStatus {
			printInstallRetryNote(st, color, name, "Homebrew failed, downloading GitHub release…")
		}
	} else {
		if printStatus {
			printInstallRetryNote(st, color, name, "primary install failed, downloading GitHub release…")
		}
	}
	if err := installer.InstallFindomainFromGitHub(ctx); err != nil {
		return firstErr
	}
	if printStatus {
		printInstallRetryNote(st, color, name, "installed from GitHub release")
	}
	return nil
}

func printInstallRetryNote(st ui.Styles, color bool, name, msg string) {
	line := fmt.Sprintf("         … %s: %s", name, msg)
	if color {
		fmt.Fprintln(os.Stdout, st.Dim.Render(line))
	} else {
		fmt.Fprintln(os.Stdout, line)
	}
	flushStdout()
}

func firstInstallErrorLine(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "unknown error"
	}
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		first := strings.TrimSpace(msg[:i])
		if strings.Contains(msg, "\n") {
			return first + " (more detail with --verbose)"
		}
		return first
	}
	return msg
}

func flushStdout() {
	_ = os.Stdout.Sync()
}
