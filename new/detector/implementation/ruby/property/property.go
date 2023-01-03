package property

import (
	"fmt"

	"github.com/bearer/curio/new/detector/types"
	"github.com/bearer/curio/new/language/tree"
	languagetypes "github.com/bearer/curio/new/language/types"
)

type Data struct {
	Name string
}

type propertyDetector struct {
	pairQuery *tree.Query
}

func New(lang languagetypes.Language) (types.Detector, error) {
	pairQuery, err := lang.CompileQuery(`(pair key: (hash_key_symbol) @key value: (_) @value) @root`)
	if err != nil {
		return nil, fmt.Errorf("error compiling pair query: %s", err)
	}

	return &propertyDetector{pairQuery: pairQuery}, nil
}

func (detector *propertyDetector) Name() string {
	return "property"
}

func (detector *propertyDetector) DetectAt(
	node *tree.Node,
	evaluator types.Evaluator,
) ([]*types.Detection, error) {
	result, err := detector.pairQuery.MatchOnceAt(node)
	if result == nil || err != nil {
		return nil, err
	}

	return []*types.Detection{{
		MatchNode: node,
		Data:      Data{Name: result["key"].Content()},
	}}, nil
}

func (detector *propertyDetector) Close() {
	detector.pairQuery.Close()
}
