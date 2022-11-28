package datatype

import (
	"strings"

	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/datatype"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report/schema"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"
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

func Discover(tree *parser.Tree, idGenerator nodeid.Generator) map[parser.NodeID]*schemadatatype.DataType {
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

	discoverClassAssignment(tree, classDataTypes)

	propertiesDatatypes := make(map[parser.NodeID]*schemadatatype.DataType)
	helperDatatypes := make(map[parser.NodeID]*schemadatatype.DataType)

	addProperties(tree, helperDatatypes)
	linkProperties(tree, propertiesDatatypes, helperDatatypes)
	discoverStructures(tree, propertiesDatatypes)
	scopeAndMergeProperties(propertiesDatatypes, classDataTypes, idGenerator)

	// merge properties and classes
	for nodeID, datatype := range classDataTypes {
		propertiesDatatypes[nodeID] = datatype
	}

	datatype.PruneMap(propertiesDatatypes)

	return propertiesDatatypes
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
