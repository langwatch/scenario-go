package examples_test

import (
	"context"
	"testing"

	"github.com/langwatch/scenario-go"
)

func Test_VegetarianRecipeAgentNormal(t *testing.T) {
	t.Parallel()

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
		result.LogResultDetails(t)
		t.Errorf("expected success but got failure")
	}
}

func Test_VegetarianRecipeAgentUberHungry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	sc := scenario.NewScenario(
		scenario.WithDescription("User is very very hungry, they say they could eat a cow"),
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
		result.LogResultDetails(t)
		t.Errorf("expected success but got failure")
	}
}
