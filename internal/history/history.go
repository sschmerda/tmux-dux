package history

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/sschmerda/tmux-commander/internal/config"
	"github.com/sschmerda/tmux-commander/internal/tmuxcmd"
)

const version = 1

type File struct {
	Version     int         `toml:"version"`
	Entries     []Entry     `toml:"commands"`
	TmuxEntries []TmuxEntry `toml:"tmux_commands"`
}

type Entry struct {
	Key      string    `toml:"key"`
	Title    string    `toml:"title"`
	Action   string    `toml:"action"`
	Command  string    `toml:"command"`
	LastUsed time.Time `toml:"last_used"`
	UseCount int       `toml:"use_count"`
}

type TmuxEntry struct {
	Key      string    `toml:"key"`
	Name     string    `toml:"name"`
	Args     string    `toml:"args"`
	LastUsed time.Time `toml:"last_used"`
	UseCount int       `toml:"use_count"`
}

func Path() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(base, "tmux-commander", "history.toml"), nil
}

func Load(path string) (File, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{Version: version}, nil
		}
		return File{}, err
	}
	var file File
	if _, err := toml.DecodeFile(path, &file); err != nil {
		return File{}, err
	}
	if file.Version == 0 {
		file.Version = version
	}
	file.normalize(-1, -1)
	return file, nil
}

func LoadDefault() (File, string, error) {
	path, err := Path()
	if err != nil {
		return File{}, "", err
	}
	file, err := Load(path)
	return file, path, err
}

func Record(path string, file File, cmd config.Command, limit int, now time.Time) (File, error) {
	return RecordWithLimits(path, file, cmd, limit, limit, now)
}

func RecordWithLimits(path string, file File, cmd config.Command, commandLimit int, tmuxLimit int, now time.Time) (File, error) {
	if commandLimit < 0 {
		commandLimit = 0
	}
	if commandLimit <= 0 || cmd.Internal != "" {
		file.Version = version
		file.normalize(commandLimit, tmuxLimit)
		return file, nil
	}
	key := config.CommandKey(cmd)
	found := false
	for i := range file.Entries {
		if file.Entries[i].Key != key {
			continue
		}
		file.Entries[i].Title = cmd.Title
		file.Entries[i].Action = cmd.Action
		file.Entries[i].Command = cmd.Command
		file.Entries[i].LastUsed = now
		file.Entries[i].UseCount++
		found = true
		break
	}
	if !found {
		file.Entries = append(file.Entries, Entry{
			Key:      key,
			Title:    cmd.Title,
			Action:   cmd.Action,
			Command:  cmd.Command,
			LastUsed: now,
			UseCount: 1,
		})
	}
	file.Version = version
	file.normalize(commandLimit, tmuxLimit)
	return file, Save(path, file)
}

func RecordTmux(path string, file File, invocation tmuxcmd.Invocation, limit int, now time.Time) (File, error) {
	return RecordTmuxWithLimits(path, file, invocation, limit, limit, now)
}

func RecordTmuxWithLimits(path string, file File, invocation tmuxcmd.Invocation, commandLimit int, tmuxLimit int, now time.Time) (File, error) {
	if tmuxLimit < 0 {
		tmuxLimit = 0
	}
	if tmuxLimit <= 0 || strings.TrimSpace(invocation.Name) == "" {
		file.Version = version
		file.normalize(commandLimit, tmuxLimit)
		return file, nil
	}
	invocation.Args = strings.TrimSpace(invocation.Args)
	key := invocation.Key()
	found := false
	for i := range file.TmuxEntries {
		if file.TmuxEntries[i].Key != key {
			continue
		}
		file.TmuxEntries[i].Name = invocation.Name
		file.TmuxEntries[i].Args = invocation.Args
		file.TmuxEntries[i].LastUsed = now
		file.TmuxEntries[i].UseCount++
		found = true
		break
	}
	if !found {
		file.TmuxEntries = append(file.TmuxEntries, TmuxEntry{
			Key:      key,
			Name:     invocation.Name,
			Args:     invocation.Args,
			LastUsed: now,
			UseCount: 1,
		})
	}
	file.Version = version
	file.normalize(commandLimit, tmuxLimit)
	return file, Save(path, file)
}

func Save(path string, file File) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".history-*.toml")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	encodeErr := toml.NewEncoder(tmp).Encode(file)
	closeErr := tmp.Close()
	if encodeErr != nil {
		_ = os.Remove(tmpPath)
		return encodeErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return closeErr
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func (f File) RecentKeys(limit int) []string {
	f.normalize(limit, -1)
	keys := make([]string, 0, len(f.Entries))
	for _, entry := range f.Entries {
		keys = append(keys, entry.Key)
		titleKey := config.CommandTitleKey(config.Command{Title: entry.Title, Action: entry.Action})
		if titleKey != "" {
			keys = append(keys, titleKey)
		}
	}
	return keys
}

func (f File) RecentTmuxKeys(limit int) []string {
	f.normalize(-1, limit)
	keys := make([]string, 0, len(f.TmuxEntries))
	for _, entry := range f.TmuxEntries {
		keys = append(keys, entry.Key)
	}
	return keys
}

func (f File) RecentTmuxInvocations(limit int) []tmuxcmd.Invocation {
	f.normalize(-1, limit)
	invocations := make([]tmuxcmd.Invocation, 0, len(f.TmuxEntries))
	for _, entry := range f.TmuxEntries {
		invocations = append(invocations, tmuxcmd.Invocation{Name: entry.Name, Args: entry.Args})
	}
	return invocations
}

func Trim(file File, limit int) File {
	file.normalize(limit, limit)
	return file
}

func TrimWithLimits(file File, commandLimit int, tmuxLimit int) File {
	file.normalize(commandLimit, tmuxLimit)
	return file
}

func (f *File) normalize(commandLimit int, tmuxLimit int) {
	f.Version = version
	seen := map[string]bool{}
	entries := f.Entries[:0]
	for _, entry := range f.Entries {
		if entry.Key == "" || seen[entry.Key] {
			continue
		}
		seen[entry.Key] = true
		entries = append(entries, entry)
	}
	f.Entries = entries
	sort.SliceStable(f.Entries, func(i, j int) bool {
		return f.Entries[i].LastUsed.After(f.Entries[j].LastUsed)
	})
	if commandLimit >= 0 && len(f.Entries) > commandLimit {
		f.Entries = f.Entries[:commandLimit]
	}

	seenTmux := map[string]bool{}
	tmuxEntries := f.TmuxEntries[:0]
	for _, entry := range f.TmuxEntries {
		if entry.Key == "" || seenTmux[entry.Key] {
			continue
		}
		seenTmux[entry.Key] = true
		tmuxEntries = append(tmuxEntries, entry)
	}
	f.TmuxEntries = tmuxEntries
	sort.SliceStable(f.TmuxEntries, func(i, j int) bool {
		return f.TmuxEntries[i].LastUsed.After(f.TmuxEntries[j].LastUsed)
	})
	if tmuxLimit >= 0 && len(f.TmuxEntries) > tmuxLimit {
		f.TmuxEntries = f.TmuxEntries[:tmuxLimit]
	}
}
