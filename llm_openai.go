package scenario

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
	"github.com/openai/openai-go/shared/constant"
)

type openAICompletion struct {
	model  string
	client openai.Client
}

// NewOpenAICompletion creates a new OpenAI completion.
func NewOpenAICompletion(model string) *openAICompletion {
	return &openAICompletion{
		model:  model,
		client: openai.NewClient(),
	}
}

// NewOpenAICompletionWithClient creates a new OpenAI completion with a specific client.
func NewOpenAICompletionWithClient(model string, client openai.Client) *openAICompletion {
	return &openAICompletion{
		model:  model,
		client: client,
	}
}

// Completion will generate a response from an LLM based on the messages, temperature, max tokens, tools, and tool choice.
func (c *openAICompletion) Completion(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error) {
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, message := range messages {
		switch message.Role {
		case MessageRoleUser:
			openaiMessages[i] = openai.UserMessage(message.Content)
		case MessageRoleAssistant:
			openaiMessages[i] = openai.AssistantMessage(message.Content)
		case MessageRoleSystem:
			openaiMessages[i] = openai.SystemMessage(message.Content)
		case MessageRoleDeveloper:
			openaiMessages[i] = openai.DeveloperMessage(message.Content)
		default:
			return nil, fmt.Errorf("unknown message role: %s", message.Role)
		}
	}

	openaiTools := make([]openai.ChatCompletionToolParam, len(tools))
	for i, tool := range tools {
		if tool.Type != ToolTypeFunction {
			return nil, fmt.Errorf("tool type is not function: %s", tool.Type)
		}

		openaiTools[i] = openai.ChatCompletionToolParam{
			Type: constant.Function(tool.Type),
			Function: shared.FunctionDefinitionParam{
				Name:        tool.Function.Name,
				Description: openai.String(tool.Function.Description),
				Strict:      param.NewOpt(tool.Function.Strict),
				Parameters:  tool.Function.Parameters,
			},
		}
	}

	params := openai.ChatCompletionNewParams{
		Messages: openaiMessages,
		Model:    shared.ChatModel(c.model),
		Tools:    openaiTools,
	}
	if temperature != nil {
		params.Temperature = openai.Float(*temperature)
	}
	if maxTokens != nil {
		params.MaxTokens = openai.Int(*maxTokens)
	}

	chatCompletion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	response := &LLMCompletionResponse{
		Choices: make([]LLMCompletionResponseChoice, len(chatCompletion.Choices)),
	}

	for i, choice := range chatCompletion.Choices {
		response.Choices[i] = LLMCompletionResponseChoice{
			Message: LLMCompletionResponseChoiceMessage{
				Content:   choice.Message.Content,
				ToolCalls: make([]ToolCall, len(choice.Message.ToolCalls)),
			},
		}

		for j, toolCall := range choice.Message.ToolCalls {
			var args map[string]any
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool call arguments (%d-%d): %w", i, j, err)
			}

			response.Choices[i].Message.ToolCalls[j] = ToolCall{
				Type: ToolType(toolCall.Type),
				ID:   toolCall.ID,
				Function: &ToolCallFunction{
					Name:      toolCall.Function.Name,
					Arguments: args,
				},
			}
		}
	}

	return response, nil
}
