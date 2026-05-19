package theme

import "strings"

type Theme struct {
	Name        string `toml:"name"`
	Background  string `toml:"background"`
	Title       string `toml:"title"`
	Header      string `toml:"header"`
	Muted       string `toml:"muted"`
	Prompt      string `toml:"prompt"`
	Query       string `toml:"query"`
	Description string `toml:"description"`
	Empty       string `toml:"empty"`
	Chip        string `toml:"chip"`
	ChipBG      string `toml:"chip_bg"`
	Glyph       string `toml:"glyph"`
	SelectedFG  string `toml:"selected_fg"`
	SelectedBG  string `toml:"selected_bg"`
}

func Default() Theme {
	return Resolve("shades-of-purple")
}

func Resolve(name string) Theme {
	switch normalize(name) {
	case "catppuccin":
		return Theme{
			Name:        "catppuccin",
			Background:  "#1e1e2e",
			Title:       "#89b4fa",
			Header:      "#cba6f7",
			Muted:       "#6c7086",
			Prompt:      "#f5c2e7",
			Query:       "#cdd6f4",
			Description: "#a6adc8",
			Empty:       "#6c7086",
			Chip:        "#94e2d5",
			ChipBG:      "#313244",
			Glyph:       "#f5c2e7",
			SelectedFG:  "#cdd6f4",
			SelectedBG:  "#313244",
		}
	case "tokyonight":
		return Theme{
			Name:        "tokyonight",
			Background:  "#1a1b26",
			Title:       "#7aa2f7",
			Header:      "#bb9af7",
			Muted:       "#565f89",
			Prompt:      "#7dcfff",
			Query:       "#c0caf5",
			Description: "#9aa5ce",
			Empty:       "#565f89",
			Chip:        "#7dcfff",
			ChipBG:      "#283457",
			Glyph:       "#7dcfff",
			SelectedFG:  "#c0caf5",
			SelectedBG:  "#283457",
		}
	case "rosepine":
		return Theme{
			Name:        "rosepine",
			Background:  "#191724",
			Title:       "#ebbcba",
			Header:      "#c4a7e7",
			Muted:       "#6e6a86",
			Prompt:      "#f6c177",
			Query:       "#e0def4",
			Description: "#908caa",
			Empty:       "#6e6a86",
			Chip:        "#9ccfd8",
			ChipBG:      "#403d52",
			Glyph:       "#f6c177",
			SelectedFG:  "#e0def4",
			SelectedBG:  "#403d52",
		}
	case "kanagawa":
		return Theme{
			Name:        "kanagawa",
			Background:  "#1f1f28",
			Title:       "#7e9cd8",
			Header:      "#957fb8",
			Muted:       "#727169",
			Prompt:      "#ffa066",
			Query:       "#dcd7ba",
			Description: "#c8c093",
			Empty:       "#727169",
			Chip:        "#7aa89f",
			ChipBG:      "#2d4f67",
			Glyph:       "#ffa066",
			SelectedFG:  "#dcd7ba",
			SelectedBG:  "#2d4f67",
		}
	case "", "shades-of-purple":
		return Theme{
			Name:        "shades-of-purple",
			Background:  "#2d2b55",
			Title:       "#fad000",
			Header:      "#ff9d00",
			Muted:       "#a599e9",
			Prompt:      "#ff9d00",
			Query:       "#ffffff",
			Description: "#b362ff",
			Empty:       "#a599e9",
			Chip:        "#9effff",
			ChipBG:      "#6943ff",
			Glyph:       "#ff9d00",
			SelectedFG:  "#ffffff",
			SelectedBG:  "#6943ff",
		}
	case "solarized":
		return Theme{
			Name:        "solarized",
			Background:  "#002b36",
			Title:       "#268bd2",
			Header:      "#b58900",
			Muted:       "#586e75",
			Prompt:      "#cb4b16",
			Query:       "#fdf6e3",
			Description: "#839496",
			Empty:       "#586e75",
			Chip:        "#2aa198",
			ChipBG:      "#073642",
			Glyph:       "#cb4b16",
			SelectedFG:  "#fdf6e3",
			SelectedBG:  "#073642",
		}
	case "gruvbox":
		return Theme{
			Name:        "gruvbox",
			Background:  "#282828",
			Title:       "#fabd2f",
			Header:      "#d3869b",
			Muted:       "#928374",
			Prompt:      "#fe8019",
			Query:       "#fbf1c7",
			Description: "#bdae93",
			Empty:       "#928374",
			Chip:        "#8ec07c",
			ChipBG:      "#504945",
			Glyph:       "#fe8019",
			SelectedFG:  "#fbf1c7",
			SelectedBG:  "#504945",
		}
	default:
		return Default()
	}
}

func ResolveWithCustom(name string, custom Theme) Theme {
	if normalize(name) != "custom" {
		return Resolve(name)
	}
	base := Default()
	base.Name = "custom"
	if custom.Background != "" {
		base.Background = custom.Background
	}
	if custom.Title != "" {
		base.Title = custom.Title
	}
	if custom.Header != "" {
		base.Header = custom.Header
	}
	if custom.Muted != "" {
		base.Muted = custom.Muted
	}
	if custom.Prompt != "" {
		base.Prompt = custom.Prompt
	}
	if custom.Query != "" {
		base.Query = custom.Query
	}
	if custom.Description != "" {
		base.Description = custom.Description
	}
	if custom.Empty != "" {
		base.Empty = custom.Empty
	}
	if custom.Chip != "" {
		base.Chip = custom.Chip
	}
	if custom.ChipBG != "" {
		base.ChipBG = custom.ChipBG
	}
	if custom.Glyph != "" {
		base.Glyph = custom.Glyph
	}
	if custom.SelectedFG != "" {
		base.SelectedFG = custom.SelectedFG
	}
	if custom.SelectedBG != "" {
		base.SelectedBG = custom.SelectedBG
	}
	return base
}

func Names() []string {
	return []string{
		"catppuccin",
		"tokyonight",
		"rosepine",
		"kanagawa",
		"shades-of-purple",
		"solarized",
		"gruvbox",
	}
}

func ConfigNames() []string {
	names := append([]string{}, Names()...)
	return append(names, "custom")
}

func normalize(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
