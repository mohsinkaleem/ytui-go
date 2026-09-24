package app

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/models"
	"github.com/mohsinkaleem/ytui-go/internal/slash"
	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/utils"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// Config holds the root model's dependencies and startup settings.
type Config struct {
	Store       *store.Store // nil disables persistence
	Client      *ytdlp.Client
	Downloads   *download.Manager
	Settings    store.Settings
	DownloadDir string // session override for Settings.DownloadDir
	Notice      string // shown as a toast on startup
}

// Model is the root bubbletea model
type Model struct {
	// State machine
	State      types.State
	PrevState  types.State // screen to return to from Loading, Download and VideoPlaying
	FormatBack types.State // screen to return to from FormatList
	Width      int
	Height     int

	// Sub-models (flat composition)
	Search     models.SearchModel
	VideoList  models.VideoListModel
	FormatList models.FormatListModel
	Download   models.DownloadModel
	Player     models.PlayerModel
	ResumeList models.ResumeListModel

	// Loading
	Spinner    spinner.Model
	LoadingMsg string

	// Status bar toast
	Toast    string
	ToastErr bool
	toastSeq int

	ticking bool // a DownloadTickMsg is pending

	Settings    store.Settings
	dirOverride string

	Fetcher       *utils.Fetcher
	DownloadMgr   *download.Manager
	PlayerManager *utils.PlayerManager
	Store         *store.Store
	SlashRegistry *slash.Registry
	YtdlpClient   *ytdlp.Client
}

// New creates a new root model
func New(cfg Config) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	m := Model{
		State:         types.StateSearchInput,
		PrevState:     types.StateSearchInput,
		FormatBack:    types.StateSearchInput,
		Search:        models.NewSearchModel(),
		VideoList:     models.NewVideoListModel(),
		FormatList:    models.NewFormatListModel(),
		Download:      models.NewDownloadModel(),
		Player:        models.NewPlayerModel(),
		ResumeList:    models.NewResumeListModel(),
		Spinner:       s,
		Toast:         cfg.Notice,
		Settings:      cfg.Settings,
		dirOverride:   cfg.DownloadDir,
		Fetcher:       utils.NewFetcher(cfg.Client),
		DownloadMgr:   cfg.Downloads,
		PlayerManager: utils.NewPlayerManager(cfg.Settings.MpvPath),
		Store:         cfg.Store,
		SlashRegistry: slash.NewRegistry(),
		YtdlpClient:   cfg.Client,
	}
	m.syncSearch()

	if cfg.Store != nil {
		if entries, err := cfg.Store.GetSearchHistory(50); err == nil {
			// Entries are newest first; push oldest first so the newest ends on top.
			for i := len(entries) - 1; i >= 0; i-- {
				q := entries[i].Title
				if q == "" {
					q = entries[i].URL
				}
				if q != "" {
					m.Search.PushHistory(q)
				}
			}
		}
	}

	return m
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	if m.Toast != "" {
		return clearToastAfter(m.toastSeq)
	}
	return nil
}

// downloadDir returns the folder downloads are saved to.
func (m *Model) downloadDir() string {
	switch {
	case m.dirOverride != "":
		return m.dirOverride
	case m.Settings.DownloadDir != "":
		return m.Settings.DownloadDir
	}
	return store.DefaultSettings().DownloadDir
}

func (m *Model) downloadOpts() ytdlp.DownloadOpts {
	return ytdlp.DownloadOpts{
		EmbedSubs:     m.Settings.EmbedSubs,
		EmbedMetadata: m.Settings.EmbedMetadata,
		EmbedChapters: m.Settings.EmbedChapters,
		OutputDir:     m.downloadDir(),
	}
}

// saveSettings persists the settings and refreshes the search screen.
func (m *Model) saveSettings() {
	m.syncSearch()
	if m.Store != nil {
		_ = m.Store.SaveSettings(m.Settings)
	}
}

// syncSearch copies the settings shown on the search screen.
func (m *Model) syncSearch() {
	m.Search.EmbedSubs = m.Settings.EmbedSubs
	m.Search.EmbedMetadata = m.Settings.EmbedMetadata
	m.Search.EmbedChapters = m.Settings.EmbedChapters
	m.Search.DownloadDir = m.downloadDir()
	m.Search.Cookies = cookiesLabel(m.Settings)
}

func cookiesLabel(s store.Settings) string {
	if s.CookiesFrom != "" {
		return s.CookiesFrom
	}
	return s.CookiesFile
}

// applyTheme pushes the current theme into components that cache styles.
func (m *Model) applyTheme() {
	m.Spinner.Style = styles.SpinnerStyle
	m.Search.ApplyTheme()
	m.VideoList.ApplyTheme()
	m.FormatList.ApplyTheme()
	m.Download.ApplyTheme()
}
