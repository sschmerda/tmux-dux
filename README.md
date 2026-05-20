# tmux-commander

`tmux-commander` is a fast Go command palette for tmux. It is designed to run as a single executable inside a tmux popup, filter commands with a fuzzy search, exit cleanly, and then dispatch the selected tmux, shell, or popup action.

## Install

Build from source:

```sh
go install github.com/stefanschmerda/tmux-commander/cmd/tmux-commander@latest
```

Local development build:

```sh
go build -o bin/tmux-commander ./cmd/tmux-commander
```

The only runtime dependency is `tmux` itself. Optional commands such as `lazygit` or `btop` are only needed if you use those entries.

## tmux Binding

Recommended binding. This lets `tmux-commander` read `[ui].width` and `[ui].height` from TOML:

```tmux
bind -n C-Space run-shell "tmux-commander popup"
```

Prefix binding variant:

```tmux
bind p run-shell "tmux-commander popup"
```

Reload tmux config:

```sh
tmux source-file ~/.tmux.conf
```

## Configuration

Configuration is TOML and is loaded from:

1. `$XDG_CONFIG_HOME/tmux-commander/config.toml`
2. `~/.config/tmux-commander/config.toml`

If no config file exists, built-in defaults are used.

```toml
[ui]
width = "40%"
height = "70%"
popup_width = "80%"
popup_height = "80%"
border = true
theme = "shades-of-purple"
glyphs = true

[[commands]]
title = "Lazygit"
description = "Open lazygit in a popup"
category = "Tools"
aliases = ["lg"]
icon = "󰊢"
action = "popup"
command = "lazygit"
popup_width = "95%"
popup_height = "90%"

[[commands]]
title = "Btop"
description = "Open btop in a popup"
category = "Tools"
aliases = ["bt"]
icon = "󰘚"
action = "popup"
command = "btop"

[[commands]]
title = "Split Horizontal"
description = "Split pane side by side"
category = "Panes"
aliases = ["sh"]
action = "tmux"
command = "split-window -h -c '#{pane_current_path}'"
```

`width` and `height` control the commander popup itself when launched with `tmux-commander popup`. `popup_width` and `popup_height` under `[ui]` control popups opened by command actions. Individual popup commands can override those defaults with their own `popup_width` and `popup_height`.

`icon` is rendered as a command glyph to the left of the command title. Omit `icon`, set it to an empty string, or set `glyphs = false` in `[ui]` to hide glyphs.

The commander popup is launched without a native tmux border and draws its own themed border. The `border` setting is retained for popup actions and future launcher options.

`[ui]` fields:

| Field | Default | Description |
| --- | --- | --- |
| `width` | `"40%"` | Commander popup width when launched with `tmux-commander popup`. |
| `height` | `"70%"` | Commander popup height when launched with `tmux-commander popup`. |
| `popup_width` | `"80%"` | Default width for spawned `popup` command actions. |
| `popup_height` | `"80%"` | Default height for spawned `popup` command actions. |
| `border` | `true` | Retained for popup action behavior and future launcher options. The commander currently draws its own border. |
| `theme` | `"shades-of-purple"` | Built-in theme name, or `"custom"` to use `[custom_theme]`. |
| `glyphs` | `true` | Enables command glyphs from `[[commands]].icon`. |

`[[commands]]` fields:

| Field | Required | Description |
| --- | --- | --- |
| `title` | Yes | Command name shown in the palette and used for fuzzy matching. |
| `description` | No | Secondary text shown next to the command title. |
| `category` | No | Group header used when the query is empty, and a weak fuzzy-match field while filtering. |
| `aliases` | No | Short searchable abbreviations rendered as chips. |
| `icon` | No | Glyph shown to the left of the command title when `[ui].glyphs` is true. |
| `action` | Yes | Dispatch type. Must be `tmux`, `shell`, or `popup`. |
| `command` | Yes | Command string used by the selected `action`. |
| `popup_width` | No | Per-command width override for `popup` actions. |
| `popup_height` | No | Per-command height override for `popup` actions. |

Select a built-in theme by setting `[ui].theme` to one of these exact names:

- `catppuccin`
- `tokyonight`
- `rosepine`
- `kanagawa`
- `shades-of-purple`
- `solarized`
- `gruvbox`

Example:

```toml
[ui]
theme = "gruvbox"
```

Set `[ui].theme` to `custom` only when you want to provide your own colors with a top-level `[custom_theme]` block:

```toml
[ui]
theme = "custom"

[custom_theme]
background = "#101018"
title = "#ffffff"
commander_border = "#ffffff"
prompt_border = "#ffffff"
header = "#ffcc66"
muted = "#7c7f93"
prompt = "#ffcc66"
query = "#ffffff"
search_bg = "#7c3aed"
search_fg = "#ffffff"
description = "#b4befe"
empty = "#7c7f93"
chip = "#94e2d5"
chip_bg = "#313244"
selected_chip = "#f5c2e7"
selected_chip_bg = "#313244"
glyph = "#f5c2e7"
match_fg = "#ffcc66"
selected_match_fg = "#f5c2e7"
selected_fg = "#ffffff"
selected_bg = "#7c3aed"
```

Any omitted custom theme field falls back to the default `shades-of-purple` value.

Custom theme fields:

| Field | Controls |
| --- | --- |
| `background` | Main commander background and spawned popup background. |
| `title` | Command titles and title-like text. |
| `commander_border` | Outer commander frame and spawned popup border. |
| `prompt_border` | Search prompt box border. |
| `header` | Category headers and normal fuzzy match highlights. |
| `muted` | Secondary muted UI text. |
| `prompt` | Prompt-colored text outside the search box. |
| `query` | General query-colored text and spawned popup foreground. |
| `search_bg` | Search prompt box fill. |
| `search_fg` | Typed search query text. |
| `description` | Command descriptions. |
| `empty` | Empty-state text. |
| `chip` | Alias chip text in normal rows. |
| `chip_bg` | Alias chip background in normal rows. |
| `selected_chip` | Alias chip text in the selected row. |
| `selected_chip_bg` | Alias chip background in the selected row. |
| `glyph` | Command glyphs and search prompt glyph. |
| `match_fg` | Fuzzy match highlight in normal rows and alias chips. |
| `selected_match_fg` | Fuzzy match highlight in the selected row. |
| `selected_fg` | Selected command text and selected glyphs. |
| `selected_bg` | Selected command row background. |

List valid theme names from the binary:

```sh
tmux-commander themes
```

The built-in `Preview Themes` palette command opens an in-app preview in the current popup. Use `Up` / `Down` or `Left` / `Right` to cycle themes, then `Enter` or `Esc` to return to the command list. The selected preview theme stays active until this `tmux-commander` popup exits, including for popup actions launched from that session. The preview does not write config; it shows the `theme = "..."` value to set permanently.

## Actions

Each command must define an `action` and a `command`.

`action = "tmux"` runs `tmux <command>` after the palette exits:

```toml
action = "tmux"
command = "split-window -h -c '#{pane_current_path}'"
```

`action = "shell"` runs the command through the user's shell after the palette exits:

```toml
action = "shell"
command = "open https://github.com"
```

`action = "popup"` opens another tmux popup after the palette exits:

```toml
action = "popup"
command = "lazygit"
popup_width = "95%"
popup_height = "90%"
```

Popup actions inherit the active theme's tmux popup style. `background` / `query` are used for the popup body, and `commander_border` is used for the popup border foreground. Full-screen terminal apps may still draw their own colors.

Popup actions start in the active tmux pane directory via `#{pane_current_path}`.

Interactive tmux prompts are intentionally dispatched after the Bubble Tea UI exits to avoid nested input conflicts.

## Controls

- Type to filter commands.
- Use `Up` / `Down` or `Ctrl-P` / `Ctrl-N` to move.
- Use `Tab` / `Shift-Tab` to jump between command categories.
- Press `Enter` to select.
- Press `Esc` or `Ctrl-C` to cancel.

In the theme preview view, use `Up` / `Down` or `Left` / `Right` to preview themes. Press `Enter` or `Esc` to return to the command list.

When the query is empty, commands are grouped by category. While filtering, category headers are hidden and results are sorted by fuzzy score. Multi-token searches such as `split pane` are supported.

## Default Commands

- Find Pane
- Split Horizontal
- Split Vertical
- Close Pane
- Zoom / Unzoom
- New Window
- Rename Window
- Close Window
- Choose Session
- New Session
- Rename Session
- Detach
- Reload Config
- Preview Themes
- Lazygit
- Btop

## Compared With eduwass/tmux-palette

This project is inspired by [`eduwass/tmux-palette`](https://github.com/eduwass/tmux-palette), which provides a Raycast-style tmux command palette with Bun, JSON configuration, popup tools, custom palettes, themes, and shell-powered palette sources.

The deliberate differences in `tmux-commander` are:

- Go single-binary distribution instead of a Bun project checkout.
- TOML configuration instead of JSON.
- A smaller first surface area focused on tmux commands, shell commands, and popup commands.
- Package boundaries for config, fuzzy matching, action dispatch, palette UI, and tmux helpers.

`tmux-palette` currently has broader customization features such as custom palettes, theme selection, JSON-powered plugin sources, and TPM-oriented installation. `tmux-commander` starts with less surface area and prioritizes low-friction native binary installation.

## Development

```sh
go test ./...
go run ./cmd/tmux-commander
go run ./cmd/tmux-commander themes
go build -o bin/tmux-commander ./cmd/tmux-commander
```

Run inside tmux for a realistic manual test:

```sh
./bin/tmux-commander popup
```

## Release Builds

Build a local binary:

```sh
go build -trimpath -ldflags="-s -w" -o bin/tmux-commander ./cmd/tmux-commander
```

Example cross-compile commands:

```sh
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/tmux-commander_darwin_arm64 ./cmd/tmux-commander
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/tmux-commander_linux_amd64 ./cmd/tmux-commander
```
