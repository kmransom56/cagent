package docanalysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPlan_ExtractsStructuredPlan(t *testing.T) {
	t.Parallel()

	content := `# Platform Migration Plan

Objective: move document ingestion off shell-based tooling.
- Implement native PDF parsing.
- Normalize Windows-hosted paths for WSL and containers.
- Validate the change with focused tests.

Constraints:
- Risk: existing Linux paths must keep working.
- Dependency: container bind mounts vary by environment.
`

	plan := BuildPlan("docs/plan.md", content, "Produce an implementation plan")

	require.Equal(t, "Platform Migration Plan", plan.Title)
	assert.Contains(t, plan.Summary, "Produce an implementation plan")
	assert.Contains(t, plan.Objectives, "Produce an implementation plan")
	assert.Contains(t, plan.Objectives, "Objective: move document ingestion off shell-based tooling.")
	assert.NotEmpty(t, plan.Risks)
	assert.NotEmpty(t, plan.RecommendedAgents)
	assert.NotEmpty(t, plan.TaskPlan)
	assert.Equal(t, "planner", plan.TaskPlan[0].Owner)
}

func TestBuildPlan_AddsFrontendSpecialistWhenNeeded(t *testing.T) {
	t.Parallel()

	content := `React dashboard refresh

Goal: update the UI workflow and validation states for the new screen.
`

	plan := BuildPlan("docs/ui.txt", content, "")

	assert.Contains(t, plan.RecommendedAgents, AgentRole{
		Name:           "frontend-specialist",
		Responsibility: "Handle UI structure, interaction flow, and visual implementation details.",
		Inputs:         []string{"User-facing requirements", "Design constraints"},
	})
	assert.Equal(t, "frontend-specialist", plan.TaskPlan[2].Owner)
}
