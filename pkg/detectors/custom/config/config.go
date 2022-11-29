package config

import "github.com/bearer/curio/pkg/commands/process/settings"

type CompiledRule struct {
	RuleName               string
	Tree                   string
	Params                 []Param
	Metavars               map[string]settings.MetaVar
	Filters                []settings.PatternFilter
	ParamParenting         bool
	RootSingularize        bool
	RootLowercase          bool
	VariableReconciliation bool
	Languages              []string
}

func (rule *CompiledRule) GetParamByPatternName(name string) *Param {
	for _, param := range rule.Params {
		if param.PatternName == name {
			return &param
		}
	}

	return nil
}

type Param struct {
	PatternName      string `yaml:"syntax_name"` // name in pattern eg: $SCRIPT
	ParamName        string `yaml:"param_name"`  // name of param eg: var1
	StringMatch      string `yaml:"string_match"`
	RegexMatch       string `yaml:"regex_match"`
	StringExtract    bool   `yaml:"string_extract"`
	ArgumentsExtract bool   `yaml:"arguments_extract"`
	ClassNameExtract bool   `yaml:"class_name_extract"`
	MatchAnything    bool   `yaml:"match_anything"`
}

func (param *Param) BuildFullName() string {
	return "param_" + param.ParamName
}
