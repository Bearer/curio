package object

import (
	"fmt"

	"github.com/bearer/curio/new/detector/types"
	"github.com/bearer/curio/new/language/tree"
	languagetypes "github.com/bearer/curio/new/language/types"
)

type Data struct {
	Name       string
	Properties []*types.Detection
}

type objectDetector struct {
	// Gathering properties
	hashPairQuery *tree.Query
	// Naming
	assignmentQuery *tree.Query
	parentPairQuery *tree.Query
}

func New(lang languagetypes.Language) (types.Detector, error) {
	// { first_name: ..., ... }
	hashPairQuery, err := lang.CompileQuery(`(hash (pair) @pair) @root`)
	if err != nil {
		return nil, fmt.Errorf("error compiling hash pair query: %s", err)
	}

	// user = <object>
	assignmentQuery, err := lang.CompileQuery(`(assignment left: (identifier) @left right: (_) @right) @root`)
	if err != nil {
		return nil, fmt.Errorf("error compiling assignment query: %s", err)
	}
	// { user: <object> }
	parentPairQuery, err := lang.CompileQuery(`(pair key: (hash_key_symbol) @key value: (_) @value) @root`)
	if err != nil {
		return nil, fmt.Errorf("error compiling parent pair query: %s", err)
	}

	return &objectDetector{
		hashPairQuery:   hashPairQuery,
		assignmentQuery: assignmentQuery,
		parentPairQuery: parentPairQuery,
	}, nil
}

func (detector *objectDetector) Name() string {
	return "object"
}

func (detector *objectDetector) DetectAt(
	node *tree.Node,
	evaluator types.Evaluator,
) ([]*types.Detection, error) {
	detections, err := detector.gatherProperties(node, evaluator)
	if len(detections) != 0 || err != nil {
		return detections, err
	}

	detections, err = detector.nameAssignedObject(node, evaluator)
	if len(detections) != 0 || err != nil {
		return detections, err
	}

	return detector.nameParentPairObject(node, evaluator)
}

func (detector *objectDetector) gatherProperties(
	node *tree.Node,
	evaluator types.Evaluator,
) ([]*types.Detection, error) {
	results, err := detector.hashPairQuery.MatchAt(node)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	var properties []*types.Detection
	for _, result := range results {
		nodeProperties, err := evaluator.ForNode(result["pair"], "property")
		if err != nil {
			return nil, err
		}

		properties = append(properties, nodeProperties...)
	}

	return []*types.Detection{{
		MatchNode: node,
		Data:      Data{Properties: properties},
	}}, nil
}

func (detector *objectDetector) nameAssignedObject(
	node *tree.Node,
	evaluator types.Evaluator,
) ([]*types.Detection, error) {
	result, err := detector.assignmentQuery.MatchOnceAt(node)
	if result == nil || err != nil {
		return nil, err
	}

	objects, err := evaluator.ForNode(result["right"], "object")
	if err != nil {
		return nil, err
	}

	var detections []*types.Detection
	for _, object := range objects {
		objectData := object.Data.(Data)

		if objectData.Name == "" {
			detections = append(detections, &types.Detection{
				MatchNode: node,
				Data: Data{
					Name:       result["left"].Content(),
					Properties: objectData.Properties,
				},
			})
		} else { // FIXME: should we remove this case?
			detections = append(detections, &types.Detection{
				MatchNode: node,
				Data: Data{
					Name: result["left"].Content(),
					Properties: []*types.Detection{{
						MatchNode: object.MatchNode,
						Data: Data{
							Name: objectData.Name,
						},
					}},
				},
			})
		}
	}

	return detections, nil
}

func (detector *objectDetector) nameParentPairObject(
	node *tree.Node,
	evaluator types.Evaluator,
) ([]*types.Detection, error) {
	result, err := detector.parentPairQuery.MatchOnceAt(node)
	if result == nil || err != nil {
		return nil, err
	}

	objects, err := evaluator.ForNode(result["value"], "object")
	if err != nil {
		return nil, err
	}

	var detections []*types.Detection
	for _, object := range objects {
		objectData := object.Data.(Data)

		detections = append(detections, &types.Detection{
			MatchNode: node,
			Data: Data{
				Name:       result["key"].Content(),
				Properties: objectData.Properties,
			},
		})
	}

	return detections, nil
}

func (detector *objectDetector) Close() {
	detector.hashPairQuery.Close()
	detector.assignmentQuery.Close()
	detector.parentPairQuery.Close()
}
