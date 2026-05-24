package history

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stefanschmerda/tmux-commander/internal/config"
)

const version = 1

type File struct {
	Version int     `toml:"version"`
	Entries []Entry `toml:"commands"`
}

type Entry struct {
	Key      string    `toml:"key"`
	Title    string    `toml:"title"`
	Action   string    `toml:"action"`
	Command  string    `toml:"command"`
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
	file.normalize(0)
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
	if limit <= 0 || cmd.Internal != "" {
		return File{Version: version}, nil
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
	file.normalize(limit)
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
	f.normalize(limit)
	keys := make([]string, 0, len(f.Entries))
	for _, entry := range f.Entries {
		keys = append(keys, entry.Key)
	}
	return keys
}

func Trim(file File, limit int) File {
	file.normalize(limit)
	return file
}

func (f *File) normalize(limit int) {
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
	if limit > 0 && len(f.Entries) > limit {
		f.Entries = f.Entries[:limit]
	}
}
