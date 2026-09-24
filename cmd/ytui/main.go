package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mohsinkaleem/ytui-go/internal/app"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/utils"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:           "ytui",
		Short:         "A beautiful TUI wrapper for yt-dlp",
		Version:       version,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          run,
	}

	rootCmd.Flags().String("theme", "", "color theme for this session ("+strings.Join(styles.ThemeNames(), ", ")+")")
	rootCmd.Flags().String("download-dir", "", "download directory for this session")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, _ []string) error {
	theme, _ := cmd.Flags().GetString("theme")
	downloadDir, _ := cmd.Flags().GetString("download-dir")
	if downloadDir != "" {
		downloadDir = utils.ExpandPath(downloadDir)
		if info, err := os.Stat(downloadDir); err != nil || !info.IsDir() {
			return fmt.Errorf("--download-dir %q is not a directory", downloadDir)
		}
	}

	client, err := ytdlp.NewClient()
	if err != nil {
		return fmt.Errorf("%w\nInstall yt-dlp: https://github.com/yt-dlp/yt-dlp#installation", err)
	}

	// Open the store; without it the app still works, just without persistence.
	settings := store.DefaultSettings()
	var notice string
	st, err := store.New()
	if err != nil {
		notice = "Could not open the database: settings and history won't be saved"
	} else {
		defer st.Close()
		if st.IsReadOnly() {
			notice = "Another ytui is running: settings and history won't be saved"
		}
		if s, err := st.GetSettings(); err == nil {
			settings = s
		}
	}

	styles.LoadTheme(settings.Theme)
	if theme != "" && !styles.LoadTheme(theme) {
		return fmt.Errorf("unknown theme %q (available: %s)", theme, strings.Join(styles.ThemeNames(), ", "))
	}
	client.SetCookiesFrom(settings.CookiesFrom)
	client.SetCookiesFile(settings.CookiesFile)

	downloads := download.NewManager(settings.MaxConcurrent, client, st)
	model := app.New(app.Config{
		Store:       st,
		Client:      client,
		Downloads:   downloads,
		Settings:    settings,
		DownloadDir: downloadDir,
		Notice:      notice,
	})

	_, runErr := tea.NewProgram(model, tea.WithAltScreen()).Run()

	if n := downloads.Shutdown(); n > 0 && st != nil && !st.IsReadOnly() {
		fmt.Fprintf(os.Stderr, "%d unfinished download(s) saved — start ytui and use /resume to continue.\n", n)
	}
	return runErr
}
