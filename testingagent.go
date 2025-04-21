package scenario

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/langwatch/scenario-go/internal/ptr"
)

var (
	testingAgentSystemMessageTemplate = mustSystemMessageCompile(`
<role>
You are pretending to be a user, you are testing an AI Agent (shown as the user role) based on a scenario.
Approach this naturally, as a human user would, with very short inputs, few words, all lowercase, imperative, not periods, like when they google or talk to chatgpt.
</role>

<goal>
Your goal (assistant) is to interact with the Agent Under Test (user) as if you were a human user to see if it can complete the scenario successfully.
</goal>

<scenario>
{{.Description}}
</scenario>

<strategy>
{{.Strategy}}
</strategy>

<success_criteria>
{{.SuccessCriteriaJSON}}
</success_criteria>

<failure_criteria>
{{.FailureCriteriaJSON}}
</failure_criteria>

<execution_flow>
1. Generate the first message to start the scenario
2. After the Agent Under Test (user) responds, generate the next message to send to the Agent Under Test, keep repeating step 2 until the criteria match
3. If the test should end, use the finish_test tool to determine if success or failure criteria have been met
</execution_flow>

<rules>
1. Test should end immediately if a failure criteria is triggered
2. Test should continue until all success criteria have been met
3. DO NOT make any judgment calls that are not explicitly listed in the success or failure criteria, withhold judgement if necessary
4. DO NOT carry over any requests yourself, YOU ARE NOT the assistant today, wait for the user to do it
</rules>
`)

	testingAgentFinishTestMessage = `
System:

<finish_test>
This is the last message, conversation has reached the maximum number of turns, give your final verdict,
if you don't have enough information to make a verdict, say inconclusive with max turns reached.
</finish_test>`
)

type testingAgentSystemMessageParams struct {
	Description         string
	Strategy            string
	SuccessCriteriaJSON string
	FailureCriteriaJSON string
}

type TestingAgent interface {
	GenerateNextMessage(
		ctx context.Context,
		description string,
		strategy string,
		successCriteria []string,
		failureCriteria []string,
		conversation []Message,
		firstMessage bool,
		lastMessage bool,
	) (*string, *Result, error)
}

type testingAgent struct {
	llmCompletion LLMCompletion
	temperature   *float64
	maxTokens     *int64
}

// NewTestingAgent creates a new testing agent.
func NewTestingAgent(
	llmCompletion LLMCompletion,
) TestingAgent {
	return &testingAgent{
		llmCompletion: llmCompletion,
		temperature:   ptr.Ptr(0.0),
		maxTokens:     nil,
	}
}

// GenerateNextMessage generates the next message to send to the agent under test.
func (t *testingAgent) GenerateNextMessage(
	ctx context.Context,
	description string,
	strategy string,
	successCriteria []string,
	failureCriteria []string,
	conversation []Message,
	firstMessage bool,
	lastMessage bool,
) (*string, *Result, error) {
	successCriteriaJSON, err := json.MarshalIndent(successCriteria, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	failureCriteriaJSON, err := json.MarshalIndent(failureCriteria, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	systemMessageParams := &testingAgentSystemMessageParams{
		Description:         description,
		Strategy:            strategy,
		SuccessCriteriaJSON: string(successCriteriaJSON),
		FailureCriteriaJSON: string(failureCriteriaJSON),
	}

	var systemMessage bytes.Buffer
	if err := testingAgentSystemMessageTemplate.Execute(&systemMessage, systemMessageParams); err != nil {
		return nil, nil, fmt.Errorf("failed to execute system message template: %w", err)
	}

	messages := []Message{{
		Role:    MessageRoleSystem,
		Content: systemMessage.String(),
	}, {
		Role:    MessageRoleAssistant,
		Content: "Hello, how can I help you today?",
	}}
	messages = append(messages, conversation...)
	if lastMessage {
		messages = append(messages, Message{
			Role:    MessageRoleUser,
			Content: testingAgentFinishTestMessage,
		})
	}

	for _, message := range messages {
		if len(message.Tools) > 0 {
			continue
		}

		switch message.Role {
		case MessageRoleAssistant:
			message.Role = MessageRoleUser
		case MessageRoleUser:
			message.Role = MessageRoleAssistant
		}
	}

	tools := []Tool{{
		Type: ToolTypeFunction,
		Function: &ToolFunction{
			Name:        "finish_test",
			Description: "Complete the test with a final verdict",
			Strict:      true,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"verdict": map[string]any{
						"type":        "string",
						"enum":        []string{"success", "failure", "inconclusive"},
						"description": "The final verdict of the test",
					},
					"reasoning": map[string]any{
						"type":        "string",
						"description": "Explanation of why this verdict was reached",
					},
					"details": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"met_criteria": map[string]any{
								"type":        "array",
								"items":       map[string]any{"type": "string"},
								"description": "List of success criteria that have been met",
							},
							"unmet_criteria": map[string]any{
								"type":        "array",
								"items":       map[string]any{"type": "string"},
								"description": "List of success criteria that have not been met",
							},
							"triggered_failures": map[string]any{
								"type":        "array",
								"items":       map[string]any{"type": "string"},
								"description": "List of failure criteria that have been triggered",
							},
						},
						"required":             []string{"met_criteria", "unmet_criteria", "triggered_failures"},
						"additionalProperties": false,
						"description":          "Detailed information about criteria evaluation",
					},
				},
				"required":             []string{"verdict", "reasoning", "details"},
				"additionalProperties": false,
			},
		},
	}}

	toolChoice := ptr.Ptr("required")
	if !lastMessage {
		toolChoice = nil
	}
	resp, err := t.llmCompletion.Completion(ctx, messages, t.temperature, t.maxTokens, tools, toolChoice)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate llm completion: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, nil, fmt.Errorf("no choices returned")
	}

	choice := resp.Choices[0]
	if len(choice.Message.ToolCalls) > 0 {
		if choice.Message.ToolCalls[0].Type != ToolTypeFunction {
			return nil, nil, fmt.Errorf("tool call is not a function")
		}

		toolCall := choice.Message.ToolCalls[0]
		if toolCall.Function.Name == "finish_test" {
			verdict, reasoning, metCriteria, unmetCriteria, triggeredFailures, err := extractFinishTestParams(toolCall)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to extract finish_test parameters: %w", err)
			}

			switch verdict {
			case "success":
				return nil, NewSuccessPartialResult(conversation, reasoning, metCriteria), nil
			case "failure":
				return nil, NewFailurePartialResult(conversation, reasoning, metCriteria, unmetCriteria, triggeredFailures), nil
			default:
				return nil, NewInconclusivePartialResult(conversation, reasoning, metCriteria, unmetCriteria, triggeredFailures), nil
			}
		}
	}

	if choice.Message.Content == "" {
		return nil, nil, fmt.Errorf("no content returned in choice")
	}

	return ptr.Ptr(choice.Message.Content), nil, nil
}

func extractFinishTestParams(toolCall ToolCall) (
	verdict string,
	reasoning string,
	metCriteria []string,
	unmetCriteria []string,
	triggeredFailures []string,
	err error,
) {
	args := toolCall.Function.Arguments

	verdict, ok := args["verdict"].(string)
	if !ok {
		err = fmt.Errorf("verdict is not a string")
		return
	}

	reasoning, ok = args["reasoning"].(string)
	if !ok {
		err = fmt.Errorf("reasoning is not a string")
		return
	}

	details, ok := args["details"].(map[string]any)
	if !ok {
		err = fmt.Errorf("details is not a map")
		return
	}

	metCriteria, err = extractStringArray(details, "met_criteria")
	if err != nil {
		return
	}

	unmetCriteria, err = extractStringArray(details, "unmet_criteria")
	if err != nil {
		return
	}

	triggeredFailures, err = extractStringArray(details, "triggered_failures")
	if err != nil {
		return
	}

	return
}

func extractStringArray(data map[string]any, key string) ([]string, error) {
	val, ok := data[key]
	if !ok {
		return nil, fmt.Errorf("%s not found", key)
	}

	// Handle []any
	if interfaceSlice, ok := val.([]any); ok {
		strSlice := make([]string, len(interfaceSlice))
		for i, item := range interfaceSlice {
			if strItem, ok := item.(string); ok {
				strSlice[i] = strItem
			} else {
				return nil, fmt.Errorf("item at index %d in %s is not a string", i, key)
			}
		}
		return strSlice, nil
	}

	// Handle []string
	if strSlice, ok := val.([]string); ok {
		return strSlice, nil
	}

	// Handle nil
	if val == nil {
		return []string{}, nil
	}

	return nil, fmt.Errorf("%s is not a valid string array, []any, or nil", key)
}

func mustSystemMessageCompile(text string) *template.Template {
	tmpl, err := template.New("system_message").Parse(text)
	if err != nil {
		panic(err)
	}

	return tmpl
}
