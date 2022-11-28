package datatype

import (
	"strings"

	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/datatype"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report"
	"github.com/bearer/curio/pkg/report/detections"
	"github.com/bearer/curio/pkg/report/detectors"
	"github.com/bearer/curio/pkg/report/schema"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"
	"github.com/smacker/go-tree-sitter/csharp"
)

var classesQuery = parser.QueryMustCompile(csharp.GetLanguage(),
	`(class_declaration
	name: (identifier) @param_id
 ) @param_class`)

var classPropertiesQuery = parser.QueryMustCompile(csharp.GetLanguage(),
	`(class_declaration
	body: (declaration_list
    	(field_declaration
        	(variable_declaration
				(variable_declarator
					(identifier) @param_property_name
				)
			) @param_property
        )
    )
 ) @param_class`)

var classGetSetPropertiesQuery = parser.QueryMustCompile(csharp.GetLanguage(),
	`(class_declaration
 body: (declaration_list
	 (property_declaration
		 name: (identifier) @param_property
	 )
 )
) @param_class`)

var classFunctionsQuery = parser.QueryMustCompile(csharp.GetLanguage(),
	`(class_declaration
	body: (declaration_list
    	(method_declaration
        	name: (identifier) @param_function_name
        )
    )
 ) @param_class`)

var functionParamatersQuery = parser.QueryMustCompile(csharp.GetLanguage(), `(parameter
	name: (identifier) @param_name
) @param_parameter`)

func Discover(report report.Report, tree *parser.Tree, idGenerator nodeid.Generator) {
	datatypes := make(map[parser.NodeID]*schemadatatype.DataType)

	// add classses
	captures := tree.QueryConventional(classesQuery)
	for _, capture := range captures {
		id := capture["param_id"].Content()
		classNode := capture["param_class"]
		datatypes[classNode.ID()] = &schemadatatype.DataType{
			Node:       classNode,
			Name:       id,
			Type:       schema.SimpleTypeObject,
			TextType:   "class",
			Properties: make(map[string]schemadatatype.DataTypable),
		}
	}

	discoverProperties(tree, datatypes)
	discoverGetSetProperties(tree, datatypes)
	discoverFunctions(tree, datatypes)
	discoverFunctionParameters(tree, datatypes)

	datatype.PruneMap(datatypes)

	report.AddDataType(detections.TypeSchema, detectors.DetectorCSharp, idGenerator, datatypes, nil)
}

func discoverProperties(tree *parser.Tree, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	// add class properties
	captures := tree.QueryConventional(classPropertiesQuery)
	for _, capture := range captures {
		classNode := capture["param_class"]
		if datatypes[classNode.ID()] == nil {
			continue
		}

		// get property name
		property := capture["param_property"]
		propertyName := capture["param_property_name"].Content()

		// get property type
		propertyTypeNode := property.ChildByFieldName("type")
		propertyType := ""
		propertyTextType := ""
		if propertyTypeNode != nil {
			propertyType = standardizeDataType(propertyTypeNode, propertyTypeNode.Content())
			propertyTextType = propertyTypeNode.Content()
		}

		datatypes[classNode.ID()].Properties[propertyName] = &schemadatatype.DataType{
			Node:     property,
			Name:     propertyName,
			Type:     propertyType,
			TextType: propertyTextType,
		}
	}
}

func discoverGetSetProperties(tree *parser.Tree, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	// add class properties
	captures := tree.QueryConventional(classGetSetPropertiesQuery)
	for _, capture := range captures {
		classNode := capture["param_class"]
		if datatypes[classNode.ID()] == nil {
			continue
		}

		// get property name
		property := capture["param_property"]
		propertyName := property.Content()

		// get property type
		propertyTypeNode := property.Parent().ChildByFieldName("type")
		propertyType := ""
		propertyTextType := ""
		if propertyTypeNode != nil {
			propertyType = standardizeDataType(propertyTypeNode, propertyTypeNode.Content())
			propertyTextType = propertyTypeNode.Content()
		}

		datatypes[classNode.ID()].Properties[propertyName] = &schemadatatype.DataType{
			Node:     property,
			Name:     propertyName,
			Type:     propertyType,
			TextType: propertyTextType,
		}
	}
}

func discoverFunctions(tree *parser.Tree, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	captures := tree.QueryConventional(classFunctionsQuery)
	for _, capture := range captures {
		classNode := capture["param_class"]
		if datatypes[classNode.ID()] == nil {
			continue
		}

		// get method name
		functionNameNode := capture["param_function_name"]
		functionName := functionNameNode.Content()

		parent := functionNameNode.Parent()

		functionTypeNode := parent.ChildByFieldName("type")
		functionType := ""
		functionTextType := ""
		if functionTypeNode != nil {
			functionType = standardizeDataType(functionTypeNode, functionTypeNode.Content())
			functionTextType = functionTypeNode.Content()
		}

		datatypes[classNode.ID()].Properties[functionName] = &schemadatatype.DataType{
			Node:     functionNameNode,
			Name:     functionName,
			Type:     functionType,
			TextType: functionTextType,
		}
	}
}

func discoverFunctionParameters(tree *parser.Tree, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	captures := tree.QueryConventional(functionParamatersQuery)
	for _, capture := range captures {
		paramNameNode := capture["param_name"]
		paramNode := capture["param_parameter"]

		paramTypeNode := paramNode.ChildByFieldName("type")
		paramType := ""
		paramTextType := ""
		if paramTypeNode != nil {
			paramTextType = paramTypeNode.Content()
			paramType = standardizeDataType(paramTypeNode, paramTypeNode.Content())
		}

		datatypes[paramNode.ID()] = &schemadatatype.DataType{
			Node:       paramNode,
			Name:       paramNameNode.Content(),
			Type:       paramType,
			TextType:   paramTextType,
			Properties: make(map[string]schemadatatype.DataTypable),
		}
	}
}

func standardizeDataType(node *parser.Node, content string) string {
	content = strings.Trim(content, " ")

	if content == "string" || content == "char" {
		return schema.SimpleTypeString
	}

	if content == "int" || content == "short" || content == "byte" || content == "uint" || content == "float" || content == "double" || content == "decimal" {
		return schema.SimpleTypeNumber
	}

	if content == "bool" {
		return schema.SimpleTypeBool
	}

	if node.Type() == "identifier" || node.Type() == "generic_name" {
		return schema.SimpleTypeObject
	}

	return schema.SimpleTypeUnknown
}
