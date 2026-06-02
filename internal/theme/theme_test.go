package theme

import "testing"

func TestResolveDefaultsToShadesOfPurple(t *testing.T) {
	got := Resolve("")
	if got.Name != "shades-of-purple" {
		t.Fatalf("theme = %q, want shades-of-purple", got.Name)
	}
}

func TestResolveKnownThemes(t *testing.T) {
	for _, name := range Names() {
		got := Resolve(name)
		if got.Name == "" || got.Background == "" || got.Title == "" || got.PaletteBorder == "" || got.PromptBorder == "" || got.Prompt == "" || got.Query == "" || got.SearchBG == "" || got.SearchFG == "" || got.Empty == "" || got.ChipBG == "" || got.SelectedChip == "" || got.SelectedChipBG == "" || got.Glyph == "" || got.MatchFG == "" || got.SelectedMatchFG == "" || got.SelectedBG == "" {
			t.Fatalf("theme %q resolved incompletely: %#v", name, got)
		}
	}
}

func TestBuiltInThemesKeepChipAndSelectionBackgroundsDistinct(t *testing.T) {
	for _, name := range Names() {
		got := Resolve(name)
		if got.ChipBG == got.SelectedBG {
			t.Fatalf("theme %q uses the same chip and selected background: %s", name, got.ChipBG)
		}
		if got.ChipBG == got.Background {
			t.Fatalf("theme %q uses the same chip and main background: %s", name, got.ChipBG)
		}
	}
}

func TestResolveUnknownThemeUsesDefault(t *testing.T) {
	got := Resolve("tokyo-night")
	if got.Name != "shades-of-purple" {
		t.Fatalf("theme = %q, want shades-of-purple", got.Name)
	}
}

func TestResolveWithCustomUsesCustomFields(t *testing.T) {
	got := ResolveWithCustom("custom", Theme{
		Background:      "#111111",
		Title:           "#eeeeee",
		PaletteBorder:   "#ddddff",
		PromptBorder:    "#ccccff",
		Prompt:          "#aaaaaa",
		Query:           "#bbbbbb",
		SearchBG:        "#444444",
		SearchFG:        "#eeeeff",
		Empty:           "#cccccc",
		ChipBG:          "#222222",
		SelectedChip:    "#ffccaa",
		SelectedChipBG:  "#332211",
		Glyph:           "#dddddd",
		MatchFG:         "#ffeeaa",
		SelectedMatchFG: "#aaffee",
		SelectedBG:      "#333333",
	})
	if got.Name != "custom" {
		t.Fatalf("theme = %q, want custom", got.Name)
	}
	if got.Background != "#111111" || got.Title != "#eeeeee" || got.PaletteBorder != "#ddddff" || got.PromptBorder != "#ccccff" || got.Prompt != "#aaaaaa" || got.Query != "#bbbbbb" || got.SearchBG != "#444444" || got.SearchFG != "#eeeeff" || got.Empty != "#cccccc" || got.ChipBG != "#222222" || got.SelectedChip != "#ffccaa" || got.SelectedChipBG != "#332211" || got.Glyph != "#dddddd" || got.MatchFG != "#ffeeaa" || got.SelectedMatchFG != "#aaffee" || got.SelectedBG != "#333333" {
		t.Fatalf("custom fields were not applied: %#v", got)
	}
	if got.Header != Default().Header {
		t.Fatalf("missing field fallback = %q, want %q", got.Header, Default().Header)
	}
}

func TestResolveWithCustomIgnoresCustomForBuiltins(t *testing.T) {
	got := ResolveWithCustom("gruvbox", Theme{Background: "#111111"})
	if got.Name != "gruvbox" || got.Background == "#111111" {
		t.Fatalf("builtin theme unexpectedly used custom values: %#v", got)
	}
}

func TestConfigNamesIncludesCustom(t *testing.T) {
	names := ConfigNames()
	if names[len(names)-1] != "custom" {
		t.Fatalf("last config theme = %q, want custom", names[len(names)-1])
	}
}
