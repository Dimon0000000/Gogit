package suggest

import "strings"

// Suggest returns Git completions for the token at cursor.
func Suggest(line string, cursor int) []Suggestion {
	context, ok := ParseContext(line, cursor)
	if !ok || len(context.WordsBefore) == 0 || context.WordsBefore[0] != "git" {
		return nil
	}

	if len(context.WordsBefore) == 1 {
		return matching(gitSubcommands, context.Prefix)
	}

	options, ok := gitOptions[context.WordsBefore[1]]
	if !ok {
		return nil
	}
	return matching(options, context.Prefix)
}

func matching(candidates []Suggestion, prefix string) []Suggestion {
	matched := make([]Suggestion, 0, len(candidates))
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate.Value, prefix) {
			matched = append(matched, candidate)
		}
	}
	return matched
}
