package tmuxcmd

import "testing"

func TestInvocationCommandLineIncludesArgs(t *testing.T) {
	invocation := Invocation{Name: "split-window", Args: "-h"}
	if got := invocation.CommandLine(); got != "split-window -h" {
		t.Fatalf("CommandLine = %q", got)
	}
}

func TestParseListCommands(t *testing.T) {
	commands := parseListCommands("split-window (splitw) [-h] [shell-command]\nkill-server\n")
	if len(commands) != 2 {
		t.Fatalf("command count = %d, want 2", len(commands))
	}
	if commands[1].Name != "split-window" || commands[1].Usage != "[-h] [shell-command]" || !commands[1].TakesArgs {
		t.Fatalf("split-window = %#v", commands[1])
	}
	if commands[0].Name != "kill-server" || commands[0].TakesArgs {
		t.Fatalf("kill-server = %#v", commands[0])
	}
	if len(commands[1].ArgHelp) == 0 {
		t.Fatalf("split-window arg help is empty: %#v", commands[1])
	}
	if commands[1].Description == "" {
		t.Fatalf("split-window description is empty: %#v", commands[1])
	}
}

func TestDefaultsIncludeArgumentUsage(t *testing.T) {
	var found bool
	for _, cmd := range Defaults() {
		if cmd.Name == "split-window" {
			found = true
			if cmd.Usage == "" || !cmd.TakesArgs {
				t.Fatalf("split-window = %#v", cmd)
			}
		}
	}
	if !found {
		t.Fatal("split-window not found")
	}
}

func TestDefaultsIncludeArgumentHelpForArgumentCommands(t *testing.T) {
	for _, cmd := range Defaults() {
		if cmd.TakesArgs && len(cmd.ArgHelp) == 0 {
			t.Fatalf("%s has no argument help", cmd.Name)
		}
	}
}

func TestDefaultsCoverTmux36aCommands(t *testing.T) {
	want := []string{
		"attach-session",
		"bind-key",
		"break-pane",
		"capture-pane",
		"choose-buffer",
		"choose-client",
		"choose-tree",
		"clear-history",
		"clear-prompt-history",
		"clock-mode",
		"command-prompt",
		"confirm-before",
		"copy-mode",
		"customize-mode",
		"delete-buffer",
		"detach-client",
		"display-menu",
		"display-message",
		"display-panes",
		"display-popup",
		"find-window",
		"has-session",
		"if-shell",
		"join-pane",
		"kill-pane",
		"kill-server",
		"kill-session",
		"kill-window",
		"last-pane",
		"last-window",
		"link-window",
		"list-buffers",
		"list-clients",
		"list-commands",
		"list-keys",
		"list-panes",
		"list-sessions",
		"list-windows",
		"load-buffer",
		"lock-client",
		"lock-server",
		"lock-session",
		"move-pane",
		"move-window",
		"new-session",
		"new-window",
		"next-layout",
		"next-window",
		"paste-buffer",
		"pipe-pane",
		"previous-layout",
		"previous-window",
		"refresh-client",
		"rename-session",
		"rename-window",
		"resize-pane",
		"resize-window",
		"respawn-pane",
		"respawn-window",
		"rotate-window",
		"run-shell",
		"save-buffer",
		"select-layout",
		"select-pane",
		"select-window",
		"send-keys",
		"send-prefix",
		"server-access",
		"set-buffer",
		"set-environment",
		"set-hook",
		"set-option",
		"set-window-option",
		"show-buffer",
		"show-environment",
		"show-hooks",
		"show-messages",
		"show-options",
		"show-prompt-history",
		"show-window-options",
		"source-file",
		"split-window",
		"start-server",
		"suspend-client",
		"swap-pane",
		"swap-window",
		"switch-client",
		"unbind-key",
		"unlink-window",
		"wait-for",
	}
	have := map[string]bool{}
	for _, cmd := range Defaults() {
		have[cmd.Name] = true
	}
	for _, name := range want {
		if !have[name] {
			t.Fatalf("missing tmux command %q", name)
		}
	}
}
