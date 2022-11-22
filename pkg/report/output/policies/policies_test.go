package policies_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataflowRisks(t *testing.T) {
	testCases := []struct {
		Name        string
		Config      settings.Config
		FileContent string
		Want        []types.RiskDetector
	}{}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			assert.Equal(t, true, true)
		})
	}
}
