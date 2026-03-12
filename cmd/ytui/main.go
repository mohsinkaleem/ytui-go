package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/spf13/cobra"

	"github.com/mohsinkaleem/ytui-go/internal/app"
	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "ytui",
		Short:   "A beautiful TUI wrapper for yt-dlp",
		Version: version,
		RunE:    run,
	}

	rootCmd.Flags().String("theme", "", "color theme (mocha, latte, monochrome)")
	rootCmd.Flags().String("download-dir", "", "default download directory")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Initialize bubble zone for mouse support
	zone.NewGlobal()

	// Apply theme if specified
	if theme, _ := cmd.Flags().GetString("theme"); theme != "" {
		if !styles.LoadTheme(theme) {
			fmt.Fprintf(os.Stderr, "Unknown theme: %s\n", theme)
			os.Exit(1)
		}
	}

	// Open store
	st, err := store.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not open database: %v\n", err)
		// Continue without persistence
	} else if st.IsReadOnly() {
		fmt.Fprintln(os.Stderr, "Note: another instance is running; this session is read-only (settings & history won't be saved).")
	}

	// Load settings from store
	if st != nil {
		settings, err := st.GetSettings()
		if err == nil && settings.Theme != "" {
			styles.LoadTheme(settings.Theme)
		}
	}

	// Create yt-dlp client
	client, err := ytdlp.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nPlease install yt-dlp: https://github.com/yt-dlp/yt-dlp\n", err)
		os.Exit(1)
	}

	// Apply cookies-from-browser setting
	if st != nil {
		if settings, err := st.GetSettings(); err == nil {
			if settings.CookiesFrom != "" {
				client.SetCookiesFrom(settings.CookiesFrom)
			} else if settings.CookiesFile != "" {
				client.SetCookiesFile(settings.CookiesFile)
			}
		}
	}

	// Create root model
	model := app.New(st, client)

	// Create program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Set program reference on model for download manager
	model.SetProgram(p)

	// Run
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Graceful shutdown
	if m, ok := finalModel.(app.Model); ok {
		m.DownloadMgr.Shutdown()
		if m.Store != nil {
			m.Store.Close()
		}
	}

	return nil
}
