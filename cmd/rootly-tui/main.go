package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rootlyhq/rootly-tui/internal/app"
	"github.com/rootlyhq/rootly-tui/internal/debug"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Define flags
	showVersion := flag.Bool("version", false, "Show version information")
	showVersionShort := flag.Bool("v", false, "Show version information (shorthand)")
	debugMode := flag.Bool("debug", false, "Enable debug logging")
	logFile := flag.String("log", "", "Write debug logs to file (implies --debug)")

	flag.Parse()

	// Check for version flag
	if *showVersion || *showVersionShort {
		fmt.Printf("rootly-tui %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Always log startup to buffer
	debug.Logger.Info("Starting rootly-tui",
		"version", version,
		"commit", commit,
	)

	// Enable debug mode (outputs to stderr/file in addition to buffer)
	if *debugMode || *logFile != "" {
		debug.Enable()
		if *logFile != "" {
			if err := debug.SetLogFile(*logFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
				os.Exit(1)
			}
			debug.Logger.Info("Logging to file", "path", *logFile)
		}
		debug.Logger.Info("Debug mode enabled")
	}

	p := tea.NewProgram(
		app.New(version),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		debug.Logger.Error("Program error", "error", err)
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
