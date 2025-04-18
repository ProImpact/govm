package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/melkeydev/govm/internal/cli"
	"github.com/melkeydev/govm/internal/model"
	"github.com/melkeydev/govm/internal/setup"
	"github.com/melkeydev/govm/internal/utils"
)

func main() {
	if len(os.Args) == 0 {
		printUsage()
		return
	}
	// Check if user is requesting version information
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("govm %s\n", utils.GetVersion())
		os.Exit(0)
	}
	if err := utils.SetupShimDirectory(); err != nil {
		fmt.Printf("Warning: Failed to set up shim directory: %v\n", err)
	}
	if len(os.Args) > 1 {
		handleCommandLine()
		return
	}
	// handleCommandLine and TUI should never throw at the same time
	launchTUI()
}

func handleCommandLine() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}
	command := os.Args[1]
	switch command {
	case "install":
		if len(os.Args) < 3 {
			fmt.Println("Error: 'install' requires a version argument")
			fmt.Println("Usage: govm install <version>")
			fmt.Println("Example: govm install 1.21")
			return
		}
		version := os.Args[2]
		version = strings.TrimPrefix(version, "go")
		cli.InstallVersion(version)
	case "use":
		if len(os.Args) < 3 {
			fmt.Println("Error: 'use' requires a version argument")
			fmt.Println("Usage: govm use <version>")
			fmt.Println("Example: govm use 1.21")
			return
		}
		version := os.Args[2]
		version = strings.TrimPrefix(version, "go")
		cli.UseVersion(version)
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Error: 'delete' requires a version argument")
			fmt.Println("Usage: govm delete <version>")
			fmt.Println("Example: govm delete 1.21")
			return
		}
		version := os.Args[2]
		version = strings.TrimPrefix(version, "go")
		cli.DeleteVersion(version)
	case "list":
		cli.ListVersions()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("GoVM - Go Version Manager")
	fmt.Println("\nUsage:")
	fmt.Println("  govm                   Launch the interactive TUI")
	fmt.Println("  govm install <version> Install a specific Go version")
	fmt.Println("  govm use <version>     Switch to a specific Go version")
	fmt.Println("  govm delete <version>  Delete a specific Go version")
	fmt.Println("  govm list              List installed Go versions")
	fmt.Println("  govm help              Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  govm install 1.21      Install Go 1.21.x (latest)")
	fmt.Println("  govm use 1.20          Switch to Go 1.20.x (latest)")
}

func launchTUI() {
	if !setup.IsShimInPath() {
		setupModel := setup.New()
		p := tea.NewProgram(setupModel, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error in setup: %v\n", err)
			os.Exit(1)
		}
	}
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#3c71a8"))
	columns := []table.Column{
		{Title: "Version", Width: 10},
		{Title: "Path", Width: 40},
		{Title: "Status", Width: 10},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	t.SetStyles(table.Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3c71a8")).
			Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3c71a8")).
			Bold(true).
			Padding(0, 1),
		Cell: lipgloss.NewStyle().
			Padding(0, 1),
	})
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}
	goVersionsDir := filepath.Join(homeDir, ".govm", "versions")
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#3c71a8")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#DDDDDD")).
		Background(lipgloss.Color("#3c71a8"))
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Go Versions"
	l.SetShowHelp(false)
	initialModel := model.Model{
		List:           l,
		Versions:       []utils.GoVersion{},
		Spinner:        s,
		Loading:        true,
		HomeDir:        homeDir,
		GoVersionsDir:  goVersionsDir,
		InstalledTable: t,
	}
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
