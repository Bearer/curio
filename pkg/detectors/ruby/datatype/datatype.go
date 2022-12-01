package datatype

import (
	"strings"

	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/datatype"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report/schema"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"
	"github.com/bearer/curio/pkg/util/pointers"
	"github.com/smacker/go-tree-sitter/ruby"
)

var classesQuery = parser.QueryMustCompile(ruby.GetLanguage(),
	`(class
		name: (constant) @param_id
	) @param_class`)

var classPropertiesQuery = parser.QueryMustCompile(ruby.GetLanguage(),
	`(class
		( call
			arguments: (argument_list
				(simple_symbol) @param_id
			)
		)
	) @param_class`)

var classFunctionsQuery = parser.QueryMustCompile(ruby.GetLanguage(),
	`(class
		( method
			name: (identifier) @param_id
		)
	) @param_class`)

func Discover(node *parser.Node, idGenerator nodeid.Generator) map[parser.NodeID]*schemadatatype.DataType {
	classDataTypes := make(map[parser.NodeID]*schemadatatype.DataType)
	// add classses
	captures := node.QueryConventional(classesQuery)
	for _, capture := range captures {
		id := capture["param_id"].Content()
		classNode := capture["param_class"]
		classDataTypes[classNode.ID()] = &schemadatatype.DataType{
			Node:       classNode,
			Name:       id,
			Assignment: pointers.String(""),
			Type:       schema.SimpleTypeObject,
			TextType:   "class",
			Properties: make(map[string]schemadatatype.DataTypable),
		}
	}

	discoverClassProperties(node, classDataTypes)
	discoverClassFunctions(node, classDataTypes)

	discoverClassAssignment(node, classDataTypes)

	propertiesDatatypes := make(map[parser.NodeID]*schemadatatype.DataType)
	helperDatatypes := make(map[parser.NodeID]*schemadatatype.DataType)

	addProperties(node, helperDatatypes)
	linkProperties(node, propertiesDatatypes, helperDatatypes)
	discoverStructures(node, propertiesDatatypes)
	scopeAndMergeProperties(propertiesDatatypes, classDataTypes, idGenerator)

	// merge properties and classes
	for nodeID, datatype := range classDataTypes {
		propertiesDatatypes[nodeID] = datatype
	}

	datatype.PruneMap(propertiesDatatypes)

	return propertiesDatatypes
}

func discoverClassProperties(node *parser.Node, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	// add class properties
	captures := node.QueryConventional(classPropertiesQuery)
	for _, capture := range captures {
		classNode := capture["param_class"]
		if datatypes[classNode.ID()] == nil {
			continue
		}

		// get property name
		propertyNode := capture["param_id"]
		propertyName := strings.TrimLeft(propertyNode.Content(), ":")

		datatypes[classNode.ID()].Properties[propertyName] = &schemadatatype.DataType{
			Node:       propertyNode,
			Name:       propertyName,
			Type:       schema.SimpleTypeUnknown,
			Properties: make(map[string]schemadatatype.DataTypable),
			TextType:   "",
		}
	}
}

func discoverClassFunctions(node *parser.Node, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	captures := node.QueryConventional(classFunctionsQuery)
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
