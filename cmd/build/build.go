// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Repo is where links into the app's code go.
const Repo = "https://github.com/spectrum-labs-tech/araldo"

// SiteURL serves the landing page and DocsURL the documentation, each from
// its own directory of the build (site and docs). Links between them are
// absolute; links within one are paths.
const (
	SiteURL = "https://araldo.dev"
	DocsURL = "https://docs.araldo.dev"
)

// The hosted plan: AppURL is its dashboard and AccountURL where people
// sign up and pay.
const (
	AppURL     = "https://app.araldo.dev"
	AccountURL = "https://account.araldo.dev"
)

// redirects sends the docs' old addresses on araldo.dev, under /docs/, to
// the same page on the docs host (Cloudflare Pages' _redirects).
const redirects = `/docs ` + DocsURL + `/ 301
/docs/ ` + DocsURL + `/ 301
/docs/* ` + DocsURL + `/:splat 301
/openapi.yaml ` + DocsURL + `/openapi.yaml 301
`

//go:embed site
var siteFS embed.FS

// Doc is one page made from a document in the app repo.
type Doc struct {
	// Src is its path in the app repo, with forward slashes.
	Src string
	// URL is its absolute path on the docs host.
	URL string
	// Section groups it in the docs index.
	Section string
	// Title is its first heading, unless set here.
	Title string
	Body  template.HTML
}

// docs lists the app repo's documents and where each goes. ADRs are added
// from the adr directory.
var docs = []Doc{
	{Src: "README.md", URL: "/overview.html", Section: "Start", Title: "Overview"},
	{Src: "docs/operations.md", URL: "/operations.html", Section: "Start"},
	{Src: "docs/architecture.md", URL: "/architecture.html", Section: "Start"},
	{Src: "docs/errors.md", URL: "/errors.html", Section: "Reference"},
	{Src: "docs/roadmap.md", URL: "/roadmap.html", Section: "Project"},
	{Src: "CONTRIBUTING.md", URL: "/contributing.html", Section: "Project"},
	{Src: "SECURITY.md", URL: "/security.html", Section: "Project"},
}

// copied are files published as they are.
var copied = map[string]string{"api/openapi.yaml": "/openapi.yaml"}

// Build writes the site from the app checkout in app into out, and returns
// how many pages it wrote.
func Build(app, out string) (int, error) {
	all, err := listDocs(app)
	if err != nil {
		return 0, err
	}
	urls := map[string]string{}
	for _, d := range all {
		urls[d.Src] = d.URL
	}
	for src, url := range copied {
		urls[src] = url
	}
	urls["docs/adr"] = "/#decisions"
	urls["docs/adr/"] = "/#decisions"

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"site":    func(p string) string { return SiteURL + p },
		"docs":    func(p string) string { return DocsURL + p },
		"app":     func(p string) string { return AppURL + p },
		"account": func(p string) string { return AccountURL + p },
	}).ParseFS(siteFS, "site/*.html")
	if err != nil {
		return 0, err
	}
	if err := os.RemoveAll(out); err != nil {
		return 0, err
	}
	site, docsDir := filepath.Join(out, "site"), filepath.Join(out, "docs")
	n := 0
	for i := range all {
		d := &all[i]
		raw, err := os.ReadFile(filepath.Join(app, filepath.FromSlash(d.Src)))
		if err != nil {
			return 0, err
		}
		title, body, err := render(raw, d.Src, urls)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", d.Src, err)
		}
		if d.Title == "" {
			d.Title = title
		}
		d.Body = body
		if err := write(docsDir, d.URL, tmpl, "doc.html", map[string]any{"Doc": d, "Docs": all}); err != nil {
			return 0, err
		}
		n++
	}
	if err := write(docsDir, "/index.html", tmpl, "docs.html", map[string]any{"Sections": sections(all)}); err != nil {
		return 0, err
	}
	if err := write(site, "/index.html", tmpl, "index.html", nil); err != nil {
		return 0, err
	}
	if err := writeFile(site, "/_redirects", []byte(redirects)); err != nil {
		return 0, err
	}
	n += 2
	for src, url := range copied {
		data, err := os.ReadFile(filepath.Join(app, filepath.FromSlash(src)))
		if err != nil {
			return 0, err
		}
		if err := writeFile(docsDir, url, data); err != nil {
			return 0, err
		}
	}
	static, err := fs.Sub(siteFS, "site/static")
	if err != nil {
		return 0, err
	}
	err = fs.WalkDir(static, ".", func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		data, err := fs.ReadFile(static, p)
		if err != nil {
			return err
		}
		for _, dir := range []string{site, docsDir} {
			if err := writeFile(dir, "/"+p, data); err != nil {
				return err
			}
		}
		return nil
	})
	return n, err
}

// listDocs is the fixed documents and every ADR but the template.
func listDocs(app string) ([]Doc, error) {
	out := append([]Doc(nil), docs...)
	entries, err := os.ReadDir(filepath.Join(app, "docs", "adr"))
	if err != nil {
		return nil, err
	}
	var adrs []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".md") && name != "README.md" && name != "template.md" {
			adrs = append(adrs, name)
		}
	}
	sort.Strings(adrs)
	for _, name := range adrs {
		out = append(out, Doc{Src: "docs/adr/" + name, URL: "/adr/" + strings.TrimSuffix(name, ".md") + ".html", Section: "Decisions"})
	}
	return out, nil
}

// Section is a group of docs in the index.
type Section struct {
	Name string
	Docs []Doc
}

func sections(all []Doc) []Section {
	var out []Section
	for _, name := range []string{"Start", "Reference", "Decisions", "Project"} {
		s := Section{Name: name}
		for _, d := range all {
			if d.Section == name {
				s.Docs = append(s.Docs, d)
			}
		}
		out = append(out, s)
	}
	return out
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID(), parser.WithASTTransformers(util.Prioritized(linkRewriter{}, 100))),
)

// render turns a document into HTML, taking its title from its first
// heading, which the page shows on its own.
func render(raw []byte, src string, urls map[string]string) (string, template.HTML, error) {
	ctx := parser.NewContext()
	ctx.Set(srcKey, src)
	ctx.Set(urlsKey, urls)
	doc := md.Parser().Parse(text.NewReader(raw), parser.WithContext(ctx))
	title := ""
	if h, ok := doc.FirstChild().(*ast.Heading); ok && h.Level == 1 {
		title = string(h.Lines().Value(raw))
		doc.RemoveChild(doc, h)
	}
	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, raw, doc); err != nil {
		return "", "", err
	}
	return title, template.HTML(buf.String()), nil //nolint:gosec // the app repo's own docs, rendered by goldmark, which escapes raw HTML
}

var (
	srcKey  = parser.NewContextKey()
	urlsKey = parser.NewContextKey()
)

// linkRewriter points links between documents at their pages, and links
// into the code at GitHub.
type linkRewriter struct{}

func (linkRewriter) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	src, _ := pc.Get(srcKey).(string)
	urls, _ := pc.Get(urlsKey).(map[string]string)
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch l := n.(type) {
		case *ast.Link:
			l.Destination = []byte(Rewrite(string(l.Destination), src, urls))
		case *ast.Image:
			l.Destination = []byte(Rewrite(string(l.Destination), src, urls))
		}
		return ast.WalkContinue, nil
	})
}

// Rewrite resolves a link written in the document at src: a link to
// another published document becomes its page, any other path in the repo
// becomes its GitHub page, and anything else (other sites, anchors) is left
// as it is.
func Rewrite(dest, src string, urls map[string]string) string {
	if dest == "" || strings.HasPrefix(dest, "#") || strings.Contains(dest, "://") || strings.HasPrefix(dest, "mailto:") {
		return dest
	}
	target, frag, _ := strings.Cut(dest, "#")
	p := path.Clean(path.Join(path.Dir(src), target))
	if strings.HasPrefix(p, "../") || p == ".." {
		return dest
	}
	if url, ok := urls[p]; ok {
		if frag != "" && !strings.Contains(url, "#") {
			url += "#" + frag
		}
		return url
	}
	kind := "blob"
	if strings.HasSuffix(target, "/") {
		kind = "tree"
	}
	url := Repo + "/" + kind + "/main/" + p
	if frag != "" {
		url += "#" + frag
	}
	return url
}

func write(out, url string, tmpl *template.Template, name string, data any) error {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return fmt.Errorf("%s: %w", url, err)
	}
	return writeFile(out, url, buf.Bytes())
}

func writeFile(out, url string, data []byte) error {
	p := filepath.Join(out, filepath.FromSlash(strings.TrimPrefix(url, "/")))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644) //nolint:gosec // a public website's files
}
