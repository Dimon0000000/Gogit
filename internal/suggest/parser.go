package suggest

import "unicode"

// ParseContext finds the token at the cursor. This first version deliberately
// handles whitespace-separated words only; shell quoting is a later milestone.
func ParseContext(line string, cursor int) (Context, bool) {
	runes := []rune(line)
	if cursor < 0 || cursor > len(runes) {
		return Context{}, false
	}

	start := cursor
	for start > 0 && !unicode.IsSpace(runes[start-1]) {
		start--
	}

	end := cursor
	for end < len(runes) && !unicode.IsSpace(runes[end]) {
		end++
	}

	return Context{
		WordsBefore: splitWords(runes[:start]),
		Token:       string(runes[start:end]),
		Prefix:      string(runes[start:cursor]),
		TokenStart:  start,
		TokenEnd:    end,
	}, true
}

func splitWords(value []rune) []string {
	fields := make([]string, 0, 4)
	for index := 0; index < len(value); {
		for index < len(value) && unicode.IsSpace(value[index]) {
			index++
		}
		if index == len(value) {
			break
		}

		start := index
		for index < len(value) && !unicode.IsSpace(value[index]) {
			index++
		}
		fields = append(fields, string(value[start:index]))
	}
	return fields
}
