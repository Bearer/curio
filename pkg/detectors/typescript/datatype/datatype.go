package datatype

import (
	"strings"

	"github.com/bearer/curio/pkg/detectors/typescript/datatype/knex"
	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/datatype"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report"
	"github.com/bearer/curio/pkg/report/detections"
	"github.com/bearer/curio/pkg/report/detectors"
	"github.com/bearer/curio/pkg/report/schema"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"

	sitter "github.com/smacker/go-tree-sitter"
)

func Discover(report report.Report, tree *parser.Tree, language *sitter.Language, idGenerator nodeid.Generator) {
	datatypesFinder := datatype.NewFinder(tree, annotateDataTypes)
	datatypesFinder.Find()

	propertyFinder := datatype.NewPropertyFinder(tree, datatypesFinder.GetValues(), annotateDataTypeProperties)
	propertyFinder.Find()

	knex.Discover(report, tree, language)

	datatype.PruneMap(datatypesFinder.GetValues())

	report.AddDataType(detections.TypeSchema, detectors.DetectorTypescript, idGenerator, datatypesFinder.GetValues(), nil)
}

func annotateDataTypes(finder *datatype.Finder, node *parser.Node, value *schemadatatype.DataType) bool {
	value.Node = node

	switch node.Type() {
	case "object_type":
		parent := node.Parent()
		if parent != nil {
			// when any of these is defined parent will take care of declaring datatype
			if parent.Type() == "interface_declaration" || parent.Type() == "type_alias_declaration" || parent.Type() == "class_declaration" || parent.Type() == "class" {
				return false
			}

			// when granparent is parameter
			grandparent := parent.Parent()
			if grandparent != nil {
				if parent.Parent().Type() == "required_parameter" || parent.Parent().Type() == "optional_parameter" {
					return false
				}
			}
		}

		// when all object has is defining generic properties
		// tags: { [key: string]: string }
		hasNonIndexSignature := false
		for i := 0; i < node.ChildCount(); i++ {
			if node.Child(i).Type() != "index_signature" {
				hasNonIndexSignature = true
				break
			}
		}
		if !hasNonIndexSignature {
			value.IsHelper = true
		}

		value.Name = ""

		value.TextType = ""
		value.Type = schema.SimpleTypeObject
		return true
	case "interface_declaration":
		name := node.ChildByFieldName("name")

		value.TextType = "interface"
		value.Type = schema.SimpleTypeObject
		value.Name = name.Content()
		return true
	case "type_alias_declaration":
		name := node.ChildByFieldName("name")
		value.Name = name.Content()
		value.TextType = "type"

		valueNode := node.ChildByFieldName("value")
		if valueNode != nil {
			if valueNode.Type() == "literal_type" {
				valueNodeChild := valueNode.Child(0)
				value.Type = standardizeDataType(valueNodeChild, valueNodeChild.Type())
				return true
			}
		}

		value.Type = schema.SimpleTypeUnknown
		return true
	case "class_declaration":
		name := node.ChildByFieldName("name")

		value.TextType = "class"
		value.Type = schema.SimpleTypeObject

		if name == nil {
			value.Name = ""
		} else {
			value.Name = name.Content()
		}

		return true
	case "class":
		name := node.ChildByFieldName("name")

		value.TextType = "class"
		value.Type = schema.SimpleTypeObject

		if name == nil {
			value.Name = ""
		} else {
			value.Name = name.Content()
		}

		return true
	case "required_parameter":
		name := node.ChildByFieldName("pattern")
		valueType := node.ChildByFieldName("type")

		if valueType == nil {
			name = node.Child(0)
			valueType = node.Child(1)
			if name == nil || name.Type() != "identifier" {
				return false
			}
			if valueType == nil {
				return false
			}
		}

		value.TextType = standardizeTextType(valueType)
		value.Type = standardizeDataType(valueType, valueType.Content())
		value.Name = name.Content()
		return true
	case "optional_parameter":
		name := node.ChildByFieldName("pattern")
		valueType := node.ChildByFieldName("type")

		if valueType == nil {
			name = node.Child(0)
			valueType = node.Child(1)
			if name == nil || name.Type() != "identifier" {
				return false
			}
			if valueType == nil {
				return false
			}
		}

		value.TextType = standardizeTextType(valueType)
		value.Type = standardizeDataType(valueType, valueType.Content())
		value.Name = name.Content()
		return true
	}

	return false
}

func annotateDataTypeProperties(finder *datatype.PropertyFinder, node *parser.Node) {
	parseProperty := func() {
		value := finder.ResolveClosestDataType(node.ID())
		if value == nil {
			return
		}

		propertyNode := node.ChildByFieldName("name")
		if propertyNode == nil {
			return
		}
		propertyName := propertyNode.Content()

		dataType := datatype.NewDataType()
		dataType.Node = propertyNode
		dataType.Name = strings.Trim(propertyName, "?")

		typeNode := node.ChildByFieldName("type")
		if typeNode != nil {
			dataType.TextType = standardizeTextType(typeNode)
			typeChildNode := typeNode.Child(0)

			if typeChildNode.Type() == "object_type" {
				childDataType := finder.ResolveClosestDataType(typeChildNode.ID())
				if childDataType != nil {
					dataType.Type = schema.SimpleTypeObject
					dataType.Properties = childDataType.Properties
					childDataType.IsHelper = true
				}
			} else if typeChildNode.Type() == "array_type" {
				dataType.Type = schema.SimpleTypeObject

				typeGrandChildNode := typeChildNode.Child(0)

				if typeGrandChildNode != nil && typeGrandChildNode.Type() == "object_type" {
					grandChildDataType := finder.ResolveClosestDataType(typeGrandChildNode.ID())
					if grandChildDataType != nil {
						dataType.Type = schema.SimpleTypeObject
						dataType.Properties = grandChildDataType.Properties
						grandChildDataType.IsHelper = true
					}
				}
			} else {
				typeContent := typeChildNode.Content()

				simpleType := standardizeDataType(typeChildNode, typeContent)
				dataType.Type = simpleType

				if simpleType != schema.SimpleTypeObject && simpleType != schema.SimpleTypeUnknown && simpleType != schema.SimpleTypeFunction {
					dataType.TextType = typeChildNode.Content()
				}
			}
		}

		valueNode := node.ChildByFieldName("value")
		if valueNode != nil {
			dataType.Type = standardizeDataType(valueNode, valueNode.Type())
		}

		value.Properties[propertyName] = &dataType
	}

	switch node.Type() {
	case "public_field_definition":
		parseProperty()
	case "property_signature":
		parseProperty()
	case "method_definition":
		value := finder.ResolveClosestDataType(node.ID())
		if value == nil {
			return
		}
		propertyNode := node.ChildByFieldName("name")
		propertyName := propertyNode.Content()

		dataType := datatype.NewDataType()
		dataType.Node = node
		dataType.Name = propertyName
		dataType.Type = schema.SimpleTypeFunction
		dataType.TextType = "function"
		value.Properties[propertyName] = &dataType
	}
}

func standardizeDataType(node *parser.Node, content string) string {
	content = strings.Trim(content, " ")

	isTypeOrNull := func(dataType string, input string) bool {
		if input == dataType {
			return true
		}
		input = strings.Trim(input, ":")
		// add support for
		// type | null
		parts := strings.Split(input, "|")

		if len(parts) == 2 {
			part0Trimmed := strings.Trim(parts[0], " ")
			part1Trimmed := strings.Trim(parts[1], " ")

			gotDataType := false
			if part0Trimmed == dataType || part1Trimmed == dataType {
				gotDataType = true
			}

			gotNull := false
			if part0Trimmed == "null" || part1Trimmed == "null" {
				gotNull = true
			}

			if gotDataType && gotNull {
				return true
			}
		}

		return false
	}

	if isTypeOrNull("string", content) {
		return schema.SimpleTypeString
	}

	if isTypeOrNull("number", content) {
		return schema.SimpleTypeNumber
	}

	if isTypeOrNull("boolean", content) {
		return schema.SimpleTypeBool
	}

	if strings.Contains(content, "Array<") || strings.Contains(content, "[") {
		return schema.SimpleTypeObject
	}

	if node.Child(0) != nil && node.Child(0).Type() == "predefined_type" {
		return schema.SimpleTypeObject
	}

	return schema.SimpleTypeUnknown
}

func standardizeTextType(node *parser.Node) string {
	if node.Type() == "type_annotation" && node.ChildCount() == 1 {
		typeNode := node.Child(0)
		if typeNode.Type() == "type_identifier" || typeNode.Type() == "predefined_type" {
			return typeNode.Content()
		}
	}
	return ""
}
