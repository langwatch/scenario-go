package scenario

import "context"

// LLMCompletion is an interface for an LLM that supports completion.
type LLMCompletion interface {
	Completion(
		ctx context.Context,
		messages []Message,
		temperature *float64,
		maxTokens *int64,
		tools []Tool,
		toolChoice *string,
	) (*LLMCompletionResponse, error)
}

type LLMCompletionResponse struct {
	Choices []LLMCompletionResponseChoice
}

type LLMCompletionResponseChoice struct {
	Message LLMCompletionResponseChoiceMessage
}

type LLMCompletionResponseChoiceMessage struct {
	Content   string
	ToolCalls []ToolCall
}
