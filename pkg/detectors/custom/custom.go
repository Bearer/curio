package custom

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/bearer/curio/pkg/commands/process/settings"
	"github.com/bearer/curio/pkg/detectors/custom/config"
	rubydetector "github.com/bearer/curio/pkg/detectors/ruby/custom_detector"
	rubydatatype "github.com/bearer/curio/pkg/detectors/ruby/datatype"
	sqldetector "github.com/bearer/curio/pkg/detectors/sql/custom_detector"
	"github.com/bearer/curio/pkg/detectors/types"
	"github.com/bearer/curio/pkg/parser"
	language "github.com/bearer/curio/pkg/parser/custom"
	parserdatatype "github.com/bearer/curio/pkg/parser/datatype"
	"github.com/bearer/curio/pkg/report/detections"
	schemadatatype "github.com/bearer/curio/pkg/report/schema/datatype"
	"github.com/bearer/curio/pkg/util/file"
	pluralize "github.com/gertd/go-pluralize"
	"golang.org/x/exp/slices"

	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/parser/sitter/sql"
	"github.com/bearer/curio/pkg/report"
	"github.com/bearer/curio/pkg/report/detectors"
	"github.com/bearer/curio/pkg/report/schema"
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/smacker/go-tree-sitter/ruby"
)

type Detector struct {
	idGenerator      nodeid.Generator
	paramIdGenerator nodeid.Generator
	compiledRules    []config.CompiledRule
	pluralize        *pluralize.Client

	Ruby language.Detector
	Sql  language.Detector
}

func New(idGenerator nodeid.Generator) types.Detector {
	detector := &Detector{
		idGenerator:      idGenerator,
		paramIdGenerator: &nodeid.IntGenerator{Counter: 1},
	}

	detector.Ruby = &rubydetector.Detector{}
	detector.Sql = &sqldetector.Detector{}
	detector.pluralize = pluralize.NewClient()

	return detector
}

func (detector *Detector) AcceptDir(dir *file.Path) (bool, error) {
	return true, nil
}

func (detector *Detector) CompileRules(rulesConfig map[string]settings.Rule) error {
	compiledRules := make([]config.CompiledRule, 0)

	for ruleName, rule := range rulesConfig {
		if rule.Disabled {
			continue
		}
		for _, lang := range rule.Languages {
			for _, rulePattern := range rule.Patterns {
				compiledRule, err := detector.compileRule(rulePattern, lang, detector.paramIdGenerator)
				if err != nil {
					return err
				}

				compiledRule.RuleName = ruleName
				compiledRule.Metavars = rule.Metavars
				compiledRule.ParamParenting = rule.ParamParenting
				if rule.ParamParenting {
					compiledRule.VariableReconciliation = false
				} else {
					compiledRule.VariableReconciliation = true
				}
				compiledRule.RootLowercase = rule.RootLowercase
				compiledRule.RootSingularize = rule.RootSingularize
				compiledRules = append(compiledRules, compiledRule)
			}
		}
	}

	sort.Slice(compiledRules, func(i, j int) bool {
		return strings.Compare(compiledRules[i].RuleName, compiledRules[j].RuleName) == -1
	})

	detector.compiledRules = compiledRules

	return nil
}

func (detector *Detector) ProcessFile(file *file.FileInfo, dir *file.Path, report report.Report) (bool, error) {
	for _, rule := range detector.compiledRules {
		err := detector.executeRule(rule, file, report, detector.idGenerator)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

func (detector *Detector) compileRule(
	rulePattern settings.RulePattern,
	lang string,
	idGenerator nodeid.Generator,
) (config.CompiledRule, error) {
	var rule config.CompiledRule
	var err error

	switch lang {
	case "ruby":
		rule, err = detector.Ruby.CompilePattern(rulePattern, detector.paramIdGenerator)
	case "sql":
		rule, err = detector.Sql.CompilePattern(rulePattern, detector.paramIdGenerator)
	default:
		return config.CompiledRule{}, errors.New("unsupported language")
	}

	if err != nil {
		return config.CompiledRule{}, err
	}

	return rule, validateRule(rule)
}

func validateRule(rule config.CompiledRule) error {
	for _, filter := range rule.Filters {
		if param := rule.GetParamByPatternName(filter.Variable); param == nil {
			return fmt.Errorf("undefined variable '%s' in filter for custom rule '%s'", filter.Variable, rule.RuleName)
		}
	}

	return nil
}

func languageMatchesFile(file *file.FileInfo, ruleLanguage string) bool {
	if file.Language == "Ruby" && ruleLanguage == "ruby" {
		return true
	}
	if ruleLanguage == "sql" {
		if file.Language != "SQL" &&
			// postgress
			file.Language != "PLpgSQL" && file.Language != "PLSQL" && file.Language != "SQLPL" &&
			// microsoft sql
			file.Language != "TSQL" {
			return false
		}
		return true
	}
	return false
}

func (detector *Detector) executeRule(rule config.CompiledRule, file *file.FileInfo, report report.Report, idGenerator nodeid.Generator) error {
	for _, lang := range rule.Languages {
		if !languageMatchesFile(file, lang) {
			continue
		}

		sitterLang := getLanguage(lang)

		tree, err := parser.ParseFile(file, file.Path, sitterLang)
		if err != nil {
			return err
		}

		query := parser.QueryMustCompile(sitterLang, rule.Tree)

		captures := tree.QueryConventional(query)

		filteredCaptures, err := filterCaptures(rule.Params, captures)
		if err != nil {
			return err
		}

		var variableReconciliation *parserdatatype.ReconciliationRequest
		if rule.VariableReconciliation {
			variableReconciliation, err = detector.buildReconciliationRequest(lang, tree)
			if err != nil {
				return err
			}
		}

		err = detector.extractData(filteredCaptures, rule, report, lang, idGenerator, variableReconciliation)
		if err != nil {
			return err
		}

		tree.Close()
	}
	return nil
}

func (detector *Detector) extractData(captures []parser.Captures, rule config.CompiledRule, report report.Report, lang string, idGenerator nodeid.Generator, variableReconciliation *parserdatatype.ReconciliationRequest) error {
	for _, capture := range captures {
		forExport := make(map[parser.NodeID]*schemadatatype.DataType)
		var parent schemadatatype.DataTypable

		if filtersMatch := matchFilters(capture, rule); !filtersMatch {
			continue
		}

		for _, param := range rule.Params {
			var paramTypes map[parser.NodeID]*schemadatatype.DataType
			var err error

			if param.ArgumentsExtract || param.ClassNameExtract {
				paramTypes, err = detector.extractArguments(lang, capture[param.BuildFullName()], idGenerator, variableReconciliation)
				if err != nil {
					return err
				}

				if rule.ParamParenting {
					// join it as children to master param and export it
					if len(forExport) == 1 {
						for _, datatype := range paramTypes {
							parent.GetProperties()[datatype.GetName()] = datatype
						}

						continue
					}

					// set parent
					if len(paramTypes) == 1 {
						for dataTypeID, datatype := range paramTypes {
							parent, err = parserdatatype.DeepestSingleChild(datatype)
							if err != nil {
								return err
							}

							forExport[dataTypeID] = datatype
							break
						}
						continue
					}
				}

				// set all for export
				for datatypeID, datatype := range paramTypes {
					forExport[datatypeID] = datatype
				}
			}

			if param.StringExtract {
				for _, metavar := range rule.Metavars {
					if metavar.Input == param.PatternName {
						matchNode := capture[param.BuildFullName()]
						content := matchNode.Content()
						searchRegexp := regexp.MustCompile(metavar.Regex)

						matches := searchRegexp.FindAllStringSubmatch(content, -1)

						for _, subgroupMatch := range matches {
							match := subgroupMatch[metavar.Output]

							matchType := &schemadatatype.DataType{
								Node:       matchNode,
								Name:       string(match),
								Type:       schema.SimpleTypeUnknown,
								Properties: make(map[string]schemadatatype.DataTypable),
							}
							matchNodeID := matchNode.ID()

							if rule.ParamParenting {
								parent.GetProperties()[matchType.Name] = matchType
								continue
							}

							forExport[matchNodeID] = matchType
						}
					}
				}
			}
		}

		detector.applyDatatypeTransformations(rule, forExport)

		report.AddDataType(
			detections.TypeCustom,
			detectors.Type(rule.RuleName),
			idGenerator,
			forExport,
			capture["rule"],
		)
	}

	return nil
}

func matchFilters(captures parser.Captures, rule config.CompiledRule) bool {
	for _, filter := range rule.Filters {
		param := rule.GetParamByPatternName(filter.Variable)
		matchNode := captures[param.BuildFullName()]
		content := matchNode.Content()

		if !slices.Contains(filter.Values, content) {
			return false
		}
	}

	return true
}

func (detector *Detector) applyDatatypeTransformations(rule config.CompiledRule, datatypes map[parser.NodeID]*schemadatatype.DataType) {
	for _, datatype := range datatypes {
		if rule.RootSingularize {
			datatype.Name = detector.pluralize.Singular(datatype.Name)
		}

		if rule.RootLowercase {
			datatype.Name = strings.ToLower(datatype.Name)
		}
	}
}

func filterCaptures(params []config.Param, captures []parser.Captures) (filtered []parser.Captures, err error) {
	for _, capture := range captures {

		// filter captures not matching rules
		shouldIgnore := false
		for _, param := range params {
			if param.StringMatch != "" {
				if capture[param.BuildFullName()].Content() != param.StringMatch {
					shouldIgnore = true
					break
				}
			}

			if param.RegexMatch != "" {
				// todo regex compilation must happen as prestep
				regex, err := regexp.Compile(param.RegexMatch)
				if err != nil {
					return nil, err
				}

				paramContent := capture[param.BuildFullName()].Content()

				if !regex.Match([]byte(paramContent)) {
					shouldIgnore = true
					break
				}

			}
		}

		if shouldIgnore {
			continue
		}

		filtered = append(filtered, capture)
	}

	return filtered, err
}

func (detector *Detector) extractArguments(language string, node *parser.Node, idGenerator nodeid.Generator, variableReconciliation *parserdatatype.ReconciliationRequest) (map[parser.NodeID]*schemadatatype.DataType, error) {
	switch language {
	case "ruby":
		return detector.Ruby.ExtractArguments(node, idGenerator, variableReconciliation)
	case "sql":
		return detector.Sql.ExtractArguments(node, idGenerator, variableReconciliation)
	}
	return nil, errors.New("unsupported language")
}

func (detector *Detector) buildReconciliationRequest(language string, tree *parser.Tree) (*parserdatatype.ReconciliationRequest, error) {
	switch language {
	case "ruby":
		allDatatypes := rubydatatype.Discover(tree.RootNode(), detector.idGenerator)
		scopedDatatypes := parserdatatype.ScopeDatatypes(allDatatypes, detector.idGenerator, rubydatatype.ScopeTerminators)
		return &parserdatatype.ReconciliationRequest{
			ScopedDatatypes:  scopedDatatypes,
			ScopeTerminators: rubydatatype.ScopeTerminators,
		}, nil
	case "sql":
		return nil, nil
	}
	return nil, errors.New("unsupported language")
}

func getLanguage(input string) *sitter.Language {
	switch input {
	case "ruby":
		return ruby.GetLanguage()
	case "sql":
		return sql.GetLanguage()
	}
	return nil
}
