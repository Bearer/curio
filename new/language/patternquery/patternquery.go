package patternquery

import (
	"github.com/bearer/curio/new/language/patternquery/builder"
	"github.com/bearer/curio/new/language/tree"
	"github.com/bearer/curio/new/language/types"
)

type Query struct {
	treeQuery       *tree.Query
	paramToVariable map[string]string
	equalParams     [][]string
	paramToContent  map[string]string
}

func Compile(lang types.Language, input string, variables []builder.Variable) (*Query, error) {
	builderResult, err := builder.Build(lang, input, variables)
	if err != nil {
		return nil, err
	}

	treeQuery, err := lang.CompileQuery(builderResult.Query)
	if err != nil {
		return nil, err
	}

	return &Query{
		treeQuery:       treeQuery,
		paramToVariable: builderResult.ParamToVariable,
		equalParams:     builderResult.EqualParams,
		paramToContent:  builderResult.ParamToContent,
	}, nil
}

func (query *Query) MatchAt(node *tree.Node) ([]tree.QueryResult, error) {
	treeResults, err := query.treeQuery.MatchAt(node)
	if err != nil {
		return nil, err
	}

	var results []tree.QueryResult

	for _, treeResult := range treeResults {
		result := query.matchAndTranslateTreeResult(treeResult)
		if result != nil {
			results = append(results, result)
		}
	}

	return results, nil
}

func (query *Query) MatchOnceAt(node *tree.Node) (tree.QueryResult, error) {
	treeResult, err := query.treeQuery.MatchOnceAt(node)
	if err != nil {
		return nil, err
	}

	return query.matchAndTranslateTreeResult(treeResult), nil
}

func (query *Query) Close() {
	query.treeQuery.Close()
}

func (query *Query) matchAndTranslateTreeResult(treeResult tree.QueryResult) tree.QueryResult {
	if treeResult == nil {
		return nil
	}

	for _, equalParams := range query.equalParams {
		value := treeResult[equalParams[0]].Content()

		for _, param := range equalParams[1:] {
			if treeResult[param].Content() != value {
				return nil
			}
		}
	}

	for param, content := range query.paramToContent {
		if treeResult[param].Content() != content {
			return nil
		}
	}

	result := make(tree.QueryResult)

	for paramName, node := range treeResult {
		result[query.paramToVariable[paramName]] = node
	}

	return result
}
