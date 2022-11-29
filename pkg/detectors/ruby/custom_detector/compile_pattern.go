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

var classNameRegex = regexp.MustCompile(`\$CLASS_NAME`)
var argumentsRegex = regexp.MustCompile(`<\$ARGUMENT>`)
var dataTypeRegex = regexp.MustCompile(`<\$DATA_TYPE>`)
var anythingRegex = regexp.MustCompile(`\$ANYTHING`)
var variableRegex = regexp.MustCompile(`\$([A-Z_]+)`)

func (detector *Detector) CompilePattern(
	rulePattern settings.RulePattern,
	idGenerator nodeid.Generator,
) (config.CompiledRule, error) {
	reworkedRule := classNameRegex.ReplaceAllString(
		rulePattern.Pattern,
		"Var_Class_Name"+idGenerator.GenerateId(),
	)
	reworkedRule = argumentsRegex.ReplaceAllString(reworkedRule, "Var_Arguments"+idGenerator.GenerateId())
	reworkedRule = dataTypeRegex.ReplaceAllString(reworkedRule, "Var_DataTypes"+idGenerator.GenerateId())
	reworkedRule = anythingRegex.ReplaceAllString(reworkedRule, "Var_Anything"+idGenerator.GenerateId())
	reworkedRule = variableRegex.ReplaceAllString(reworkedRule, "var_Variable_$1")

	tree, err := parser.ParseBytes(&file.FileInfo{}, &file.Path{}, []byte(reworkedRule), language, 0)
	if err != nil {
		return config.CompiledRule{}, err
	}
	defer tree.Close()

	rule := &config.CompiledRule{
		Languages: []string{"ruby"},
		Params:    make([]config.Param, 0),
		Filters:   rulePattern.Filters,
	}

	custom.GenerateTreeSitterQuery(tree.RootNode().Child(0), idGenerator, rule, detector, false)

	return *rule, nil
}
