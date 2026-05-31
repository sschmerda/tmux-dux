# tmux-commander

`tmux-commander` is a fast Go command palette for tmux. It is designed to run as a single executable inside a tmux popup, filter commands with a fuzzy search, exit cleanly, and then dispatch the selected tmux, shell, popup, or current-pane shell action.

## Install

With TPM:

```tmux
set -g @plugin 'sschmerda/tmux-commander'
set -g @tmux-commander-key 'Space'
```

Press `prefix` + `I` to install. The TPM plugin downloads the latest release binary into the plugin directory. `@tmux-commander-key` is optional; no key is bound unless you set it explicitly. The example above binds `prefix` + `Space` to `tmux-commander popup`.

For a global binding that does not require the tmux prefix, bind it manually after the TPM plugin declaration:

```tmux
bind-key -n C-Space run-shell '"${TMUX_COMMANDER_BIN:-tmux-commander}" popup'
```

With the release installer:

```sh
curl -fsSL https://raw.githubusercontent.com/sschmerda/tmux-commander/main/scripts/install.sh | sh
```

Install a specific release:

```sh
TMUX_COMMANDER_VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/sschmerda/tmux-commander/main/scripts/install.sh | sh
```

The installer supports macOS and Linux on `amd64` and `arm64`, matching the precompiled GitHub release archives. It installs to `~/.local/bin` by default. Override that with `TMUX_COMMANDER_INSTALL_DIR`.

Verify a downloaded release archive with GitHub artifact attestations:

```sh
gh attestation verify tmux-commander_linux_arm64.tar.gz \
  --repo sschmerda/tmux-commander
```

The installer verifies SHA256 checksums automatically. Artifact attestation verification additionally proves that a release archive was produced by this repository's GitHub Actions release workflow.

Build from source:

```sh
go install github.com/sschmerda/tmux-commander/cmd/tmux-commander@latest
```

Local development build:

```sh
go build -o bin/tmux-commander ./cmd/tmux-commander
```

The only runtime dependency is `tmux` itself. Optional commands such as `lazygit` or `btop` are only needed if you use those entries.

## tmux Binding

Recommended binding when installing outside TPM. This lets `tmux-commander` read `[ui].width` and `[ui].height` from TOML:

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

Command definitions may either live inline in `config.toml` or in a sibling file:

1. `$XDG_CONFIG_HOME/tmux-commander/commands.toml`
2. `~/.config/tmux-commander/commands.toml`

If `commands.toml` exists, its `[[commands]]` replace any inline `[[commands]]` from `config.toml`. This keeps app settings small while allowing long command catalogs. `commands.toml` may only define `[[commands]]`; keep `[ui]`, `[keys]`, and `[custom_theme]` in `config.toml`.

If no config file exists, built-in defaults are used. If no command definitions exist, built-in commands are used.

`config.toml`:

```toml
[ui]
width = "40%"
height = "80%"
popup_width = "80%"
popup_height = "80%"
border = true
theme = "shades-of-purple"
glyphs = true
show_description = true
show_toggle_hint = true
tmux_description = true
recent_commands = true
recent_limit = 10
tmux_recent_limit = 10

[keys]
tmux_mode = "ctrl+t"
move_up = "ctrl+p"
move_down = "ctrl+n"
scroll_up = "ctrl+y"
scroll_down = "ctrl+e"
half_page_up = "ctrl+u"
half_page_down = "ctrl+d"
next_category = "tab"
previous_category = "shift+tab"
```

`commands.toml`:

```toml
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

[[commands]]
title = "New Session"
description = "Prompt for a new detached session"
category = "Sessions"
action = "tmux"
command = "new-session -d -s {{input}}"
prompt = "session_name"
```

`width` and `height` control the commander popup itself when launched with `tmux-commander popup`. `popup_width` and `popup_height` under `[ui]` control popups opened by command actions. Individual popup commands can override those defaults with their own `popup_width` and `popup_height`.

`icon` is rendered as a command glyph to the left of the command title. Omit `icon`, set it to an empty string, or set `glyphs = false` in `[ui]` to hide glyphs.

The commander popup is launched without a native tmux border and draws its own themed border. The `border` setting is retained for popup actions and future launcher options.

`[ui]` fields:

| Field | Default | Description |
| --- | --- | --- |
| `width` | `"40%"` | Commander popup width when launched with `tmux-commander popup`. |
| `height` | `"80%"` | Commander popup height when launched with `tmux-commander popup`. |
| `popup_width` | `"80%"` | Default width for spawned `popup` command actions. |
| `popup_height` | `"80%"` | Default height for spawned `popup` command actions. |
| `border` | `true` | Retained for popup action behavior and future launcher options. The commander currently draws its own border. |
| `theme` | `"shades-of-purple"` | Built-in theme name, or `"custom"` to use `[custom_theme]`. |
| `glyphs` | `true` | Enables command glyphs from `[[commands]].icon`. |
| `show_description` | `true` | Shows command descriptions next to command titles. Set to `false` for a denser command list. |
| `show_toggle_hint` | `true` | Shows the mode line below the search field, including the configured tmux-command toggle key. |
| `tmux_description` | `true` | Shows short command descriptions in tmux-command mode. Set to `false` for a denser tmux command list. |
| `recent_commands` | `true` | Enables the recent-command section and recency boost while filtering. |
| `recent_limit` | `10` | Maximum number of recent commands to keep in state and show in the palette. Set to `0` to disable recents. |
| `tmux_recent_limit` | `10` | Maximum number of recent tmux-command mode entries to keep in state and show in tmux-command mode. Set to `0` to disable tmux-command recents. |

`[keys]` fields:

| Field | Default | Description |
| --- | --- | --- |
| `tmux_mode` | `"ctrl+t"` | Toggles between configured commands and tmux-command mode. The active key is shown in the palette hint and in `Show Controls`. |
| `move_up` | `"ctrl+p"` | Moves the selection up. The `Up` arrow always remains available. |
| `move_down` | `"ctrl+n"` | Moves the selection down. The `Down` arrow always remains available. |
| `scroll_up` | `"ctrl+y"` | Scrolls the visible list up by one row without moving the selection at the list edge. |
| `scroll_down` | `"ctrl+e"` | Scrolls the visible list down by one row without moving the selection at the list edge. |
| `half_page_up` | `"ctrl+u"` | Moves the selection up by half a page. |
| `half_page_down` | `"ctrl+d"` | Moves the selection down by half a page. |
| `next_category` | `"tab"` | Jumps to the next command category when the query is empty. |
| `previous_category` | `"shift+tab"` | Jumps to the previous command category when the query is empty. |

Key names are normalized, so `ctrl-y` and `ctrl+y` are equivalent.

`[[commands]]` fields:

| Field | Required | Description |
| --- | --- | --- |
| `title` | Yes | Command name shown in the palette and used for fuzzy matching. |
| `description` | No | Secondary text shown next to the command title. |
| `category` | No | Group header used when the query is empty, and a weak fuzzy-match field while filtering. |
| `aliases` | No | Short searchable abbreviations rendered as chips. |
| `icon` | No | Glyph shown to the left of the command title when `[ui].glyphs` is true. |
| `action` | Yes | Dispatch type. Must be `tmux`, `shell`, `popup`, or `current_shell`. |
| `command` | Yes | Command string used by the selected `action`. |
| `prompt` | No | Built-in in-popup input prompt. Supported values: `session_name`, `window_name`, `target_index`, `file_path`, `command`, `search_query`, `count`. |
| `show_output` | No | Run a non-interactive `tmux` or `shell` command inside commander and show stdout/stderr in a themed output view. |
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

## Recent Commands

When `[ui].recent_commands` is enabled, selecting a command writes a bounded history file:

1. `$XDG_STATE_HOME/tmux-commander/history.toml`
2. `~/.local/state/tmux-commander/history.toml`

The file is created lazily after the first command selection. It is capped to `[ui].recent_limit`, so it cannot grow beyond the configured number of entries.

Example state file:

```toml
version = 1

[[commands]]
key = "popup:lazygit"
title = "Lazygit"
action = "popup"
command = "lazygit"
last_used = 2026-05-21T10:15:00Z
use_count = 3
```

Commands are identified by `action:command`, not by title. Reusing an existing command updates its `last_used` timestamp, increments `use_count`, and moves it to the top without creating a duplicate. A new command is added at the top and the oldest entry is trimmed only when the list would exceed `recent_limit`.

With an empty search query, recent commands are shown first under a `Recent` heading, followed by a subtle divider and the normal categorized command list. Recent commands still appear again in their normal category. While filtering, all commands are searched normally, but recent commands receive a small score boost.

The same history file also stores recent commands launched from tmux-command mode under a separate `[[tmux_commands]]` section. These entries keep the command arguments, so rerunning a recent `split-window -h` entry does not require retyping `-h`. The tmux-command recent list is capped by `[ui].tmux_recent_limit`, separately from `[ui].recent_limit`.

## Actions

Each command must define an `action` and a `command`.

User-defined actions are terminal handoffs. Selecting a `tmux`, `shell`, `popup`, or `current_shell` command exits the commander first, then dispatches the command. This keeps tmux prompts and interactive terminal programs from competing with the Bubble Tea input loop.

`action = "tmux"` runs `tmux <command>` after the palette exits. The `command` value should omit the leading `tmux`.

```toml
action = "tmux"
command = "split-window -h -c '#{pane_current_path}'"
```

Commands may define `prompt` to collect one value inside the commander popup before dispatch. The input screen uses the same styling as the tmux-command argument prompt.

```toml
action = "tmux"
command = "new-session -d -s {{input}}"
prompt = "session_name"
```

`{{input}}` inserts a shell-quoted value, which is the right default for names and paths. `{{raw_input}}` inserts the typed value without quoting for advanced command templates.

Built-in prompt names:

| Prompt | Use for |
| --- | --- |
| `session_name` | New, rename, or switch-session commands. |
| `window_name` | New or rename-window commands. |
| `target_index` | Commands that target a tmux index, such as moving a window. |
| `file_path` | Save/load commands and commands that need a filesystem path. |
| `command` | Free-form tmux or shell command input. |
| `search_query` | Search commands such as `find-window`. |
| `count` | Numeric repeat-count commands, such as moving a window several positions. |

`action = "shell"` runs the command through the user's shell. By default the palette exits before dispatch, which is best for quick side effects such as copying text to a tmux buffer or opening a URL. If the command produces output you want to read, add `show_output = true` so commander stays open and renders stdout/stderr in an internal output view.

```toml
action = "shell"
command = "open https://github.com"
```

`show_output = true` is intended for non-interactive `tmux` or `shell` commands. Press `Esc` or `q` to return to the command list.

```toml
action = "shell"
command = "date"
show_output = true
```

`action = "current_shell"` sends the command text to the active tmux pane and presses Enter after the palette exits. Use it when you want output to stay in the pane's normal shell history and scrollback.

```toml
action = "current_shell"
command = "git status"
```

`action = "popup"` opens another tmux popup after the palette exits and runs the command inside it. Use it for interactive terminal programs, fullscreen TUIs, or commands with output you want to read without writing into the active pane.

```toml
action = "popup"
command = "lazygit"
popup_width = "95%"
popup_height = "90%"
```

Popup actions inherit the active theme's tmux popup style. `background` / `query` are used for the popup body, and `commander_border` is used for the popup border foreground. Full-screen terminal apps may still draw their own colors.

Popup actions start in the active tmux pane directory via `#{pane_current_path}`.

Interactive tmux prompts are intentionally dispatched after the Bubble Tea UI exits to avoid nested input conflicts.

Built-in internal commands manage `tmux-commander` itself. In-app commands such as `Preview Themes`, `Show Controls`, `Clear Recent Commands`, and `List Config Path` stay inside the commander popup and return to the same query, selection, and scroll position. `Reload Config` reloads TOML and restarts the palette inside the same popup while preserving that position. `Open / Edit Config` is the exception: it exits the commander and opens `$EDITOR` in a tmux popup because the editor is an external interactive program.

## Controls

- Type to filter commands.
- Use `Up` / `Down`, or the configured movement keys, `Ctrl-P` / `Ctrl-N` by default, to move.
- Use the configured one-row scroll keys, `Ctrl-Y` / `Ctrl-E` by default, to scroll the visible list.
- Use the configured half-page keys, `Ctrl-U` / `Ctrl-D` by default, to move by half a page.
- Use the configured category keys, `Tab` / `Shift-Tab` by default, to jump between command categories.
- Press the configured toggle key, `Ctrl-T` by default, to switch between the configured command palette and tmux-command mode. The active key is shown below the search field.
- Press `Enter` to select.
- Press `Esc` or `Ctrl-C` to cancel.

The `Show Controls` internal command displays the active key bindings after user config has been applied.

In tmux-command mode, the palette fuzzy-searches tmux command names. Selecting an argument-capable tmux command opens an argument input view inside commander. Press `Enter` from that view to run `tmux <command> <arguments>`, or `Esc` to return to the tmux command list without losing the previous selection.

In the theme preview view, use `Up` / `Down` or `Left` / `Right` to preview themes. Press `Enter` or `Esc` to return to the command list at the previous position.

Internal message views such as `Show Controls`, `Clear Recent Commands`, and `List Config Path` stay inside the commander popup. Press `Esc` or `q` to return to the command list at the previous position.

When the query is empty, commands are grouped by recent use and category. While filtering, category headers are hidden and results are sorted by fuzzy score with a small recency boost. Multi-token searches such as `split pane` are supported.

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
- Preview Themes
- Clear Recent Commands
- List Config Path
- Show Controls
- Open / Edit Config
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

`tmux-palette` currently has broader customization features such as custom palettes, theme selection, and JSON-powered plugin sources. `tmux-commander` starts with less surface area and prioritizes low-friction native binary installation, including TPM installation backed by precompiled release binaries.

## Development

```sh
go test ./...
go run ./cmd/tmux-commander
go run ./cmd/tmux-commander themes
go run ./cmd/tmux-commander version
go build -o bin/tmux-commander ./cmd/tmux-commander
```

Run inside tmux for a realistic manual test:

```sh
./bin/tmux-commander popup
```

## Release Builds

Release builds are handled by GoReleaser and GitHub Actions. Pushing a `v*` tag runs tests, builds Linux/macOS `amd64` and `arm64` archives, generates `checksums.txt`, publishes a GitHub release, and creates artifact attestations for the release archives.

Create a release:

```sh
git tag v0.1.0
git push origin v0.1.0
```

Run a local snapshot with GoReleaser:

```sh
goreleaser release --snapshot --clean
```

Verify a published release archive:

```sh
gh attestation verify tmux-commander_linux_arm64.tar.gz \
  --repo sschmerda/tmux-commander
```

Build a local binary without GoReleaser:

```sh
go build -trimpath -ldflags="-s -w" -o bin/tmux-commander ./cmd/tmux-commander
```

Example cross-compile commands:

```sh
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/tmux-commander_darwin_arm64 ./cmd/tmux-commander
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/tmux-commander_linux_amd64 ./cmd/tmux-commander
```

Homebrew can be added later by extending `.goreleaser.yaml` with a `brews` publisher that writes to a tap repository.
