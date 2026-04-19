package docanalysis

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

type Plan struct {
	Title             string      `json:"title"`
	Summary           string      `json:"summary"`
	Objectives        []string    `json:"objectives"`
	Risks             []string    `json:"risks"`
	RecommendedAgents []AgentRole `json:"recommended_agents"`
	TaskPlan          []Task      `json:"task_plan"`
}

type AgentRole struct {
	Name           string   `json:"name"`
	Responsibility string   `json:"responsibility"`
	Inputs         []string `json:"inputs"`
}

type Task struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Owner       string   `json:"owner"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
}

func BuildPlan(sourcePath, content, goal string) Plan {
	lines := normalizedLines(content)
	title := detectTitle(sourcePath, lines)
	summary := buildSummary(lines, goal)
	objectives := detectObjectives(lines, goal)
	risks := detectRisks(lines)
	contextTags := detectContextTags(strings.ToLower(strings.Join(lines, "\n") + "\n" + goal))
	agents := buildAgentRoles(contextTags)
	tasks := buildTasks(title, goal, objectives, contextTags)

	return Plan{
		Title:             title,
		Summary:           summary,
		Objectives:        objectives,
		Risks:             risks,
		RecommendedAgents: agents,
		TaskPlan:          tasks,
	}
}

func normalizedLines(content string) []string {
	rawLines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

func detectTitle(sourcePath string, lines []string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
	}
	for _, line := range lines {
		clean := strings.Trim(line, "-*0123456789. ")
		if clean != "" {
			return truncate(clean, 100)
		}
	}
	base := filepath.Base(sourcePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func buildSummary(lines []string, goal string) string {
	if goal != "" {
		return truncate(fmt.Sprintf("Analyze the document and produce a delivery plan focused on: %s. %s", goal, firstSentences(lines, 2)), 420)
	}
	return truncate(firstSentences(lines, 3), 420)
}

func firstSentences(lines []string, limit int) string {
	joined := strings.Join(lines, " ")
	parts := splitSentences(joined)
	var kept []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kept = append(kept, part)
		if len(kept) == limit {
			break
		}
	}
	if len(kept) == 0 && len(lines) > 0 {
		return truncate(lines[0], 240)
	}
	return strings.Join(kept, " ")
}

func splitSentences(value string) []string {
	var parts []string
	start := 0
	for index, r := range value {
		if r != '.' && r != '!' && r != '?' {
			continue
		}
		part := strings.TrimSpace(value[start : index+1])
		if part != "" {
			parts = append(parts, part)
		}
		start = index + 1
	}
	if start < len(value) {
		part := strings.TrimSpace(value[start:])
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func detectObjectives(lines []string, goal string) []string {
	seen := map[string]struct{}{}
	var objectives []string
	if goal != "" {
		objectives = append(objectives, goal)
		seen[goal] = struct{}{}
	}
	for _, line := range lines {
		clean := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789.[]() "))
		lower := strings.ToLower(clean)
		if clean == "" {
			continue
		}
		if strings.Contains(lower, "objective") || strings.Contains(lower, "goal") || strings.Contains(lower, "requirement") || strings.Contains(lower, "deliverable") || strings.Contains(lower, "must ") || strings.Contains(lower, "should ") || strings.HasPrefix(lower, "build ") || strings.HasPrefix(lower, "create ") || strings.HasPrefix(lower, "implement ") {
			clean = truncate(clean, 160)
			if _, ok := seen[clean]; ok {
				continue
			}
			objectives = append(objectives, clean)
			seen[clean] = struct{}{}
		}
		if len(objectives) >= 5 {
			break
		}
	}
	if len(objectives) == 0 && len(lines) > 0 {
		objectives = append(objectives, truncate(lines[0], 160))
	}
	return objectives
}

func detectRisks(lines []string) []string {
	seen := map[string]struct{}{}
	var risks []string
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "risk") || strings.Contains(lower, "constraint") || strings.Contains(lower, "dependency") || strings.Contains(lower, "assumption") || strings.Contains(lower, "blocker") {
			candidate := truncate(strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. ")), 180)
			if _, ok := seen[candidate]; ok || candidate == "" {
				continue
			}
			risks = append(risks, candidate)
			seen[candidate] = struct{}{}
		}
		if len(risks) >= 4 {
			break
		}
	}
	return risks
}

func detectContextTags(content string) []string {
	tags := make([]string, 0, 6)
	appendTag := func(tag string, markers ...string) {
		for _, marker := range markers {
			if strings.Contains(content, marker) {
				tags = append(tags, tag)
				return
			}
		}
	}

	appendTag("api", "api", "endpoint", "http", "rest", "graphql")
	appendTag("frontend", "ui", "frontend", "react", "page", "screen")
	appendTag("data", "database", "schema", "sql", "storage", "data model")
	appendTag("integration", "integration", "queue", "event", "service bus", "kafka", "webhook")
	appendTag("migration", "migrate", "upgrade", "legacy", "modernize")
	appendTag("testing", "test", "validation", "acceptance", "qa")

	slices.Sort(tags)
	return slices.Compact(tags)
}

func buildAgentRoles(contextTags []string) []AgentRole {
	agents := []AgentRole{
		{
			Name:           "planner",
			Responsibility: "Translate the document into scoped milestones, dependencies, and handoff points.",
			Inputs:         []string{"Document summary", "Objectives", "Risks"},
		},
		{
			Name:           "implementer",
			Responsibility: "Execute code and configuration changes implied by the document.",
			Inputs:         []string{"Task plan", "Relevant source files"},
		},
		{
			Name:           "validator",
			Responsibility: "Verify the resulting changes with focused tests, builds, or smoke checks.",
			Inputs:         []string{"Changed files", "Acceptance criteria"},
		},
	}

	if slices.Contains(contextTags, "frontend") {
		agents = append(agents, AgentRole{
			Name:           "frontend-specialist",
			Responsibility: "Handle UI structure, interaction flow, and visual implementation details.",
			Inputs:         []string{"User-facing requirements", "Design constraints"},
		})
	}
	if slices.Contains(contextTags, "data") || slices.Contains(contextTags, "integration") {
		agents = append(agents, AgentRole{
			Name:           "integration-specialist",
			Responsibility: "Own schema, storage, and external system integration changes.",
			Inputs:         []string{"Data contracts", "Connection details", "Migration notes"},
		})
	}

	return agents
}

func buildTasks(title, goal string, objectives, contextTags []string) []Task {
	focus := title
	if goal != "" {
		focus = goal
	}

	inputs := []string{"Source document"}
	if len(objectives) > 0 {
		inputs = append(inputs, objectives[0])
	}

	tasks := []Task{
		{
			ID:          "1",
			Title:       "Clarify scope",
			Description: fmt.Sprintf("Extract the concrete scope, acceptance criteria, and boundaries for %s.", focus),
			Owner:       "planner",
			Inputs:      inputs,
			Outputs:     []string{"Scoped milestone list", "Open questions"},
		},
		{
			ID:          "2",
			Title:       "Map affected areas",
			Description: "Locate the code, configuration, and documents that must change to satisfy the requested outcome.",
			Owner:       "planner",
			Inputs:      []string{"Scoped milestone list", "Repository structure"},
			Outputs:     []string{"Affected file list", "Dependency map"},
		},
		{
			ID:          "3",
			Title:       "Implement changes",
			Description: "Apply the required implementation changes in small, testable slices.",
			Owner:       preferredImplementer(contextTags),
			Inputs:      []string{"Affected file list", "Acceptance criteria"},
			Outputs:     []string{"Code changes", "Updated configuration"},
		},
		{
			ID:          "4",
			Title:       "Validate outcome",
			Description: "Run focused verification to confirm the implementation matches the document intent.",
			Owner:       "validator",
			Inputs:      []string{"Code changes", "Acceptance criteria"},
			Outputs:     []string{"Validation results", "Follow-up fixes if needed"},
		},
	}

	if slices.Contains(contextTags, "migration") {
		tasks = append(tasks, Task{
			ID:          "5",
			Title:       "Plan migration rollout",
			Description: "Sequence rollout, fallback, and compatibility checks for the migration path.",
			Owner:       "integration-specialist",
			Inputs:      []string{"Dependency map", "Validation results"},
			Outputs:     []string{"Rollout checklist", "Rollback notes"},
		})
	}

	return tasks
}

func preferredImplementer(contextTags []string) string {
	if slices.Contains(contextTags, "frontend") {
		return "frontend-specialist"
	}
	if slices.Contains(contextTags, "data") || slices.Contains(contextTags, "integration") {
		return "integration-specialist"
	}
	return "implementer"
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(compactWhitespace(value))
	if len(value) <= limit {
		return value
	}
	return strings.TrimSpace(value[:limit-3]) + "..."
}

func compactWhitespace(value string) string {
	return strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " ")
}
