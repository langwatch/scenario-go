package scenario

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Scenario is the interface for a scenario.
type Scenario interface {
	Run(ctx context.Context) (*Result, error)
}

// scenario is the default implementation of the Scenario interface.
type scenario struct {
	description     string
	strategy        string
	agent           Agent
	testingAgent    TestingAgent
	successCriteria []string
	failureCriteria []string
	maxTurns        int

	conversation []Message
}

// NewScenario creates a new scenario with the given options.
func NewScenario(opts ...ScenarioOption) Scenario {
	s := &scenario{
		strategy:        "Start with a first message and guide the conversation to play out the scenario.",
		successCriteria: []string{},
		failureCriteria: []string{},
		maxTurns:        10,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Run executes the scenario.
func (s *scenario) Run(ctx context.Context) (*Result, error) {
	if s.agent == nil {
		return &Result{Success: false}, errors.New("agent not set")
	}

	testStart := time.Now()
	agentDuration := time.Duration(0)

	initialMessage, initialResult, err := s.testingAgent.GenerateNextMessage(ctx, s.description, s.strategy, s.successCriteria, s.failureCriteria, s.conversation, true, false)
	if err != nil {
		return &Result{Success: false}, fmt.Errorf("failed to generate initial message: %w", err)
	}
	if initialResult != nil {
		return initialResult, fmt.Errorf("initial message generated a result which is unexpected: %v", initialResult)
	}

	currentMessage := initialMessage
	for iteration := range s.maxTurns {
		lastIteration := iteration == s.maxTurns-1
		s.conversation = append(s.conversation, Message{
			Role:    "user",
			Content: *currentMessage,
		})

		agentStart := time.Now()
		agentMessages, err := s.agent.Run(ctx, *currentMessage)
		if err != nil {
			return &Result{Success: false}, fmt.Errorf("failed to run agent: %w", err)
		}
		if len(agentMessages) == 0 {
			return &Result{Success: false}, errors.New("no messages returned from agent")
		}

		// Remove first messages if they are user or system messages
		if len(agentMessages) > 0 && agentMessages[0].Role == MessageRoleSystem {
			agentMessages = agentMessages[1:]
		}
		if len(agentMessages) > 0 && agentMessages[0].Role == MessageRoleUser {
			agentMessages = agentMessages[1:]
		}

		agentDuration += time.Since(agentStart)
		s.conversation = append(s.conversation, agentMessages...)

		nextMessage, result, err := s.testingAgent.GenerateNextMessage(ctx, s.description, s.strategy, s.successCriteria, s.failureCriteria, s.conversation, false, lastIteration)
		if err != nil {
			return &Result{Success: false}, fmt.Errorf("failed to generate next message: %w", err)
		}
		if result != nil {
			result.AgentDurationNSec = agentDuration
			result.TotalDurationNSec = time.Since(testStart)

			return result, nil
		}

		currentMessage = nextMessage
	}

	return &Result{
		Success:           false,
		Conversation:      s.conversation,
		Reasoning:         fmt.Sprintf("The conversation did not end in a failure after %d turns.", s.maxTurns),
		MetCriteria:       []string{},
		UnmetCriteria:     []string{},
		TriggeredFailures: []string{},
		TotalDurationNSec: time.Since(testStart),
		AgentDurationNSec: agentDuration,
	}, nil
}
