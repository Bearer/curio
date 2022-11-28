package datatype

import (
	"strings"

	"github.com/bearer/curio/pkg/parser"
	parserdatatype "github.com/bearer/curio/pkg/parser/datatype"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report/schema"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"
	"github.com/smacker/go-tree-sitter/python"
)

// person.city
var attributeQuery = parser.QueryMustCompile(python.GetLanguage(),
	`(attribute
		object: (_) @param_child
		attribute: (identifier) @param_id
	  ) @param_parent`)

var subscriptQuery = parser.QueryMustCompile(python.GetLanguage(),
	`(subscript
		value: (_) @param_child
		subscript: (string) @param_id
	  ) @param_parent`)

func addProperties(tree *parser.Tree, helperDatatypes map[parser.NodeID]*schemadatatype.DataType) {
	// add element references
	captures := tree.QueryConventional(attributeQuery)
	for _, capture := range captures {
		childNode := capture["param_child"]
		if childNode.Type() == "identifier" {
			id := childNode.Content()
			helperDatatypes[childNode.ID()] = &schemadatatype.DataType{
				Node:       childNode,
				Name:       id,
				Type:       schema.SimpleTypeUnknown,
				TextType:   "",
				Properties: make(map[string]schemadatatype.DataTypable),
				UUID:       "",
			}
		}

		elementNode := capture["param_parent"]
		idNode := capture["param_id"]

		helperDatatypes[elementNode.ID()] = &schemadatatype.DataType{
			Node:       idNode,
			Name:       idNode.Content(),
			Type:       schema.SimpleTypeUnknown,
			TextType:   "",
			Properties: make(map[string]schemadatatype.DataTypable),
			UUID:       "",
		}
	}

	// add calls
	captures = tree.QueryConventional(subscriptQuery)
	for _, capture := range captures {
		childNode := capture["param_child"]
		if childNode.Type() == "identifier" {
			id := childNode.Content()
			helperDatatypes[childNode.ID()] = &schemadatatype.DataType{
				Node:       childNode,
				Name:       id,
				Type:       schema.SimpleTypeUnknown,
				TextType:   "",
				Properties: make(map[string]schemadatatype.DataTypable),
				UUID:       "",
			}
		}

		elementNode := capture["param_parent"]
		idNode := capture["param_id"]
		id := strings.Trim(idNode.Content(), "'")
		id = strings.Trim(id, `"`)

		helperDatatypes[elementNode.ID()] = &schemadatatype.DataType{
			Node:       idNode,
			Name:       id,
			Type:       schema.SimpleTypeUnknown,
			TextType:   "",
			Properties: make(map[string]schemadatatype.DataTypable),
			UUID:       "",
		}
	}
}

func linkProperties(tree *parser.Tree, datatypes, helperDatatypes map[parser.NodeID]*schemadatatype.DataType) {
	for _, helperType := range helperDatatypes {
		node := helperType.Node
		parent := node.Parent()
		// add root node

		if parent.Type() == "attribute" {
			object := parent.ChildByFieldName("object")

			// add root attributes
			if object != nil && object.ID() == node.ID() {
				datatypes[node.ID()] = helperType
				continue
			}

			// link child attribtue chains
			if object.Type() == "attribute" || object.Type() == "subscript" || object.Type() == "identifier" {
				_, hasID := helperDatatypes[object.ID()]
				if hasID {
					helperDatatypes[object.ID()].Properties[helperType.Name] = helperType
					continue
				}
			}
		}

		if parent.Type() == "subscript" {
			value := parent.ChildByFieldName("value")

			// add root attributes
			if value.ID() == node.ID() {
				datatypes[node.ID()] = helperType
				continue
			}

			// link child subscript chains
			if value.Type() == "attribute" || value.Type() == "subscript" || value.Type() == "identifier" {
				_, hasID := helperDatatypes[value.ID()]
				if hasID {
					helperDatatypes[value.ID()].Properties[helperType.Name] = helperType
					continue
				}
			}
		}

		// link to root document
		datatypes[node.ID()] = helperType

	}
}

func scopeAndMergeProperties(propertiesDatatypes, classDataTypes map[parser.NodeID]*schemadatatype.DataType, idGenerator nodeid.Generator) {
	// self with class name
	for key, datatype := range propertiesDatatypes {
		if datatype.Name == "self" {
			class, err := datatype.Node.FindParent("class_definition")
			if err != nil {
				continue
			}

			nameNode := class.ChildByFieldName("name")
			if nameNode == nil {
				continue
			}

			datatype.Name = nameNode.Content()
			classDataTypes[datatype.Node.ID()] = datatype
			delete(propertiesDatatypes, key)
		}
	}
	// scope replaced properties
	terminatorKeywords := []string{"module"}
	parserdatatype.ScopeDatatypes(classDataTypes, idGenerator, terminatorKeywords)

	// // scoped data
	terminatorKeywords = []string{"module", "function_definition", "lambda"}
	// pull all scope terminator nodes
	parserdatatype.ScopeDatatypes(propertiesDatatypes, idGenerator, terminatorKeywords)
}
