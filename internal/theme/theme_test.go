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
		if got.Name == "" || got.Background == "" || got.Title == "" || got.Prompt == "" || got.Query == "" || got.Empty == "" || got.ChipBG == "" || got.Glyph == "" || got.SelectedBG == "" {
			t.Fatalf("theme %q resolved incompletely: %#v", name, got)
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
		Background: "#111111",
		Title:      "#eeeeee",
		Prompt:     "#aaaaaa",
		Query:      "#bbbbbb",
		Empty:      "#cccccc",
		ChipBG:     "#222222",
		Glyph:      "#dddddd",
		SelectedBG: "#333333",
	})
	if got.Name != "custom" {
		t.Fatalf("theme = %q, want custom", got.Name)
	}
	if got.Background != "#111111" || got.Title != "#eeeeee" || got.Prompt != "#aaaaaa" || got.Query != "#bbbbbb" || got.Empty != "#cccccc" || got.ChipBG != "#222222" || got.Glyph != "#dddddd" || got.SelectedBG != "#333333" {
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
