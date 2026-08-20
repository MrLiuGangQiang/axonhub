package responses

import (
	"strings"

	"github.com/tidwall/gjson"
)

// ResponsesLiteAllTurnsSupported reports whether a model accepts the all_turns
// reasoning context that Responses Lite requires. Some Codex models only support
// "auto" and "current_turn", so sending all_turns to them causes an upstream 400.
func ResponsesLiteAllTurnsSupported(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "gpt-5.3-codex-spark":
		return false
	default:
		return true
	}
}

// ResponsesLiteSupportsTools reports whether a raw Responses request body's tools
// are compatible with Responses Lite. Responses Lite only accepts function tools,
// custom tools, and client-executed web search; other tool types (image_generation,
// namespace, google, server-executed web_search, etc.) cause an upstream HTTP 400.
func ResponsesLiteSupportsTools(body []byte) bool {
	if len(body) == 0 {
		return true
	}

	tools := gjson.GetBytes(body, "tools")
	if !tools.Exists() || !tools.IsArray() {
		return true
	}

	supported := true
	tools.ForEach(func(_, tool gjson.Result) bool {
		switch strings.ToLower(strings.TrimSpace(tool.Get("type").String())) {
		case "function", "custom":
			// always supported
		case "web_search":
			// only client-executed web search is supported by Responses Lite
			if !strings.EqualFold(strings.TrimSpace(tool.Get("execution_mode").String()), "client") {
				supported = false
			}
		default:
			supported = false
		}
		return supported
	})
	return supported
}
