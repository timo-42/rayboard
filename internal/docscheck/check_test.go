package docscheck

import (
	"testing"
	"testing/fstest"
)

func TestCheckEmbeddedDocs(t *testing.T) {
	report := Check()
	if len(report.Errors) != 0 {
		t.Fatalf("expected embedded docs to pass, got %v", report.Errors)
	}
	if report.FilesChecked < len(requiredDocs) {
		t.Fatalf("expected at least %d docs files, got %d", len(requiredDocs), report.FilesChecked)
	}
	if report.LinksChecked == 0 {
		t.Fatal("expected local docs links to be checked")
	}
}

func TestCheckFSReportsMissingRequiredLink(t *testing.T) {
	docs := fstest.MapFS{
		"README.md": {
			Data: []byte(`# Docs

[Runtime](runtime-config.md)
`),
		},
		"runtime-config.md": {Data: []byte("# Runtime\n")},
	}

	report := CheckFS(docs)

	if len(report.Errors) == 0 {
		t.Fatal("expected docs check errors")
	}
	if !containsError(report.Errors, "missing required docs file api.md") {
		t.Fatalf("expected missing required file error, got %v", report.Errors)
	}
}

func TestCheckFSReportsBrokenLocalLink(t *testing.T) {
	docs := requiredDocsFS()
	docs["README.md"] = &fstest.MapFile{Data: []byte(`# Docs

[Runtime](runtime-config.md)
[Missing](missing.md)
` + upstreamLinksText())}

	report := CheckFS(docs)

	if !containsError(report.Errors, "README.md links missing docs file missing.md") {
		t.Fatalf("expected broken link error, got %v", report.Errors)
	}
}

func requiredDocsFS() fstest.MapFS {
	docs := fstest.MapFS{}
	for _, name := range requiredDocs {
		docs[name] = &fstest.MapFile{Data: []byte("# " + name + "\n")}
	}
	return docs
}

func upstreamLinksText() string {
	out := ""
	for _, link := range requiredUpstreamLinks {
		out += "\n" + link
	}
	return out
}

func containsError(errors []string, want string) bool {
	for _, err := range errors {
		if err == want {
			return true
		}
	}
	return false
}
