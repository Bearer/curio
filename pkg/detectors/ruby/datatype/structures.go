package datatype

import (
	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/report/schema"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"
	"github.com/smacker/go-tree-sitter/ruby"
)

// Foo = Struct.new(foo: "foo", bar: "bar")
// Foo.new(foo: "foo", bar: "bar")
var structuresQuery = parser.QueryMustCompile(ruby.GetLanguage(), `
(call
		receiver: (constant) @receiver_id
		method: (identifier) @method_id
		arguments: (argument_list) @param_arguments
) @param_call
`)

func discoverStructures(tree *parser.Tree, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	// add class properties
	captures := tree.QueryConventional(structuresQuery)
	for _, capture := range captures {
		methodNode := capture["method_id"]
		methodID := methodNode.Content()
		if methodID != "new" {
			continue
		}

		receiver := capture["receiver_id"]

		// add parent name
		parentName := ""
		if receiver.Content() == "Struct" {
			callNode := capture["param_call"]
			assigmentNode := callNode.Parent()
			if assigmentNode.Type() != "assignment" {
				continue
			}

			leftNode := assigmentNode.ChildByFieldName("left")
			if leftNode == nil {
				continue
			}

			if leftNode.Type() != "constant" && leftNode.Type() != "identifier" {
				continue
			}

			parentName = leftNode.Content()
		} else {
			parentName = receiver.Content()
		}

		parent := &schemadatatype.DataType{
			Node:       receiver,
			Name:       parentName,
			Type:       schema.SimpleTypeUnknown,
			Properties: make(map[string]schemadatatype.DataTypable),
			TextType:   "",
		}

		arguments := capture["param_arguments"]

		// add child properties
		for i := 0; i < arguments.ChildCount(); i++ {
			pair := arguments.Child(i)

			if pair.Type() != "pair" {
				continue
			}

			key := pair.ChildByFieldName("key")
			if key.Type() != "hash_key_symbol" {
				continue
			}

			propertyName := key.Content()

			parent.Properties[propertyName] = &schemadatatype.DataType{
				Node:       key,
				Name:       propertyName,
				Type:       schema.SimpleTypeUnknown,
				TextType:   "",
				Properties: make(map[string]schemadatatype.DataTypable),
			}
		}

		datatypes[receiver.ID()] = parent
	}
}
