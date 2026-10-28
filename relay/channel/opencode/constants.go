package opencode

const ChannelName = "opencode"

// ModelList is a small, administrator-facing catalog of models available on
// OpenCode Go. Channels can add any other model id exposed by
// GET https://opencode.ai/inference/v1/models.
var ModelList = []string{
	"glm-5.3-flash",
	"glm-5.3",
	"glm-5.2",
	"kimi-k3",
	"kimi-k2.7-code",
	"minimax-m3",
	"deepseek-v4.1-flash",
	"deepseek-v4-pro",
	"gpt-6-luna",
	"gpt-5.6-luna",
	"claude-haiku-5-5",
	"gemini-4-argon",
	"qwen3.8-max",
	"qwen3.8-flash",
	"grok-4.7",
}
