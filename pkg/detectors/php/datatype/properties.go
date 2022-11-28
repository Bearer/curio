package datatype

import (
	"github.com/bearer/curio/pkg/parser"
	parserdatatype "github.com/bearer/curio/pkg/parser/datatype"
	"github.com/bearer/curio/pkg/parser/nodeid"
	php "github.com/bearer/curio/pkg/parser/sitter/php2"
	"github.com/bearer/curio/pkg/report/schema"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"
	sitter "github.com/smacker/go-tree-sitter"
)

var propertiesQuery = parser.QueryMustCompile(php.GetLanguage(),
	`(member_access_expression
		object: (_) @param_property
	)@param_object`)

var propertiesFunctionsQuery = parser.QueryMustCompile(php.GetLanguage(),
	`(member_call_expression
		object: (_) @param_property
	)@param_object`)

func addProperties(tree *parser.Tree, helperDatatypes map[parser.NodeID]*schemadatatype.DataType) {
	// add root propertie
	var parseQuery = func(query *sitter.Query) {
		captures := tree.QueryConventional(query)
		for _, capture := range captures {
			propertyNode := capture["param_property"]
			if propertyNode.Type() == "variable_name" {
				id := propertyNode.Child(0).Content()
				helperDatatypes[propertyNode.ID()] = &schemadatatype.DataType{
					Node:       propertyNode,
					Name:       id,
					Type:       schema.SimpleTypeUnknown,
					TextType:   "",
					Properties: make(map[string]schemadatatype.DataTypable),
					UUID:       "",
				}
			}
			objectNode := capture["param_object"]
			idNode := objectNode.ChildByFieldName("name")
			if idNode == nil {
				continue
			}

			if idNode.Type() == "variable_name" {
				idNode = idNode.Child(0)
			}

			id := idNode.Content()

			helperDatatypes[objectNode.ID()] = &schemadatatype.DataType{
				Node:       objectNode,
				Name:       id,
				Type:       schema.SimpleTypeUnknown,
				TextType:   "",
				Properties: make(map[string]schemadatatype.DataTypable),
				UUID:       "",
			}
		}
	}

	parseQuery(propertiesQuery)
	parseQuery(propertiesFunctionsQuery)
}

func linkProperties(tree *parser.Tree, datatypes, helperDatatypes map[parser.NodeID]*schemadatatype.DataType) {
	for _, helperType := range helperDatatypes {
		node := helperType.Node

		// add root node
		if node.Type() == "variable_name" {
			datatypes[node.ID()] = helperType
			continue
		} else {
			// set node to name which has column number
			helperType.Node = node.ChildByFieldName("name")
		}

		// chain proprety access
		object := node.ChildByFieldName("object")

		if object.Type() == "member_access_expression" || object.Type() == "variable_name" {
			helperDatatypes[object.ID()].Properties[helperType.Name] = helperType
			continue
		}

		// link to root document
		datatypes[node.ID()] = helperType

	}
}

func scopeAndMergeProperties(propertiesDatatypes, classDataTypes map[parser.NodeID]*schemadatatype.DataType, idGenerator nodeid.Generator) {
	// replace this with class name
	for key, datatype := range propertiesDatatypes {
		if datatype.Name == "this" {
			class, err := datatype.Node.FindParent("class_declaration")
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
	terminatorKeywords := []string{"program"}
	parserdatatype.ScopeDatatypes(classDataTypes, idGenerator, terminatorKeywords)

	// scoped data
	terminatorKeywords = []string{"program", "method_declaration", "anonymous_function_creation_expression", "function_definition"}
	// pull all scope terminator nodes
	parserdatatype.ScopeDatatypes(propertiesDatatypes, idGenerator, terminatorKeywords)
}
