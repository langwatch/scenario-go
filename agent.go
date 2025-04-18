package scenario

import "context"

// Agent is the interface your agent should implement to be used with the scenario package.
type Agent interface {
	// Run runs the agent.
	Run(ctx context.Context, message string) ([]Message, error)
}
