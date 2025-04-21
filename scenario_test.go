package scenario

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAgent is a mock implementation of the Agent interface.
type mockAgent struct {
	runFunc func(ctx context.Context, message string) ([]Message, error)
}

func (m *mockAgent) Run(ctx context.Context, message string) ([]Message, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, message)
	}
	// Default behavior: respond with a simple message
	return []Message{
		{Role: MessageRoleAssistant, Content: "Agent response to: " + message},
	}, nil
}

// mockTestingAgent is a mock implementation of the TestingAgent interface.
type mockTestingAgent struct {
	generateNextMessageFunc func(
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

func (m *mockTestingAgent) GenerateNextMessage(
	ctx context.Context,
	description string,
	strategy string,
	successCriteria []string,
	failureCriteria []string,
	conversation []Message,
	firstMessage bool,
	lastMessage bool,
) (*string, *Result, error) {
	if m.generateNextMessageFunc != nil {
		return m.generateNextMessageFunc(ctx, description, strategy, successCriteria, failureCriteria, conversation, firstMessage, lastMessage)
	}
	// Default behavior: always succeed after one turn
	if firstMessage {
		msg := "Initial user message"
		return &msg, nil, nil
	}
	// On the second call (not first message)
	res := NewSuccessPartialResult(
		conversation,
		"Test succeeded",
		[]string{"Success criteria met"},
	)
	return nil, res, nil
}

// TestScenario_Run_Success tests a successful scenario run.
func TestScenario_Run_Success(t *testing.T) {
	ctx := context.Background()
	mockAgentInst := &mockAgent{}               // Use default simple response
	mockTestingAgentInst := &mockTestingAgent{} // Use default success behavior

	s := NewScenario(
		WithDescription("Test Description"),
		WithAgent(mockAgentInst),
		WithTestingAgent(mockTestingAgentInst),
		WithSuccessCriteria("Success criteria met"),
		WithMaxTurns(2), // Ensure it finishes within default mock behavior
	)

	result, err := s.Run(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Equal(t, "Test succeeded", result.Reasoning)
	assert.Contains(t, result.MetCriteria, "Success criteria met")
	assert.Less(t, time.Duration(0), result.TotalDurationNSec)
	assert.Less(t, time.Duration(0), result.AgentDurationNSec)
	// Check conversation (initial user message + agent response)
	require.Len(t, result.Conversation, 2)
	assert.Equal(t, MessageRoleUser, result.Conversation[0].Role)
	assert.Equal(t, "Initial user message", result.Conversation[0].Content)
	assert.Equal(t, MessageRoleAssistant, result.Conversation[1].Role)
	assert.Equal(t, "Agent response to: Initial user message", result.Conversation[1].Content)
}

// TestScenario_Run_MaxTurns tests a scenario reaching max turns without success.
func TestScenario_Run_MaxTurns(t *testing.T) {
	ctx := context.Background()
	mockAgentInst := &mockAgent{} // Use default simple response
	turnCounter := 0

	mockTestingAgentInst := &mockTestingAgent{
		generateNextMessageFunc: func(ctx context.Context, description string, strategy string, successCriteria []string, failureCriteria []string, conversation []Message, firstMessage bool, lastMessage bool) (*string, *Result, error) {
			turnCounter++
			// Keep generating messages without checking lastMessage
			msg := fmt.Sprintf("User message turn %d", turnCounter)
			return &msg, nil, nil
		},
	}

	maxTurns := 3
	s := NewScenario(
		WithDescription("Max Turns Test"),
		WithAgent(mockAgentInst),
		WithTestingAgent(mockTestingAgentInst),
		WithMaxTurns(maxTurns),
	)

	result, err := s.Run(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.False(t, result.Success)
	assert.Contains(t, result.Reasoning, fmt.Sprintf("The conversation did not end in a failure after %d turns.", maxTurns))
	assert.Empty(t, result.MetCriteria)
	assert.Empty(t, result.UnmetCriteria)
	assert.Empty(t, result.TriggeredFailures)
	assert.Len(t, result.Conversation, maxTurns*2) // User msg + Agent response per turn
}

// TestScenario_Run_Failure tests a scenario run that ends in failure.
func TestScenario_Run_Failure(t *testing.T) {
	ctx := context.Background()
	mockAgentInst := &mockAgent{}
	mockTestingAgentInst := &mockTestingAgent{
		generateNextMessageFunc: func(ctx context.Context, description string, strategy string, successCriteria []string, failureCriteria []string, conversation []Message, firstMessage bool, lastMessage bool) (*string, *Result, error) {
			if firstMessage {
				msg := "Initial user message"
				return &msg, nil, nil
			}
			// Fail on the second turn
			res := NewFailurePartialResult(
				conversation,
				"Test failed",
				[]string{}, // No met criteria
				[]string{}, // No unmet success criteria
				[]string{"Failure criteria triggered"},
			)
			return nil, res, nil
		},
	}

	s := NewScenario(
		WithDescription("Failure Test"),
		WithAgent(mockAgentInst),
		WithTestingAgent(mockTestingAgentInst),
		WithFailureCriteria("Failure criteria triggered"),
		WithMaxTurns(5),
	)

	result, err := s.Run(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.False(t, result.Success)
	assert.Equal(t, "Test failed", result.Reasoning)
	assert.Empty(t, result.MetCriteria)
	assert.Empty(t, result.UnmetCriteria)
	assert.Contains(t, result.TriggeredFailures, "Failure criteria triggered")
	assert.Len(t, result.Conversation, 2) // User initial, Agent response
}

// TestScenario_Run_Agent_Error tests a scenario where the agent returns an error.
func TestScenario_Run_Agent_Error(t *testing.T) {
	ctx := context.Background()
	agentError := errors.New("agent failed")
	mockAgentInst := &mockAgent{
		runFunc: func(ctx context.Context, message string) ([]Message, error) {
			return nil, agentError
		},
	}
	mockTestingAgentInst := &mockTestingAgent{ // Only need initial message
		generateNextMessageFunc: func(ctx context.Context, description string, strategy string, successCriteria []string, failureCriteria []string, conversation []Message, firstMessage bool, lastMessage bool) (*string, *Result, error) {
			if firstMessage {
				msg := "Initial user message"
				return &msg, nil, nil
			}
			t.Fatal("GenerateNextMessage should not be called after agent error")
			return nil, nil, nil
		},
	}

	s := NewScenario(
		WithDescription("Agent Error Test"),
		WithAgent(mockAgentInst),
		WithTestingAgent(mockTestingAgentInst),
	)

	result, err := s.Run(ctx)

	require.Error(t, err)
	// Check that the original error is wrapped
	require.ErrorContains(t, err, "failed to run agent:")
	require.ErrorIs(t, err, agentError)
	require.NotNil(t, result) // Should still return a result struct
	assert.False(t, result.Success)
	// Conversation should contain only the initial user message before the agent error
	require.Len(t, result.Conversation, 0)
}

// TestScenario_Run_TestingAgent_InitialError tests a scenario where the testing agent fails to generate the initial message.
func TestScenario_Run_TestingAgent_InitialError(t *testing.T) {
	ctx := context.Background()
	testingAgentError := errors.New("testing agent initial error")
	mockAgentInst := &mockAgent{} // Agent shouldn't be called
	mockTestingAgentInst := &mockTestingAgent{
		generateNextMessageFunc: func(ctx context.Context, description string, strategy string, successCriteria []string, failureCriteria []string, conversation []Message, firstMessage bool, lastMessage bool) (*string, *Result, error) {
			if firstMessage {
				return nil, nil, testingAgentError
			}
			t.Fatal("GenerateNextMessage should not be called after initial error")
			return nil, nil, nil
		},
	}

	s := NewScenario(
		WithDescription("Testing Agent Initial Error Test"),
		WithAgent(mockAgentInst),
		WithTestingAgent(mockTestingAgentInst),
	)

	result, err := s.Run(ctx)

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to generate initial message:")
	require.ErrorIs(t, err, testingAgentError)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Empty(t, result.Conversation) // No messages should have been added
}

// TestScenario_Run_TestingAgent_NextError tests a scenario where the testing agent fails to generate a subsequent message.
func TestScenario_Run_TestingAgent_NextError(t *testing.T) {
	ctx := context.Background()
	testingAgentError := errors.New("testing agent next error")
	mockAgentInst := &mockAgent{} // Agent runs once
	mockTestingAgentInst := &mockTestingAgent{
		generateNextMessageFunc: func(ctx context.Context, description string, strategy string, successCriteria []string, failureCriteria []string, conversation []Message, firstMessage bool, lastMessage bool) (*string, *Result, error) {
			if firstMessage {
				msg := "Initial user message"
				return &msg, nil, nil
			}
			// Error on the second call (after agent responds)
			return nil, nil, testingAgentError
		},
	}

	s := NewScenario(
		WithDescription("Testing Agent Next Error Test"),
		WithAgent(mockAgentInst),
		WithTestingAgent(mockTestingAgentInst),
	)

	result, err := s.Run(ctx)

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to generate next message:")
	require.ErrorIs(t, err, testingAgentError)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	// Conversation should contain the initial user message and the agent's response
	require.Len(t, result.Conversation, 0)
}

// TestScenario_Run_NoAgent tests running without setting an agent.
func TestScenario_Run_NoAgent(t *testing.T) {
	ctx := context.Background()
	mockTestingAgentInst := &mockTestingAgent{} // Testing agent setup doesn't matter here

	// Deliberately don't set the agent
	s := NewScenario(
		WithDescription("No Agent Test"),
		WithTestingAgent(mockTestingAgentInst),
		// Missing WithAgent(...)
	)

	result, err := s.Run(ctx)

	require.Error(t, err)
	require.EqualError(t, err, "agent not set")
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Empty(t, result.Conversation)
}

// TestScenario_Run_AgentReturnsNoMessages tests when the agent returns an empty slice of messages.
func TestScenario_Run_AgentReturnsNoMessages(t *testing.T) {
	ctx := context.Background()
	mockAgentInst := &mockAgent{
		runFunc: func(ctx context.Context, message string) ([]Message, error) {
			return []Message{}, nil // Return empty slice
		},
	}
	mockTestingAgentInst := &mockTestingAgent{ // Only need initial message
		generateNextMessageFunc: func(ctx context.Context, description string, strategy string, successCriteria []string, failureCriteria []string, conversation []Message, firstMessage bool, lastMessage bool) (*string, *Result, error) {
			if firstMessage {
				msg := "Initial user message"
				return &msg, nil, nil
			}
			t.Fatal("GenerateNextMessage should not be called after agent error")
			return nil, nil, nil
		},
	}

	s := NewScenario(
		WithDescription("Agent No Messages Test"),
		WithAgent(mockAgentInst),
		WithTestingAgent(mockTestingAgentInst),
	)

	result, err := s.Run(ctx)

	require.Error(t, err)
	require.EqualError(t, err, "no messages returned from agent")
	require.NotNil(t, result)
	assert.False(t, result.Success)
	// Conversation should contain only the initial user message
	require.Len(t, result.Conversation, 0)
}
