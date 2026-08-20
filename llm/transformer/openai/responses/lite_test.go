package responses

import "testing"

func TestResponsesLiteAllTurnsSupported(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{model: "gpt-5", want: true},
		{model: "gpt-5.4-mini", want: true},
		{model: "gpt-5.6-sol", want: true},
		{model: "gpt-5.3-codex-spark", want: false},
		{model: "GPT-5.3-CODEX-SPARK", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			if got := ResponsesLiteAllTurnsSupported(tt.model); got != tt.want {
				t.Fatalf("ResponsesLiteAllTurnsSupported(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestResponsesLiteSupportsTools(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{name: "no tools", body: `{"model":"gpt-5.4-mini","input":"hi"}`, want: true},
		{name: "function tool", body: `{"tools":[{"type":"function","name":"f","parameters":{}}]}`, want: true},
		{name: "custom tool", body: `{"tools":[{"type":"custom","name":"c"}]}`, want: true},
		{name: "client-executed web search", body: `{"tools":[{"type":"web_search","execution_mode":"client"}]}`, want: true},
		{name: "server web search without execution mode", body: `{"tools":[{"type":"web_search"}]}`, want: false},
		{name: "server-executed web search", body: `{"tools":[{"type":"web_search","execution_mode":"server"}]}`, want: false},
		{name: "image generation tool", body: `{"tools":[{"type":"image_generation"}]}`, want: false},
		{name: "namespace tool", body: `{"tools":[{"type":"namespace"}]}`, want: false},
		{name: "mixed supported and unsupported", body: `{"tools":[{"type":"function"},{"type":"image_generation"}]}`, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResponsesLiteSupportsTools([]byte(tt.body)); got != tt.want {
				t.Fatalf("ResponsesLiteSupportsTools() = %v, want %v", got, tt.want)
			}
		})
	}
}
