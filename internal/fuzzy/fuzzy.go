package fuzzy

import (
	"sort"
	"strings"

	"github.com/stefanschmerda/tmux-commander/internal/config"
)

type Match struct {
	Command config.Command
	Score   int
}

func Filter(commands []config.Command, query string) []Match {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		matches := make([]Match, 0, len(commands))
		for _, cmd := range commands {
			matches = append(matches, Match{Command: cmd})
		}
		return matches
	}

	tokens := strings.Fields(query)
	matches := make([]Match, 0, len(commands))
	for _, cmd := range commands {
		text := searchableText(cmd)
		score := 0
		ok := true
		for _, token := range tokens {
			tokenScore := scoreToken(text, token)
			if tokenScore == 0 {
				ok = false
				break
			}
			score += tokenScore
		}
		if ok {
			matches = append(matches, Match{Command: cmd, Score: score})
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].Command.Title < matches[j].Command.Title
		}
		return matches[i].Score > matches[j].Score
	})
	return matches
}

func searchableText(cmd config.Command) string {
	parts := []string{cmd.Title, initials(cmd.Title), cmd.Description, cmd.Category, cmd.Icon}
	parts = append(parts, cmd.Aliases...)
	return strings.ToLower(strings.Join(parts, " "))
}

func initials(s string) string {
	words := strings.Fields(s)
	var b strings.Builder
	for _, word := range words {
		b.WriteByte(strings.ToLower(word)[0])
	}
	return b.String()
}

func scoreToken(text, token string) int {
	if strings.Contains(text, token) {
		return 100 + len(token)
	}
	score := 0
	pos := 0
	for _, r := range token {
		idx := strings.IndexRune(text[pos:], r)
		if idx < 0 {
			return 0
		}
		score += 3
		pos += idx + 1
	}
	return score
}
