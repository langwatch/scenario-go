package scenario

type ScenarioOption func(*scenario)

// WithDescription sets the scenario's description.
func WithDescription(description string) ScenarioOption {
	return func(s *scenario) {
		s.description = description
	}
}

// WithStrategy sets the scenario's strategy.
func WithStrategy(strategy string) ScenarioOption {
	return func(s *scenario) {
		s.strategy = strategy
	}
}

// WithMaxTurns sets the scenario's max turns.
func WithMaxTurns(maxTurns int) ScenarioOption {
	return func(s *scenario) {
		s.maxTurns = maxTurns
	}
}

// WithAgent configures the scenario with a Agent dependency.
func WithAgent(agent Agent) ScenarioOption {
	return func(s *scenario) {
		s.agent = agent
	}
}

// WithTestingAgent configures the scenario with a TestingAgent dependency.
func WithTestingAgent(testingAgent TestingAgent) ScenarioOption {
	return func(s *scenario) {
		s.testingAgent = testingAgent
	}
}

// WithSuccessCriteria sets the scenario's success criteria.
func WithSuccessCriteria(criteria ...string) ScenarioOption {
	return func(s *scenario) {
		s.successCriteria = criteria
	}
}

// WithFailureCriteria sets the scenario's failure criteria.
func WithFailureCriteria(criteria ...string) ScenarioOption {
	return func(s *scenario) {
		s.failureCriteria = criteria
	}
}
