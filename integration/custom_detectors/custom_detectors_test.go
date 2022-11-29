package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/bearer/curio/integration/internal/testhelper"
)

func newScanTest(language, name, filename string) testhelper.TestCase {
	arguments := []string{"scan", filepath.Join("testdata", language, filename), "--report=dataflow", "--format=yaml"}
	options := testhelper.TestCaseOptions{StartWorker: true}

	return testhelper.NewTestCase(name, arguments, options)
}

func TestCustomDetectors(t *testing.T) {
	tests := []testhelper.TestCase{
		newScanTest("ruby", "detect_rails_session", "detect_rails_session.rb"),
		newScanTest("ruby", "detect_ruby_logger", "detect_ruby_logger.rb"),
		newScanTest("ruby", "ruby_file_detection", "ruby_file_detection.rb"),
		newScanTest("ruby", "ruby_http_detection", "ruby_http_detection.rb"),
	}

	testhelper.RunTests(t, tests)
}
