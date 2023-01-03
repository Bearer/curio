package set

import (
	"fmt"

	"github.com/bearer/curio/new/detector/types"
	"github.com/bearer/curio/new/language/tree"
)

type set struct {
	detectors map[string]types.Detector
}

func New(detectors []types.Detector) (types.DetectorSet, error) {
	detectorMap := make(map[string]types.Detector)

	for _, detector := range detectors {
		name := detector.Name()

		if _, existing := detectorMap[name]; existing {
			return nil, fmt.Errorf("duplicate detector '%s'", name)
		}

		detectorMap[name] = detector
	}

	return &set{
		detectors: detectorMap,
	}, nil
}

func (set *set) DetectAt(
	node *tree.Node,
	detectorType string,
	evaluator types.Evaluator,
) ([]*types.Detection, error) {
	detector, ok := set.detectors[detectorType]
	if !ok {
		return nil, fmt.Errorf("detector type '%s' not registered", detectorType)
	}

	return detector.DetectAt(node, evaluator)
}
