package datatype

import (
	objectdetector "github.com/bearer/curio/new/detector/implementation/ruby/object"
	propertydetector "github.com/bearer/curio/new/detector/implementation/ruby/property"
	"github.com/bearer/curio/new/detector/types"
	"github.com/bearer/curio/new/language/tree"
	languagetypes "github.com/bearer/curio/new/language/types"
)

type Data struct {
	Name string
}

type datatypeDetector struct{}

func New(lang languagetypes.Language) (types.Detector, error) {
	return &datatypeDetector{}, nil
}

func (detector *datatypeDetector) Name() string {
	return "datatype"
}

func (detector *datatypeDetector) DetectAt(
	node *tree.Node,
	evaluator types.Evaluator,
) ([]*types.Detection, error) {
	objectDetections, err := evaluator.ForNode(node, "object")
	if err != nil {
		return nil, err
	}

	var result []*types.Detection

	for _, object := range objectDetections {
		data := object.Data.(objectdetector.Data)
		if data.Name != "user" {
			continue
		}

		for _, property := range data.Properties {
			propertyData := property.Data.(propertydetector.Data)
			if propertyData.Name == "first_name" {
				result = append(result, &types.Detection{
					ContextNode: node,
					MatchNode:   property.MatchNode,
					Data:        Data{Name: "Person Name"},
				})
			}
		}
	}

	return result, nil
}

func (detector *datatypeDetector) Close() {}
