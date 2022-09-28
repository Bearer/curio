package csharp_test

import (
	"path/filepath"
	"testing"

	"github.com/bradleyjkemp/cupaloy"

	"github.com/bearer/curio/pkg/detectors"
	"github.com/bearer/curio/pkg/detectors/csharp"
	"github.com/bearer/curio/pkg/parser/nodeid"

	"github.com/bearer/curio/pkg/detectors/internal/testhelper"
	detectortypes "github.com/bearer/curio/pkg/report/detectors"
)

const detectorType = detectortypes.DetectorCSharp

func TestDetectorReportInterfaces(t *testing.T) {
	var registrations = []detectors.InitializedDetector{{
		Type:     detectorType,
		Detector: csharp.New(&nodeid.IntGenerator{Counter: 0})}}
	detectorReport := testhelper.Extract(t, filepath.Join("testdata", "project"), registrations, detectorType)

	cupaloy.SnapshotT(t, detectorReport.Detections)
}

func TestDetectorReportDataTypes(t *testing.T) {
	var registrations = []detectors.InitializedDetector{{
		Type:     detectorType,
		Detector: csharp.New(&nodeid.IntGenerator{Counter: 0})}}
	detectorReport := testhelper.Extract(t, filepath.Join("testdata", "datatypes"), registrations, detectorType)

	cupaloy.SnapshotT(t, detectorReport.Detections)
}
