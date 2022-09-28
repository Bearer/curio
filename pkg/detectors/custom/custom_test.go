package custom_test

import (
	_ "embed"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"gopkg.in/yaml.v2"

	"github.com/bearer/curio/pkg/detectors"
	"github.com/bearer/curio/pkg/detectors/custom"
	"github.com/bearer/curio/pkg/detectors/custom/config"
	"github.com/bearer/curio/pkg/parser/nodeid"

	"github.com/bearer/curio/pkg/detectors/internal/testhelper"
	detectortypes "github.com/bearer/curio/pkg/report/detectors"
)

const detectorType = detectortypes.DetectorCustom

//go:embed testdata/config/ruby_loggers.yml
var configRubyLoggers []byte

//go:embed testdata/config/rails_encrypts.yml
var configRailsEncrypts []byte

//go:embed testdata/config/sql_create_table.yml
var configSQLCreateTable []byte

func TestRubyLoggersJSON(t *testing.T) {
	var rulesConfig config.Config

	detector := custom.New(&nodeid.IntGenerator{Counter: 0})
	err := yaml.Unmarshal(configRubyLoggers, &rulesConfig)
	if err != nil {
		t.Fatal(err)
		return
	}
	customDetector := detector.(*custom.Detector)
	err = customDetector.CompileRules(rulesConfig)
	if err != nil {
		t.Fatal(err)
	}

	var registrations = []detectors.InitializedDetector{{
		Type:     detectorType,
		Detector: detector}}
	detectorReport := testhelper.Extract(t, filepath.Join("testdata", "ruby", "loggers"), registrations, detectorType)

	bytes, err := json.MarshalIndent(detectorReport.CustomDetections, "", "\t")

	if err != nil {
		t.Fatal(err)
	}

	cupaloy.SnapshotT(t, string(bytes))
}

func TestRailsEncryptsJSON(t *testing.T) {
	var rulesConfig config.Config

	detector := custom.New(&nodeid.IntGenerator{Counter: 0})
	err := yaml.Unmarshal(configRailsEncrypts, &rulesConfig)
	if err != nil {
		t.Fatal(err)
		return
	}
	customDetector := detector.(*custom.Detector)
	err = customDetector.CompileRules(rulesConfig)
	if err != nil {
		t.Fatal(err)
	}

	var registrations = []detectors.InitializedDetector{{
		Type:     detectorType,
		Detector: detector}}
	detectorReport := testhelper.Extract(t, filepath.Join("testdata", "rails", "encrypts"), registrations, detectorType)

	bytes, err := json.MarshalIndent(detectorReport.CustomDetections, "", "\t")

	if err != nil {
		t.Fatal(err)
	}

	cupaloy.SnapshotT(t, string(bytes))
}

func TestSQLCreateTableJSON(t *testing.T) {
	var rulesConfig config.Config

	detector := custom.New(&nodeid.IntGenerator{Counter: 0})
	err := yaml.Unmarshal(configSQLCreateTable, &rulesConfig)
	if err != nil {
		t.Fatal(err)
		return
	}
	customDetector := detector.(*custom.Detector)
	err = customDetector.CompileRules(rulesConfig)
	if err != nil {
		t.Fatal(err)
	}

	var registrations = []detectors.InitializedDetector{{
		Type:     detectorType,
		Detector: detector}}
	detectorReport := testhelper.Extract(t, filepath.Join("testdata", "sql", "create_table"), registrations, detectorType)

	bytes, err := json.MarshalIndent(detectorReport.CustomDetections, "", "\t")

	if err != nil {
		t.Fatal(err)
	}

	cupaloy.SnapshotT(t, string(bytes))
}
