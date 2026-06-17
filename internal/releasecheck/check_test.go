package releasecheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckCurrentRepository(t *testing.T) {
	report := Check("../..")
	if len(report.Errors) != 0 {
		t.Fatalf("expected release check to pass, got %v", report.Errors)
	}
	if report.DocsFilesChecked == 0 || report.DocsLinksChecked == 0 || report.FilesChecked != 2 {
		t.Fatalf("unexpected report counts: %#v", report)
	}
}

func TestCheckReportsMissingWorkflowArtifact(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Makefile", strings.Join(requiredMakefileSnippets, "\n"))
	writeFile(t, root, workflowPath, strings.ReplaceAll(strings.Join(requiredWorkflowSnippets, "\n"), "rayboard-linux-amd64-${{ github.sha }}", "rayboard-linux-${{ github.sha }}"))

	report := Check(root)

	if !containsError(report.Errors, workflowPath+" missing required release wiring \"name: rayboard-linux-amd64-${{ github.sha }}\"") {
		t.Fatalf("expected missing linux artifact name error, got %v", report.Errors)
	}
}

func TestCheckReportsMissingWorkflowReleaseVerify(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Makefile", strings.Join(requiredMakefileSnippets, "\n"))
	writeFile(t, root, workflowPath, strings.ReplaceAll(strings.Join(requiredWorkflowSnippets, "\n"), "make verify-release", "make verify-docs"))

	report := Check(root)

	if !containsError(report.Errors, workflowPath+" missing required release wiring \"make verify-release\"") {
		t.Fatalf("expected missing verify-release error, got %v", report.Errors)
	}
}

func writeFile(t *testing.T, root string, name string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func containsError(errors []string, want string) bool {
	for _, err := range errors {
		if err == want {
			return true
		}
	}
	return false
}
