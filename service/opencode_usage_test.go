package service

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendOpenCodeUsageForLog(t *testing.T) {
	other := model.NewLogOther()
	appendOpenCodeUsageForLog(other, &dto.Usage{
		Cost: 0.0012,
		OpenCode: &dto.OpenCodeUsage{
			BillingSource: "go",
			Allowance:     0.0012,
			ListCost:      0.0015,
		},
	})

	payload := other.JSONString()
	require.NotEmpty(t, payload)
	assert.Contains(t, payload, `"opencode"`)
	assert.Contains(t, payload, `"opencode_cost":0.0012`)
	assert.Contains(t, payload, `"billing_source":"go"`)
	assert.Contains(t, payload, `"allowance":0.0012`)
}
