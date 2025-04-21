package scenario

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test types to match the LLM interface
type CompletionResponse struct {
	Choices []CompletionChoice
}

type CompletionChoice struct {
	Message Message
}

// mockLLMCompletion is a mock implementation of LLMCompletion interface
type mockLLMCompletion struct {
	completionFunc func(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error)
}

func (m *mockLLMCompletion) Completion(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error) {
	if m.completionFunc != nil {
		return m.completionFunc(ctx, messages, temperature, maxTokens, tools, toolChoice)
	}
	return nil, nil
}

func TestTestingAgent_GenerateNextMessage_FirstMessage(t *testing.T) {
	ctx := context.Background()
	expectedMessage := "hello there"

	mockLLM := &mockLLMCompletion{
		completionFunc: func(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error) {
			// Verify system message and role swapping
			require.Greater(t, len(messages), 1)
			assert.Equal(t, MessageRoleSystem, messages[0].Role)
			assert.Contains(t, messages[0].Content, "pretending to be a user")

			return &LLMCompletionResponse{
				Choices: []LLMCompletionResponseChoice{
					{
						Message: LLMCompletionResponseChoiceMessage{
							Content: expectedMessage,
						},
					},
				},
			}, nil
		},
	}

	agent := NewTestingAgent(mockLLM)
	msg, result, err := agent.GenerateNextMessage(
		ctx,
		"Test description",
		"Test strategy",
		[]string{"success1"},
		[]string{"failure1"},
		[]Message{},
		true,
		false,
	)

	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, expectedMessage, *msg)
	assert.Nil(t, result)
}

func TestTestingAgent_GenerateNextMessage_Success(t *testing.T) {
	ctx := context.Background()
	mockLLM := &mockLLMCompletion{
		completionFunc: func(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error) {
			toolCalls := []ToolCall{
				{
					Type: ToolTypeFunction,
					Function: &ToolCallFunction{
						Name: "finish_test",
						Arguments: map[string]interface{}{
							"verdict":   "success",
							"reasoning": "All criteria met",
							"details": map[string]interface{}{
								"met_criteria":       []string{"success1"},
								"unmet_criteria":     []string{},
								"triggered_failures": []string{},
							},
						},
					},
				},
			}

			return &LLMCompletionResponse{
				Choices: []LLMCompletionResponseChoice{
					{
						Message: LLMCompletionResponseChoiceMessage{
							ToolCalls: toolCalls,
						},
					},
				},
			}, nil
		},
	}

	agent := NewTestingAgent(mockLLM)
	conversation := []Message{
		{Role: MessageRoleUser, Content: "initial message"},
		{Role: MessageRoleAssistant, Content: "response"},
	}

	msg, result, err := agent.GenerateNextMessage(
		ctx,
		"Test description",
		"Test strategy",
		[]string{"success1"},
		[]string{"failure1"},
		conversation,
		false,
		true,
	)

	require.NoError(t, err)
	require.Nil(t, msg)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "All criteria met", result.Reasoning)
	assert.Contains(t, result.MetCriteria, "success1")
	assert.Empty(t, result.UnmetCriteria)
	assert.Empty(t, result.TriggeredFailures)
}

func TestTestingAgent_GenerateNextMessage_Failure(t *testing.T) {
	ctx := context.Background()
	mockLLM := &mockLLMCompletion{
		completionFunc: func(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error) {
			toolCalls := []ToolCall{
				{
					Type: ToolTypeFunction,
					Function: &ToolCallFunction{
						Name: "finish_test",
						Arguments: map[string]interface{}{
							"verdict":   "failure",
							"reasoning": "Failure criteria triggered",
							"details": map[string]interface{}{
								"met_criteria":       []string{},
								"unmet_criteria":     []string{"success1"},
								"triggered_failures": []string{"failure1"},
							},
						},
					},
				},
			}

			return &LLMCompletionResponse{
				Choices: []LLMCompletionResponseChoice{
					{
						Message: LLMCompletionResponseChoiceMessage{
							ToolCalls: toolCalls,
						},
					},
				},
			}, nil
		},
	}

	agent := NewTestingAgent(mockLLM)
	conversation := []Message{
		{Role: MessageRoleUser, Content: "initial message"},
		{Role: MessageRoleAssistant, Content: "response"},
	}

	msg, result, err := agent.GenerateNextMessage(
		ctx,
		"Test description",
		"Test strategy",
		[]string{"success1"},
		[]string{"failure1"},
		conversation,
		false,
		true,
	)

	require.NoError(t, err)
	require.Nil(t, msg)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "Failure criteria triggered", result.Reasoning)
	assert.Empty(t, result.MetCriteria)
	assert.Contains(t, result.UnmetCriteria, "success1")
	assert.Contains(t, result.TriggeredFailures, "failure1")
}

func TestTestingAgent_GenerateNextMessage_Inconclusive(t *testing.T) {
	ctx := context.Background()
	mockLLM := &mockLLMCompletion{
		completionFunc: func(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error) {
			toolCalls := []ToolCall{
				{
					Type: ToolTypeFunction,
					Function: &ToolCallFunction{
						Name: "finish_test",
						Arguments: map[string]interface{}{
							"verdict":   "inconclusive",
							"reasoning": "Max turns reached",
							"details": map[string]interface{}{
								"met_criteria":       []string{},
								"unmet_criteria":     []string{"success1"},
								"triggered_failures": []string{},
							},
						},
					},
				},
			}

			return &LLMCompletionResponse{
				Choices: []LLMCompletionResponseChoice{
					{
						Message: LLMCompletionResponseChoiceMessage{
							ToolCalls: toolCalls,
						},
					},
				},
			}, nil
		},
	}

	agent := NewTestingAgent(mockLLM)
	conversation := []Message{
		{Role: MessageRoleUser, Content: "initial message"},
		{Role: MessageRoleAssistant, Content: "response"},
	}

	msg, result, err := agent.GenerateNextMessage(
		ctx,
		"Test description",
		"Test strategy",
		[]string{"success1"},
		[]string{"failure1"},
		conversation,
		false,
		true,
	)

	require.NoError(t, err)
	require.Nil(t, msg)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "Max turns reached", result.Reasoning)
	assert.Empty(t, result.MetCriteria)
	assert.Contains(t, result.UnmetCriteria, "success1")
	assert.Empty(t, result.TriggeredFailures)
}

func TestTestingAgent_GenerateNextMessage_Error_NoChoices(t *testing.T) {
	ctx := context.Background()
	mockLLM := &mockLLMCompletion{
		completionFunc: func(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error) {
			return &LLMCompletionResponse{
				Choices: []LLMCompletionResponseChoice{},
			}, nil
		},
	}

	agent := NewTestingAgent(mockLLM)
	msg, result, err := agent.GenerateNextMessage(
		ctx,
		"Test description",
		"Test strategy",
		[]string{"success1"},
		[]string{"failure1"},
		[]Message{},
		true,
		false,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no choices returned")
	assert.Nil(t, msg)
	assert.Nil(t, result)
}

func TestTestingAgent_GenerateNextMessage_Error_EmptyContent(t *testing.T) {
	ctx := context.Background()
	mockLLM := &mockLLMCompletion{
		completionFunc: func(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error) {
			return &LLMCompletionResponse{
				Choices: []LLMCompletionResponseChoice{
					{
						Message: LLMCompletionResponseChoiceMessage{
							Content: "",
						},
					},
				},
			}, nil
		},
	}

	agent := NewTestingAgent(mockLLM)
	msg, result, err := agent.GenerateNextMessage(
		ctx,
		"Test description",
		"Test strategy",
		[]string{"success1"},
		[]string{"failure1"},
		[]Message{},
		true,
		false,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no content returned in choice")
	assert.Nil(t, msg)
	assert.Nil(t, result)
}

func TestTestingAgent_GenerateNextMessage_Error_InvalidToolCall(t *testing.T) {
	ctx := context.Background()
	mockLLM := &mockLLMCompletion{
		completionFunc: func(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error) {
			toolCalls := []ToolCall{
				{
					Type: "invalid_type",
				},
			}

			return &LLMCompletionResponse{
				Choices: []LLMCompletionResponseChoice{
					{
						Message: LLMCompletionResponseChoiceMessage{
							ToolCalls: toolCalls,
						},
					},
				},
			}, nil
		},
	}

	agent := NewTestingAgent(mockLLM)
	msg, result, err := agent.GenerateNextMessage(
		ctx,
		"Test description",
		"Test strategy",
		[]string{"success1"},
		[]string{"failure1"},
		[]Message{},
		false,
		true,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "tool call is not a function")
	assert.Nil(t, msg)
	assert.Nil(t, result)
}

func TestTestingAgent_GenerateNextMessage_Error_InvalidFinishTestParams(t *testing.T) {
	ctx := context.Background()
	mockLLM := &mockLLMCompletion{
		completionFunc: func(ctx context.Context, messages []Message, temperature *float64, maxTokens *int64, tools []Tool, toolChoice *string) (*LLMCompletionResponse, error) {
			toolCalls := []ToolCall{
				{
					Type: ToolTypeFunction,
					Function: &ToolCallFunction{
						Name: "finish_test",
						Arguments: map[string]interface{}{
							"verdict": 123, // Invalid type
							"details": map[string]interface{}{
								"met_criteria":       []string{},
								"unmet_criteria":     []string{},
								"triggered_failures": []string{},
							},
						},
					},
				},
			}

			return &LLMCompletionResponse{
				Choices: []LLMCompletionResponseChoice{
					{
						Message: LLMCompletionResponseChoiceMessage{
							ToolCalls: toolCalls,
						},
					},
				},
			}, nil
		},
	}

	agent := NewTestingAgent(mockLLM)
	msg, result, err := agent.GenerateNextMessage(
		ctx,
		"Test description",
		"Test strategy",
		[]string{"success1"},
		[]string{"failure1"},
		[]Message{},
		false,
		true,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract finish_test parameters")
	assert.Nil(t, msg)
	assert.Nil(t, result)
}
