package customdetector

import (
	"strings"

	"github.com/bearer/curio/pkg/detectors/custom/config"
	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/sitter/sql"
)

var language = sql.GetLanguage()

type Detector struct {
}

func (detector *Detector) IsParam(node *parser.Node) (isTerminating bool, shouldIgnore bool, param *config.Param) {
	if node.Type() == "identifier" {
		content := node.Content()
		if strings.Index(content, "Var_Column_Id") == 0 {
			param = &config.Param{
				ArgumentsExtract: true,
			}
			isTerminating = true
			return
		}

		if strings.Index(content, "Var_Function_Name") == 0 {
			param = &config.Param{
				ArgumentsExtract: true,
			}
			isTerminating = true
			return
		}

		if strings.Index(node.Content(), "Var_Table_Name") == 0 {
			param = &config.Param{
				ArgumentsExtract: true,
			}
			isTerminating = true
			return
		}

		if strings.Index(node.Content(), "Var_Anything") == 0 {
			param = &config.Param{
				MatchAnything: true,
			}
			shouldIgnore = true
			return
		}
	}

	if node.Type() == "content" && strings.Index(node.Content(), ` Var_Script`) == 0 {
		param = &config.Param{
			StringExtract: true,
			PatternName:   "SCRIPT",
		}
		isTerminating = true
		return
	}

	return false, false, nil
}
