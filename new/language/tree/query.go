package tree

import (
	"errors"

	sitter "github.com/smacker/go-tree-sitter"
)

type Query struct {
	sitterQuery *sitter.Query
}

type QueryResult map[string]*Node

func CompileQuery(sitterLanguage *sitter.Language, input string) (*Query, error) {
	sitterQuery, err := sitter.NewQuery([]byte(input), sitterLanguage)
	if err != nil {
		return nil, err
	}

	return &Query{sitterQuery: sitterQuery}, nil
}

func (query *Query) MatchAt(node *Node) ([]QueryResult, error) {
	cursor := sitter.NewQueryCursor()
	defer cursor.Close()

	cursor.SetPointRange(node.sitterNode.StartPoint(), node.sitterNode.EndPoint())
	cursor.Exec(query.sitterQuery, node.sitterNode)

	var results []QueryResult

	for {
		match, found := cursor.NextMatch()
		if !found {
			break
		}

		result := make(QueryResult)
		for _, capture := range match.Captures {
			result[query.sitterQuery.CaptureNameForId(capture.Index)] = node.tree.wrap(capture.Node)
		}

		resultRoot, rootExists := result["root"]
		if !rootExists {
			return nil, errors.New("missing @root capture in tree sitter query")
		}

		// FIXME
		if node.Equal(resultRoot) {
			results = append(results, result)
		}
	}

	return results, nil
}

func (query *Query) MatchOnceAt(node *Node) (QueryResult, error) {
	results, err := query.MatchAt(node)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}
	if len(results) > 1 {
		return nil, errors.New("query returned more than one result")
	}

	return results[0], nil
}

func (query *Query) Close() {
	query.sitterQuery.Close()
}
