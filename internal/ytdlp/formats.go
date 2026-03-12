package ytdlp

import (
	"sort"

	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// SortFormats sorts formats by height desc, then ABR desc, then filesize desc
func SortFormats(formats []Format) []Format {
	sorted := make([]Format, len(formats))
	copy(sorted, formats)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Height != sorted[j].Height {
			return sorted[i].Height > sorted[j].Height
		}
		if sorted[i].ABR != sorted[j].ABR {
			return sorted[i].ABR > sorted[j].ABR
		}
		return sorted[i].EffectiveSize() > sorted[j].EffectiveSize()
	})
	return sorted
}

// FilterByType filters formats by video+audio, video-only, or audio-only
func FilterByType(formats []Format, tab types.FormatTab) []Format {
	var filtered []Format
	for _, f := range formats {
		switch tab {
		case types.FormatTabVideo:
			if f.HasBoth() || f.IsVideoOnly() {
				filtered = append(filtered, f)
			}
		case types.FormatTabAudio:
			if f.IsAudioOnly() {
				filtered = append(filtered, f)
			}
		default:
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// SuggestBestFormats returns common quality presets
func SuggestBestFormats(formats []Format) []types.FormatCombo {
	var combos []types.FormatCombo

	// Best overall
	combos = append(combos, types.FormatCombo{Label: "Best", FormatID: "bestvideo+bestaudio/best"})

	// Find specific resolutions
	resMap := make(map[int]Format)
	for _, f := range formats {
		if f.Height > 0 {
			if existing, ok := resMap[f.Height]; !ok || f.EffectiveSize() > existing.EffectiveSize() {
				resMap[f.Height] = f
			}
		}
	}

	for _, h := range []int{2160, 1440, 1080, 720, 480, 360} {
		if f, ok := resMap[h]; ok {
			combos = append(combos, types.FormatCombo{
				Label:    f.Resolution,
				FormatID: f.FormatID,
			})
		}
	}

	// Best audio only
	var bestAudio Format
	for _, f := range formats {
		if f.IsAudioOnly() && f.ABR > bestAudio.ABR {
			bestAudio = f
		}
	}
	if bestAudio.FormatID != "" {
		combos = append(combos, types.FormatCombo{
			Label:    "Audio only (best)",
			FormatID: bestAudio.FormatID,
		})
	}

	return combos
}

// ConvertFormatsToItems converts ytdlp formats to types.FormatItem
func ConvertFormatsToItems(formats []Format) []types.FormatItem {
	items := make([]types.FormatItem, len(formats))
	for i, f := range formats {
		items[i] = types.FormatItem{
			FormatID:   f.FormatID,
			Ext:        f.Ext,
			Resolution: f.Resolution,
			Width:      f.Width,
			Height:     f.Height,
			FPS:        f.FPS,
			VCodec:     f.VCodec,
			ACodec:     f.ACodec,
			Filesize:   f.EffectiveSize(),
			TBR:        f.TBR,
			VBR:        f.VBR,
			ABR:        f.ABR,
			FormatNote: f.FormatNote,
			Protocol:   f.Protocol,
		}
	}
	return items
}
