package styles

import (
	"testing"
)

func TestBuiltinThemes(t *testing.T) {
	themes := map[string]Theme{
		"mocha":      CatppuccinMocha,
		"latte":      CatppuccinLatte,
		"monochrome": Monochrome,
	}
	for name, theme := range themes {
		t.Run(name, func(t *testing.T) {
			if theme.Name == "" {
				t.Error("theme Name should not be empty")
			}
			if theme.Base == "" {
				t.Error("theme Base should not be empty")
			}
			if theme.Text == "" {
				t.Error("theme Text should not be empty")
			}
			if theme.Accent == "" {
				t.Error("theme Accent should not be empty")
			}
		})
	}
}

func TestLoadTheme(t *testing.T) {
	tests := []struct {
		name      string
		themeName string
		wantName  string
		wantOK    bool
	}{
		{"mocha", "mocha", "mocha", true},
		{"latte", "latte", "latte", true},
		{"monochrome", "monochrome", "monochrome", true},
		{"nonexistent", "nonexistent", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok := LoadTheme(tt.themeName)
			if ok != tt.wantOK {
				t.Errorf("LoadTheme(%q) = %v, want %v", tt.themeName, ok, tt.wantOK)
			}
			if tt.wantOK && CurrentTheme.Name != tt.wantName {
				t.Errorf("CurrentTheme.Name = %q, want %q", CurrentTheme.Name, tt.wantName)
			}
		})
	}
}

func TestCurrentThemeInitialized(t *testing.T) {
	if CurrentTheme.Name == "" {
		t.Error("CurrentTheme should be initialized")
	}
}

func TestRebuildStylesDoesNotPanic(t *testing.T) {
	themes := []string{"mocha", "latte", "monochrome"}
	for _, name := range themes {
		t.Run(name, func(t *testing.T) {
			LoadTheme(name)
		})
	}
}

func TestThemeNames(t *testing.T) {
	names := ThemeNames()
	if len(names) == 0 {
		t.Error("ThemeNames() should return at least one theme")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		in    string
		width int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello w…"},
		{"日本語のタイトル", 7, "日本語…"},
		{"x", 0, "x"},
	}
	for _, tt := range tests {
		if got := Truncate(tt.in, tt.width); got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
		}
	}
}

func TestTruncateLeft(t *testing.T) {
	if got := TruncateLeft("/a/b/c/file.mp4", 10); got != "…/file.mp4" {
		t.Errorf("TruncateLeft = %q, want %q", got, "…/file.mp4")
	}
	if got := TruncateLeft("short", 10); got != "short" {
		t.Errorf("TruncateLeft = %q, want unchanged", got)
	}
}
