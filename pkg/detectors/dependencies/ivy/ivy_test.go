package ivy_test

import (
	"path/filepath"
	"testing"

	"github.com/bearer/curio/pkg/detectors/internal/testhelper"
	"github.com/bearer/curio/pkg/report/detectors"
	"github.com/bradleyjkemp/cupaloy"
)

const detectorType = detectors.DetectorDependencies

var registrations = testhelper.RegistrationFor(detectorType)

func TestDependenciesReport(t *testing.T) {
	report := testhelper.Extract(t, filepath.Join("testdata"), registrations, detectorType)
	cupaloy.SnapshotT(t, report.Dependencies)
}
