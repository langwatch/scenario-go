![scenario](https://github.com/langwatch/scenario-go/raw/main/assets/scenario-wide.webp)

<div align="center">
<!-- Discord, pkg dev, Docs, etc links -->
</div>

# Scenario (Go): Use an Agent to test your Agent

Scenario is a library for testing agents end-to-end as a human would, but without having to manually do it. The automated testing agent covers every single scenario for you.

You define the scenarios, and the testing agent will simulate your users as it follows them, it will keep chatting and evaluating your agent until it reaches the desired goal or detects an unexpected behavior.

## Getting Started

Install scenario:

```bash
go get github.com/langwatch/scenario-go
```

Now create your first scenario and save it as `examples/vegetarian_recipe_agent_test.go`: 

```go
package examples_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/langwatch/scenario-go"
	"github.com/openai/openai-go"
)

func TestVegetarianRecipeAgent(t *testing.T) {
	ctx := context.Background()
	sc := scenario.NewScenario(
		scenario.WithDescription("User is looking for a dinner idea"),
		scenario.WithAgent(NewVegetarianRecipeAgent()),
		scenario.WithTestingAgent(scenario.NewTestingAgent(scenario.NewOpenAICompletion("gpt-4o-mini"))),
		scenario.WithSuccessCriteria(
			"Recipe agent generates a vegetarian recipe",
			"Recipe includes a list of ingredients",
			"Recipe includes step-by-step cooking instructions",
		),
		scenario.WithFailureCriteria(
			"The recipe is not vegetarian or includes meat",
			"The agent asks more than two follow-up questions",
		),
	)

	result, err := sc.Run(ctx)
	if err != nil {
		t.Fatalf("scenario failed to run: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success but got failure")
	}
}

type VegetarianRecipeAgent struct {
	history []scenario.Message
	client  openai.Client
}

func NewVegetarianRecipeAgent() *VegetarianRecipeAgent {
	return &VegetarianRecipeAgent{
		history: []scenario.Message{{
			Role: "system",
			Content: `
You are a vegetarian recipe agent.
Given the user request, ask AT MOST ONE follow-up question, then provide a complete recipe.
Keep your responses concise and focused.`,
		}},
		client: openai.NewClient(),
	}
}

func (a *VegetarianRecipeAgent) Run(ctx context.Context, message string) ([]scenario.Message, error) {
	a.history = append(a.history, scenario.Message{
		Role:    "user",
		Content: message,
	})

	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, len(a.history))
	for i, message := range a.history {
		switch message.Role {
		case scenario.MessageRoleSystem:
			openaiMessages[i] = openai.SystemMessage(message.Content)
		case scenario.MessageRoleUser:
			openaiMessages[i] = openai.UserMessage(message.Content)
		case scenario.MessageRoleAssistant:
			openaiMessages[i] = openai.AssistantMessage(message.Content)
		}
	}

	chatCompletion, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openaiMessages,
		Model:    openai.ChatModelGPT4o,
	})
	if err != nil {
		return nil, err
	}

	resp := scenario.Message{
		Role:    "assistant",
		Content: chatCompletion.Choices[0].Message.Content,
	}
	a.history = append(a.history, resp)
	return []scenario.Message{resp}, nil
}

```

Create a `.env` file and put your OpenAI API key in it:

```bash
OPENAI_API_KEY=<your-api-key>
```

