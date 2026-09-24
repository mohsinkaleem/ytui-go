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
