package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"macventory/components/collector"
	"macventory/components/fileutil"
	"macventory/components/model"
	"macventory/components/report"
)

// Run processes the supplied CLI arguments and returns an exit code.
func Run(args []string, version string) int {
	if len(args) == 0 {
		printHelp(os.Stdout)
		return 0
	}

	switch args[0] {
	case "scan", "start":
		return runScan(args[0], args[1:], version)
	case "report":
		return runReport(args[1:])
	case "check":
		return runCheck(args[1:])
	case "version", "--version", "-v":
		fmt.Printf("Macventory %s\n", version)
		return 0
	case "help", "--help", "-h":
		printHelp(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		printHelp(os.Stderr)
		return 2
	}
}

// runScan scans the Mac and writes an inventory report.
func runScan(name string, args []string, version string) int {
	var cfg model.Config

	desktopDirectory, err := userDesktopDirectory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "locate Desktop directory: %v\n", err)
		return 1
	}

	defaultOutput := filepath.Join(
		desktopDirectory,
		"macventory-"+time.Now().Format("2006-01-02-150405")+".md",
	)

	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	flags.StringVar(
		&cfg.Output,
		"output",
		defaultOutput,
		"path of the Markdown report",
	)
	flags.DurationVar(
		&cfg.CommandTimeout,
		"timeout",
		20*time.Second,
		"maximum runtime for each external command",
	)
	flags.BoolVar(
		&cfg.IncludeSystemApps,
		"include-system-apps",
		false,
		"include applications under /System",
	)

	flags.Usage = func() {
		fmt.Fprintf(
			flags.Output(),
			"Usage: macventory %s [options]\n\n",
			name,
		)
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if flags.NArg() != 0 {
		fmt.Fprintf(
			os.Stderr,
			"unexpected argument: %s\n",
			flags.Arg(0),
		)
		return 2
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	sections := collector.CollectAll(
		ctx,
		cfg,
		collector.All(),
	)

	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "scan cancelled")
		return 130
	}

	contents := report.Render(sections, version)

	if err := fileutil.AtomicWrite(
		cfg.Output,
		[]byte(contents),
		0o600,
	); err != nil {
		fmt.Fprintf(os.Stderr, "write report: %v\n", err)
		return 1
	}

	absolutePath, err := filepath.Abs(cfg.Output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve report path: %v\n", err)
		return 1
	}

	fmt.Printf("Inventory written to %s\n", absolutePath)
	return 0
}

// runReport displays information about the latest inventory report.
func runReport(args []string) int {
	desktopDirectory, err := userDesktopDirectory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "locate Desktop directory: %v\n", err)
		return 1
	}

	flags := flag.NewFlagSet("report", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	var openReport bool
	var directory string

	flags.BoolVar(
		&openReport,
		"open",
		false,
		"open the latest report in the default application",
	)
	flags.StringVar(
		&directory,
		"directory",
		desktopDirectory,
		"directory containing inventory reports",
	)

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if flags.NArg() != 0 {
		fmt.Fprintf(
			os.Stderr,
			"unexpected argument: %s\n",
			flags.Arg(0),
		)
		return 2
	}

	latest, info, err := report.Latest(directory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "report: %v\n", err)
		return 1
	}

	absolutePath, err := filepath.Abs(latest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve report path: %v\n", err)
		return 1
	}

	fmt.Printf(
		"Latest report: %s\nModified: %s\nSize: %d bytes\n",
		absolutePath,
		info.ModTime().Format(time.RFC3339),
		info.Size(),
	)

	if openReport {
		if runtime.GOOS != "darwin" {
			fmt.Fprintln(
				os.Stderr,
				"opening reports automatically is only supported on macOS",
			)
			return 2
		}

		if err := exec.Command("open", absolutePath).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "open report: %v\n", err)
			return 1
		}
	}

	return 0
}

// runCheck displays the availability of optional collector tools.
func runCheck(args []string) int {
	flags := flag.NewFlagSet("check", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if flags.NArg() != 0 {
		fmt.Fprintf(
			os.Stderr,
			"unexpected argument: %s\n",
			flags.Arg(0),
		)
		return 2
	}

	tools := []string{
		"system_profiler",
		"brew",
		"mas",
		"npm",
		"pnpm",
		"yarn",
		"pipx",
		"uv",
		"gem",
		"cargo",
		"go",
		"code",
		"cursor",
		"docker",
		"xcodebuild",
		"xcrun",
		"terraform",
		"gcloud",
		"aws",
	}

	fmt.Println("Collector readiness")
	fmt.Println("-------------------")

	for _, tool := range tools {
		path, err := exec.LookPath(tool)
		if err != nil {
			fmt.Printf("%-18s unavailable\n", tool)
			continue
		}

		fmt.Printf("%-18s available  %s\n", tool, path)
	}

	return 0
}

// userDesktopDirectory returns the current user's Desktop directory.
func userDesktopDirectory() (string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDirectory, "Desktop"), nil
}

// printHelp prints the main command help.
func printHelp(output io.Writer) {
	fmt.Fprint(output, `Macventory creates a portable software inventory report for a Mac.
Usage:
    macventory <command> [options]

Commands:
    scan       Scan the Mac and create a Markdown report
    start      Alias for scan
    report     Show information about the latest report
    check      Show which optional collectors are available
    version    Print the Macventory version
    help       Show help

Examples:
    macventory scan
    macventory start
    macventory report
    macventory report --open
    macventory scan --output ~/Desktop/my-mac-inventory.md
`)
}
