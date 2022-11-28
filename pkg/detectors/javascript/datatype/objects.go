package datatype

import (
	"strings"

	"github.com/bearer/curio/pkg/parser"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"

	"github.com/bearer/curio/pkg/report/schema"
	"github.com/smacker/go-tree-sitter/javascript"
)

var objectsQuery = parser.QueryMustCompile(javascript.GetLanguage(),
	`(object
	(pair) @param_pair
) @param_object`)
var nestedObjectsQuery = parser.QueryMustCompile(javascript.GetLanguage(),
	`(object
	(pair
		value: (object) @param_child_object
	) @param_pair
) @param_object`)

func addObjects(tree *parser.Tree, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	// add object and properties
	captures := tree.QueryConventional(objectsQuery)
	for _, capture := range captures {
		objectNode := capture["param_object"]

		propertyNode := capture["param_pair"].ChildByFieldName("key")

		if propertyNode == nil {
			continue
		}

		propertyName := propertyNode.Content()

		// create common object
		if datatypes[objectNode.ID()] == nil {
			datatypes[objectNode.ID()] = &schemadatatype.DataType{
				Node:       objectNode,
				Name:       "",
				Type:       schema.SimpleTypeUnknown,
				TextType:   "",
				Properties: make(map[string]schemadatatype.DataTypable),
			}
		}

		// add property
		datatypes[objectNode.ID()].Properties[propertyName] = &schemadatatype.DataType{
			Node:       propertyNode,
			Name:       propertyNode.Content(),
			Type:       schema.SimpleTypeUnknown,
			TextType:   "",
			Properties: make(map[string]schemadatatype.DataTypable),
		}
	}

	// helperTypes := make(map[parser.NodeID]*schemadatatype.DataType)

	var keysToDelete []parser.NodeID

	// link nested objects to their properties
	captures = tree.QueryConventional(nestedObjectsQuery)
	for _, capture := range captures {
		propertyValueNode := capture["param_child_object"]
		objectNode := capture["param_object"]

		propertyNode := capture["param_pair"].ChildByFieldName("key")
		if propertyNode == nil {
			continue
		}

		propertyName := propertyNode.Content()

		// sometimes objects only have function names in that case we should ignore them
		_, hasPropertyValue := datatypes[propertyValueNode.ID()]
		if !hasPropertyValue {
			continue
		}
		_, hasObjectValue := datatypes[objectNode.ID()]
		if !hasObjectValue {
			continue
		}

		// link property
		datatypes[objectNode.ID()].Properties[propertyName] = datatypes[propertyValueNode.ID()]

		// update property name
		datatypes[objectNode.ID()].Properties[propertyName].SetName(propertyName)

		// mark root node key for deletion
		keysToDelete = append(keysToDelete, propertyValueNode.ID())
	}

	for _, v := range keysToDelete {
		delete(datatypes, v)
	}

	// add root object names
	for _, datatype := range datatypes {
		node := datatype.Node

		parent := node.Parent()

		if parent.Type() == "assignment_expression" {
			left := parent.ChildByFieldName("left")

			if left.Type() == "member_expression" {
				property := left.ChildByFieldName("property")
				datatype.Name = property.Content()
				continue
			}

			if left.Type() == "identifier" {
				datatype.Name = left.Content()
				continue
			}

			if left.Type() == "subscript_expression" {
				index := left.ChildByFieldName("index")

				if index != nil && index.Type() == "member_expression" {
					property := index.ChildByFieldName("property")

					if property != nil && property.Type() == "property_identifier" {
						datatype.Name = property.Content()
						continue
					}
				}
			}
		}

		if parent.Type() == "variable_declarator" {
			identifier := parent.ChildByFieldName("name")
			datatype.Name = identifier.Content()
			continue
		}

		if parent.Type() == "arguments" {
			grandparent := parent.Parent()

			if grandparent.Type() == "subscript_expression" {
				index := grandparent.ChildByFieldName("index")

				if index == nil {
					continue
				}

				if index.Type() == "string" || index.Type() == "number" {
					datatype.Name = index.Content()
					continue
				}
			}

			if grandparent.Type() == "call_expression" {
				function := grandparent.ChildByFieldName("function")

				if function.Type() == "identifier" {
					datatype.Name = function.Content()
					continue
				}

				if function.Type() == "member_expression" {
					property := function.ChildByFieldName("property")
					datatype.Name = property.Content()
					continue
				}
			}

		}

		if parent.Type() == "return_statement" {
			grandparent := parent.Parent()

			if grandparent.Type() == "statement_block" {
				grandGrandParent := grandparent.Parent()

				if grandGrandParent.Type() == "function_declaration" {
					identifier := grandGrandParent.ChildByFieldName("name")

					if identifier != nil {
						datatype.Name = identifier.Content()
						continue
					}
				}

				// fallback to assigment expression
				assigmentExpression := grandGrandParent.Parent()
				if assigmentExpression != nil && objectAssigment(datatype, assigmentExpression) {
					continue
				}
			}
		}

		if parent.Type() == "parenthesized_expression" {
			grandparent := parent.Parent()
			if grandparent != nil && grandparent.Type() == "arrow_function" {
				grandGrandParent := grandparent.Parent()

				if grandGrandParent == nil {
					continue
				}

				if grandGrandParent.Type() == "variable_declarator" {
					name := grandGrandParent.ChildByFieldName("name")

					if name != nil && name.Type() == "identifier" {
						datatype.Name = name.Content()
						continue
					}
				}

				if grandGrandParent.Type() == "assignment_expression" {
					if objectAssigment(datatype, grandGrandParent) {
						continue
					}
				}
			}
		}
	}
}

func objectAssigment(datatype *schemadatatype.DataType, node *parser.Node) bool {
	left := node.ChildByFieldName("left")

	if left == nil {
		return true
	}

	if left.Type() == "member_expression" {
		property := left.ChildByFieldName("property")

		if property != nil && property.Type() == "property_identifier" {
			datatype.Name = property.Content()
			return true
		}
	}

	if left.Type() == "subscript_expression" {
		index := left.ChildByFieldName("index")

		if index == nil {
			return true
		}

		if index.Type() == "string" {
			content := index.Content()
			content = strings.Trim(content, `"`)
			content = strings.Trim(content, `'`)
			content = strings.Trim(content, "`")
			datatype.Name = content
			return true
		}

		if index.Type() == "number" {
			datatype.Name = index.Content()
			return true
		}
	}

	return false
}
