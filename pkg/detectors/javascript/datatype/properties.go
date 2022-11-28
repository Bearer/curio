package datatype

import (
	"github.com/bearer/curio/pkg/parser"
	parserdatatype "github.com/bearer/curio/pkg/parser/datatype"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"

	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report/schema"
	"github.com/smacker/go-tree-sitter/javascript"
)

var nestedPropertiesQuery = parser.QueryMustCompile(javascript.GetLanguage(),
	`(member_expression
property: (property_identifier) @param_property
)@param_object`)

var rootPropertiesQuery = parser.QueryMustCompile(javascript.GetLanguage(),
	`(member_expression
		object: (identifier) @param_property
	)`)

var rootThisPropertiesQuery = parser.QueryMustCompile(javascript.GetLanguage(),
	`(member_expression
	object: (this) @param_property
)`)

func addProperties(tree *parser.Tree, helperDatatypes map[parser.NodeID]*schemadatatype.DataType) {
	// add root propertie
	captures := tree.QueryConventional(rootPropertiesQuery)
	for _, capture := range captures {
		rootPropertyNode := capture["param_property"]
		id := capture["param_property"].Content()
		helperDatatypes[rootPropertyNode.ID()] = &schemadatatype.DataType{
			Node:       rootPropertyNode,
			Name:       id,
			Type:       schema.SimpleTypeUnknown,
			TextType:   "",
			Properties: make(map[string]schemadatatype.DataTypable),
			UUID:       "",
		}
	}

	// add this root properties
	captures = tree.QueryConventional(rootThisPropertiesQuery)
	for _, capture := range captures {
		rootPropertyNode := capture["param_property"]
		id := capture["param_property"].Content()
		helperDatatypes[rootPropertyNode.ID()] = &schemadatatype.DataType{
			Node:       rootPropertyNode,
			Name:       id,
			Type:       schema.SimpleTypeUnknown,
			TextType:   "",
			Properties: make(map[string]schemadatatype.DataTypable),
			UUID:       "",
		}
	}

	// add properties-
	captures = tree.QueryConventional(nestedPropertiesQuery)
	for _, capture := range captures {
		objectNode := capture["param_object"]
		propertyNode := capture["param_property"]
		id := capture["param_property"].Content()
		helperDatatypes[objectNode.ID()] = &schemadatatype.DataType{
			Node:       propertyNode,
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
		// add root node
		if node.Type() == "identifier" {
			datatypes[node.ID()] = helperType
			continue
		}

		if node.Type() == "this" {
			datatypes[node.ID()] = helperType
			continue
		}

		parent := node.Parent()
		if parent.Type() == "member_expression" {
			// link to root node
			object := parent.ChildByFieldName("object")
			if object != nil && (object.Type() == "identifier" || object.Type() == "this") {
				helperDatatypes[object.ID()].Properties[helperType.Name] = helperType
				continue
			}

			// link to chain
			if object != nil && object.Type() == "member_expression" {
				helperDatatypes[object.ID()].Properties[helperType.Name] = helperType
				continue
			}
		}

		// link to root document
		datatypes[node.ID()] = helperType
	}
}

func scopeProperties(datatypes map[parser.NodeID]*schemadatatype.DataType, idGenerator nodeid.Generator) {
	terminatorKeywords := []string{"program", "function", "arrow_function", "method_definition"}
	// pull all scope terminator nodes
	parserdatatype.ScopeDatatypes(datatypes, idGenerator, terminatorKeywords)
}
