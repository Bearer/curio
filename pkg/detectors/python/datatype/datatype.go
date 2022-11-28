package datatype

import (
	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/datatype"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report"
	"github.com/bearer/curio/pkg/report/detections"
	"github.com/bearer/curio/pkg/report/detectors"
	"github.com/bearer/curio/pkg/report/schema"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"
	"github.com/smacker/go-tree-sitter/python"
)

var classesQuery = parser.QueryMustCompile(python.GetLanguage(),
	`(class_definition
		name: (identifier) @param_id
	  ) @param_class`)

var classPropertiesQuery = parser.QueryMustCompile(python.GetLanguage(),
	`(class_definition
		body: (block
			(expression_statement) @param_element
		)
	  ) @param_class`)

var classFunctionsQuery = parser.QueryMustCompile(python.GetLanguage(),
	`(class_definition
		name: (identifier) @param_object
		body: (block
			(function_definition
			  name: (identifier) @param_id
		  ) @param_element
		)
	  ) @param_class`)

func Discover(report report.Report, tree *parser.Tree, idGenerator nodeid.Generator) {
	classDataTypes := make(map[parser.NodeID]*schemadatatype.DataType)
	// add classses
	captures := tree.QueryConventional(classesQuery)
	for _, capture := range captures {
		id := capture["param_id"].Content()
		classNode := capture["param_class"]
		classDataTypes[classNode.ID()] = &schemadatatype.DataType{
			Node:       classNode,
			Name:       id,
			Type:       schema.SimpleTypeObject,
			TextType:   "class",
			Properties: make(map[string]schemadatatype.DataTypable),
		}
	}

	discoverClassProperties(tree, classDataTypes)
	discoverClassFunctions(tree, classDataTypes)

	propertiesDatatypes := make(map[parser.NodeID]*schemadatatype.DataType)
	helperDatatypes := make(map[parser.NodeID]*schemadatatype.DataType)

	addProperties(tree, helperDatatypes)
	linkProperties(tree, propertiesDatatypes, helperDatatypes)
	scopeAndMergeProperties(propertiesDatatypes, classDataTypes, idGenerator)

	// merge properties and classes
	for nodeID, datatype := range classDataTypes {
		propertiesDatatypes[nodeID] = datatype
	}

	datatype.PruneMap(propertiesDatatypes)

	report.AddDataType(detections.TypeSchema, detectors.DetectorPython, idGenerator, propertiesDatatypes, nil)
}

func discoverClassProperties(tree *parser.Tree, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	// add class properties
	captures := tree.QueryConventional(classPropertiesQuery)
	for _, capture := range captures {
		classNode := capture["param_class"]
		if datatypes[classNode.ID()] == nil {
			continue
		}

		// get property name
		propertyNode := capture["param_element"]
		child := propertyNode.Child(0)
		if child == nil {
			continue
		}

		propertyName := ""
		if child.Type() == "assignment" {
			left := child.ChildByFieldName("left")

			if left != nil && left.Type() == "identifier" {
				propertyName = left.Content()
			}
		}

		if child.Type() == "identifier" {
			propertyName = child.Content()
		}

		if propertyName == "" {
			continue
		}

		datatypes[classNode.ID()].Properties[propertyName] = &schemadatatype.DataType{
			Node:       propertyNode,
			Name:       propertyName,
			Type:       schema.SimpleTypeUnknown,
			Properties: make(map[string]schemadatatype.DataTypable),
			TextType:   "",
		}
	}
}

func discoverClassFunctions(tree *parser.Tree, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	captures := tree.QueryConventional(classFunctionsQuery)
	for _, capture := range captures {
		classNode := capture["param_class"]
		if datatypes[classNode.ID()] == nil {
			continue
		}

		// get method name
		functionNameNode := capture["param_id"]
		functionName := functionNameNode.Content()

		datatypes[classNode.ID()].Properties[functionName] = &schemadatatype.DataType{
			Node:     functionNameNode,
			Name:     functionName,
			Type:     schema.SimpleTypeUnknown,
			TextType: "",
		}
	}
}
