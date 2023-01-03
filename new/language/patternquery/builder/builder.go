package builder

import (
	"fmt"
	"strings"

	"github.com/bearer/curio/new/language/tree"
	"github.com/bearer/curio/new/language/types"
	"github.com/bearer/curio/pkg/parser/nodeid"
)

type Variable struct {
	NodeType   string
	DummyValue string
	Name       string
}

type Result struct {
	Query           string
	ParamToVariable map[string]string
	EqualParams     [][]string
	ParamToContent  map[string]string
}

type builder struct {
	stringBuilder    strings.Builder
	idGenerator      nodeid.Generator
	variables        []Variable
	variableToParams map[string][]string
	paramToContent   map[string]string
}

func Build(lang types.Language, input string, variables []Variable) (*Result, error) {
	tree, err := lang.Parse(input)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	if tree.RootNode().ChildCount() != 1 {
		return nil, fmt.Errorf("expecting 1 node but got %d", tree.RootNode().ChildCount())
	}

	builder := builder{
		stringBuilder:    strings.Builder{},
		idGenerator:      &nodeid.IntGenerator{},
		variables:        variables,
		variableToParams: make(map[string][]string),
		paramToContent:   make(map[string]string),
	}

	return builder.build(tree.RootNode().Child(0)), nil
}

func (builder *builder) build(rootNode *tree.Node) *Result {
	builder.write("(")
	builder.compileNode(rootNode)
	builder.write(" @root")
	builder.write(")")

	paramToVariable, equalParams := builder.processVariableToParams()

	return &Result{
		Query:           builder.stringBuilder.String(),
		ParamToVariable: paramToVariable,
		EqualParams:     equalParams,
		ParamToContent:  builder.paramToContent,
	}
}

func (builder *builder) compileNode(node *tree.Node) error {
	if node.IsError() {
		return fmt.Errorf(
			"error parsing pattern at %d:%d: %s",
			node.LineNumber(),
			node.ColumnNumber(),
			node.Content(),
		)
	}

	if variable := builder.getVariableFor(node); variable != nil {
		builder.compileVariableNode(variable)
	} else if !node.IsNamed() {
		builder.compileUnnamedNode(node)
	} else if node.ChildCount() == 0 {
		builder.compileLeafNode(node)
	} else {
		return builder.compileNodeWithChildren(node)
	}

	return nil
}

// variable nodes match their type and capture their content
func (builder *builder) compileVariableNode(variable *Variable) {
	paramName := builder.newParam()
	builder.variableToParams[variable.Name] = append(builder.variableToParams[variable.Name], paramName)

	builder.write("(")
	builder.write(variable.NodeType)
	builder.write(") @")
	builder.write(paramName)
}

// Un-named nodes match their content as a literal
func (builder *builder) compileUnnamedNode(node *tree.Node) {
	builder.write(`"`)
	builder.write(escapeQueryString(node.Content()))
	builder.write(`"`)
}

// Leaves match their type and content
func (builder *builder) compileLeafNode(node *tree.Node) {
	paramName := builder.newParam()
	builder.paramToContent[paramName] = node.Content()

	builder.write("(")
	builder.write(node.Type())
	builder.write(") @")
	builder.write(paramName)
}

// Nodes with children match their type and child nodes
func (builder *builder) compileNodeWithChildren(node *tree.Node) error {
	builder.write("(")
	builder.write(node.Type())

	for i := 0; i < node.ChildCount(); i++ {
		builder.write(" ")

		if err := builder.compileNode(node.Child(i)); err != nil {
			return err
		}
	}

	builder.write(")")

	return nil
}

func (builder *builder) processVariableToParams() (map[string]string, [][]string) {
	paramToVariable := make(map[string]string)
	var equalParams [][]string

	for variableName, paramNames := range builder.variableToParams {
		if len(paramNames) > 1 {
			equalParams = append(equalParams, paramNames)
		}

		paramToVariable[paramNames[0]] = variableName
	}

	return paramToVariable, equalParams
}

func (builder *builder) getVariableFor(node *tree.Node) *Variable {
	for _, variable := range builder.variables {
		if node.Content() == variable.DummyValue {
			return &variable
		}
	}

	return nil
}

func (builder *builder) write(value string) {
	builder.stringBuilder.WriteString(value)
}

func (builder *builder) newParam() string {
	return "param" + builder.idGenerator.GenerateId()
}

func escapeQueryString(value string) string {
	return value // FIXME
}
