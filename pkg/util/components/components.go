package components

import (
	"regexp"

	"github.com/bearer/curio/pkg/util/normalize_key"
	"github.com/bearer/curio/pkg/util/regex"
)

var keyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bbucket\b`),
	regexp.MustCompile(`\bstore\b`),
	regexp.MustCompile(`\bstorage\b`),
}

func KeyIsRelevant(key string) bool {
	return regex.AnyMatch(keyPatterns, normalize_key.Normalize(key))
}
