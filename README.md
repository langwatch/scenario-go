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

Now create your first test scenario and save it as `vegetarianrecipeagent_test.go`: 

```go
package examples_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/langwatch/scenario-go"
	"github.com/openai/openai-go"
)

func Test_VegetarianRecipeAgent(t *testing.T) {
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

Now run it with go test:

```bash
go test test_vegetarian_recipe_agent.go
```

You can find a fully working example in [examples/vegetarianrecipeagent_test.py](examples/vegetarianrecipeagent_test.go).

## Customize strategy and max_turns

You can customize how should the testing agent go about testing by defining a Strategy. You can also limit the maximum number of turns the scenario will take by setting the MaxTurns option, this defaults to 10.

For example, in this Lovable Clone scenario test:

```go
sc := scenario.NewScenario(
    scenario.WithDescription("User wants to create a new landing page for their dog walking startup"),
    scenario.WithAgent(lovable_agent),
    scenario.WithStrategy("send the first message to generate the landing page, then a single follow up request to extend it, then give your final verdict"),
    scenario.WithSuccessCriteria([
        "agent reads the files before go and making changes",
        "agent modified the index.css file",
        "agent modified the Index.tsx file",
        "agent created a comprehensive landing page",
        "agent extended the landing page with a new section",
    ],
    scenario.WithFailureCriteria([
        "agent says it can't read the file",
        "agent produces incomplete code or is too lazy to finish",
    ],
    scenario.WithMaxTurns(5),
)

result, err := sc.Run(ctx)
```

## Completions

The testing agent uses the `LLMCompletion` interface to perform LLM interactions. You can
use any LLM you want, you just need to implement the `LLMCompletion` interface. Scenario
ships with an implementation of OpenAI under `OpenAICompletion` that you can use as a
reference. View it [here](https://github.com/langwatch/scenario-go/blob/main/llm_openai.go).

## Contributing

We welcome contributions!
