package customdetector

import (
	"regexp"

	"github.com/bearer/curio/pkg/commands/process/settings"
	"github.com/bearer/curio/pkg/detectors/custom/config"
	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/custom"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/util/file"
)

var tableNameRegex = regexp.MustCompile(`\$TABLE_NAME`)
var columnsRegex = regexp.MustCompile(`[<]?\$COLUMN[>]?`)
var functionNameRegex = regexp.MustCompile(`\$FUNCTION_NAME`)
var scriptRegex = regexp.MustCompile(`\$SCRIPT`)
var anythingRegex = regexp.MustCompile(`\$ANTYHING`)

func (detector *Detector) CompilePattern(
	rulePattern settings.RulePattern,
	idGenerator nodeid.Generator,
) (config.CompiledRule, error) {
	reworkedRule := tableNameRegex.ReplaceAllLiteral(
		[]byte(rulePattern.Pattern),
		[]byte("Var_Table_Name"+idGenerator.GenerateId()),
	)
	reworkedRule = columnsRegex.ReplaceAllLiteral(reworkedRule, []byte("Var_Column_Id"+idGenerator.GenerateId()+" Var_Type_id"+idGenerator.GenerateId()))
	reworkedRule = functionNameRegex.ReplaceAllLiteral(reworkedRule, []byte("Var_Function_Name"+idGenerator.GenerateId()+"()"))
	reworkedRule = scriptRegex.ReplaceAllLiteral(reworkedRule, []byte(`$$ Var_Script`+idGenerator.GenerateId()+` $$`))
	reworkedRule = anythingRegex.ReplaceAllLiteral(reworkedRule, []byte(`Var_Anything`+idGenerator.GenerateId()))

	tree, err := parser.ParseBytes(&file.FileInfo{}, &file.Path{}, []byte(reworkedRule), language, 0)
	if err != nil {
		return config.CompiledRule{}, err
	}
	defer tree.Close()

	compiledRule := &config.CompiledRule{
		Languages: []string{"sql"},
		Params:    make([]config.Param, 0),
		Filters:   rulePattern.Filters,
	}

	custom.GenerateTreeSitterQuery(tree.RootNode().Child(0), idGenerator, compiledRule, detector, false)

	return *compiledRule, nil
}
