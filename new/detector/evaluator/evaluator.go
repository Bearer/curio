package evaluator

import (
	"fmt"
	"strings"

	"github.com/bearer/curio/new/detector/types"
	"github.com/bearer/curio/new/language/tree"
	langtree "github.com/bearer/curio/new/language/tree"
	languagetypes "github.com/bearer/curio/new/language/types"
	"golang.org/x/exp/slices"
)

type evaluator struct {
	lang               languagetypes.Language
	detectorSet        types.DetectorSet
	detectionCache     map[langtree.NodeID]map[string][]*types.Detection
	executingDetectors map[langtree.NodeID][]string
}

func New(
	lang languagetypes.Language,
	detectorSet types.DetectorSet,
	tree *langtree.Tree,
) types.Evaluator {
	detectionCache := make(map[langtree.NodeID]map[string][]*types.Detection)

	return &evaluator{
		lang:               lang,
		detectorSet:        detectorSet,
		detectionCache:     detectionCache,
		executingDetectors: make(map[langtree.NodeID][]string),
	}
}

func (evaluator *evaluator) ForTree(rootNode *langtree.Node, detectorType string) ([]*types.Detection, error) {
	var result []*types.Detection

	if err := rootNode.Walk(func(node *langtree.Node) error {
		detections, err := evaluator.nonUnifiedNodeDetections(node, detectorType)
		if err != nil {
			return err
		}

		result = append(result, detections...)

		for _, unifiedNode := range node.UnifiedNodes() {
			unifiedNodeDetections, err := evaluator.ForTree(unifiedNode, detectorType)
			if err != nil {
				return err
			}

			result = append(result, unifiedNodeDetections...)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (evaluator *evaluator) ForNode(
	node *langtree.Node,
	detectorType string,
) ([]*types.Detection, error) {
	detections, err := evaluator.nonUnifiedNodeDetections(node, detectorType)
	if err != nil {
		return nil, err
	}

	for _, unifiedNode := range node.UnifiedNodes() {
		unifiedNodeDetections, err := evaluator.nonUnifiedNodeDetections(unifiedNode, detectorType)
		if err != nil {
			return nil, err
		}

		detections = append(detections, unifiedNodeDetections...)
	}

	return detections, nil
}

func (evaluator *evaluator) nonUnifiedNodeDetections(
	node *langtree.Node,
	detectorType string,
) ([]*types.Detection, error) {
	nodeDetections, ok := evaluator.detectionCache[node.ID()]
	if !ok {
		err := evaluator.detectAtNode(node, detectorType)
		if err != nil {
			return nil, err
		}

		nodeDetections = evaluator.detectionCache[node.ID()]
	}

	if detections, ok := nodeDetections[detectorType]; ok {
		return detections, nil
	}

	evaluator.detectAtNode(node, detectorType)
	return nodeDetections[detectorType], nil
}

func (evaluator *evaluator) TreeHas(rootNode *langtree.Node, detectorType string) (bool, error) {
	hasDetection := false

	if err := rootNode.Walk(func(node *langtree.Node) error {
		var err error
		hasDetection, err = evaluator.NodeHas(node, detectorType)
		if err != nil {
			return err
		}

		if hasDetection {
			return langtree.ErrTerminateWalk
		}

		return nil
	}); err != nil {
		return false, err
	}

	return hasDetection, nil
}

func (evaluator *evaluator) NodeHas(node *langtree.Node, detectorType string) (bool, error) {
	detections, err := evaluator.ForNode(node, detectorType)
	if err != nil {
		return false, err
	}

	return len(detections) != 0, nil
}

func (evaluator *evaluator) detectAtNode(node *langtree.Node, detectorType string) error {
	return evaluator.withCycleProtection(node, detectorType, func() error {
		detections, err := evaluator.detectorSet.DetectAt(node, detectorType, evaluator)
		if err != nil {
			return err
		}

		nodeDetections, ok := evaluator.detectionCache[node.ID()]
		if !ok {
			nodeDetections = make(map[string][]*types.Detection)
			evaluator.detectionCache[node.ID()] = nodeDetections
		}

		nodeDetections[detectorType] = detections

		return nil
	})
}

func (evaluator *evaluator) withCycleProtection(node *tree.Node, detectorType string, body func() error) error {
	nodeID := node.ID()

	executingDetectors := evaluator.executingDetectors[nodeID]
	if slices.Contains(evaluator.executingDetectors[nodeID], detectorType) {
		return fmt.Errorf(
			"cycle found in detector usage: [%s > %s]\nnode type: %s, content:\n%s",
			strings.Join(executingDetectors, " > "),
			detectorType,
			node.Type(),
			node.Content(),
		)
	}

	evaluator.executingDetectors[nodeID] = append(evaluator.executingDetectors[nodeID], detectorType)

	if err := body(); err != nil {
		return err
	}

	if len(evaluator.executingDetectors[nodeID]) == 1 {
		delete(evaluator.executingDetectors, nodeID)
	} else {
		executingDetectors := evaluator.executingDetectors[nodeID]
		evaluator.executingDetectors[nodeID] = executingDetectors[:len(executingDetectors)-1]
	}

	return nil
}
