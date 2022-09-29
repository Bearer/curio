package json

import (
	"github.com/bearer/curio/pkg/detectors/openapi/queries"
	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report/schema/schemahelper"
	"github.com/smacker/go-tree-sitter/javascript"
)

var queryOperationId = parser.QueryMustCompile(javascript.GetLanguage(), `
(_
	(
      pair
        key:
            (string) @helperOperationId
            (#match? @helperOperationId "^\"operationId\"$")
         value:
            (string) @param_operation_id
	)
	(
      pair
        key:
            (string) @helperParameters
            (#match? @helperParameters "^\"parameters\"$")
         value:
          (array)  @param_parameters
	)
)
`)

func AnnotateOperationId(nodeIDMap *nodeid.Map, tree *parser.Tree, foundValues map[parser.Node]*schemahelper.Schema) error {
	return queries.AnnotateOperationId(queries.OperationIdRequest{
		Tree:        tree,
		FoundValues: foundValues,
		Query:       queryOperationId,
		ChildMatch:  OperationIdChildMatcher{},
		NodeIDMap:   nodeIDMap,
	})
}

type OperationIdChildMatcher struct {
}

func (childMatcher OperationIdChildMatcher) Match(input *parser.Node) *parser.Node {
	return input
}
