package datatype

import (
	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/datatype"
	"github.com/bearer/curio/pkg/report/detections"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"

	"github.com/bearer/curio/pkg/parser/nodeid"
	php "github.com/bearer/curio/pkg/parser/sitter/php2"
	"github.com/bearer/curio/pkg/report"
	"github.com/bearer/curio/pkg/report/detectors"
	"github.com/bearer/curio/pkg/report/schema"
)

var classesQuery = parser.QueryMustCompile(php.GetLanguage(),
	`(class_declaration
		name: (name) @param_id
	) @param_class`)

var classPropertiesQuery = parser.QueryMustCompile(php.GetLanguage(),
	`(class_declaration
		body:
		(declaration_list
			(property_declaration
				(property_element
					(variable_name
						(name) @param_id
					)
				)
			)
		)
	) @param_class`)

var classFunctionsQuery = parser.QueryMustCompile(php.GetLanguage(),
	`(class_declaration
		body: (declaration_list
			(method_declaration
				name: (name) @param_id
			)
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

	report.AddDataType(detections.TypeSchema, detectors.DetectorPHP, idGenerator, propertiesDatatypes, nil)
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
		propertyNode := capture["param_id"]
		propertyName := propertyNode.Content()

		datatypes[classNode.ID()].Properties[propertyName] = &schemadatatype.DataType{
			Node:     propertyNode,
			Name:     propertyName,
			Type:     schema.SimpleTypeUnknown,
			TextType: "",
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
