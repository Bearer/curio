package base

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/bearer/curio/new/language/patternquery"
	patternquerybuilder "github.com/bearer/curio/new/language/patternquery/builder"
	"github.com/bearer/curio/new/language/tree"
	"github.com/bearer/curio/new/language/types"
)

type Language struct {
	sitterLanguage          *sitter.Language
	analyzeFlow             func(rootNode *tree.Node)
	extractPatternVariables func(input string) (string, []patternquerybuilder.Variable, error)
}

func New(
	sitterLanguage *sitter.Language,
	analyzeFlow func(rootNode *tree.Node),
	extractPatternVariables func(input string) (string, []patternquerybuilder.Variable, error),
) *Language {
	return &Language{
		sitterLanguage:          sitterLanguage,
		analyzeFlow:             analyzeFlow,
		extractPatternVariables: extractPatternVariables,
	}
}

func (lang *Language) Parse(input string) (*tree.Tree, error) {
	tree, err := tree.Parse(lang.sitterLanguage, input)
	if err != nil {
		return nil, err
	}

	lang.analyzeFlow(tree.RootNode())
	return tree, nil
}

func (lang *Language) CompileQuery(input string) (*tree.Query, error) {
	return tree.CompileQuery(lang.sitterLanguage, input)
}

func (lang *Language) CompilePatternQuery(input string) (types.PatternQuery, error) {
	inputWithoutVariables, params, err := lang.extractPatternVariables(input)
	if err != nil {
		return nil, err
	}

	return patternquery.Compile(lang, inputWithoutVariables, params)
}
