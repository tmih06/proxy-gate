package opencode

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/claude"
	"github.com/QuantumNous/new-api/relay/channel/gemini"
	"github.com/QuantumNous/new-api/relay/channel/openai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"

	"github.com/gin-gonic/gin"
)

// Adaptor routes an OpenCode request to the upstream API family that serves
// the requested model. OpenCode exposes separate OpenAI, Anthropic, and Gemini
// compatible paths under https://opencode.ai/inference.
type Adaptor struct {
	openai openai.Adaptor
	claude claude.Adaptor
	gemini gemini.Adaptor
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	a.openai.Init(info)
	a.claude.Init(info)
	a.gemini.Init(info)
}

func (a *Adaptor) family(info *relaycommon.RelayInfo) types.RelayFormat {
	if info == nil {
		return types.RelayFormatOpenAI
	}
	format := info.GetFinalRequestRelayFormat()
	// Some callers keep RelayFormatOpenAI while switching RelayMode to
	// Responses (for example the chat-completions-via-Responses path).
	if format == types.RelayFormatOpenAI && info.RelayMode == relayconstant.RelayModeResponses {
		return types.RelayFormatOpenAIResponses
	}
	return format
}

func openCodeBaseURL(info *relaycommon.RelayInfo) string {
	if info == nil {
		return constant.GetChannelBaseURL(constant.ChannelTypeOpenCode)
	}
	baseURL := strings.TrimRight(strings.TrimSpace(info.ChannelBaseUrl), "/")
	if baseURL == "" {
		baseURL = constant.GetChannelBaseURL(constant.ChannelTypeOpenCode)
	}
	return baseURL
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	baseURL := openCodeBaseURL(info)
	cloned := *info
	switch a.family(info) {
	case types.RelayFormatClaude:
		cloned.ChannelBaseUrl = baseURL + "/anthropic"
		return a.claude.GetRequestURL(&cloned)
	case types.RelayFormatGemini:
		cloned.ChannelBaseUrl = baseURL + "/google"
		return a.gemini.GetRequestURL(&cloned)
	case types.RelayFormatOpenAI, types.RelayFormatOpenAIResponses:
		cloned.ChannelBaseUrl = baseURL + "/openai"
		return a.openai.GetRequestURL(&cloned)
	default:
		return "", fmt.Errorf("opencode channel: unsupported relay format %q", a.family(info))
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, header *http.Header, info *relaycommon.RelayInfo) error {
	switch a.family(info) {
	case types.RelayFormatClaude:
		if err := a.claude.SetupRequestHeader(c, header, info); err != nil {
			return err
		}
		// OpenCode's Anthropic-compatible endpoint authenticates with a
		// bearer token. The Claude adaptor also sets x-api-key, which is not
		// one of OpenCode's accepted credentials.
		header.Del("x-api-key")
	case types.RelayFormatGemini:
		if err := a.gemini.SetupRequestHeader(c, header, info); err != nil {
			return err
		}
		// OpenCode's Gemini-compatible endpoint authenticates with a bearer
		// token, not x-goog-api-key.
		header.Del("x-goog-api-key")
	case types.RelayFormatOpenAI, types.RelayFormatOpenAIResponses:
		if err := a.openai.SetupRequestHeader(c, header, info); err != nil {
			return err
		}
	default:
		return fmt.Errorf("opencode channel: unsupported relay format %q", a.family(info))
	}
	header.Set("Authorization", "Bearer "+info.ApiKey)
	return nil
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	return a.openai.ConvertOpenAIRequest(c, info, request)
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	return a.claude.ConvertClaudeRequest(c, info, request)
}

func (a *Adaptor) ConvertGeminiRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeminiChatRequest) (any, error) {
	return a.gemini.ConvertGeminiRequest(c, info, request)
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	return a.openai.ConvertOpenAIResponsesRequest(c, info, request)
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, errors.New("opencode channel: rerank endpoint not supported")
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	return nil, errors.New("opencode channel: embeddings endpoint not supported")
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	return nil, errors.New("opencode channel: audio endpoint not supported")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	return nil, errors.New("opencode channel: image endpoint not supported")
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (any, *types.NewAPIError) {
	switch a.family(info) {
	case types.RelayFormatClaude:
		return a.claude.DoResponse(c, resp, info)
	case types.RelayFormatGemini:
		return a.gemini.DoResponse(c, resp, info)
	case types.RelayFormatOpenAI, types.RelayFormatOpenAIResponses:
		return a.openai.DoResponse(c, resp, info)
	default:
		return nil, types.NewError(fmt.Errorf("opencode channel: unsupported relay format %q", a.family(info)), types.ErrorCodeInvalidRequest)
	}
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
