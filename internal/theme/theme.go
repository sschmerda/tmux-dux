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
			Title:       "#cdd6f4",
			Header:      "#f9e2af",
			Muted:       "#6c7086",
			Prompt:      "#f9e2af",
			Query:       "#cdd6f4",
			Description: "#a6adc8",
			Empty:       "#6c7086",
			Chip:        "#cdd6f4",
			ChipBG:      "#313244",
			Glyph:       "#f9e2af",
			SelectedFG:  "#f5f7ff",
			SelectedBG:  "#45475a",
		}
	case "tokyonight":
		return Theme{
			Name:        "tokyonight",
			Background:  "#1a1b26",
			Title:       "#c0caf5",
			Header:      "#e0af68",
			Muted:       "#565f89",
			Prompt:      "#e0af68",
			Query:       "#c0caf5",
			Description: "#9aa5ce",
			Empty:       "#565f89",
			Chip:        "#c0caf5",
			ChipBG:      "#24283b",
			Glyph:       "#e0af68",
			SelectedFG:  "#ffffff",
			SelectedBG:  "#292e42",
		}
	case "rosepine":
		return Theme{
			Name:        "rosepine",
			Background:  "#191724",
			Title:       "#e0def4",
			Header:      "#f6c177",
			Muted:       "#6e6a86",
			Prompt:      "#f6c177",
			Query:       "#e0def4",
			Description: "#908caa",
			Empty:       "#6e6a86",
			Chip:        "#e0def4",
			ChipBG:      "#26233a",
			Glyph:       "#f6c177",
			SelectedFG:  "#ffffff",
			SelectedBG:  "#403d52",
		}
	case "kanagawa":
		return Theme{
			Name:        "kanagawa",
			Background:  "#1f1f28",
			Title:       "#dcd7ba",
			Header:      "#e6c384",
			Muted:       "#727169",
			Prompt:      "#e6c384",
			Query:       "#dcd7ba",
			Description: "#c8c093",
			Empty:       "#727169",
			Chip:        "#dcd7ba",
			ChipBG:      "#2a2a37",
			Glyph:       "#e6c384",
			SelectedFG:  "#ffffff",
			SelectedBG:  "#363646",
		}
	case "", "shades-of-purple":
		return Theme{
			Name:        "shades-of-purple",
			Background:  "#2d2b55",
			Title:       "#d7d3ff",
			Header:      "#fad000",
			Muted:       "#a599e9",
			Prompt:      "#fad000",
			Query:       "#ffffff",
			Description: "#b8b0e8",
			Empty:       "#a599e9",
			Chip:        "#d7d3ff",
			ChipBG:      "#1e1e3f",
			Glyph:       "#fad000",
			SelectedFG:  "#ffffff",
			SelectedBG:  "#403b75",
		}
	case "solarized":
		return Theme{
			Name:        "solarized",
			Background:  "#002b36",
			Title:       "#eee8d5",
			Header:      "#b58900",
			Muted:       "#586e75",
			Prompt:      "#b58900",
			Query:       "#fdf6e3",
			Description: "#839496",
			Empty:       "#586e75",
			Chip:        "#eee8d5",
			ChipBG:      "#073642",
			Glyph:       "#b58900",
			SelectedFG:  "#fdf6e3",
			SelectedBG:  "#164955",
		}
	case "gruvbox":
		return Theme{
			Name:        "gruvbox",
			Background:  "#282828",
			Title:       "#ebdbb2",
			Header:      "#fabd2f",
			Muted:       "#928374",
			Prompt:      "#fabd2f",
			Query:       "#fbf1c7",
			Description: "#bdae93",
			Empty:       "#928374",
			Chip:        "#ebdbb2",
			ChipBG:      "#3c3836",
			Glyph:       "#fabd2f",
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
