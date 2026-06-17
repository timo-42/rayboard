package releasecheck

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/timo-42/rayboard/internal/docscheck"
)

const workflowPath = ".github/workflows/crossbuild.yml"

var requiredWorkflowSnippets = []string{
	"name: Cross Build",
	"go test ./...",
	"make verify-docs",
	"make verify-release",
	"make build-cross",
	"actions/upload-artifact@v4",
	"name: rayboard-mac-arm64-${{ github.sha }}",
	"path: dist/rayboard-darwin-arm64",
	"name: rayboard-linux-amd64-${{ github.sha }}",
	"path: dist/rayboard-linux-amd64",
	"if-no-files-found: error",
}

var requiredMakefileSnippets = []string{
	".PHONY: verify-docs",
	".PHONY: build-cross",
	"GOOS=darwin GOARCH=arm64 go build -o $(DIST)/$(APP)-darwin-arm64 $(CMD)",
	"GOOS=linux GOARCH=amd64 go build -o $(DIST)/$(APP)-linux-amd64 $(CMD)",
}

// Report is the result of validating release-time repository wiring.
type Report struct {
	DocsFilesChecked int
	DocsLinksChecked int
	FilesChecked     int
	Errors           []string
}

// Check validates docs plus release/build workflow wiring from the current repository root.
func Check(root string) Report {
	report := Report{}
	docsReport := docscheck.Check()
	report.DocsFilesChecked = docsReport.FilesChecked
	report.DocsLinksChecked = docsReport.LinksChecked
	for _, err := range docsReport.Errors {
		report.addError("docs: %s", err)
	}
	root = findRoot(root)
	report.checkFileContains(root, workflowPath, requiredWorkflowSnippets)
	report.checkFileContains(root, "Makefile", requiredMakefileSnippets)
	sort.Strings(report.Errors)
	return report
}

func findRoot(root string) string {
	current, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	for {
		if _, err := os.Stat(filepath.Join(current, filepath.FromSlash(workflowPath))); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return root
		}
		current = parent
	}
}

func (report *Report) checkFileContains(root string, name string, snippets []string) {
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		report.addError("read %s: %v", name, err)
		return
	}
	report.FilesChecked++
	text := string(content)
	for _, snippet := range snippets {
		if !strings.Contains(text, snippet) {
			report.addError("%s missing required release wiring %q", name, snippet)
		}
	}
}

func (report *Report) addError(format string, args ...any) {
	report.Errors = append(report.Errors, fmt.Sprintf(format, args...))
}
