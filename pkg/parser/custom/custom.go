package custom

import (
	"github.com/bearer/curio/pkg/detectors/custom/config"
	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report/schema/datatype"
	"github.com/bearer/curio/pkg/util/file"
)

type Detector interface {
	ExtractArguments(node *parser.Node, idGenerator nodeid.Generator, fileinfo *file.FileInfo, filepath *file.Path) (map[parser.NodeID]*datatype.DataType, error)
	CompilePattern(Rule string, idGenerator nodeid.Generator) (config.CompiledRule, error)
	IsParam(node *parser.Node) (bool, bool, *config.Param)
}

func GenerateTreeSitterQuery(node *parser.Node, idGenerator nodeid.Generator, rule *config.CompiledRule, detector Detector) {
	if node.Type() == "ERROR" {
		return
	}

	end, shouldIgnore, param := detector.IsParam(node)

	if shouldIgnore {
		return
	}

	varName := "var"
	paramVar := "param_" + varName

	if rule.Tree != "" {
		rule.Tree += " "
	}
	rule.Tree += "("
	rule.Tree += node.Type() + ""

	assignedID := ""

	if param != nil {
		assignedID = idGenerator.GenerateId()
		paramName := paramVar + assignedID
		rule.Tree += ")"
		rule.Tree += " @" + paramName
		param.ParamName = varName + assignedID
		rule.Params = append(rule.Params, *param)
		if end {
			return
		}
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)

		GenerateTreeSitterQuery(child, idGenerator, rule, detector)
	}

	rule.Tree += ") @rule"
}
