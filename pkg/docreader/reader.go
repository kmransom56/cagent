package docreader

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ledongthuc/pdf"
)

var pdfTextReader = readPDFText

type PathResolution struct {
	RequestedPath string   `json:"requested_path"`
	ResolvedPath  string   `json:"resolved_path"`
	Warnings      []string `json:"warnings,omitempty"`
}

// ResolvePath normalizes a user or config-supplied path into a runtime-accessible path.
func ResolvePath(basePath, rawPath string) PathResolution {
	return resolvePath(basePath, rawPath, runtime.GOOS, os.LookupEnv)
}

// ReadText reads a document and returns text content.
// Text-based files are returned as-is; PDFs are extracted natively.
func ReadText(path string) (string, error) {
	if strings.EqualFold(filepath.Ext(path), ".pdf") {
		text, err := pdfTextReader(path)
		if err != nil {
			return "", fmt.Errorf("reading pdf %s: %w", path, err)
		}
		return text, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", path, err)
	}

	return string(data), nil
}

func readPDFText(path string) (string, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	textReader, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}

	content, err := io.ReadAll(textReader)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func resolvePath(basePath, rawPath, goos string, lookupEnv func(string) (string, bool)) PathResolution {
	resolution := PathResolution{RequestedPath: rawPath, ResolvedPath: rawPath}
	if rawPath == "" {
		return resolution
	}
	isWindowsPath := isWindowsAbsolutePath(rawPath)

	type candidate struct {
		path    string
		warning string
	}

	var existing []candidate
	var fallback []candidate
	seen := map[string]struct{}{}
	addCandidate := func(pathValue, warning string) {
		if pathValue == "" {
			return
		}
		cleaned := cleanForOS(goos, pathValue)
		if _, ok := seen[cleaned]; ok {
			return
		}
		seen[cleaned] = struct{}{}
		entry := candidate{path: cleaned, warning: warning}
		if _, err := os.Stat(cleaned); err == nil {
			existing = append(existing, entry)
		} else {
			fallback = append(fallback, entry)
		}
	}

	if isWindowsPath && goos != "windows" {
		if mappedPath, ok := applyPathMap(rawPath, goos, lookupEnv); ok {
			addCandidate(mappedPath, "resolved path using CAGENT_PATH_MAP")
		}

		if remapped, ok := remapWindowsPathIntoBase(basePath, rawPath, goos); ok {
			addCandidate(remapped, "remapped Windows host path into the current workspace")
		}

		addCandidate(windowsToWSLPath(rawPath), "translated Windows path into WSL mount form")
		addCandidate(normalizeNativePath(basePath, rawPath, goos), "")
	} else {
		addCandidate(normalizeNativePath(basePath, rawPath, goos), "")

		if mappedPath, ok := applyPathMap(rawPath, goos, lookupEnv); ok {
			addCandidate(mappedPath, "resolved path using CAGENT_PATH_MAP")
		}
	}

	selected := candidate{path: cleanForOS(goos, rawPath)}
	if len(existing) > 0 {
		selected = existing[0]
	} else if len(fallback) > 0 {
		selected = fallback[0]
	}

	resolution.ResolvedPath = selected.path
	if selected.warning != "" {
		resolution.Warnings = append(resolution.Warnings, selected.warning)
	}

	return resolution
}

func normalizeNativePath(basePath, rawPath, goos string) string {
	if isAbsoluteForOS(goos, rawPath) || isWindowsAbsolutePath(rawPath) {
		return rawPath
	}
	if basePath == "" {
		return rawPath
	}
	return joinForOS(goos, basePath, rawPath)
}

func applyPathMap(rawPath, goos string, lookupEnv func(string) (string, bool)) (string, bool) {
	mappings, ok := lookupEnv("CAGENT_PATH_MAP")
	if !ok || strings.TrimSpace(mappings) == "" {
		return "", false
	}

	normalizedRaw := strings.ToLower(strings.ReplaceAll(rawPath, `\`, `/`))
	for _, entry := range strings.Split(mappings, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}

		from := strings.TrimSpace(parts[0])
		to := strings.TrimSpace(parts[1])
		if from == "" || to == "" {
			continue
		}

		normalizedFrom := strings.ToLower(strings.ReplaceAll(from, `\`, `/`))
		if !strings.HasPrefix(normalizedRaw, normalizedFrom) {
			continue
		}

		remainder := strings.TrimPrefix(strings.ReplaceAll(rawPath, `\`, `/`), strings.ReplaceAll(from, `\`, `/`))
		remainder = strings.TrimPrefix(remainder, "/")
		if remainder == "" {
			return cleanForOS(goos, to), true
		}
		return joinForOS(goos, to, remainder), true
	}

	return "", false
}

func remapWindowsPathIntoBase(basePath, rawPath, goos string) (string, bool) {
	if basePath == "" {
		return "", false
	}

	baseName := strings.ToLower(filepath.Base(filepath.Clean(basePath)))
	if baseName == "." || baseName == string(filepath.Separator) || baseName == "" {
		return "", false
	}

	segments := splitWindowsPathSegments(rawPath)
	for index := len(segments) - 1; index >= 0; index-- {
		if strings.ToLower(segments[index]) != baseName {
			continue
		}

		if index == len(segments)-1 {
			return basePath, true
		}

		return joinForOS(goos, append([]string{basePath}, segments[index+1:]...)...), true
	}

	return "", false
}

func splitWindowsPathSegments(rawPath string) []string {
	normalized := strings.ReplaceAll(rawPath, `\`, `/`)
	if len(normalized) >= 2 && normalized[1] == ':' {
		normalized = normalized[2:]
	}
	normalized = strings.TrimPrefix(normalized, "/")
	if normalized == "" {
		return nil
	}
	return strings.Split(normalized, "/")
}

func windowsToWSLPath(rawPath string) string {
	normalized := strings.ReplaceAll(rawPath, `\`, "/")
	drive := strings.ToLower(normalized[:1])
	remainder := strings.TrimPrefix(normalized[2:], "/")
	if remainder == "" {
		return path.Join("/mnt", drive)
	}
	return path.Join("/mnt", drive, remainder)
}

func cleanForOS(goos, pathValue string) string {
	if goos == "windows" {
		return filepath.Clean(pathValue)
	}
	return path.Clean(strings.ReplaceAll(pathValue, `\`, "/"))
}

func joinForOS(goos string, parts ...string) string {
	if goos == "windows" {
		return filepath.Join(parts...)
	}
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized = append(normalized, strings.ReplaceAll(part, `\`, "/"))
	}
	return path.Join(normalized...)
}

func isAbsoluteForOS(goos, pathValue string) bool {
	if goos == "windows" {
		return filepath.IsAbs(pathValue)
	}
	return strings.HasPrefix(strings.ReplaceAll(pathValue, `\`, "/"), "/")
}

func isWindowsAbsolutePath(pathValue string) bool {
	return len(pathValue) >= 3 && ((pathValue[0] >= 'A' && pathValue[0] <= 'Z') || (pathValue[0] >= 'a' && pathValue[0] <= 'z')) && pathValue[1] == ':' && (pathValue[2] == '\\' || pathValue[2] == '/')
}
