package policies

import (
	"context"

	"github.com/bearer/curio/pkg/classification/db"
	"github.com/bearer/curio/pkg/commands/process/settings"
	"github.com/bearer/curio/pkg/report/output/dataflow"
	"github.com/open-policy-agent/opa/rego"
	"github.com/rs/zerolog/log"
)

type PolicyInput struct {
	PolicyName        string             `json:"policy_name"`
	PolicyId          string             `json:"policy_id"`
	PolicyDescription string             `json:"policy_description"`
	Dataflow          *dataflow.DataFlow `json:"dataflow"`
	DataCategories    []db.DataCategory  `json:"data_categories"`
}

func GetOutput(dataflow *dataflow.DataFlow, config settings.Config) ([]rego.Vars, error) {
	ctx := context.TODO()

	var result []rego.Vars

	for _, policy := range config.Policies {
		options := []func(r *rego.Rego){rego.Query(policy.Query)}
		for _, module := range policy.Modules {
			options = append(options, rego.Module(module.Name, module.Content))
		}

		r := rego.New(options...)
		query, err := r.PrepareForEval(ctx)
		if err != nil {
			return nil, err
		}

		// Create a prepared query that can be evaluated.
		rs, err := query.Eval(
			ctx,
			rego.EvalInput(
				PolicyInput{
					PolicyName:        policy.Name,
					PolicyDescription: policy.Description,
					PolicyId:          policy.Id,
					Dataflow:          dataflow,
					DataCategories:    db.Default().DataCategories,
				},
			),
		)
		if err != nil {
			return nil, err
		}

		log.Debug().Msgf("result %#v", rs)

		if len(rs) > 0 {
			result = append(result, rs[0].Bindings)
		}
	}

	return result, nil
}
