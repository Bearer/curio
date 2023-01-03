package types

import "github.com/bearer/curio/new/language/tree"

type PatternQuery interface {
	MatchAt(node *tree.Node) ([]tree.QueryResult, error)
	MatchOnceAt(node *tree.Node) (tree.QueryResult, error)
	Close()
}

type Language interface {
	Parse(input string) (*tree.Tree, error)
	CompileQuery(input string) (*tree.Query, error)
	CompilePatternQuery(input string) (PatternQuery, error)
}
