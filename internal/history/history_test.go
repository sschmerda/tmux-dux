package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stefanschmerda/tmux-commander/internal/config"
	"github.com/stefanschmerda/tmux-commander/internal/tmuxcmd"
)

func TestPathUsesXDGStateHome(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/state-root")
	path, err := Path()
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}
	want := filepath.Join("/tmp/state-root", "tmux-commander", "history.toml")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestPathFallsBackToLocalState(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := Path()
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}
	want := filepath.Join(home, ".local", "state", "tmux-commander", "history.toml")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestRecordDeduplicatesAndMovesCommandToTop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.toml")
	base := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	lazygit := config.Command{Title: "Lazygit", Action: "popup", Command: "lazygit"}
	btop := config.Command{Title: "Btop", Action: "popup", Command: "btop"}

	file, err := Record(path, File{Version: version}, lazygit, 10, base)
	if err != nil {
		t.Fatalf("Record returned error: %v", err)
	}
	file, err = Record(path, file, btop, 10, base.Add(time.Minute))
	if err != nil {
		t.Fatalf("Record returned error: %v", err)
	}
	file, err = Record(path, file, lazygit, 10, base.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("Record returned error: %v", err)
	}

	if len(file.Entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(file.Entries))
	}
	if file.Entries[0].Title != "Lazygit" || file.Entries[0].UseCount != 2 {
		t.Fatalf("first entry = %#v", file.Entries[0])
	}
	if file.Entries[1].Title != "Btop" {
		t.Fatalf("second entry = %#v", file.Entries[1])
	}
}

func TestRecordTrimsToLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.toml")
	base := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	file := File{Version: version}
	for i, title := range []string{"One", "Two", "Three"} {
		var err error
		file, err = Record(path, file, config.Command{Title: title, Action: "tmux", Command: title}, 2, base.Add(time.Duration(i)*time.Minute))
		if err != nil {
			t.Fatalf("Record returned error: %v", err)
		}
	}
	if len(file.Entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(file.Entries))
	}
	if file.Entries[0].Title != "Three" || file.Entries[1].Title != "Two" {
		t.Fatalf("entries = %#v", file.Entries)
	}
}

func TestRecordTmuxDeduplicatesByNameAndArgs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.toml")
	base := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	splitHorizontal := tmuxcmd.Invocation{Name: "split-window", Args: "-h"}
	splitVertical := tmuxcmd.Invocation{Name: "split-window", Args: "-v"}

	file, err := RecordTmux(path, File{Version: version}, splitHorizontal, 10, base)
	if err != nil {
		t.Fatalf("RecordTmux returned error: %v", err)
	}
	file, err = RecordTmux(path, file, splitVertical, 10, base.Add(time.Minute))
	if err != nil {
		t.Fatalf("RecordTmux returned error: %v", err)
	}
	file, err = RecordTmux(path, file, splitHorizontal, 10, base.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("RecordTmux returned error: %v", err)
	}

	if len(file.TmuxEntries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(file.TmuxEntries))
	}
	if file.TmuxEntries[0].Name != "split-window" || file.TmuxEntries[0].Args != "-h" || file.TmuxEntries[0].UseCount != 2 {
		t.Fatalf("first entry = %#v", file.TmuxEntries[0])
	}
	if file.TmuxEntries[1].Args != "-v" {
		t.Fatalf("second entry = %#v", file.TmuxEntries[1])
	}
}

func TestRecentTmuxInvocationsPreservesArgs(t *testing.T) {
	now := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	file := File{Version: version, TmuxEntries: []TmuxEntry{{
		Key:      "split-window -h",
		Name:     "split-window",
		Args:     "-h",
		LastUsed: now,
		UseCount: 1,
	}}}

	invocations := file.RecentTmuxInvocations(10)
	if len(invocations) != 1 || invocations[0].Name != "split-window" || invocations[0].Args != "-h" {
		t.Fatalf("invocations = %#v", invocations)
	}
}

func TestTrimWithLimitsUsesSeparateCommandAndTmuxLimits(t *testing.T) {
	base := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	file := File{
		Version: version,
		Entries: []Entry{
			{Key: "tmux:one", Title: "One", LastUsed: base},
			{Key: "tmux:two", Title: "Two", LastUsed: base.Add(time.Minute)},
		},
		TmuxEntries: []TmuxEntry{
			{Key: "split-window -h", Name: "split-window", Args: "-h", LastUsed: base},
			{Key: "new-window", Name: "new-window", LastUsed: base.Add(time.Minute)},
			{Key: "kill-pane", Name: "kill-pane", LastUsed: base.Add(2 * time.Minute)},
		},
	}

	trimmed := TrimWithLimits(file, 1, 2)
	if len(trimmed.Entries) != 1 || trimmed.Entries[0].Title != "Two" {
		t.Fatalf("command entries = %#v", trimmed.Entries)
	}
	if len(trimmed.TmuxEntries) != 2 || trimmed.TmuxEntries[0].Name != "kill-pane" || trimmed.TmuxEntries[1].Name != "new-window" {
		t.Fatalf("tmux entries = %#v", trimmed.TmuxEntries)
	}
}

func TestTrimBoundsExistingFile(t *testing.T) {
	base := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	file := File{Version: version, Entries: []Entry{
		{Key: "tmux:one", Title: "One", LastUsed: base},
		{Key: "tmux:two", Title: "Two", LastUsed: base.Add(time.Minute)},
	}}
	trimmed := Trim(file, 1)
	if len(trimmed.Entries) != 1 || trimmed.Entries[0].Title != "Two" {
		t.Fatalf("trimmed = %#v", trimmed)
	}
}

func TestLoadMissingReturnsEmptyHistory(t *testing.T) {
	file, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if file.Version != version || len(file.Entries) != 0 {
		t.Fatalf("file = %#v", file)
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.toml")
	now := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	file := File{Version: version, Entries: []Entry{{
		Key:      "popup:lazygit",
		Title:    "Lazygit",
		Action:   "popup",
		Command:  "lazygit",
		LastUsed: now,
		UseCount: 3,
	}}}
	if err := Save(path, file); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("history file was not written: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(loaded.Entries) != 1 || loaded.Entries[0].Title != "Lazygit" || loaded.Entries[0].UseCount != 3 {
		t.Fatalf("loaded = %#v", loaded)
	}
}
