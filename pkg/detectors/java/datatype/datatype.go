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
	"github.com/smacker/go-tree-sitter/java"
)

var classesQuery = parser.QueryMustCompile(java.GetLanguage(),
	`(class_declaration
		name: (identifier) @param_name
	) @param_class`)

var classPropertiesQuery = parser.QueryMustCompile(java.GetLanguage(),
	`(class_declaration
		body: (class_body
		  (field_declaration
			  type: (_) @param_type
			  (variable_declarator
			  (identifier) @param_id
			)
		  ) @param_node
		)
	)@param_class
	`)

var classFunctionsQuery = parser.QueryMustCompile(java.GetLanguage(),
	`(class_declaration
		body: (class_body
			(method_declaration
				name: (identifier) @param_id
			) @param_node
		)
	) @param_class`)

func Discover(report report.Report, tree *parser.Tree, idGenerator nodeid.Generator) {
	datatypes := make(map[parser.NodeID]*schemadatatype.DataType)

	// add classses
	captures := tree.QueryConventional(classesQuery)
	for _, capture := range captures {
		name := capture["param_name"].Content()
		classNode := capture["param_class"]

		datatypes[classNode.ID()] = &schemadatatype.DataType{
			Node:       classNode,
			Name:       name,
			Type:       schema.SimpleTypeObject,
			TextType:   "class",
			Properties: make(map[string]schemadatatype.DataTypable),
		}
	}

	discoverProperties(tree, datatypes)
	discoverFunctions(tree, datatypes)

	datatype.PruneMap(datatypes)

	report.AddDataType(detections.TypeSchema, detectors.DetectorJava, idGenerator, datatypes, nil)
}

func discoverProperties(tree *parser.Tree, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	// add class properties
	captures := tree.QueryConventional(classPropertiesQuery)
	for _, capture := range captures {
		classNode := capture["param_class"]
		if datatypes[classNode.ID()] == nil {
			continue
		}

		// get node
		propertyNode := capture["param_node"]

		// get property name
		propertyName := capture["param_id"].Content()

		// get property type
		propertyTypeNode := capture["param_type"]
		propertyType := standardizeDataType(propertyTypeNode, propertyTypeNode.Content())
		propertyTextType := propertyTypeNode.Content()

		datatypes[classNode.ID()].Properties[propertyName] = &schemadatatype.DataType{
			Node:       propertyNode,
			Name:       propertyName,
			Type:       propertyType,
			TextType:   propertyTextType,
			Properties: make(map[string]schemadatatype.DataTypable),
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

		// get node
		functionNode := capture["param_node"]

		// get method name
		functionNameNode := capture["param_id"]
		functionName := functionNameNode.Content()

		datatypes[classNode.ID()].Properties[functionName] = &schemadatatype.DataType{
			Node:       functionNode,
			Name:       functionName,
			Type:       schema.SimpleTypeFunction,
			TextType:   "",
			Properties: make(map[string]schemadatatype.DataTypable),
		}
	}
}

func standardizeDataType(node *parser.Node, content string) string {
	content = strings.Trim(content, " ")

	if content == "String" || content == "char" {
		return schema.SimpleTypeString
	}

	if content == "int" || content == "short" || content == "float" || content == "double" {
		return schema.SimpleTypeNumber
	}

	if content == "byte" {
		return schema.SimpleTypeBinary
	}

	if content == "boolean" {
		return schema.SimpleTypeBool
	}

	if node.Type() == "type_identifier" {
		return schema.SimpleTypeObject
	}

	return schema.SimpleTypeUnknown
}
