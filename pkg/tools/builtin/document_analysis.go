package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/cagent/pkg/docanalysis"
	"github.com/docker/cagent/pkg/docreader"
	"github.com/docker/cagent/pkg/tools"
)

const ToolNameAnalyzeDocument = "analyze_document"

type DocumentAnalysisTool struct {
	workingDir string
}

var (
	_ tools.ToolSet      = (*DocumentAnalysisTool)(nil)
	_ tools.Instructable = (*DocumentAnalysisTool)(nil)
)

type AnalyzeDocumentArgs struct {
	Path string `json:"path" jsonschema:"Path to the document to analyze"`
	Goal string `json:"goal,omitempty" jsonschema:"Optional planning goal or desired output focus"`
}

func NewDocumentAnalysisTool(workingDir string) *DocumentAnalysisTool {
	return &DocumentAnalysisTool{workingDir: workingDir}
}

func (t *DocumentAnalysisTool) Instructions() string {
	return `## Document Analysis Tool Instructions

Use this tool when you have a requirements, design, migration, or planning document and need a deterministic execution plan.

- Provide the document path.
- Add a goal when you want the task plan biased toward a specific outcome.
- The tool extracts text natively, including PDFs, and returns a structured plan with objectives, risks, recommended agents, and execution tasks.`
}

func (t *DocumentAnalysisTool) Tools(context.Context) ([]tools.Tool, error) {
	return []tools.Tool{
		{
			Name:         ToolNameAnalyzeDocument,
			Category:     "knowledge",
			Description:  "Analyze a document or PDF and turn it into a structured agent and task plan.",
			Parameters:   tools.MustSchemaFor[AnalyzeDocumentArgs](),
			OutputSchema: tools.MustSchemaFor[docanalysis.Plan](),
			Handler:      tools.NewHandler(t.handleAnalyzeDocument),
			Annotations: tools.ToolAnnotations{
				ReadOnlyHint: true,
				Title:        "Analyze Document",
			},
		},
	}, nil
}

func (t *DocumentAnalysisTool) handleAnalyzeDocument(_ context.Context, args AnalyzeDocumentArgs) (*tools.ToolCallResult, error) {
	if args.Path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	resolution := docreader.ResolvePath(t.workingDir, args.Path)
	content, err := docreader.ReadText(resolution.ResolvedPath)
	if err != nil {
		return nil, fmt.Errorf("reading document: %w", err)
	}

	plan := docanalysis.BuildPlan(resolution.ResolvedPath, content, args.Goal)
	payload, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("marshaling plan: %w", err)
	}

	meta := map[string]any{
		"source_path":    resolution.ResolvedPath,
		"warnings":       resolution.Warnings,
		"content_length": len(content),
	}

	return &tools.ToolCallResult{
		Output: string(payload),
		Meta:   meta,
	}, nil
}
