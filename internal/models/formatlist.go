package models

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mohsinkaleem/ytui-go/internal/styles"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// FormatListModel represents the format selection screen
type FormatListModel struct {
	List        list.Model
	Video       types.VideoItem
	Formats     []types.FormatItem
	ActiveTab   types.FormatTab
	CustomInput textinput.Model
	Combos      []types.FormatCombo
	ComboIdx    int
	Width       int
	Height      int

	widths colWidths
	counts [2]int // number of video and audio formats, for the tab labels
}

// colStrings holds one table row: id, quality, ext, vcodec, acodec, fps, resolution, size.
type colStrings [8]string

// colWidths holds the display width of each table column.
type colWidths [8]int

var formatHeaders = colStrings{"ID", "QUALITY", "EXT", "VIDEO", "AUDIO", "FPS", "RESOLUTION", "SIZE"}

// formatItemDelegate renders format items with aligned columns
type formatItemDelegate struct {
	widths colWidths
}

func (d formatItemDelegate) Height() int                             { return 1 }
func (d formatItemDelegate) Spacing() int                            { return 0 }
func (d formatItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// formatCols extracts the display columns for a FormatItem.
func formatCols(item types.FormatItem) colStrings {
	fps := ""
	if item.FPS > 0 {
		fps = fmt.Sprintf("%.0f", item.FPS)
	}
	res := ""
	if item.Width > 0 && item.Height > 0 {
		res = fmt.Sprintf("%d×%d", item.Width, item.Height)
	}
	return colStrings{item.FormatID, item.QualityLabel(), item.Ext, codecName(item.VCodec), codecName(item.ACodec), fps, res, item.HumanSize()}
}

// codecName shortens codec strings like "avc1.640028" to their family name.
func codecName(codec string) string {
	if codec == "" || codec == "none" {
		return "—"
	}
	name, _, _ := strings.Cut(codec, ".")
	return name
}

// computeColWidths calculates the width of each column across the header and items.
func computeColWidths(items []list.Item) colWidths {
	var w colWidths
	for i, h := range formatHeaders {
		w[i] = lipgloss.Width(h)
	}
	for _, li := range items {
		if f, ok := li.(types.FormatItem); ok {
			for i, c := range formatCols(f) {
				w[i] = max(w[i], lipgloss.Width(c))
			}
		}
	}
	return w
}

// row joins cols into an aligned table line.
func (w colWidths) row(cols colStrings) string {
	var b strings.Builder
	for i, c := range cols {
		if i > 0 {
			b.WriteString("  ")
		}
		if i < len(cols)-1 {
			c = padRight(c, w[i])
		}
		b.WriteString(c)
	}
	return b.String()
}

func (d formatItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(types.FormatItem)
	if !ok {
		return
	}

	line := styles.Truncate(d.widths.row(formatCols(item)), m.Width()-3)
	if index == m.Index() {
		fmt.Fprint(w, styles.ListSelectedItemStyle.Render(styles.ListSelectedTitleStyle.Render(line)))
	} else {
		fmt.Fprint(w, styles.ListItemStyle.Render(styles.TextStyle.Render(line)))
	}
}

// NewFormatListModel creates a new format list model
func NewFormatListModel() FormatListModel {
	l := newList(formatItemDelegate{widths: computeColWidths(nil)})
	l.SetFilteringEnabled(false)
	l.SetStatusBarItemName("format", "formats")

	ti := textinput.New()
	ti.Placeholder = "e.g. 137+140 or bestvideo[height<=720]+bestaudio"
	ti.CharLimit = 200
	ti.Width = 50
	ti.Prompt = "❯ "

	m := FormatListModel{
		List:        l,
		ActiveTab:   types.FormatTabVideo,
		CustomInput: ti,
		widths:      computeColWidths(nil),
	}
	m.ApplyTheme()
	return m
}

// ApplyTheme re-applies the current theme to the custom format input.
func (m *FormatListModel) ApplyTheme() {
	styleInput(&m.CustomInput)
}

// SetFormats sets the available formats and video info
func (m *FormatListModel) SetFormats(formats []types.FormatItem, video types.VideoItem) {
	m.Formats = formats
	m.Video = video
	m.Combos = suggestFormatCombos(formats)
	m.ComboIdx = 0
	m.CustomInput.SetValue("")

	m.counts = [2]int{}
	for _, f := range formats {
		switch {
		case isVideoFormat(f):
			m.counts[0]++
		case f.IsAudioOnly():
			m.counts[1]++
		}
	}

	tab := types.FormatTabVideo
	if m.counts[0] == 0 && m.counts[1] > 0 {
		tab = types.FormatTabAudio // e.g. music sites
	}
	m.setTab(tab)
}

// isVideoFormat reports whether f belongs on the Video tab (unknown codecs included).
func isVideoFormat(f types.FormatItem) bool {
	return f.HasVideoAndAudio() || f.IsVideoOnly() || (f.VCodec == "" && f.ACodec == "")
}

// SetSize updates dimensions
func (m *FormatListModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.CustomInput.Width = max(min(w, 72)-lipgloss.Width(m.CustomInput.Prompt)-1, 10)
	m.resizeList()
}

// resizeList fits the list below the header lines drawn by View.
func (m *FormatListModel) resizeList() {
	header := 6 // title, details, blank, tabs, blank, table header
	if m.ActiveTab == types.FormatTabVideo {
		header++ // auto-merge note
	}
	m.List.SetSize(m.Width, max(m.Height-header, 1))
}

// NextTab cycles to next tab
func (m *FormatListModel) NextTab() {
	m.setTab((m.ActiveTab + 1) % 3)
}

// PrevTab cycles to previous tab
func (m *FormatListModel) PrevTab() {
	m.setTab((m.ActiveTab + 2) % 3)
}

// MoveCombo moves the preset cursor on the Custom tab.
func (m *FormatListModel) MoveCombo(delta int) {
	m.ComboIdx = max(min(m.ComboIdx+delta, len(m.Combos)-1), 0)
}

func (m *FormatListModel) setTab(tab types.FormatTab) {
	m.ActiveTab = tab
	if tab == types.FormatTabCustom {
		m.CustomInput.Focus()
	} else {
		m.CustomInput.Blur()
	}

	var items []list.Item
	for _, f := range m.Formats {
		if (tab == types.FormatTabVideo && isVideoFormat(f)) || (tab == types.FormatTabAudio && f.IsAudioOnly()) {
			items = append(items, f)
		}
	}
	m.widths = computeColWidths(items)
	m.List.SetDelegate(formatItemDelegate{widths: m.widths})
	m.List.SetItems(items)
	m.List.ResetSelected() // a stale cursor past the end would select nothing
	m.resizeList()
}

// SelectedFormat returns the selected format
func (m *FormatListModel) SelectedFormat() (types.FormatItem, bool) {
	item := m.List.SelectedItem()
	if item == nil {
		return types.FormatItem{}, false
	}
	f, ok := item.(types.FormatItem)
	return f, ok
}

// GetFormatID returns the yt-dlp format selector for the current selection,
// or "" when nothing is selected. Video-only formats are paired with the best audio.
func (m *FormatListModel) GetFormatID() string {
	if m.ActiveTab == types.FormatTabCustom {
		if custom := strings.TrimSpace(m.CustomInput.Value()); custom != "" {
			return custom
		}
		if m.ComboIdx >= 0 && m.ComboIdx < len(m.Combos) {
			return m.Combos[m.ComboIdx].FormatID
		}
		return ""
	}
	f, ok := m.SelectedFormat()
	if !ok {
		return ""
	}
	if f.IsVideoOnly() {
		return f.FormatID + "+bestaudio"
	}
	return f.FormatID
}

// View renders the format list screen
func (m *FormatListModel) View() string {
	var b strings.Builder

	b.WriteString(styles.VideoTitleStyle.Render(styles.Truncate(m.Video.Title, m.Width)))
	b.WriteString("\n")
	b.WriteString(styles.VideoDetailStyle.Render(styles.Truncate(videoDetails(m.Video), m.Width)))
	b.WriteString("\n\n")

	tabs := []string{fmt.Sprintf("Video (%d)", m.counts[0]), fmt.Sprintf("Audio (%d)", m.counts[1]), "Custom"}
	for i, tab := range tabs {
		if types.FormatTab(i) == m.ActiveTab {
			b.WriteString(styles.TabActiveStyle.Render(tab))
		} else {
			b.WriteString(styles.TabInactiveStyle.Render(tab))
		}
	}
	b.WriteString("\n\n")

	if m.ActiveTab == types.FormatTabCustom {
		b.WriteString(m.customView())
		return b.String()
	}

	if m.ActiveTab == types.FormatTabVideo {
		b.WriteString(styles.MutedStyle.Render(styles.Truncate("Video-only formats (AUDIO —) are merged with the best audio automatically", m.Width)))
		b.WriteString("\n")
	}
	b.WriteString(styles.TableHeaderStyle.Render(styles.Truncate("   "+m.widths.row(formatHeaders), m.Width)))
	b.WriteString("\n")
	b.WriteString(m.List.View())

	return b.String()
}

func (m *FormatListModel) customView() string {
	var b strings.Builder

	if len(m.Combos) > 0 {
		b.WriteString(styles.BoldStyle.Render("Presets"))
		b.WriteString("\n")

		labelW, idW := 0, 0
		for _, c := range m.Combos {
			labelW = max(labelW, lipgloss.Width(c.Label))
			idW = max(idW, lipgloss.Width(c.FormatID))
		}
		// A typed format wins over the presets, so only show the cursor when it's empty.
		usePreset := strings.TrimSpace(m.CustomInput.Value()) == ""
		for i, c := range m.Combos {
			cursor, labelStyle := "  ", styles.TextStyle
			if usePreset && i == m.ComboIdx {
				cursor, labelStyle = styles.AccentStyle.Render("❯ "), styles.ListSelectedTitleStyle
			}
			line := cursor + labelStyle.Render(padRight(c.Label, labelW)) + "  " + styles.MutedStyle.Render(padRight(c.FormatID, idW))
			if c.Size > 0 {
				line += "  " + styles.SpeedStyle.Render("~"+formatBytes(c.Size))
			}
			b.WriteString(styles.Truncate(line, m.Width))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.BoldStyle.Render("Custom format"))
	b.WriteString("\n")
	b.WriteString(m.CustomInput.View())
	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render(styles.Truncate("Any yt-dlp -f selector; format IDs are listed on the Video and Audio tabs", m.Width)))
	return b.String()
}

// suggestFormatCombos generates recommended video+audio combinations from available formats
func suggestFormatCombos(formats []types.FormatItem) []types.FormatCombo {
	var combos []types.FormatCombo

	// Helper: find format by ID and return its size
	sizeOf := func(fid string) int64 {
		for _, f := range formats {
			if f.FormatID == fid {
				return f.Filesize
			}
		}
		return 0
	}

	// Best overall
	combos = append(combos, types.FormatCombo{
		Label:    "Best video + best audio",
		FormatID: "bestvideo+bestaudio/best",
	})

	// Find best audio format
	var bestAudioID string
	var bestABR float64
	var bestAudioSize int64
	for _, f := range formats {
		if f.IsAudioOnly() && f.ABR > bestABR {
			bestABR = f.ABR
			bestAudioID = f.FormatID
			bestAudioSize = f.Filesize
		}
	}

	// Find best video-only for each resolution, pair with best audio
	type vidEntry struct {
		item types.FormatItem
	}
	resMap := make(map[int]vidEntry)
	for _, f := range formats {
		if f.IsVideoOnly() && f.Height > 0 {
			if existing, ok := resMap[f.Height]; !ok ||
				f.Filesize > existing.item.Filesize ||
				(f.Filesize == existing.item.Filesize && f.TBR > existing.item.TBR) {
				resMap[f.Height] = vidEntry{item: f}
			}
		}
	}

	for _, h := range []int{2160, 1440, 1080, 720, 480, 360} {
		if v, ok := resMap[h]; ok && bestAudioID != "" {
			label := fmt.Sprintf("%dp video + audio", h)
			formatID := fmt.Sprintf("%s+%s", v.item.FormatID, bestAudioID)
			var size int64
			if v.item.Filesize > 0 && bestAudioSize > 0 {
				size = v.item.Filesize + bestAudioSize
			}
			combos = append(combos, types.FormatCombo{Label: label, FormatID: formatID, Size: size})
		}
	}

	// Audio-only options
	if bestAudioID != "" {
		label := fmt.Sprintf("Audio only (%.0fkbps)", bestABR)
		combos = append(combos, types.FormatCombo{Label: label, FormatID: bestAudioID, Size: sizeOf(bestAudioID)})
	}

	// Smallest combined
	combos = append(combos, types.FormatCombo{
		Label:    "Smallest video + audio",
		FormatID: "worstvideo+worstaudio/worst",
	})

	// Estimate "best" combo size by summing best video + best audio sizes
	var bestVidSize int64
	for _, h := range []int{2160, 1440, 1080, 720, 480, 360} {
		if v, ok := resMap[h]; ok && v.item.Filesize > 0 {
			bestVidSize = v.item.Filesize
			break
		}
	}
	if bestVidSize > 0 && bestAudioSize > 0 {
		combos[0].Size = bestVidSize + bestAudioSize
	}

	return combos
}
