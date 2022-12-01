package customdetector

import (
	"strings"

	"github.com/bearer/curio/pkg/detectors/ruby/datatype"
	"github.com/bearer/curio/pkg/parser"
	parserdatatype "github.com/bearer/curio/pkg/parser/datatype"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report/schema"
	"github.com/rs/zerolog/log"

	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"
)

func (detector *Detector) ExtractArguments(node *parser.Node, idGenerator nodeid.Generator, variableReconciliation *parserdatatype.ReconciliationRequest) (map[parser.NodeID]*schemadatatype.DataType, error) {
	extractedDatatypes, err := detector.extractArguments(node, idGenerator)
	if err != nil {
		return nil, err
	}

	for _, dt := range extractedDatatypes {
		log.Error().Msgf("extract args dt: %#v", dt)
	}

	if variableReconciliation != nil {
		parserdatatype.VariableReconciliation(extractedDatatypes, variableReconciliation)
	}

	for _, dt := range extractedDatatypes {
		log.Error().Msgf("extract args reconciled dt: %#v", dt)
	}

	return extractedDatatypes, nil
}

func (detector *Detector) extractArguments(node *parser.Node, idGenerator nodeid.Generator) (map[parser.NodeID]*schemadatatype.DataType, error) {
	if node == nil {
		return nil, nil
	}

	joinedDatatypes := make(map[parser.NodeID]*schemadatatype.DataType)

	// handle class name
	if node.Type() == "constant" {
		datatype := &schemadatatype.DataType{
			Node:       node,
			Name:       node.Content(),
			Type:       schema.SimpleTypeObject,
			Properties: make(map[string]schemadatatype.DataTypable),
		}
		joinedDatatypes[datatype.Node.ID()] = datatype
		return joinedDatatypes, nil
	}

	for i := 0; i < node.ChildCount(); i++ {
		singleArgument := node.Child(i)

		if singleArgument.Type() == "identifier" || singleArgument.Type() == "simple_symbol" || singleArgument.Type() == "bare_symbol" {
			content := singleArgument.Content()

			if singleArgument.Type() == "simple_symbol" {
				content = strings.TrimLeft(content, ":")
			}

			datatype := &schemadatatype.DataType{
				Node:       singleArgument,
				Name:       content,
				Type:       schema.SimpleTypeUnknown,
				Properties: make(map[string]schemadatatype.DataTypable),
			}
			joinedDatatypes[datatype.Node.ID()] = datatype
			continue
		}
	}

	complexDatatypes := datatype.Discover(node, idGenerator)
	for nodeID, target := range complexDatatypes {
		_, exists := joinedDatatypes[nodeID]
		if exists {
			continue
		}

		joinedDatatypes[nodeID] = target
	}

	return joinedDatatypes, nil
}
