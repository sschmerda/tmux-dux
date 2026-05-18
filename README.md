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

Launch the palette in a popup:

```tmux
bind -n C-Space display-popup -E -w 75% -h 70% tmux-commander
```

Prefix binding variant:

```tmux
bind p display-popup -E -w 75% -h 70% tmux-commander
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
width = "75%"
height = "70%"
popup_width = "80%"
popup_height = "80%"
border = true

[[commands]]
title = "Lazygit"
description = "Open lazygit in a popup"
category = "Tools"
aliases = ["lg"]
icon = "git"
popup = "lazygit"

[[commands]]
title = "Btop"
description = "Open btop in a popup"
category = "Tools"
aliases = ["bt"]
icon = "cpu"
popup = "btop"

[[commands]]
title = "Split Horizontal"
description = "Split pane side by side"
category = "Panes"
aliases = ["sh"]
tmux = "split-window -h -c '#{pane_current_path}'"
```

## Actions

Each command must define exactly one action field.

`tmux` runs a tmux command after the palette exits:

```toml
tmux = "split-window -h -c '#{pane_current_path}'"
```

`shell` runs a shell command after the palette exits:

```toml
shell = "open https://github.com"
```

`popup` opens another tmux popup after the palette exits:

```toml
popup = "lazygit"
```

Interactive tmux prompts are intentionally dispatched after the Bubble Tea UI exits to avoid nested input conflicts.

## Controls

- Type to filter commands.
- Use `Up` / `Down` or `Ctrl-P` / `Ctrl-N` to move.
- Press `Enter` to select.
- Press `Esc` or `Ctrl-C` to cancel.

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
go build -o bin/tmux-commander ./cmd/tmux-commander
```

Run inside tmux for a realistic manual test:

```sh
tmux display-popup -E -w 75% -h 70% ./bin/tmux-commander
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
