package custom

import (
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/bearer/curio/new/detector/types"
	"github.com/bearer/curio/new/language/tree"
	languagetypes "github.com/bearer/curio/new/language/types"
)

type Data struct {
	Datatypes []*types.Detection
}

type Filter struct {
	Variable  string
	Values    []string
	Detection string
}

type Rule struct {
	Pattern string
	Filters []Filter
}

type customDetector struct {
	detectorType string
	patternQuery languagetypes.PatternQuery
	filters      []Filter
}

func New(lang languagetypes.Language, detectorType string, rule Rule) (types.Detector, error) {
	patternQuery, err := lang.CompilePatternQuery(rule.Pattern)
	if err != nil {
		return nil, fmt.Errorf("error compiling pattern: %s", err)
	}

	// TODO: validate filters against pattern

	return &customDetector{
		detectorType: detectorType,
		patternQuery: patternQuery,
		filters:      rule.Filters,
	}, nil
}

func (detector *customDetector) Name() string {
	return detector.detectorType
}

func (detector *customDetector) DetectAt(
	node *tree.Node,
	evaluator types.Evaluator,
) ([]*types.Detection, error) {
	results, err := detector.patternQuery.MatchAt(node)
	if err != nil {
		return nil, err
	}

	var detections []*types.Detection

	for _, result := range results {
		filtersMatch, datatypeDetections, err := detector.matchFilters(result, evaluator)
		if err != nil {
			return nil, err
		}

		if !filtersMatch {
			continue
		}

		detections = append(detections, &types.Detection{
			MatchNode: node,
			Data: Data{
				Datatypes: datatypeDetections,
			},
		})
	}

	return detections, nil
}

func (detector *customDetector) matchFilters(
	result tree.QueryResult,
	evaluator types.Evaluator,
) (bool, []*types.Detection, error) {
	var datatypeDetections []*types.Detection

	for _, filter := range detector.filters {
		node, ok := result[filter.Variable]
		// shouldn't happen if filters are validated against pattern
		if !ok {
			return false, nil, nil
		}

		if len(filter.Values) != 0 && !slices.Contains(filter.Values, node.Content()) {
			return false, nil, nil
		}

		if filter.Detection == "datatype" {
			filterDetections, err := evaluator.ForTree(node, "datatype")
			if err != nil {
				return false, nil, err
			}

			datatypeDetections = append(datatypeDetections, filterDetections...)
		} else if filter.Detection != "" {
			hasDetection, err := evaluator.TreeHas(node, filter.Detection)
			if err != nil || !hasDetection {
				return false, nil, err
			}
		}
	}

	return true, datatypeDetections, nil
}

func (detector *customDetector) Close() {
	detector.patternQuery.Close()
}
