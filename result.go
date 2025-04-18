package scenario

import (
	"testing"
	"time"
)

// Result is the result of a scenario.
type Result struct {
	// Success is true if the scenario was successful.
	Success bool

	// Conversation is the conversation between the user and the assistant.
	Conversation []Message

	// Reasoning is the reasoning for the result given by the assistant.
	Reasoning string

	// MetCriteria is the criteria that were met by the assistant.
	MetCriteria []string

	// UnmetCriteria is the criteria that were not met by the assistant.
	UnmetCriteria []string

	// TriggeredFailures is the failures that were triggered by the assistant.
	TriggeredFailures []string

	// TotalDurationNSec is the total duration of the scenario, in nanoseconds.
	TotalDurationNSec time.Duration

	// AgentDurationNSec is the duration of your agent within the scenario, in nanoseconds.
	AgentDurationNSec time.Duration
}

// NewSuccessPartialResult creates a new success result without the total time elapsed and agent time elapsed.
func NewSuccessPartialResult(
	conversation []Message,
	reasoning string,
	metCriteria []string,
) *Result {
	return &Result{
		Success:      true,
		Conversation: conversation,
		Reasoning:    reasoning,
		MetCriteria:  metCriteria,
	}
}

// NewFailurePartialResult creates a new failure result without the total time elapsed and agent time elapsed.
func NewFailurePartialResult(
	conversation []Message,
	reasoning string,
	metCriteria []string,
	unmetCriteria []string,
	triggeredFailures []string,
) *Result {
	return &Result{
		Success:           false,
		Conversation:      conversation,
		Reasoning:         reasoning,
		MetCriteria:       metCriteria,
		UnmetCriteria:     unmetCriteria,
		TriggeredFailures: triggeredFailures,
	}
}

// NewInconclusivePartialResult creates a new inconclusive result without the total time elapsed and agent time elapsed.
func NewInconclusivePartialResult(
	conversation []Message,
	reasoning string,
	metCriteria []string,
	unmetCriteria []string,
	triggeredFailures []string,
) *Result {
	return &Result{
		Success:           false,
		Conversation:      conversation,
		Reasoning:         reasoning,
		MetCriteria:       metCriteria,
		UnmetCriteria:     unmetCriteria,
		TriggeredFailures: triggeredFailures,
	}
}

// LogResultDetails logs detailed information about the Result struct. It's useful to call
// this in your tests on failure to get more context about the result, which will aid you
// with debugging.
func LogResultDetails(t *testing.T, res Result) {
	t.Helper() // so that failures are attributed to the correct line in tests

	// Option 1: Format key fields manually.
	t.Logf("Test Result Details:")
	t.Logf("Success: %v", res.Success)
	t.Logf("Reasoning: %s", res.Reasoning)
	t.Logf("Met Criteria: %v", res.MetCriteria)
	t.Logf("Unmet Criteria: %v", res.UnmetCriteria)
	t.Logf("Triggered Failures: %v", res.TriggeredFailures)
	t.Logf("Total Duration (ns): %v", res.TotalDurationNSec)
	t.Logf("Agent Duration (ns): %v", res.AgentDurationNSec)
}
