package frontend

import (
	"bytes"
	"html"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"

	projectdocs "github.com/timo-42/rayboard/docs"
)

var docsLinkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

type docPage struct {
	Slug  string
	File  string
	Title string
	HTML  template.HTML
	Pages []docNavItem
}

type docNavItem struct {
	Slug    string
	Title   string
	Current bool
}

func docsHandler() http.Handler {
	pages, err := loadDocPages()
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "docs unavailable", http.StatusInternalServerError)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/docs")
		slug = strings.Trim(slug, "/")
		if slug == "" {
			slug = "README"
		}
		slug = strings.TrimSuffix(slug, ".html")
		slug = strings.TrimSuffix(slug, ".md")
		page, ok := pages[slug]
		if !ok {
			http.NotFound(w, r)
			return
		}
		renderDocs(w, page)
	})
}

func loadDocPages() (map[string]docPage, error) {
	entries, err := fs.ReadDir(projectdocs.Files, ".")
	if err != nil {
		return nil, err
	}
	items := make([]docNavItem, 0, len(entries))
	pages := map[string]docPage{}
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".md" {
			continue
		}
		content, err := projectdocs.Files.ReadFile(entry.Name())
		if err != nil {
			return nil, err
		}
		slug := strings.TrimSuffix(entry.Name(), ".md")
		title := docTitle(content, slug)
		page := docPage{
			Slug:  slug,
			File:  entry.Name(),
			Title: title,
			HTML:  markdownHTML(string(content)),
		}
		pages[slug] = page
		items = append(items, docNavItem{Slug: slug, Title: title})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Slug == "README" {
			return true
		}
		if items[j].Slug == "README" {
			return false
		}
		return items[i].Title < items[j].Title
	})
	for slug, page := range pages {
		page.Pages = make([]docNavItem, 0, len(items))
		for _, item := range items {
			item.Current = item.Slug == slug
			page.Pages = append(page.Pages, item)
		}
		pages[slug] = page
	}
	return pages, nil
}

func docTitle(content []byte, fallback string) string {
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return strings.ReplaceAll(fallback, "-", " ")
}

func markdownHTML(markdown string) template.HTML {
	var out bytes.Buffer
	lines := strings.Split(markdown, "\n")
	inCode := false
	inList := false
	inTable := false
	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inList {
				out.WriteString("</ul>\n")
				inList = false
			}
			if inTable {
				out.WriteString("</tbody></table>\n")
				inTable = false
			}
			if inCode {
				out.WriteString("</code></pre>\n")
				inCode = false
			} else {
				out.WriteString("<pre><code>")
				inCode = true
			}
			continue
		}
		if inCode {
			out.WriteString(html.EscapeString(line))
			out.WriteByte('\n')
			continue
		}
		if trimmed == "" {
			if inList {
				out.WriteString("</ul>\n")
				inList = false
			}
			if inTable {
				out.WriteString("</tbody></table>\n")
				inTable = false
			}
			continue
		}
		if isTableSeparator(trimmed) {
			continue
		}
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			if inList {
				out.WriteString("</ul>\n")
				inList = false
			}
			cells := tableCells(trimmed)
			if !inTable {
				out.WriteString("<table><tbody>\n")
				inTable = true
			}
			out.WriteString("<tr>")
			for _, cell := range cells {
				out.WriteString("<td>")
				out.WriteString(inlineMarkdown(cell))
				out.WriteString("</td>")
			}
			out.WriteString("</tr>\n")
			continue
		}
		if inTable {
			out.WriteString("</tbody></table>\n")
			inTable = false
		}
		if strings.HasPrefix(trimmed, "- ") {
			if !inList {
				out.WriteString("<ul>\n")
				inList = true
			}
			out.WriteString("<li>")
			out.WriteString(inlineMarkdown(strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))))
			out.WriteString("</li>\n")
			continue
		}
		if inList {
			out.WriteString("</ul>\n")
			inList = false
		}
		if level, text, ok := heading(trimmed); ok {
			out.WriteString("<h")
			out.WriteByte(byte('0' + level))
			out.WriteString(">")
			out.WriteString(inlineMarkdown(text))
			out.WriteString("</h")
			out.WriteByte(byte('0' + level))
			out.WriteString(">\n")
			continue
		}
		out.WriteString("<p>")
		out.WriteString(inlineMarkdown(trimmed))
		out.WriteString("</p>\n")
	}
	if inCode {
		out.WriteString("</code></pre>\n")
	}
	if inList {
		out.WriteString("</ul>\n")
	}
	if inTable {
		out.WriteString("</tbody></table>\n")
	}
	return template.HTML(out.String())
}

func heading(line string) (int, string, bool) {
	for level := 1; level <= 6; level++ {
		prefix := strings.Repeat("#", level) + " "
		if strings.HasPrefix(line, prefix) {
			return level, strings.TrimSpace(strings.TrimPrefix(line, prefix)), true
		}
	}
	return 0, "", false
}

func inlineMarkdown(text string) string {
	escaped := html.EscapeString(text)
	return docsLinkPattern.ReplaceAllStringFunc(escaped, func(match string) string {
		groups := docsLinkPattern.FindStringSubmatch(match)
		if len(groups) != 3 {
			return match
		}
		href := strings.TrimSuffix(groups[2], ".md")
		if !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") && !strings.HasPrefix(href, "#") {
			href = "/docs/" + strings.TrimPrefix(href, "./")
		}
		return `<a href="` + html.EscapeString(href) + `">` + groups[1] + `</a>`
	})
}

func isTableSeparator(line string) bool {
	line = strings.Trim(line, "| ")
	if line == "" {
		return false
	}
	for _, part := range strings.Split(line, "|") {
		part = strings.TrimSpace(part)
		if strings.Trim(part, "-:") != "" {
			return false
		}
	}
	return true
}

func tableCells(line string) []string {
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, 0, len(parts))
	for _, part := range parts {
		cells = append(cells, strings.TrimSpace(part))
	}
	return cells
}

func renderDocs(w http.ResponseWriter, page docPage) {
	tpl, err := template.ParseFS(assets, "templates/docs.html")
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tpl.Execute(w, page)
}
