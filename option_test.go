package scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a new scenario for testing
func newTestScenario() Scenario {
	return NewScenario()
}

func TestWithDescription(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	WithDescription("test")(sc)
	assert.Equal(t, "test", sc.description)
}

func TestWithStrategy(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	WithStrategy("test")(sc)
	assert.Equal(t, "test", sc.strategy)
}

func TestWithMaxTurns(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	WithMaxTurns(10)(sc)
	assert.Equal(t, 10, sc.maxTurns)
}

func TestWithAgent(t *testing.T) {
	s := newTestScenario()
	mockAgent := &mockAgent{}
	sc := s.(*scenario)

	opt := WithAgent(mockAgent)
	opt(sc)

	assert.Equal(t, mockAgent, sc.agent)
}

func TestWithTestingAgent(t *testing.T) {
	s := newTestScenario()
	mockTestingAgent := &mockTestingAgent{}
	sc := s.(*scenario)

	opt := WithTestingAgent(mockTestingAgent)
	opt(sc)

	assert.Equal(t, mockTestingAgent, sc.testingAgent)
}

func TestWithSuccessCriteria(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	WithSuccessCriteria("test")(sc)
	assert.Equal(t, []string{"test"}, sc.successCriteria)
}

func TestWithFailureCriteria(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	WithFailureCriteria("test")(sc)
	assert.Equal(t, []string{"test"}, sc.failureCriteria)
}

func TestMultipleOptions(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	description := "test description"
	strategy := "test strategy"
	maxTurns := 5
	mockAgent := &mockAgent{}
	mockTestingAgent := &mockTestingAgent{}
	successCriteria := []string{"success1", "success2"}
	failureCriteria := []string{"failure1", "failure2"}

	// Apply all options
	opts := []ScenarioOption{
		WithDescription(description),
		WithStrategy(strategy),
		WithMaxTurns(maxTurns),
		WithAgent(mockAgent),
		WithTestingAgent(mockTestingAgent),
		WithSuccessCriteria(successCriteria...),
		WithFailureCriteria(failureCriteria...),
	}

	// Apply each option
	for _, opt := range opts {
		opt(sc)
	}

	// Verify all fields are set correctly
	assert.Equal(t, description, sc.description)
	assert.Equal(t, strategy, sc.strategy)
	assert.Equal(t, maxTurns, sc.maxTurns)
	assert.Equal(t, mockAgent, sc.agent)
	assert.Equal(t, mockTestingAgent, sc.testingAgent)
	assert.Equal(t, successCriteria, sc.successCriteria)
	assert.Equal(t, failureCriteria, sc.failureCriteria)
}

func TestNewScenario(t *testing.T) {
	description := "test description"
	strategy := "test strategy"
	maxTurns := 5
	mockAgent := &mockAgent{}
	mockTestingAgent := &mockTestingAgent{}
	successCriteria := []string{"success1"}
	failureCriteria := []string{"failure1"}

	// Test creating a new scenario with all options
	s := NewScenario(
		WithDescription(description),
		WithStrategy(strategy),
		WithMaxTurns(maxTurns),
		WithAgent(mockAgent),
		WithTestingAgent(mockTestingAgent),
		WithSuccessCriteria(successCriteria...),
		WithFailureCriteria(failureCriteria...),
	)

	require.NotNil(t, s)
	sc := s.(*scenario) // Type assertion in test is okay
	assert.Equal(t, description, sc.description)
	assert.Equal(t, strategy, sc.strategy)
	assert.Equal(t, maxTurns, sc.maxTurns)
	assert.Equal(t, mockAgent, sc.agent)
	assert.Equal(t, mockTestingAgent, sc.testingAgent)
	assert.Equal(t, successCriteria, sc.successCriteria)
	assert.Equal(t, failureCriteria, sc.failureCriteria)

	// Test creating a new scenario with no options
	s = NewScenario()
	require.NotNil(t, s)
	sc = s.(*scenario) // Type assertion in test is okay
	assert.Empty(t, sc.description)
	assert.Equal(t, "Start with a first message and guide the conversation to play out the scenario.", sc.strategy)
	assert.Equal(t, 10, sc.maxTurns) // Default value
	assert.Nil(t, sc.agent)
	assert.Nil(t, sc.testingAgent)
	assert.Empty(t, sc.successCriteria)
	assert.Empty(t, sc.failureCriteria)
}

func TestWithMaxTurns_Negative(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	maxTurns := -5

	opt := WithMaxTurns(maxTurns)
	opt(sc)

	// Negative max turns should be stored as-is, validation should happen at runtime
	assert.Equal(t, maxTurns, sc.maxTurns)
}

func TestWithSuccessCriteria_Duplicates(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	criteria := []string{"success1", "success1", "success2", "success2"}

	opt := WithSuccessCriteria(criteria...)
	opt(sc)

	// Duplicates should be preserved as-is
	assert.Equal(t, criteria, sc.successCriteria)
}

func TestWithFailureCriteria_Duplicates(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	criteria := []string{"failure1", "failure1", "failure2", "failure2"}

	opt := WithFailureCriteria(criteria...)
	opt(sc)

	// Duplicates should be preserved as-is
	assert.Equal(t, criteria, sc.failureCriteria)
}

func TestWithSuccessCriteria_EmptyStrings(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	criteria := []string{"success1", "", "success2", ""}

	opt := WithSuccessCriteria(criteria...)
	opt(sc)

	// Empty strings should be preserved as-is
	assert.Equal(t, criteria, sc.successCriteria)
}

func TestWithFailureCriteria_EmptyStrings(t *testing.T) {
	s := newTestScenario()
	sc := s.(*scenario)
	criteria := []string{"failure1", "", "failure2", ""}

	opt := WithFailureCriteria(criteria...)
	opt(sc)

	// Empty strings should be preserved as-is
	assert.Equal(t, criteria, sc.failureCriteria)
}

// Test chaining multiple options in a single call
func TestOptionChaining(t *testing.T) {
	s := newTestScenario()
	description := "test description"
	strategy := "test strategy"
	maxTurns := 5

	// Chain multiple options in a single call
	opt := func(s *scenario) {
		WithDescription(description)(s)
		WithStrategy(strategy)(s)
		WithMaxTurns(maxTurns)(s)
	}
	opt(s.(*scenario))

	assert.Equal(t, description, s.(*scenario).description)
	assert.Equal(t, strategy, s.(*scenario).strategy)
	assert.Equal(t, maxTurns, s.(*scenario).maxTurns)
}

// Test that nil options are handled gracefully
func TestNilOption(t *testing.T) {
	s := newTestScenario()
	var nilOpt ScenarioOption

	// Should not panic
	assert.NotPanics(t, func() {
		if nilOpt != nil {
			nilOpt(s.(*scenario))
		}
	})
}
