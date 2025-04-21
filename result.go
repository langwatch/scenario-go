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
func (r *Result) LogResultDetails(t *testing.T) {
	t.Helper()

	t.Logf("Test Result Details:")
	t.Logf("Success: %v", r.Success)
	t.Logf("Reasoning: %s", r.Reasoning)
	t.Logf("Met Criteria: %v", r.MetCriteria)
	t.Logf("Unmet Criteria: %v", r.UnmetCriteria)
	t.Logf("Triggered Failures: %v", r.TriggeredFailures)
	t.Logf("Total Duration (ns): %v", r.TotalDurationNSec)
	t.Logf("Agent Duration (ns): %v", r.AgentDurationNSec)
}
