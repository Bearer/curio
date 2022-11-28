package customdetector

import (
	"strings"

	"github.com/bearer/curio/pkg/detectors/ruby/datatype"
	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report/schema"

	parserdatatype "github.com/bearer/curio/pkg/parser/datatype"

	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"
	"github.com/bearer/curio/pkg/util/file"
)

func (detector *Detector) ExtractArguments(node *parser.Node, idGenerator nodeid.Generator, fileinfo *file.FileInfo, filepath *file.Path) (map[parser.NodeID]*schemadatatype.DataType, error) {
	extractedDatatypes, err := detector.extractArguments(node, idGenerator, fileinfo, filepath)
	if err != nil {
		return nil, err
	}

	allDatatypes := datatype.Discover(node.Tree().RootNode(), idGenerator)
	parserdatatype.VariableReconciliation(extractedDatatypes, allDatatypes, datatype.ScopeTerminators)

	return extractedDatatypes, nil
}

func (detector *Detector) extractArguments(node *parser.Node, idGenerator nodeid.Generator, fileinfo *file.FileInfo, filepath *file.Path) (map[parser.NodeID]*schemadatatype.DataType, error) {
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

		singleArgumentDatatypes := datatype.Discover(singleArgument, idGenerator)

		for nodeID, target := range singleArgumentDatatypes {
			joinedDatatypes[nodeID] = target
		}

	}

	return joinedDatatypes, nil
}
