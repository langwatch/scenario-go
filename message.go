package scenario

// MessageRole is the role of a message.
type MessageRole string

const (
	// MessageRoleUser is the role of a user.
	MessageRoleUser MessageRole = "user"

	// MessageRoleAssistant is the role of an assistant.
	MessageRoleAssistant MessageRole = "assistant"

	// MessageRoleSystem is the role of a system.
	MessageRoleSystem MessageRole = "system"

	// MessageRoleDeveloper is the role of a developer.
	MessageRoleDeveloper MessageRole = "developer"
)

// ToolType is the type of a tool.
type ToolType string

const (
	// ToolTypeFunction is the type of a function tool.
	ToolTypeFunction ToolType = "function"
)

// Message is a message in a conversation.
type Message struct {
	// Role is the role of the message.
	Role MessageRole

	// Content is the content of the message.
	Content string

	// ToolCalls contains the tool calls available to the message.
	ToolCalls []any
}

// Tool represents a tool that can be used in a message.
type Tool struct {
	// Type is the type of the tool.
	Type ToolType

	// Function defines the function to call.
	Function *ToolFunction
}

// ToolFunction represents the function definition of a tool.
type ToolFunction struct {
	// Name is the name of the function.
	Name string

	// Description is the description of the function.
	Description string

	// Strict is whether the function is strict.
	Strict bool

	// Parameters is the parameters of the function.
	Parameters map[string]any
}

// ToolCall is a tool call in a message.
type ToolCall struct {
	ID       string
	Type     ToolType
	Function *ToolCallFunction
}

type ToolCallFunction struct {
	Name      string
	Arguments map[string]any
}
