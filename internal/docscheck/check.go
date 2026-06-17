package docscheck

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"

	projectdocs "github.com/timo-42/rayboard/docs"
)

var markdownLinkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

var requiredDocs = []string{
	"README.md",
	"runtime-config.md",
	"auth-rbac.md",
	"user-guide.md",
	"admin-guide.md",
	"api.md",
	"frontend.md",
	"demo-seed.md",
	"automation-lua.md",
	"development.md",
	"operations.md",
}

var requiredUpstreamLinks = []string{
	"https://cel.dev/",
	"https://github.com/google/cel-go",
	"https://github.com/yuin/gopher-lua",
	"https://pkg.go.dev/github.com/robfig/cron/v3",
	"https://github.com/robfig/cron",
	"https://htmx.org/",
	"https://sortablejs.github.io/Sortable/",
	"https://codemirror.net/",
	"https://openrouter.ai/docs",
	"https://github.com/containrrr/shoutrrr",
	"https://containrrr.dev/shoutrrr/",
	"https://www.sqlite.org/fts5.html",
}

// Report is the result of checking the embedded documentation.
type Report struct {
	FilesChecked int
	LinksChecked int
	Errors       []string
}

// Check validates the embedded documentation index and local markdown links.
func Check() Report {
	return CheckFS(projectdocs.Files)
}

// CheckFS validates a documentation filesystem. It is exported for tests.
func CheckFS(docs fs.FS) Report {
	report := Report{}
	entries, err := fs.ReadDir(docs, ".")
	if err != nil {
		report.addError("read docs: %v", err)
		return report
	}

	files := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".md" {
			continue
		}
		content, err := fs.ReadFile(docs, entry.Name())
		if err != nil {
			report.addError("read %s: %v", entry.Name(), err)
			continue
		}
		files[entry.Name()] = string(content)
	}
	report.FilesChecked = len(files)

	report.checkRequiredFiles(files)
	report.checkREADMEIndex(files)
	report.checkUpstreamLinks(files)
	report.checkLocalLinks(files)
	sort.Strings(report.Errors)
	return report
}

func (report *Report) checkRequiredFiles(files map[string]string) {
	for _, name := range requiredDocs {
		if _, ok := files[name]; !ok {
			report.addError("missing required docs file %s", name)
		}
	}
}

func (report *Report) checkREADMEIndex(files map[string]string) {
	readme, ok := files["README.md"]
	if !ok {
		return
	}
	for _, name := range requiredDocs {
		if name == "README.md" {
			continue
		}
		if !strings.Contains(readme, "("+name+")") {
			report.addError("README.md does not link required docs file %s", name)
		}
	}
}

func (report *Report) checkUpstreamLinks(files map[string]string) {
	all := strings.Builder{}
	for name, content := range files {
		all.WriteString("\n--- ")
		all.WriteString(name)
		all.WriteString(" ---\n")
		all.WriteString(content)
	}
	docsText := all.String()
	for _, link := range requiredUpstreamLinks {
		if !strings.Contains(docsText, link) {
			report.addError("missing required upstream docs link %s", link)
		}
	}
}

func (report *Report) checkLocalLinks(files map[string]string) {
	for source, content := range files {
		for _, match := range markdownLinkPattern.FindAllStringSubmatch(content, -1) {
			if len(match) != 3 {
				continue
			}
			target := strings.TrimSpace(match[2])
			if shouldSkipLink(target) {
				continue
			}
			report.LinksChecked++
			target = strings.SplitN(target, "#", 2)[0]
			target = strings.TrimPrefix(target, "./")
			target = path.Clean(path.Join(path.Dir(source), target))
			if path.Ext(target) == "" {
				target += ".md"
			}
			if _, ok := files[target]; !ok {
				report.addError("%s links missing docs file %s", source, match[2])
			}
		}
	}
}

func shouldSkipLink(target string) bool {
	target = strings.TrimSpace(target)
	if target == "" || strings.HasPrefix(target, "#") {
		return true
	}
	lower := strings.ToLower(target)
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "mailto:")
}

func (report *Report) addError(format string, args ...any) {
	report.Errors = append(report.Errors, fmt.Sprintf(format, args...))
}
