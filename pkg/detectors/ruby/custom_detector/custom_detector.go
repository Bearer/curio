package customdetector

import (
	"strings"

	"github.com/bearer/curio/pkg/detectors/custom/config"
	"github.com/bearer/curio/pkg/parser"
	"github.com/smacker/go-tree-sitter/ruby"
)

var language = ruby.GetLanguage()

type Detector struct {
}

func (detector *Detector) IsParam(node *parser.Node) (isTerminating bool, shouldIgnore bool, param *config.Param) {
	if strings.Index(node.Content(), "Var_Anything") == 0 {
		param = &config.Param{
			ArgumentsExtract: true,
			MatchAnything:    true,
		}

		isTerminating = true
		return
	}

	if node.Type() == "constant" || node.Type() == "identifier" {
		nodeContent := node.Content()

		// log.Error().Msgf("Type %s Content %s", node.Type(), node.Content())
		// get class names
		if strings.Index(nodeContent, "Var_Class_Name") == 0 {
			param = &config.Param{
				ClassNameExtract: true,
			}
			isTerminating = true
			return
		}

		if strings.Index(nodeContent, "var_Variable_") == 0 {
			param = &config.Param{
				PatternName:   nodeContent[13:],
				MatchAnything: true,
			}
			isTerminating = true
			return
		}

		// get simple string identifiers
		param = &config.Param{
			StringMatch: nodeContent,
		}
		isTerminating = true
		return
	}

	if node.Child(0) != nil && strings.Index(node.Child(0).Content(), "Var_DataTypes") == 0 {
		param = &config.Param{
			ArgumentsExtract: true,
		}

		isTerminating = true
		return
	}

	if node.Type() == "symbol_array" && node.Child(0) != nil && node.Child(0).Type() == "bare_symbol" && strings.Index(node.Child(0).Content(), "Var_Arguments") == 0 {
		param = &config.Param{
			ArgumentsExtract: true,
			// StringExtract:    true,
		}
		isTerminating = true
		return
	}

	if node.Type() == "argument_list" && node.Child(0).Type() == "constant" && strings.Index(node.Child(0).Content(), "Var_Arguments") == 0 {
		param = &config.Param{
			ArgumentsExtract: true,
		}
		isTerminating = true
		return
	}

	return false, false, nil
}
