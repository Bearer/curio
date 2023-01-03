package regex

import (
	"regexp"
	"strings"
)

// AnyMatch returns true if any of the regexes match the given string
func AnyMatch(exprs []*regexp.Regexp, str string) bool {
	for _, expr := range exprs {
		if expr.MatchString(str) {
			return true
		}
	}

	return false
}

func ReplaceAllWithSubmatches(
	pattern *regexp.Regexp,
	input string,
	replace func(submatches []string) (string, error),
) (string, error) {
	indices := pattern.FindAllStringSubmatchIndex(input, -1)

	start := 0
	result := strings.Builder{}

	for _, match := range indices {
		result.WriteString(input[start:match[0]])

		submatches := make([]string, len(match)/2)
		for i := 0; i < len(match); i += 2 {
			if match[i] != -1 {
				submatches[i/2] = input[match[i]:match[i+1]]
			}
		}

		replacement, err := replace(submatches)
		if err != nil {
			return "", err
		}

		result.WriteString(replacement)
		start = match[1]
	}

	if start < len(input) {
		result.WriteString(input[start:])
	}

	return result.String(), nil
}
