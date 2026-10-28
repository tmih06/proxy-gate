package opencode

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/relayconvert"
	"github.com/QuantumNous/new-api/relaykit/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openCodeRelayInfo(format types.RelayFormat) *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType:    constant.ChannelTypeOpenCode,
			ChannelBaseUrl: "https://opencode.ai/inference",
			ApiKey:         "oc_sk_test",
		},
		RelayFormat: format,
	}
}

func TestOpenCodeChannelRegistration(t *testing.T) {
	apiType, ok := common.ChannelType2APIType(constant.ChannelTypeOpenCode)
	require.True(t, ok)
	assert.Equal(t, constant.APITypeOpenCode, apiType)
	assert.Equal(t, "OpenCode", constant.GetChannelTypeName(constant.ChannelTypeOpenCode))
	assert.Equal(t, "https://opencode.ai/inference", constant.GetChannelBaseURL(constant.ChannelTypeOpenCode))

	endpoints := common.GetEndpointTypesByChannelType(constant.ChannelTypeOpenCode, "glm-5.3")
	assert.Contains(t, endpoints, constant.EndpointTypeOpenAI)
	assert.Contains(t, endpoints, constant.EndpointTypeOpenAIResponse)
	assert.Contains(t, endpoints, constant.EndpointTypeAnthropic)
	assert.Contains(t, endpoints, constant.EndpointTypeGemini)
}

func TestGetRequestURLRoutesToOpenCodeAPIFamilies(t *testing.T) {
	adaptor := &Adaptor{}
	tests := []struct {
		name      string
		info      *relaycommon.RelayInfo
		wantURL   string
		wantError bool
	}{
		{
			name: "openai chat completions",
			info: func() *relaycommon.RelayInfo {
				info := openCodeRelayInfo(types.RelayFormatOpenAI)
				info.RequestURLPath = "/v1/chat/completions"
				return info
			}(),
			wantURL: "https://opencode.ai/inference/openai/v1/chat/completions",
		},
		{
			name: "openai responses",
			info: func() *relaycommon.RelayInfo {
				info := openCodeRelayInfo(types.RelayFormatOpenAIResponses)
				info.RelayMode = relayconstant.RelayModeResponses
				info.RequestURLPath = "/v1/responses"
				return info
			}(),
			wantURL: "https://opencode.ai/inference/openai/v1/responses",
		},
		{
			name: "anthropic messages",
			info: func() *relaycommon.RelayInfo {
				info := openCodeRelayInfo(types.RelayFormatClaude)
				return info
			}(),
			wantURL: "https://opencode.ai/inference/anthropic/v1/messages",
		},
		{
			name: "gemini generate content",
			info: func() *relaycommon.RelayInfo {
				info := openCodeRelayInfo(types.RelayFormatGemini)
				info.RelayMode = relayconstant.RelayModeGemini
				info.UpstreamModelName = "gemini-4-argon"
				return info
			}(),
			wantURL: "https://opencode.ai/inference/google/v1beta/models/gemini-4-argon:generateContent",
		},
		{
			name: "unsupported compaction endpoint",
			info: func() *relaycommon.RelayInfo {
				info := openCodeRelayInfo(types.RelayFormatOpenAIResponsesCompaction)
				return info
			}(),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := adaptor.GetRequestURL(tt.info)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, got)
		})
	}
}

func TestSetupRequestHeaderUsesBearerForEveryFamily(t *testing.T) {
	adaptor := &Adaptor{}
	tests := []struct {
		name   string
		format types.RelayFormat
	}{
		{name: "openai", format: types.RelayFormatOpenAI},
		{name: "responses", format: types.RelayFormatOpenAIResponses},
		{name: "anthropic", format: types.RelayFormatClaude},
		{name: "gemini", format: types.RelayFormatGemini},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			header := make(http.Header)

			require.NoError(t, adaptor.SetupRequestHeader(c, &header, openCodeRelayInfo(tt.format)))
			assert.Equal(t, "Bearer oc_sk_test", header.Get("Authorization"))
			assert.Empty(t, header.Get("x-api-key"))
			assert.Empty(t, header.Get("x-goog-api-key"))
			if tt.format == types.RelayFormatClaude {
				assert.NotEmpty(t, header.Get("anthropic-version"))
			}
		})
	}
}

func TestOpenCodeUsageFieldsUnmarshalAcrossFamilies(t *testing.T) {
	t.Run("openai chat", func(t *testing.T) {
		var response dto.OpenAITextResponse
		require.NoError(t, common.Unmarshal([]byte(`{
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 5,
				"total_tokens": 15,
				"cost": 0.0012,
				"opencode": {
					"billing_source": "go",
					"list_cost": 0.0015,
					"allowance": 0.0012,
					"tokens": {"input": 10, "output": 5, "cache_read": 2, "cache_write_5m": 1}
				}
			}
		}`), &response))
		require.NotNil(t, response.Usage.OpenCode)
		assert.Equal(t, "go", response.Usage.OpenCode.BillingSource)
		assert.InDelta(t, 0.0012, response.Usage.OpenCode.Allowance, 0.000001)
		require.NotNil(t, response.Usage.OpenCode.Tokens)
		assert.Equal(t, 2, response.Usage.OpenCode.Tokens.CacheRead)
		cost, ok := response.Usage.Cost.(float64)
		require.True(t, ok)
		assert.InDelta(t, 0.0012, cost, 0.000001)
	})

	t.Run("anthropic", func(t *testing.T) {
		var response dto.ClaudeResponse
		require.NoError(t, common.Unmarshal([]byte(`{
			"usage": {
				"input_tokens": 10,
				"output_tokens": 5,
				"cost": 0.002,
				"opencode": {"billing_source": "go", "allowance": 0.002}
			}
		}`), &response))
		require.NotNil(t, response.Usage)
		require.NotNil(t, response.Usage.OpenCode)
		assert.Equal(t, "go", response.Usage.OpenCode.BillingSource)
	})

	t.Run("gemini", func(t *testing.T) {
		var response dto.GeminiChatResponse
		require.NoError(t, common.Unmarshal([]byte(`{
			"usageMetadata": {
				"promptTokenCount": 10,
				"candidatesTokenCount": 5,
				"totalTokenCount": 15,
				"cost": 0.003,
				"opencode": {"billing_source": "go", "allowance": 0.003}
			}
		}`), &response))
		require.NotNil(t, response.UsageMetadata.OpenCode)
		assert.Equal(t, "go", response.UsageMetadata.OpenCode.BillingSource)
	})
}

func TestConvertOpenAIRequestPreservesStreamOptions(t *testing.T) {
	adaptor := &Adaptor{}
	info := openCodeRelayInfo(types.RelayFormatOpenAI)
	request := &dto.GeneralOpenAIRequest{
		Model:         "glm-5.3",
		Stream:        common.GetPointer(true),
		StreamOptions: &dto.StreamOptions{IncludeUsage: true},
	}

	converted, err := adaptor.ConvertOpenAIRequest(nil, info, request)
	require.NoError(t, err)
	openAIRequest, ok := converted.(*dto.GeneralOpenAIRequest)
	require.True(t, ok)
	require.NotNil(t, openAIRequest.StreamOptions)
	assert.True(t, openAIRequest.StreamOptions.IncludeUsage)
}

func TestDoResponseReturnsOpenCodeUsage(t *testing.T) {
	adaptor := &Adaptor{}
	info := openCodeRelayInfo(types.RelayFormatOpenAI)
	info.UpstreamModelName = "glm-5.3"
	adaptor.Init(info)

	body := `{
		"id": "chatcmpl-opencode",
		"model": "glm-5.3",
		"choices": [],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 5,
			"total_tokens": 15,
			"cost": 0.0012,
			"opencode": {"billing_source": "go", "allowance": 0.0012}
		}
	}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	usageAny, apiErr := adaptor.DoResponse(c, resp, info)
	require.Nil(t, apiErr)
	usage, ok := usageAny.(*dto.Usage)
	require.True(t, ok)
	require.NotNil(t, usage.OpenCode)
	assert.Equal(t, "go", usage.OpenCode.BillingSource)
	assert.InDelta(t, 0.0012, usage.OpenCode.Allowance, 0.000001)
}

func TestOpenCodeUsageSurvivesUsageConverters(t *testing.T) {
	t.Run("anthropic", func(t *testing.T) {
		mapped := relayconvert.UsageFromClaudeAPIUsage(&dto.ClaudeUsage{
			InputTokens:  10,
			OutputTokens: 5,
			Cost:         0.002,
			OpenCode:     &dto.OpenCodeUsage{BillingSource: "go", Allowance: 0.002},
		})
		require.NotNil(t, mapped)
		require.NotNil(t, mapped.OpenCode)
		assert.Equal(t, "go", mapped.OpenCode.BillingSource)
	})

	t.Run("gemini", func(t *testing.T) {
		mapped := relayconvert.UsageFromGeminiMetadata(&dto.GeminiUsageMetadata{
			PromptTokenCount:     10,
			CandidatesTokenCount: 5,
			TotalTokenCount:      15,
			Cost:                 0.003,
			OpenCode:             &dto.OpenCodeUsage{BillingSource: "go", Allowance: 0.003},
		}, 0)
		require.NotNil(t, mapped)
		require.NotNil(t, mapped.OpenCode)
		assert.Equal(t, "go", mapped.OpenCode.BillingSource)
	})
}

func TestMergeUsageNonZeroKeepsOpenCodeAccounting(t *testing.T) {
	current := &dto.Usage{}
	incoming := &dto.Usage{
		OpenCode: &dto.OpenCodeUsage{BillingSource: "go", Allowance: 0.004},
	}
	merged := dto.MergeUsageNonZero(current, incoming)
	require.NotNil(t, merged.OpenCode)
	assert.Equal(t, 0.004, merged.OpenCode.Allowance)
}
