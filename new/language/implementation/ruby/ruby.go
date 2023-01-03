package ruby

import (
	"fmt"
	"regexp"

	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/ssoroka/slice"
	"golang.org/x/exp/slices"

	"github.com/bearer/curio/new/language/base"
	patternquerybuilder "github.com/bearer/curio/new/language/patternquery/builder"
	"github.com/bearer/curio/new/language/tree"
	"github.com/bearer/curio/pkg/util/regex"
)

var (
	variableLookupParents = []string{"pair", "argument_list"}

	// $<name:type> or $<name>
	patternQueryVariableRegex = regexp.MustCompile(`\$<(?P<name>[^>:]+)(?::(?P<type>[^>]+))?>`)
	allowedPatternQueryTypes  = []string{"identifier", "constant", "_"}
)

func Get() *base.Language {
	return base.New(ruby.GetLanguage(), analyzeFlow, extractPatternVariables)
}

func analyzeFlow(rootNode *tree.Node) {
	scope := make(map[string]*tree.Node)

	rootNode.Walk(func(node *tree.Node) error {
		switch node.Type() {
		case "method":
			scope = make(map[string]*tree.Node)
		case "assignment":
			left := node.ChildByFieldName("left")
			right := node.ChildByFieldName("right")

			if left.Type() == "identifier" {
				scope[left.Content()] = node

				node.UnifyWith(right)
			}
		case "identifier":
			parent := node.Parent()
			if parent != nil && slice.Contains(variableLookupParents, parent.Type()) {
				scopedNode := scope[node.Content()]
				if scopedNode != nil {
					node.UnifyWith(scopedNode)
				}
			}
		}

		return nil
	})
}

func extractPatternVariables(input string) (string, []patternquerybuilder.Variable, error) {
	nameIndex := patternQueryVariableRegex.SubexpIndex("name")
	typeIndex := patternQueryVariableRegex.SubexpIndex("type")
	i := 0

	var params []patternquerybuilder.Variable

	replaced, err := regex.ReplaceAllWithSubmatches(patternQueryVariableRegex, input, func(submatches []string) (string, error) {
		nodeType := submatches[typeIndex]
		if nodeType == "" {
			nodeType = "_"
		}

		if !slices.Contains(allowedPatternQueryTypes, nodeType) {
			return "", fmt.Errorf("invalid node type '%s' in pattern query", nodeType)
		}

		dummyValue := produceDummyValue(i, nodeType)

		params = append(params, patternquerybuilder.Variable{
			Name:       submatches[nameIndex],
			NodeType:   nodeType,
			DummyValue: dummyValue,
		})

		i += 1

		return dummyValue, nil
	})
	if err != nil {
		return "", nil, err
	}

	return replaced, params, nil
}

func produceDummyValue(i int, nodeType string) string {
	switch nodeType {
	case "identifier":
		return "curioVar" + fmt.Sprint(i)
	default:
		return "CurioVar" + fmt.Sprint(i)
	}
}
