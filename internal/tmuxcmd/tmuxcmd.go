package tmuxcmd

import (
	"os/exec"
	"sort"
	"strings"
)

type Command struct {
	Name        string
	Usage       string
	Description string
	ArgHelp     []string
	TakesArgs   bool
}

type Invocation struct {
	Name string
	Args string
}

func (i Invocation) CommandLine() string {
	if strings.TrimSpace(i.Args) == "" {
		return i.Name
	}
	return i.Name + " " + strings.TrimSpace(i.Args)
}

func (i Invocation) Key() string {
	return strings.ToLower(strings.TrimSpace(i.Name + " " + i.Args))
}

func Load() []Command {
	output, err := exec.Command("tmux", "list-commands").Output()
	if err != nil {
		return Defaults()
	}
	commands := parseListCommands(string(output))
	if len(commands) == 0 {
		return Defaults()
	}
	return enrich(commands)
}

func Defaults() []Command {
	commands := defaultCommands()
	sortCommands(commands)
	return enrich(commands)
}

func defaultCommands() []Command {
	commands := []Command{
		{Name: "attach-session", Usage: "[-dErx] [-c working-directory] [-f flags] [-t target-session]", Description: "Attach to a tmux session", TakesArgs: true},
		{Name: "bind-key", Usage: "[-nr] [-N note] [-T key-table] key [command [argument ...]]", Description: "Bind a key to a tmux command", TakesArgs: true},
		{Name: "break-pane", Usage: "[-abdP] [-F format] [-n window-name] [-s src-pane] [-t dst-window]", Description: "Move a pane into its own window", TakesArgs: true},
		{Name: "capture-pane", Usage: "[-aepPqCJMN] [-b buffer-name] [-E end-line] [-S start-line] [-t target-pane]", Description: "Capture pane contents", TakesArgs: true},
		{Name: "choose-buffer", Usage: "[-NryZ] [-F format] [-f filter] [-K key-format] [-O sort-order] [-t target-pane] [template]", Description: "Open tmux buffer chooser", TakesArgs: true},
		{Name: "choose-client", Usage: "[-NryZ] [-F format] [-f filter] [-K key-format] [-O sort-order] [-t target-pane] [template]", Description: "Open tmux client chooser", TakesArgs: true},
		{Name: "choose-tree", Usage: "[-GNrswyZ] [-F format] [-f filter] [-K key-format] [-O sort-order] [-t target-pane] [template]", Description: "Choose sessions, windows, or panes", TakesArgs: true},
		{Name: "clear-history", Usage: "[-H] [-t target-pane]", Description: "Clear pane history", TakesArgs: true},
		{Name: "clear-prompt-history", Usage: "[-T prompt-type]", Description: "Clear status prompt history", TakesArgs: true},
		{Name: "clock-mode", Usage: "[-t target-pane]", Description: "Display a large clock", TakesArgs: true},
		{Name: "command-prompt", Usage: "[-1bFiklN] [-I inputs] [-p prompts] [-t target-client] [-T prompt-type] [template]", Description: "Open tmux command prompt", TakesArgs: true},
		{Name: "confirm-before", Usage: "[-by] [-c confirm-key] [-p prompt] [-t target-client] command", Description: "Confirm before running a command", TakesArgs: true},
		{Name: "copy-mode", Usage: "[-deHMqSu] [-s src-pane] [-t target-pane]", Description: "Enter copy mode", TakesArgs: true},
		{Name: "customize-mode", Usage: "[-NZ] [-F format] [-f filter] [-t target-pane] [template]", Description: "Open tmux customize mode", TakesArgs: true},
		{Name: "delete-buffer", Usage: "[-b buffer-name]", Description: "Delete a paste buffer", TakesArgs: true},
		{Name: "detach-client", Usage: "[-aP] [-E shell-command] [-s target-session] [-t target-client]", Description: "Detach a tmux client", TakesArgs: true},
		{Name: "display-menu", Usage: "[-OM] [-b border-lines] [-c target-client] [-C starting-choice] [-H selected-style] [-s style] [-S border-style] [-t target-pane] [-T title] [-x position] [-y position] name key command [name key command ...]", Description: "Display a tmux menu", TakesArgs: true},
		{Name: "display-message", Usage: "[-aCIlNpv] [-c target-client] [-d delay] [-t target-pane] [message]", Description: "Show a tmux message", TakesArgs: true},
		{Name: "display-panes", Usage: "[-bN] [-d duration] [-t target-client] [template]", Description: "Display pane indicators", TakesArgs: true},
		{Name: "display-popup", Usage: "[-BCEkN] [-b border-lines] [-c target-client] [-d start-directory] [-e environment] [-h height] [-s style] [-S border-style] [-t target-pane] [-T title] [-w width] [-x position] [-y position] [shell-command [argument ...]]", Description: "Open a tmux popup", TakesArgs: true},
		{Name: "find-window", Usage: "[-iCNrTZ] [-t target-pane] match-string", Description: "Find windows by name or content", TakesArgs: true},
		{Name: "has-session", Usage: "[-t target-session]", Description: "Check whether a session exists", TakesArgs: true},
		{Name: "if-shell", Usage: "[-bF] [-t target-pane] shell-command command [command]", Description: "Run a tmux command conditionally", TakesArgs: true},
		{Name: "join-pane", Usage: "[-bdfhv] [-l size] [-s src-pane] [-t dst-pane]", Description: "Join a pane into another window", TakesArgs: true},
		{Name: "kill-pane", Usage: "[-a] [-t target-pane]", Description: "Kill a pane", TakesArgs: true},
		{Name: "kill-server", Description: "Kill the tmux server"},
		{Name: "kill-session", Usage: "[-aC] [-t target-session]", Description: "Kill a session", TakesArgs: true},
		{Name: "kill-window", Usage: "[-a] [-t target-window]", Description: "Kill a window", TakesArgs: true},
		{Name: "last-pane", Usage: "[-deZ] [-t target-window]", Description: "Select the previously active pane", TakesArgs: true},
		{Name: "last-window", Usage: "[-t target-session]", Description: "Select the previously active window", TakesArgs: true},
		{Name: "link-window", Usage: "[-abdk] [-s src-window] [-t dst-window]", Description: "Link a window into another session", TakesArgs: true},
		{Name: "list-buffers", Usage: "[-F format] [-f filter]", Description: "List paste buffers", TakesArgs: true},
		{Name: "list-clients", Usage: "[-F format] [-f filter] [-t target-session]", Description: "List attached clients", TakesArgs: true},
		{Name: "list-commands", Usage: "[-F format] [command]", Description: "List tmux commands", TakesArgs: true},
		{Name: "list-keys", Usage: "[-1aN] [-P prefix-string] [-T key-table] [key]", Description: "List key bindings", TakesArgs: true},
		{Name: "list-panes", Usage: "[-as] [-F format] [-f filter] [-t target]", Description: "List panes", TakesArgs: true},
		{Name: "list-sessions", Usage: "[-F format] [-f filter]", Description: "List sessions", TakesArgs: true},
		{Name: "list-windows", Usage: "[-a] [-F format] [-f filter] [-t target-session]", Description: "List windows", TakesArgs: true},
		{Name: "load-buffer", Usage: "[-w] [-b buffer-name] [-t target-client] path", Description: "Load a paste buffer from a file", TakesArgs: true},
		{Name: "lock-client", Usage: "[-t target-client]", Description: "Lock a client", TakesArgs: true},
		{Name: "lock-server", Description: "Lock all clients"},
		{Name: "lock-session", Usage: "[-t target-session]", Description: "Lock all clients attached to a session", TakesArgs: true},
		{Name: "move-pane", Usage: "[-bdfhv] [-l size] [-s src-pane] [-t dst-pane]", Description: "Move a pane into another pane's window", TakesArgs: true},
		{Name: "move-window", Usage: "[-abrdk] [-s src-window] [-t dst-window]", Description: "Move a window", TakesArgs: true},
		{Name: "new-session", Usage: "[-AdDEPX] [-c start-directory] [-e environment] [-f flags] [-F format] [-n window-name] [-s session-name] [-t group-name] [-x width] [-y height] [shell-command [argument ...]]", Description: "Create a new session", TakesArgs: true},
		{Name: "new-window", Usage: "[-abdkPS] [-c start-directory] [-e environment] [-F format] [-n window-name] [-t target-window] [shell-command [argument ...]]", Description: "Create a new window", TakesArgs: true},
		{Name: "next-layout", Usage: "[-t target-window]", Description: "Switch to the next layout", TakesArgs: true},
		{Name: "next-window", Usage: "[-a] [-t target-session]", Description: "Select the next window", TakesArgs: true},
		{Name: "pipe-pane", Usage: "[-IOo] [-t target-pane] [shell-command]", Description: "Pipe pane input or output to a shell command", TakesArgs: true},
		{Name: "paste-buffer", Usage: "[-dpr] [-b buffer-name] [-s separator] [-t target-pane]", Description: "Paste a buffer", TakesArgs: true},
		{Name: "previous-layout", Usage: "[-t target-window]", Description: "Switch to the previous layout", TakesArgs: true},
		{Name: "previous-window", Usage: "[-a] [-t target-session]", Description: "Select the previous window", TakesArgs: true},
		{Name: "refresh-client", Usage: "[-cDLRSU] [-A pane:state] [-B name:what:format] [-C size] [-f flags] [-l [target-pane]] [-r pane:report] [-t target-client] [adjustment]", Description: "Refresh or control an attached client", TakesArgs: true},
		{Name: "rename-session", Usage: "[-t target-session] new-name", Description: "Rename a session", TakesArgs: true},
		{Name: "rename-window", Usage: "[-t target-window] new-name", Description: "Rename a window", TakesArgs: true},
		{Name: "resize-pane", Usage: "[-DLMRTUZ] [-t target-pane] [-x width] [-y height] [adjustment]", Description: "Resize a pane", TakesArgs: true},
		{Name: "resize-window", Usage: "[-aADLRU] [-t target-window] [-x width] [-y height] [adjustment]", Description: "Resize a window", TakesArgs: true},
		{Name: "respawn-pane", Usage: "[-k] [-c start-directory] [-e environment] [-t target-pane] [shell-command [argument ...]]", Description: "Respawn a pane", TakesArgs: true},
		{Name: "respawn-window", Usage: "[-k] [-c start-directory] [-e environment] [-t target-window] [shell-command [argument ...]]", Description: "Respawn a window", TakesArgs: true},
		{Name: "rotate-window", Usage: "[-DUZ] [-t target-window]", Description: "Rotate panes in a window", TakesArgs: true},
		{Name: "run-shell", Usage: "[-bCE] [-c start-directory] [-d delay] [-t target-pane] [shell-command]", Description: "Run a shell command from tmux", TakesArgs: true},
		{Name: "save-buffer", Usage: "[-a] [-b buffer-name] path", Description: "Save a paste buffer to a file", TakesArgs: true},
		{Name: "select-layout", Usage: "[-Enop] [-t target-pane] [layout-name]", Description: "Select a window layout", TakesArgs: true},
		{Name: "select-pane", Usage: "[-DdeLlMmRUZ] [-T title] [-t target-pane]", Description: "Select or modify a pane", TakesArgs: true},
		{Name: "select-window", Usage: "[-lnpT] [-t target-window]", Description: "Select a window", TakesArgs: true},
		{Name: "send-keys", Usage: "[-FHKlMRX] [-c target-client] [-N repeat-count] [-t target-pane] key ...", Description: "Send keys to a pane", TakesArgs: true},
		{Name: "send-prefix", Usage: "[-2] [-t target-pane]", Description: "Send the prefix key to a pane", TakesArgs: true},
		{Name: "server-access", Usage: "[-adlrw] [user]", Description: "Change tmux server access permissions", TakesArgs: true},
		{Name: "set-buffer", Usage: "[-aw] [-b buffer-name] [-t target-client] [-n new-buffer-name] data", Description: "Set a paste buffer", TakesArgs: true},
		{Name: "set-environment", Usage: "[-Fhgru] [-t target-session] name [value]", Description: "Set an environment variable", TakesArgs: true},
		{Name: "set-hook", Usage: "[-agpRuw] [-t target-pane] hook-name [command]", Description: "Set a tmux hook", TakesArgs: true},
		{Name: "set-option", Usage: "[-aFgopqsuUw] [-t target-pane] option [value]", Description: "Set a tmux option", TakesArgs: true},
		{Name: "set-window-option", Usage: "[-aFgoqu] [-t target-window] option [value]", Description: "Set a window option", TakesArgs: true},
		{Name: "show-buffer", Usage: "[-b buffer-name]", Description: "Show a paste buffer", TakesArgs: true},
		{Name: "show-environment", Usage: "[-hgs] [-t target-session] [name]", Description: "Show environment variables", TakesArgs: true},
		{Name: "show-hooks", Usage: "[-gpw] [-t target-pane] [hook]", Description: "Show tmux hooks", TakesArgs: true},
		{Name: "show-messages", Usage: "[-JT] [-t target-client]", Description: "Show tmux messages", TakesArgs: true},
		{Name: "show-options", Usage: "[-AgHpqsvw] [-t target-pane] [option]", Description: "Show tmux options", TakesArgs: true},
		{Name: "show-prompt-history", Usage: "[-T prompt-type]", Description: "Show status prompt history", TakesArgs: true},
		{Name: "show-window-options", Usage: "[-AgHpqv] [-t target-window] [option]", Description: "Show window options", TakesArgs: true},
		{Name: "source-file", Usage: "[-Fnqv] [-t target-pane] path ...", Description: "Source a tmux config file", TakesArgs: true},
		{Name: "split-window", Usage: "[-bdfhIvPZ] [-c start-directory] [-e environment] [-F format] [-l size] [-t target-pane] [shell-command [argument ...]]", Description: "Split a pane", TakesArgs: true},
		{Name: "start-server", Description: "Start the tmux server"},
		{Name: "suspend-client", Usage: "[-t target-client]", Description: "Suspend a tmux client", TakesArgs: true},
		{Name: "swap-pane", Usage: "[-dDUZ] [-s src-pane] [-t dst-pane]", Description: "Swap panes", TakesArgs: true},
		{Name: "swap-window", Usage: "[-d] [-s src-window] [-t dst-window]", Description: "Swap windows", TakesArgs: true},
		{Name: "switch-client", Usage: "[-ElnprZ] [-c target-client] [-t target-session] [-T key-table]", Description: "Switch a client", TakesArgs: true},
		{Name: "unbind-key", Usage: "[-anq] [-T key-table] key", Description: "Remove a key binding", TakesArgs: true},
		{Name: "unlink-window", Usage: "[-k] [-t target-window]", Description: "Unlink a window from a session", TakesArgs: true},
		{Name: "wait-for", Usage: "[-L | -S | -U] channel", Description: "Wait for or signal a named tmux channel", TakesArgs: true},
	}
	return commands
}

func parseListCommands(output string) []Command {
	lines := strings.Split(output, "\n")
	commands := make([]Command, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, usage := splitCommandLine(line)
		if name == "" {
			continue
		}
		commands = append(commands, Command{
			Name:      name,
			Usage:     usage,
			TakesArgs: usage != "",
		})
	}
	sortCommands(commands)
	return enrich(commands)
}

func enrich(commands []Command) []Command {
	descriptions := fallbackDescriptions()
	for i := range commands {
		if commands[i].Description == "" {
			commands[i].Description = descriptions[commands[i].Name]
		}
		commands[i].ArgHelp = argumentHelp(commands[i])
	}
	return commands
}

func fallbackDescriptions() map[string]string {
	descriptions := map[string]string{}
	for _, cmd := range defaultCommands() {
		descriptions[cmd.Name] = cmd.Description
	}
	return descriptions
}

func argumentHelp(cmd Command) []string {
	switch cmd.Name {
	case "attach-session":
		return []string{"-d: detach other clients from the session", "-r: attach read-only", "-x: send SIGHUP to the parent of detached clients", "-c working-directory: set session working directory", "-f flags: set client flags", "-t target-session: session to attach"}
	case "bind-key":
		return []string{"-n: bind without requiring prefix", "-r: key may repeat", "-N note: attach a note to the binding", "-T key-table: bind in a specific key table", "key: key to bind", "command ...: tmux command to run"}
	case "break-pane":
		return []string{"-d: do not select the new window", "-P: print information about the new window", "-F format: output format for -P", "-n window-name: name for the new window", "-s src-pane: pane to break out", "-t dst-window: destination window"}
	case "capture-pane":
		return []string{"-p: print captured text instead of storing it", "-b buffer-name: store capture in a named buffer", "-S start-line: first line to capture", "-E end-line: last line to capture", "-t target-pane: pane to capture"}
	case "choose-buffer":
		return []string{"-N: start without preview", "-r: reverse sort order", "-y: skip confirmation prompts", "-Z: zoom while chooser is active", "-f filter: initial filter expression", "-F format: format each buffer row", "-K key-format: format shortcut keys", "-O sort-order: sort by time, name, or size", "-t target-pane: pane where chooser opens", "template: command run with %% replaced by the selected buffer name"}
	case "choose-client":
		return []string{"-N: start without preview", "-r: reverse sort order", "-y: skip confirmation prompts", "-Z: zoom while chooser is active", "-f filter: initial filter expression", "-F format: format each client row", "-K key-format: format shortcut keys", "-O sort-order: sort by name, size, creation, or activity", "-t target-pane: pane where chooser opens", "template: command run with %% replaced by the selected client name"}
	case "choose-tree":
		return []string{"-G: include all sessions in session groups", "-N: start without preview", "-r: reverse sort order", "-s: start with sessions collapsed", "-w: start with windows collapsed", "-y: skip confirmation prompts", "-Z: zoom while chooser is active", "-f filter: initial filter expression", "-F format: format each session/window/pane row", "-K key-format: format shortcut keys", "-O sort-order: sort by index, name, or activity time", "-t target-pane: pane where chooser opens", "template: command run with %% or %1 replaced by the selected session, window, or pane target"}
	case "customize-mode":
		return []string{"-N: start without preview", "-Z: zoom while customize mode is active", "-f filter: initial filter expression", "-F format: format each option row", "-t target-pane: pane where customize mode opens", "template: command run for the selected option"}
	case "clear-history":
		return []string{"-H: also remove hyperlinks", "-t target-pane: pane whose history should be cleared"}
	case "clear-prompt-history", "show-prompt-history":
		return []string{"-T prompt-type: prompt history type such as command, search, target, or window-target"}
	case "clock-mode":
		return []string{"-t target-pane: pane where clock mode should open"}
	case "command-prompt":
		return []string{"-p prompts: comma-separated prompt labels", "-I inputs: initial prompt values", "-T prompt-type: completion/history type", "-F: expand template as a format", "-1: accept one key only", "-k: accept one key and translate to key name", "-N: accept numeric input only", "template: command template using %% or %1..%9"}
	case "confirm-before":
		return []string{"-p prompt: confirmation prompt text", "-c confirm-key: key that confirms", "-y: Enter confirms by default", "-b: show prompt in the background", "command: tmux command to run after confirmation"}
	case "copy-mode":
		return []string{"-e: exit copy mode at bottom", "-M: begin mouse drag copy mode", "-q: cancel if already in copy mode", "-s src-pane: copy from source pane", "-t target-pane: pane to put into copy mode", "-u: scroll one page up"}
	case "delete-buffer", "show-buffer":
		return []string{"-b buffer-name: buffer to use"}
	case "detach-client":
		return []string{"-a: detach all clients except target", "-P: send SIGHUP to parent process", "-E shell-command: replace client with shell command", "-s target-session: detach clients attached to session", "-t target-client: client to detach"}
	case "display-menu":
		return []string{"-T title: menu title", "-x position: horizontal position", "-y position: vertical position", "-s style: menu style", "-S border-style: border style", "-H selected-style: selected item style", "name key command ...: menu items"}
	case "display-message":
		return []string{"-p: print to stdout", "-d delay: display duration", "-t target-pane: pane for format data", "-c target-client: client to display on", "-a: list format variables", "message: message or format string"}
	case "display-panes":
		return []string{"-b: do not block other commands", "-N: keep pane indicators until timeout", "-d duration: indicator duration", "-t target-client: client to display on", "template: command run with %% replaced by pane ID"}
	case "display-popup":
		return []string{"-E: close when command exits", "-B: no border", "-b border-lines: border character style", "-d start-directory: popup working directory", "-e environment: set environment variable", "-w width: popup width", "-h height: popup height", "-x position / -y position: popup position", "shell-command ...: command to run inside popup"}
	case "find-window":
		return []string{"-i: ignore case", "-C: match visible content", "-N: match window name", "-T: match pane title", "-r: regular expression", "-Z: zoom result", "-t target-pane: starting target", "match-string: text or pattern to search"}
	case "has-session":
		return []string{"-t target-session: session to check"}
	case "if-shell":
		return []string{"-b: run shell command in background", "-F: treat expanded shell-command as condition", "-t target-pane: target for formats", "shell-command: condition command", "command: command if true", "command: optional command if false"}
	case "join-pane", "move-pane":
		return []string{"-h: split horizontally", "-v: split vertically", "-b: place before or above target", "-d: do not select moved pane", "-l size: size of new pane", "-s src-pane: pane to move", "-t dst-pane: destination pane"}
	case "kill-pane", "kill-window":
		return []string{"-a: kill all except target", "-t target: pane or window to kill"}
	case "kill-session":
		return []string{"-a: kill all sessions except target", "-C: clear alerts", "-t target-session: session to kill"}
	case "last-pane":
		return []string{"-d: disable input to pane", "-e: enable input to pane", "-Z: keep zoomed state", "-t target-window: window containing panes"}
	case "last-window", "next-layout", "next-window", "previous-layout", "previous-window":
		return []string{"-t target-session/window: session or window target", "-a: choose next/previous window with an alert when supported"}
	case "link-window", "move-window":
		return []string{"-a: place after destination", "-b: place before destination", "-d: do not select linked/moved window", "-k: kill destination if it exists", "-r: renumber windows when supported", "-s src-window: source window", "-t dst-window: destination window"}
	case "list-buffers":
		return []string{"-F format: output format", "-f filter: filter expression"}
	case "list-clients":
		return []string{"-F format: output format", "-f filter: filter expression", "-t target-session: list clients attached to a session"}
	case "list-commands":
		return []string{"-F format: output format", "command: optional command name to describe"}
	case "list-keys":
		return []string{"-1: list only the first matching key", "-a: list all key tables", "-N: include key notes", "-P prefix-string: prefix each line", "-T key-table: key table to list", "key: optional key to list"}
	case "list-panes":
		return []string{"-a: list panes in all sessions", "-s: list panes in the target session", "-F format: output format", "-f filter: filter expression", "-t target: pane, window, or session to list"}
	case "list-sessions":
		return []string{"-F format: output format", "-f filter: filter expression"}
	case "list-windows":
		return []string{"-a: list windows in all sessions", "-F format: output format", "-f filter: filter expression", "-t target-session: session whose windows should be listed"}
	case "load-buffer":
		return []string{"-w: also send buffer to clipboard", "-b buffer-name: target buffer name", "-t target-client: client for clipboard", "path: file to load, or - for stdin"}
	case "lock-client", "lock-session", "suspend-client":
		return []string{"-t target: client or session to affect"}
	case "new-session":
		return []string{"-d: do not attach", "-A: attach if session already exists", "-c start-directory: working directory", "-e environment: set environment variable", "-n window-name: initial window name", "-s session-name: new session name", "-t group-name: session group", "-x width / -y height: initial size", "shell-command ...: initial command"}
	case "new-window":
		return []string{"-a: insert after target", "-b: insert before target", "-d: do not select new window", "-k: kill target if it exists", "-S: select existing named window", "-c start-directory: working directory", "-e environment: set environment variable", "-n window-name: new window name", "-t target-window: target position", "shell-command ...: command to run"}
	case "paste-buffer":
		return []string{"-d: delete buffer after pasting", "-p: bracket paste", "-r: do not replace linefeeds", "-b buffer-name: buffer to paste", "-s separator: line separator", "-t target-pane: pane to paste into"}
	case "pipe-pane":
		return []string{"-I: pipe command output into pane", "-O: pipe pane output to command", "-o: open only if no pipe exists", "-t target-pane: pane to pipe", "shell-command: command connected to pane"}
	case "refresh-client":
		return []string{"-S: update status line only", "-C size: set control mode client size", "-A pane:state: control pane pause state", "-B name:what:format: set format subscription", "-f flags: set client flags", "-t target-client: client to refresh"}
	case "rename-session", "rename-window":
		return []string{"-t target: session or window to rename", "new-name: replacement name"}
	case "resize-pane":
		return []string{"-D/-U/-L/-R: resize down/up/left/right", "-Z: toggle pane zoom", "-M: begin mouse resizing", "-T: trim lines below cursor", "-x width: absolute width", "-y height: absolute height", "-t target-pane: pane to resize", "adjustment: cells or percentage"}
	case "resize-window":
		return []string{"-D/-U/-L/-R: resize down/up/left/right", "-A: largest attached session size", "-a: smallest attached session size", "-x width: absolute width", "-y height: absolute height", "-t target-window: window to resize", "adjustment: cells"}
	case "respawn-pane", "respawn-window":
		return []string{"-k: kill existing command first", "-c start-directory: working directory", "-e environment: set environment variable", "-t target: pane or window to respawn", "shell-command ...: replacement command"}
	case "rotate-window":
		return []string{"-D: rotate down", "-U: rotate up", "-Z: keep zoomed state", "-t target-window: window to rotate"}
	case "run-shell":
		return []string{"-b: run in background", "-C: run as tmux command", "-E: redirect stderr to stdout", "-c start-directory: working directory", "-d delay: delay in seconds", "-t target-pane: pane for output", "shell-command: command to run"}
	case "save-buffer":
		return []string{"-a: append to file", "-b buffer-name: buffer to save", "path: destination file, or - for stdout"}
	case "select-layout":
		return []string{"-E: spread panes evenly", "-n: next layout", "-o: previous layout", "-p: print current layout", "-t target-pane: target pane/window", "layout-name: layout to select"}
	case "select-pane":
		return []string{"-D/-U/-L/-R: select neighboring pane", "-Z: toggle zoom", "-d: disable input", "-e: enable input", "-m: mark pane", "-M: clear marked pane", "-T title: set pane title", "-t target-pane: pane to select"}
	case "select-window":
		return []string{"-l: last window", "-n: next window", "-p: previous window", "-T: toggle activity alert", "-t target-window: window to select"}
	case "send-keys":
		return []string{"-l: send keys literally", "-R: reset terminal state", "-X: send copy-mode command", "-N repeat-count: repeat keys", "-t target-pane: pane to receive keys", "key ...: keys or text to send"}
	case "send-prefix":
		return []string{"-2: send the secondary prefix", "-t target-pane: pane to receive prefix"}
	case "server-access":
		return []string{"-a: add access", "-d: delete access", "-l: list access", "-r: make read-only", "-w: make writable", "user: system user"}
	case "set-buffer":
		return []string{"-a: append to buffer", "-w: also send to clipboard", "-b buffer-name: buffer to set", "-n new-buffer-name: rename buffer", "-t target-client: client for clipboard", "data: buffer contents"}
	case "set-environment":
		return []string{"-g: global environment", "-h: hidden variable", "-r: remove variable", "-u: unset variable", "-F: expand formats", "-t target-session: session environment", "variable [value]: variable to set"}
	case "set-hook":
		return []string{"-a: append hook", "-g: global hook", "-R: run hook immediately", "-u: unset hook", "-w: window hook", "-t target-pane: target scope", "hook-name [command]: hook and command"}
	case "set-option", "set-window-option":
		return []string{"-g: global option", "-u: unset option", "-a: append to option", "-F: expand formats", "-q: quiet errors", "-t target: pane/window/session target", "option [value]: option to set"}
	case "show-environment":
		return []string{"-g: global environment", "-h: include hidden variables", "-s: shell output format", "-t target-session: session environment", "variable: optional variable name"}
	case "show-hooks":
		return []string{"-g: global hooks", "-p: pane hooks", "-w: window hooks", "-t target-pane: target scope", "hook: optional hook name"}
	case "show-options", "show-window-options":
		return []string{"-g: global options", "-v: value only", "-q: quiet errors", "-H: include hooks where supported", "-t target: pane/window/session target", "option: optional option name"}
	case "source-file":
		return []string{"-F: expand formats in path", "-n: parse only", "-q: quiet missing files", "-v: verbose parsed commands", "-t target-pane: target for parsed commands", "path ...: files or globs to source"}
	case "split-window":
		return []string{"-h: split left/right", "-v: split top/bottom", "-d: do not select new pane", "-c start-directory: working directory", "-e environment: set environment variable", "-l size: pane size", "-t target-pane: pane to split", "shell-command ...: command to run in new pane"}
	case "swap-pane":
		return []string{"-d: do not select swapped pane", "-D: swap with next pane", "-U: swap with previous pane", "-Z: keep zoomed state", "-s src-pane: source pane", "-t dst-pane: destination pane"}
	case "swap-window":
		return []string{"-d: do not select swapped window", "-s src-window: source window", "-t dst-window: destination window"}
	case "switch-client":
		return []string{"-l: last session", "-n: next session", "-p: previous session", "-r: toggle read-only", "-Z: keep zoomed state", "-c target-client: client to switch", "-t target-session: session to switch to", "-T key-table: switch key table"}
	case "unbind-key":
		return []string{"-a: remove all bindings", "-n: root table binding", "-q: quiet missing binding", "-T key-table: key table", "key: key to unbind"}
	case "unlink-window":
		return []string{"-k: kill window if only linked once", "-t target-window: window to unlink"}
	case "wait-for":
		return []string{"-L: lock channel", "-S: signal channel", "-U: unlock channel", "channel: named wait channel"}
	default:
		return genericArgumentHelp(cmd.Usage)
	}
}

func genericArgumentHelp(usage string) []string {
	if strings.TrimSpace(usage) == "" {
		return nil
	}
	help := []string{}
	for _, item := range []struct {
		needle string
		text   string
	}{
		{"target-client", "target-client: tmux client to affect"},
		{"target-session", "target-session: tmux session to affect"},
		{"target-window", "target-window: tmux window to affect"},
		{"target-pane", "target-pane: tmux pane to affect"},
		{"src-pane", "src-pane: source pane"},
		{"dst-pane", "dst-pane: destination pane"},
		{"src-window", "src-window: source window"},
		{"dst-window", "dst-window: destination window"},
		{"format", "format: tmux format string"},
		{"filter", "filter: tmux filter expression"},
		{"buffer-name", "buffer-name: named paste buffer"},
		{"shell-command", "shell-command: command run by the shell"},
		{"command", "command: tmux command string"},
		{"path", "path: filesystem path"},
		{"option", "option: tmux option name"},
		{"value", "value: option or variable value"},
		{"key", "key: tmux key name"},
		{"template", "template: command template"},
		{"environment", "environment: VARIABLE=value"},
		{"width", "width: width in cells or percent where supported"},
		{"height", "height: height in cells or percent where supported"},
	} {
		if strings.Contains(usage, item.needle) {
			help = append(help, item.text)
		}
	}
	if len(help) == 0 {
		help = append(help, "arguments follow the tmux usage line above")
	}
	return help
}

func splitCommandLine(line string) (string, string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", ""
	}
	name := fields[0]
	usage := strings.TrimSpace(strings.TrimPrefix(line, name))
	if strings.HasPrefix(usage, "(") {
		if end := strings.Index(usage, ")"); end >= 0 {
			usage = strings.TrimSpace(usage[end+1:])
		}
	}
	return name, usage
}

func sortCommands(commands []Command) {
	sort.SliceStable(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
}
