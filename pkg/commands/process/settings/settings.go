package settings

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"github.com/bearer/curio/pkg/flag"
	"github.com/bearer/curio/pkg/util/rego"
)

type Config struct {
	Worker         flag.WorkerOptions `mapstructure:"worker" json:"worker" yaml:"worker"`
	Scan           flag.ScanOptions   `mapstructure:"scan" json:"scan" yaml:"scan"`
	Report         flag.ReportOptions `mapstructure:"report" json:"report" yaml:"report"`
	CustomDetector map[string]Rule    `mapstructure:"custom_detector" json:"custom_detector" yaml:"custom_detector"`
	Policies       map[string]*Policy `mapstructure:"policies" json:"policies" yaml:"policies"`
	Target         string             `mapstructure:"target" json:"target" yaml:"target"`
}

type PolicyLevel string

var LevelCritical = "critical"
var LevelHigh = "high"
var LevelMedium = "medium"
var LevelLow = "low"

type Modules []*PolicyModule

type Policy struct {
	Query string `mapstructure:"query" json:"query" yaml:"query"`
	Id    string `mapstructure:"id" json:"id" yaml:"id"`
	// DisplayId   string      `mapstructure:"display_id" json:"display_id" yaml:"display_id"`
	Name        string      `mapstructure:"name" json:"name" yaml:"name"`
	Description string      `mapstructure:"description" json:"description" yaml:"description"`
	Level       PolicyLevel `mapstructure:"level" json:"level" yaml:"level"`
	Modules     Modules     `mapstructure:"modules" json:"modules" yaml:"modules"`
}

type PolicyModule struct {
	Path    string `mapstructure:"path" json:"path,omitempty" yaml:"path,omitempty"`
	Name    string `mapstructure:"name" json:"name" yaml:"name"`
	Content string `mapstructure:"content" json:"content" yaml:"content"`
}

func (modules Modules) ToRegoModules() (output []rego.Module) {
	for _, module := range modules {
		output = append(output, rego.Module{
			Name:    module.Name,
			Content: module.Content,
		})
	}
	return
}

type PatternFilter struct {
	Variable       string   `mapstructure:"variable" json:"variable" yaml:"variable"`
	Values         []string `mapstructure:"values" json:"values" yaml:"values"`
	Minimum        *int     `mapstructure:"minimum" json:"minimum" yaml:"minimum"`
	Maximum        *int     `mapstructure:"maximum" json:"maximum" yaml:"maximum"`
	MatchViolation bool     `mapstructure:"match_violation" json:"match_violation" yaml:"match_violation"`
}

type RulePattern struct {
	Pattern string          `mapstructure:"pattern" json:"pattern" yaml:"pattern"`
	Filters []PatternFilter `mapstructure:"filters" json:"filters" yaml:"filters"`
}

type Rule struct {
	Id             string        `mapstructure:"id" json:"id" yaml:"id"`
	Disabled       bool          `mapstructure:"disabled" json:"disabled" yaml:"disabled"`
	Type           string        `mapstructure:"type" json:"type" yaml:"type"`
	Languages      []string      `mapstructure:"languages" json:"languages" yaml:"languages"`
	ParamParenting bool          `mapstructure:"param_parenting" json:"param_parenting" yaml:"param_parenting"`
	Patterns       []RulePattern `mapstructure:"patterns" json:"patterns" yaml:"patterns"`

	RootSingularize bool `mapstructure:"root_singularize" yaml:"root_singularize" `
	RootLowercase   bool `mapstructure:"root_lowercase" yaml:"root_lowercase"`

	Metavars       map[string]MetaVar `mapstructure:"metavars" json:"metavars" yaml:"metavars"`
	Stored         bool               `mapstructure:"stored" json:"stored" yaml:"stored"`
	DetectPresence bool               `mapstructure:"detect_presence" json:"detect_presence" yaml:"detect_presence"`
	OmitParent     bool               `mapstructure:"omit_parent" json:"omit_parent" yaml:"omit_parent"`
}

type Processor struct {
	Query   string  `mapstructure:"query" json:"query" yaml:"query"`
	Modules Modules `mapstructure:"modules" json:"modules" yaml:"modules"`
}

type MetaVar struct {
	Input  string `mapstructure:"input" json:"input" yaml:"input"`
	Output int    `mapstructure:"output" json:"output" yaml:"output"`
	Regex  string `mapstructure:"regex" json:"regex" yaml:"regex"`
}

//go:embed policies/*
var policiesFs embed.FS

//go:embed processors/*
var processorsFs embed.FS

var CustomDetectorKey string = "scan.custom_detector"
var PoliciesKey string = "scan.policies"

func FromOptions(opts flag.Options) (Config, error) {
	detectors := DefaultCustomDetector()

	policies := DefaultPolicies()

	// validate detector options
	onlyDetector := opts.DetectorOptions.OnlyDetector
	skipDetector := opts.DetectorOptions.SkipDetector

	// TODO: fix these and use detector ID
	validDetectors := make(map[string]bool)
	for key := range detectors {
		validDetectors[key] = true
	}

	var invalidDetectors []string
	for key := range onlyDetector {
		if !validDetectors[key] {
			invalidDetectors = append(invalidDetectors, key)
		}
	}

	for key := range skipDetector {
		if !validDetectors[key] {
			invalidDetectors = append(invalidDetectors, key)
		}
	}

	if len(invalidDetectors) > 0 {
		return Config{}, fmt.Errorf("unknown detectors %s", invalidDetectors)
	}

	// apply detector options
	for key := range detectors {
		if len(onlyDetector) > 0 && !onlyDetector[key] {
			delete(detectors, key)
			continue
		}

		if skipDetector[key] {
			delete(detectors, key)
			continue
		}
	}

	externalDetectors, err := LoadExternalDetectors(opts.ExternalDetectorDir)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load external detectors %w", err)
	}

	for ruleName, rule := range externalDetectors {
		_, ok := detectors[ruleName]
		if ok {
			return Config{}, fmt.Errorf("tried to overwrite default custom detector %s with external detector", ruleName)
		}

		detectors[ruleName] = rule
	}

	// validate policy options
	onlyPolicy := opts.PolicyOptions.OnlyPolicy
	skipPolicy := opts.PolicyOptions.SkipPolicy

	var invalidPolicyIds []string
	for key := range onlyPolicy {
		if policies[key] == nil {
			invalidPolicyIds = append(invalidPolicyIds, key)
		}
	}

	for key := range skipPolicy {
		if policies[key] == nil {
			invalidPolicyIds = append(invalidPolicyIds, key)
		}
	}

	if len(invalidPolicyIds) > 0 {
		return Config{}, fmt.Errorf("unknown policy ids %s", invalidPolicyIds)
	}

	// apply policy options
	for key := range policies {
		policy := policies[key]

		if len(onlyPolicy) > 0 && !onlyPolicy[policy.Id] {
			delete(policies, key)
			continue
		}

		if skipPolicy[policy.Id] {
			delete(policies, key)
			continue
		}

		for _, module := range policy.Modules {
			if module.Path != "" {
				content, err := policiesFs.ReadFile(module.Path)
				if err != nil {
					return Config{}, err
				}
				module.Content = string(content)
			}
		}
	}

	// apply external policies
	externalPolicies, err := LoadExternalPolicies(opts.ExternalPolicyDir)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load external policies %w", err)
	}

	for policyName, policy := range externalPolicies {
		_, ok := policies[policyName]
		if ok {
			return Config{}, fmt.Errorf("tried to overwrite default policy %s with external detector", policyName)
		}

		policies[policyName] = policy
	}

	config := Config{
		Worker:         opts.WorkerOptions,
		CustomDetector: detectors,
		Scan:           opts.ScanOptions,
		Report:         opts.ReportOptions,
		Policies:       policies,
	}

	return config, nil
}

func (rulePattern *RulePattern) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try to parse as a string
	var pattern string
	if err := unmarshal(&pattern); err == nil {
		rulePattern.Pattern = pattern
		return nil
	}

	// Wasn't a string so it must be the structured format
	type rawRulePattern RulePattern
	return unmarshal((*rawRulePattern)(rulePattern))
}

func DefaultCustomDetector() map[string]Rule {
	policiesDir := "policies"
	rules := make(map[string]Rule)

	// policies dir
	dirEntries, err := policiesFs.ReadDir(policiesDir)
	if err != nil {
		log.Fatal().Msgf("failed to read policies dir %e", err)
	}

	// each policy
	for _, entry := range dirEntries {
		policyId := entry.Name()
		if !strings.HasPrefix(policyId, "CR-") {
			// not an actual policy dir
			continue
		}

		policyDir := policiesDir + "/" + policyId
		policyDirEntries, err := policiesFs.ReadDir(policyDir)
		if err != nil {
			log.Fatal().Msgf("failed to read policy dir %s %e", policyDir, err)
		}

		// each language
		for _, entry := range policyDirEntries {
			language := entry.Name()
			if filepath.Ext(language) != "" {
				// is a file not a folder
				continue
			}

			customDetectorFile := policyDir + "/" + language + "/detector.yml"
			customDetector, err := policiesFs.ReadFile(customDetectorFile)
			if err != nil {
				log.Fatal().Msgf("failed to read custom detector file %s %e", customDetectorFile, err)
			}

			var rule Rule
			err = yaml.Unmarshal(customDetector, &rule)
			if err != nil {
				log.Fatal().Msgf("failed to unmarshal custom detector file %s %e", customDetectorFile, err)
			}

			rule.Id = policyId
			rules[rule.Id] = rule
		}
	}

	return rules
}

func DefaultPolicies() map[string]*Policy {
	policies := make(map[string]*Policy)
	policiesDir := "policies"

	// policies dir
	dirEntries, err := policiesFs.ReadDir(policiesDir)
	if err != nil {
		log.Fatal().Msgf("failed to read policies dir %e", err)
	}

	// each policy
	for _, entry := range dirEntries {
		policyId := entry.Name()
		if !strings.HasPrefix(policyId, "CR-") {
			// not an actual policy dir
			continue
		}

		policyFilename := policiesDir + "/" + policyId + "/rule.yml"
		policyFile, err := policiesFs.ReadFile(policyFilename)
		if err != nil {
			log.Fatal().Msgf("failed to read policy file %s %e", policyFilename, err)
		}

		var policy Policy
		err = yaml.Unmarshal(policyFile, &policy)
		if err != nil {
			log.Fatal().Msgf("failed to unmarshal policy file %s %e", policyFilename, err)
		}

		policy.Id = policyId
		policies[policy.Id] = &policy
	}

	return policies
}

func ProcessorRegoModuleText(processorName string) (string, error) {
	processorPath := fmt.Sprintf("processors/%s.rego", processorName)
	data, err := processorsFs.ReadFile(processorPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
