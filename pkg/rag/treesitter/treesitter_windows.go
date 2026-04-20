//go:build windows && cgo

package treesitter

import "github.com/docker/docker-agent/pkg/rag/chunk"

// DocumentProcessor falls back to text chunking on Windows because the
// go-tree-sitter dependency used for code-aware chunking is not currently
// available in this environment.
type DocumentProcessor struct {
	textFallback *chunk.TextDocumentProcessor
}

// NewDocumentProcessor creates a text-only fallback processor for Windows.
func NewDocumentProcessor(chunkSize, chunkOverlap int, respectWordBoundaries bool) *DocumentProcessor {
	return &DocumentProcessor{
		textFallback: chunk.NewTextDocumentProcessor(chunkSize, chunkOverlap, respectWordBoundaries),
	}
}

// Process implements chunk.DocumentProcessor.
func (p *DocumentProcessor) Process(path string, content []byte) ([]chunk.Chunk, error) {
	return p.textFallback.Process(path, content)
}
