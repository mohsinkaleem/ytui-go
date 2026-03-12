package models

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
}

// colWidths defines fixed column widths for format table: quality, ext, vcodec, acodec, fps, resolution, size
// They are computed dynamically from the current item set.
type colWidths [7]int

// formatItemDelegate renders format items with aligned columns
type formatItemDelegate struct {
	widths colWidths
}

func (d formatItemDelegate) Height() int                             { return 1 }
func (d formatItemDelegate) Spacing() int                            { return 0 }
func (d formatItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// formatCols extracts the 7 display columns for a FormatItem.
func formatCols(item types.FormatItem) [7]string {
	// quality
	quality := item.FormatNote
	if item.Height > 0 {
		quality = fmt.Sprintf("%dp", item.Height)
	} else if item.ABR > 0 {
		quality = fmt.Sprintf("%.0fkbps", item.ABR)
	}

	// ext
	ext := item.Ext

	// vcodec
	vcodec := "—"
	if item.VCodec != "" && item.VCodec != "none" {
		vcodec = item.VCodec
	}

	// acodec
	acodec := "—"
	if item.ACodec != "" && item.ACodec != "none" {
		acodec = item.ACodec
	}

	// fps
	fps := ""
	if item.FPS > 0 {
		fps = fmt.Sprintf("%.0ffps", item.FPS)
	}

	// resolution
	res := ""
	if item.Width > 0 && item.Height > 0 {
		res = fmt.Sprintf("%d×%d", item.Width, item.Height)
	}

	// size
	size := item.HumanSize()

	return [7]string{quality, ext, vcodec, acodec, fps, res, size}
}

// computeColWidths calculates the max width of each column across all items.
func computeColWidths(items []list.Item) colWidths {
	// minimum widths so headers / empty lists still look reasonable
	mins := colWidths{5, 4, 6, 6, 3, 7, 4}
	var w colWidths
	for i, m := range mins {
		w[i] = m
	}
	for _, li := range items {
		f, ok := li.(types.FormatItem)
		if !ok {
			continue
		}
		cols := formatCols(f)
		for i, c := range cols {
			if len(c) > w[i] {
				w[i] = len(c)
			}
		}
	}
	return w
}

func (d formatItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(types.FormatItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	cols := formatCols(item)
	widths := d.widths

	// Build aligned line using fixed-width columns separated by two spaces
	line := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %-*s  %-*s  %s",
		widths[0], cols[0],
		widths[1], cols[1],
		widths[2], cols[2],
		widths[3], cols[3],
		widths[4], cols[4],
		widths[5], cols[5],
		cols[6],
	)

	if isSelected {
		content := styles.ListSelectedTitleStyle.Render(line)
		fmt.Fprint(w, styles.ListSelectedItemStyle.Render(content))
	} else {
		fmt.Fprint(w, styles.ListItemStyle.Render(styles.ListTitleStyle.Render(line)))
	}
}

// NewFormatListModel creates a new format list model
func NewFormatListModel() FormatListModel {
	delegate := formatItemDelegate{widths: computeColWidths(nil)}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	ti := textinput.New()
	ti.Placeholder = "e.g., 140+137"
	ti.CharLimit = 100
	ti.Width = 30
	ti.Prompt = "❯ "
	ti.PromptStyle = styles.InputPromptStyle

	return FormatListModel{
		List:        l,
		ActiveTab:   types.FormatTabVideo,
		CustomInput: ti,
	}
}

// SetFormats sets the available formats and video info
func (m *FormatListModel) SetFormats(formats []types.FormatItem, video types.VideoItem) {
	m.Formats = formats
	m.Video = video
	m.ActiveTab = types.FormatTabVideo
	m.Combos = suggestFormatCombos(formats)
	m.ComboIdx = 0
	m.updateListItems()
}

// SetSize updates dimensions
func (m *FormatListModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.List.SetSize(w, h-10) // reserve space for header + tabs
}

// NextTab cycles to next tab
func (m *FormatListModel) NextTab() {
	m.ActiveTab = (m.ActiveTab + 1) % 3
	if m.ActiveTab == types.FormatTabCustom {
		m.CustomInput.Focus()
	} else {
		m.CustomInput.Blur()
	}
	m.updateListItems()
}

// PrevTab cycles to previous tab
func (m *FormatListModel) PrevTab() {
	if m.ActiveTab == 0 {
		m.ActiveTab = 2
	} else {
		m.ActiveTab--
	}
	if m.ActiveTab == types.FormatTabCustom {
		m.CustomInput.Focus()
	} else {
		m.CustomInput.Blur()
	}
	m.updateListItems()
}

func (m *FormatListModel) updateListItems() {
	var items []list.Item
	headerExtra := 0
	for _, f := range m.Formats {
		switch m.ActiveTab {
		case types.FormatTabVideo:
			headerExtra = 2 // auto-merge note + blank line
			if f.HasVideoAndAudio() || f.IsVideoOnly() {
				items = append(items, f)
			}
		case types.FormatTabAudio:
			if f.IsAudioOnly() {
				items = append(items, f)
			}
		default:
			items = append(items, f)
		}
	}
	// Compute aligned column widths from the visible item set
	m.List.SetDelegate(formatItemDelegate{widths: computeColWidths(items)})
	m.List.SetItems(items)
	m.List.SetSize(m.Width, m.Height-10-headerExtra)
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

// IsVideoOnlyFormat checks if a formatID corresponds to a video-only format
func (m *FormatListModel) IsVideoOnlyFormat(formatID string) bool {
	for _, f := range m.Formats {
		if f.FormatID == formatID {
			return f.IsVideoOnly()
		}
	}
	return false
}

// GetFormatID returns the format ID to use for download
func (m *FormatListModel) GetFormatID() string {
	if m.ActiveTab == types.FormatTabCustom {
		customVal := strings.TrimSpace(m.CustomInput.Value())
		if customVal != "" {
			return customVal
		}
		// Use selected combo
		if m.ComboIdx >= 0 && m.ComboIdx < len(m.Combos) {
			return m.Combos[m.ComboIdx].FormatID
		}
		return "bestvideo+bestaudio/best"
	}
	if f, ok := m.SelectedFormat(); ok {
		return f.FormatID
	}
	return "best"
}

// View renders the format list screen
func (m *FormatListModel) View() string {
	var b strings.Builder

	// Video info header
	b.WriteString(styles.VideoTitleStyle.Render(m.Video.Title))
	b.WriteString("\n")

	details := []string{}
	if m.Video.DurationString != "" {
		details = append(details, "⏱ "+m.Video.DurationString)
	}
	if m.Video.ViewCount > 0 {
		details = append(details, "👁 "+formatViewCount(m.Video.ViewCount))
	}
	if m.Video.Channel != "" {
		details = append(details, "📺 "+m.Video.Channel)
	}
	if len(details) > 0 {
		b.WriteString(styles.VideoDetailStyle.Render(strings.Join(details, " • ")))
	}
	b.WriteString("\n\n")

	// Tab bar
	tabs := []string{"Video", "Audio", "Custom"}
	for i, tab := range tabs {
		if types.FormatTab(i) == m.ActiveTab {
			b.WriteString(styles.TabActiveStyle.Render(tab))
		} else {
			b.WriteString(styles.TabInactiveStyle.Render(tab))
		}
		if i < len(tabs)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("\n\n")

	// Content
	if m.ActiveTab == types.FormatTabCustom {
		// Recommended combinations
		if len(m.Combos) > 0 {
			b.WriteString(styles.BoldStyle.Render("Recommended:"))
			b.WriteString("\n\n")
			for i, c := range m.Combos {
				cursor := "  "
				var line string
				label := fmt.Sprintf("%-28s", c.Label)
				fmtID := styles.MutedStyle.Render(c.FormatID)
				sizeStr := ""
				if c.Size > 0 {
					sizeStr = "  " + styles.SpeedStyle.Render("~"+comboHumanSize(c.Size))
				}
				if i == m.ComboIdx {
					cursor = styles.AccentStyle.Render("❯ ")
					line = cursor + styles.ListSelectedTitleStyle.Render(label) + "  " + fmtID + sizeStr
				} else {
					line = cursor + styles.ListTitleStyle.Render(label) + "  " + fmtID + sizeStr
				}
				b.WriteString(line)
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}

		b.WriteString(styles.BoldStyle.Render("Custom format:"))
		b.WriteString("\n")
		b.WriteString(m.CustomInput.View())
		b.WriteString("\n\n")
		b.WriteString(styles.MutedStyle.Render("↑/↓ to select combo  •  type a custom format  •  enter to download"))
	} else {
		// Video tab note about auto-merge
		if m.ActiveTab == types.FormatTabVideo {
			b.WriteString(styles.MutedStyle.Render("Formats with no(-) audio are auto-merged with best audio"))
			b.WriteString("\n\n")
		}
		b.WriteString(m.List.View())
	}

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

// comboHumanSize converts a byte count to a human-readable string for combo labels.
func comboHumanSize(n int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case n >= GB:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(GB))
	case n >= MB:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(MB))
	case n >= KB:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(KB))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
