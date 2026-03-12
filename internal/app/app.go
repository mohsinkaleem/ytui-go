package app

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/download"
	"github.com/mohsinkaleem/ytui-go/internal/models"
	"github.com/mohsinkaleem/ytui-go/internal/slash"
	"github.com/mohsinkaleem/ytui-go/internal/store"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
	"github.com/mohsinkaleem/ytui-go/internal/utils"
	"github.com/mohsinkaleem/ytui-go/internal/ytdlp"
)

// Model is the root bubbletea model
type Model struct {
	// State machine
	State     types.State
	PrevState types.State
	Width     int
	Height    int

	// Sub-models (flat composition)
	Search     models.SearchModel
	VideoList  models.VideoListModel
	FormatList models.FormatListModel
	Download   models.DownloadModel
	Player     models.PlayerModel
	ResumeList models.ResumeListModel

	// Shared state
	SelectedVideo  types.VideoItem
	SelectedVideos []types.VideoItem // for batch format download
	CurrentQuery   string

	// Loading
	Spinner     spinner.Model
	LoadingType string

	// UX
	ErrMsg   string
	ToastMsg string

	// Managers
	SearchManager  *utils.SearchManager
	FormatsManager *utils.FormatsManager
	PlaylistMgr    *utils.PlaylistManager
	DownloadMgr    *download.Manager
	PlayerManager  *utils.PlayerManager

	// Persistence
	Store *store.Store

	// Slash commands
	SlashRegistry *slash.Registry

	// yt-dlp client
	YtdlpClient *ytdlp.Client

	// Program reference (set after init)
	Program *tea.Program
}

// New creates a new root model
func New(st *store.Store, client *ytdlp.Client) Model {
	registry := slash.NewRegistry()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.CurrentTheme.Pink)

	searchMgr := utils.NewSearchManager(client)
	formatsMgr := utils.NewFormatsManager(client)
	playlistMgr := utils.NewPlaylistManager(client)
	playerMgr := utils.NewPlayerManager("")

	// Download manager
	var dlMgr *download.Manager
	if st != nil {
		settings, _ := st.GetSettings()
		dlMgr = download.NewManager(settings.MaxConcurrent, client, st)
	} else {
		dlMgr = download.NewManager(3, client, nil)
	}

	model := Model{
		State:          types.StateSearchInput,
		Search:         models.NewSearchModel(registry),
		VideoList:      models.NewVideoListModel(),
		FormatList:     models.NewFormatListModel(),
		Download:       models.NewDownloadModel(),
		Player:         models.NewPlayerModel(),
		ResumeList:     models.NewResumeListModel(),
		Spinner:        s,
		SearchManager:  searchMgr,
		FormatsManager: formatsMgr,
		PlaylistMgr:    playlistMgr,
		DownloadMgr:    dlMgr,
		PlayerManager:  playerMgr,
		Store:          st,
		SlashRegistry:  registry,
		YtdlpClient:    client,
	}

	// Load search history from store
	if st != nil {
		// Load embed settings from store
		if settings, err := st.GetSettings(); err == nil {
			model.Search.EmbedSubs = settings.EmbedSubs
			model.Search.EmbedMetadata = settings.EmbedMetadata
			model.Search.EmbedChapters = settings.EmbedChapters
		}

		entries, err := st.GetSearchHistory(50)
		if err == nil {
			seen := make(map[string]bool)
			for _, e := range entries {
				q := e.Title
				if q == "" {
					q = e.URL
				}
				if q != "" && !seen[q] {
					seen[q] = true
					model.Search.History = append(model.Search.History, q)
				}
			}
		}
	}

	return model
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		m.Search.Input.Focus(),
	)
}

// SetProgram sets the program reference on the model and managers
func (m *Model) SetProgram(p *tea.Program) {
	m.Program = p
	m.DownloadMgr.SetProgram(p)
	m.DownloadMgr.Start()
}
