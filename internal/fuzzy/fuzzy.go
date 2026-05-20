package fuzzy

import (
	"sort"
	"strings"

	sahilfuzzy "github.com/sahilm/fuzzy"
	"github.com/stefanschmerda/tmux-commander/internal/config"
)

type Match struct {
	Command      config.Command
	Score        int
	TitleIndexes []int
	AliasIndexes map[string][]int
}

type fieldKind int

const (
	fieldTitle fieldKind = iota
	fieldAlias
	fieldCategory
	fieldInitials
)

const (
	titleWeight    = 1000
	aliasWeight    = 900
	initialsWeight = 700
	categoryWeight = 250
)

type field struct {
	kind   fieldKind
	value  string
	alias  string
	weight int
}

func Filter(commands []config.Command, query string) []Match {
	query = strings.TrimSpace(query)
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
		match := Match{
			Command:      cmd,
			AliasIndexes: map[string][]int{},
		}
		ok := true
		for _, token := range tokens {
			fieldMatch, found := bestFieldMatch(token, searchableFields(cmd))
			if !found {
				ok = false
				break
			}
			match.Score += fieldMatch.score
			mergeFieldMatch(&match, fieldMatch)
		}
		if ok {
			match.TitleIndexes = sortedUnique(match.TitleIndexes)
			for alias, indexes := range match.AliasIndexes {
				match.AliasIndexes[alias] = sortedUnique(indexes)
			}
			matches = append(matches, match)
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

type fieldMatch struct {
	field   field
	indexes []int
	score   int
}

func bestFieldMatch(token string, fields []field) (fieldMatch, bool) {
	var best fieldMatch
	found := false
	for _, field := range fields {
		matches := sahilfuzzy.Find(token, []string{field.value})
		if len(matches) == 0 {
			continue
		}
		score := matches[0].Score + field.weight*len(matches[0].MatchedIndexes)
		if !found || score > best.score {
			best = fieldMatch{
				field:   field,
				indexes: matches[0].MatchedIndexes,
				score:   score,
			}
			found = true
		}
	}
	return best, found
}

func searchableFields(cmd config.Command) []field {
	fields := []field{
		{kind: fieldTitle, value: cmd.Title, weight: titleWeight},
		{kind: fieldCategory, value: cmd.Category, weight: categoryWeight},
		{kind: fieldInitials, value: initials(cmd.Title), weight: initialsWeight},
	}
	for _, alias := range cmd.Aliases {
		fields = append(fields, field{kind: fieldAlias, value: alias, alias: alias, weight: aliasWeight})
	}
	return fields
}

func mergeFieldMatch(match *Match, fieldMatch fieldMatch) {
	switch fieldMatch.field.kind {
	case fieldTitle:
		match.TitleIndexes = append(match.TitleIndexes, fieldMatch.indexes...)
	case fieldAlias:
		match.AliasIndexes[fieldMatch.field.alias] = append(match.AliasIndexes[fieldMatch.field.alias], fieldMatch.indexes...)
	}
}

func sortedUnique(indexes []int) []int {
	if len(indexes) == 0 {
		return nil
	}
	sort.Ints(indexes)
	result := indexes[:0]
	previous := -1
	for _, index := range indexes {
		if index == previous {
			continue
		}
		result = append(result, index)
		previous = index
	}
	return result
}

func initials(s string) string {
	words := strings.Fields(s)
	var b strings.Builder
	for _, word := range words {
		b.WriteByte(strings.ToLower(word)[0])
	}
	return b.String()
}
