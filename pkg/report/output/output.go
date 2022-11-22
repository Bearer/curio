package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bearer/curio/pkg/commands/process/settings"
	"github.com/bearer/curio/pkg/flag"
	"github.com/bearer/curio/pkg/report/output/dataflow"
	"github.com/bearer/curio/pkg/report/output/detectors"
	"github.com/bearer/curio/pkg/report/output/policies"
	"github.com/bearer/curio/pkg/report/output/stats"
	"github.com/bearer/curio/pkg/types"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"

	"github.com/rs/zerolog"
)

var ErrUndefinedFormat = errors.New("undefined output format")

func ReportJSON(report types.Report, output *zerolog.Event, config settings.Config) error {
	outputDetections, err := getReportOutput(report, config)
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(&outputDetections)
	if err != nil {
		return fmt.Errorf("failed to json marshal detections: %w", err)
	}

	output.Msg(string(jsonBytes))

	return nil
}

func ReportYAML(report types.Report, output *zerolog.Event, config settings.Config) error {
	ouputDetections, err := getReportOutput(report, config)
	if err != nil {
		return err
	}

	jsonBytes, err := yaml.Marshal(&ouputDetections)
	if err != nil {
		return fmt.Errorf("failed to json marshal detections: %w", err)
	}

	output.Msg(string(jsonBytes))

	return nil
}

func ReportTable(report types.Report, output *zerolog.Event, config settings.Config) error {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	outputDetections, err := getReportOutput(report, config)
	if err != nil {
		return err
	}

	for _, detection := range outputDetections {

	}

	table.SetHeader([]string{"Name", "Sign", "Rating"})

	table.Append([]string{"A", "The Good", "500"})
	table.Append([]string{"B", "The Very very Bad Man", "288"})
	table.Append([]string{"C", "The Ugly", "120"})
	table.Append([]string{"D", "The Gopher", "800"})

	table.Render()

	output.Msg(tableString.String())

	return nil
}

func getReportOutput(report types.Report, config settings.Config) (any, error) {
	if config.Report.Report == flag.ReportDetectors {
		return detectors.GetOutput(report)
	} else if config.Report.Report == flag.ReportDataFlow {
		detections, err := detectors.GetOutput(report)
		if err != nil {
			return nil, err
		}

		return dataflow.GetOutput(detections, config)

	} else if config.Report.Report == flag.ReportPolicies {
		detections, err := detectors.GetOutput(report)
		if err != nil {
			return nil, err
		}

		dataflow, err := dataflow.GetOutput(detections, config)
		if err != nil {
			return nil, err
		}

		return policies.GetOutput(dataflow, config)
	} else if config.Report.Report == flag.ReportStats {
		lineOfCodeOutput, err := stats.GoclocDetectorOutput(config.Scan.Target)
		if err != nil {
			return nil, err
		}

		detectorsOutput, err := detectors.GetOutput(report)
		if err != nil {
			return nil, err
		}

		dataflowOutput, err := dataflow.GetOutput(detectorsOutput, config)
		if err != nil {
			return nil, err
		}

		return stats.GetOutput(lineOfCodeOutput, dataflowOutput, config)
	}

	return nil, ErrUndefinedFormat
}
