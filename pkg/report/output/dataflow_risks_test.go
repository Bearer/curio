package output_test

import (
	"os"
	"testing"

	"github.com/bearer/curio/pkg/commands/process/settings"
	"github.com/bearer/curio/pkg/report/output"
	"github.com/bearer/curio/pkg/report/output/dataflow"
	"github.com/bearer/curio/pkg/report/output/dataflow/types"
	globaltypes "github.com/bearer/curio/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestDataflowRisks(t *testing.T) {
	config := settings.Config{
		CustomDetector: map[string]settings.Rule{
			"ruby_leak": {
				Stored: true,
			},
		},
	}
	testCases := []struct {
		Name        string
		Config      settings.Config
		FileContent string
		Want        []types.RiskDetector
	}{
		{
			Name:        "single detection",
			Config:      config,
			FileContent: `{"type": "custom_classified", "detector_type":"rails_leak", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}`,
			Want: []types.RiskDetector{
				{
					DetectorID: "rails_leak",
					DataTypes: []types.RiskDatatype{
						{
							Name:   "Username",
							Stored: false,
							Locations: []types.RiskLocation{
								{Filename: "./users.rb", LineNumber: 25},
							},
						},
					},
				},
			},
		},
		{
			Name:   "single detection - duplicates",
			Config: config,
			FileContent: `{"type": "custom_classified", "detector_type":"rails_leak", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}
{"type": "custom_classified", "detector_type":"rails_leak", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}`,
			Want: []types.RiskDetector{
				{
					DetectorID: "rails_leak",
					DataTypes: []types.RiskDatatype{
						{
							Name:   "Username",
							Stored: false,
							Locations: []types.RiskLocation{
								{Filename: "./users.rb", LineNumber: 25},
							},
						},
					},
				},
			},
		},
		{
			Name:        "single detection - stored",
			Config:      config,
			FileContent: `{"type": "custom_classified", "detector_type":"ruby_leak", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}`,
			Want: []types.RiskDetector{
				{
					DetectorID: "ruby_leak",
					DataTypes: []types.RiskDatatype{
						{
							Name:   "Username",
							Stored: true,
							Locations: []types.RiskLocation{
								{Filename: "./users.rb", LineNumber: 25},
							},
						},
					},
				},
			},
		},
		{
			Name:   "single detection - multiple occurences - deterministic output",
			Config: config,
			FileContent: `{"type": "custom_classified", "detector_type":"rails_leak", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}
			{"type": "custom_classified", "detector_type":"rails_leak", "source": {"filename": "./users.rb", "line_number": 2}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}`,
			Want: []types.RiskDetector{
				{
					DetectorID: "rails_leak",
					DataTypes: []types.RiskDatatype{
						{
							Name:   "Username",
							Stored: false,
							Locations: []types.RiskLocation{
								{Filename: "./users.rb", LineNumber: 2},
								{Filename: "./users.rb", LineNumber: 25},
							},
						},
					},
				},
			},
		},
		{
			Name:   "multiple detections - same detector - deterministic output",
			Config: config,
			FileContent: `{"type": "custom_classified", "detector_type":"rails_leak", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}
{"type": "custom_classified", "detector_type":"rails_leak", "source": {"filename": "./address.rb", "line_number": 2}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Physical Address"} ,"decision":{"state": "valid"}}}}`,
			Want: []types.RiskDetector{
				{
					DetectorID: "rails_leak",
					DataTypes: []types.RiskDatatype{
						{
							Name: "Physical Address",
							Locations: []types.RiskLocation{
								{Filename: "./address.rb", LineNumber: 2},
							},
						},
						{
							Name: "Username",
							Locations: []types.RiskLocation{
								{Filename: "./users.rb", LineNumber: 25},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			file, err := os.CreateTemp("", "*test.jsonlines")
			if err != nil {
				t.Fatalf("failed to create tmp file for report %s", err)
				return
			}
			defer os.Remove(file.Name())
			_, err = file.Write([]byte(test.FileContent))
			if err != nil {
				t.Fatalf("failed to write to tmp file %s", err)
				return
			}
			file.Close()

			detections, err := output.GetDetectorsOutput(globaltypes.Report{
				Path: file.Name(),
			})
			if err != nil {
				t.Fatalf("failed to get detectors output %s", err)
				return
			}

			dataflow, err := dataflow.GetOuput(detections, test.Config)
			if err != nil {
				t.Fatalf("failed to get detectors output %s", err)
				return
			}

			assert.Equal(t, test.Want, dataflow.Risks)
		})
	}
}
