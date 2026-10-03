// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewrite(t *testing.T) {
	t.Parallel()
	urls := map[string]string{
		"docs/operations.md":     "/docs/operations.html",
		"docs/adr/0017-media.md": "/docs/adr/0017-media.html",
		"api/openapi.yaml":       "/openapi.yaml",
		"docs/adr":               "/docs/#decisions",
	}
	tests := []struct {
		dest, src, want string
	}{
		{"docs/operations.md", "README.md", "/docs/operations.html"},
		{"docs/operations.md#ai-assistants-mcp", "README.md", "/docs/operations.html#ai-assistants-mcp"},
		{"0017-media.md", "docs/adr/0018-engagement.md", "/docs/adr/0017-media.html"},
		{"adr/0017-media.md", "docs/roadmap.md", "/docs/adr/0017-media.html"},
		{"../api/openapi.yaml", "docs/operations.md", "/openapi.yaml"},
		{"docs/adr/", "README.md", "/docs/#decisions"},
		{"LICENSE", "README.md", Repo + "/blob/main/LICENSE"},
		{"../internal/platform/rules.go", "docs/architecture.md", Repo + "/blob/main/internal/platform/rules.go"},
		{"../internal/platform/", "docs/architecture.md", Repo + "/tree/main/internal/platform"},
		{"https://taskfile.dev", "README.md", "https://taskfile.dev"},
		{"#run-it", "README.md", "#run-it"},
		{"mailto:security@example.com", "SECURITY.md", "mailto:security@example.com"},
		{"../../outside", "docs/roadmap.md", "../../outside"},
	}
	for _, tt := range tests {
		if got := Rewrite(tt.dest, tt.src, urls); got != tt.want {
			t.Errorf("Rewrite(%q, %q) = %q, want %q", tt.dest, tt.src, got, tt.want)
		}
	}
}

// fakeApp writes the files Build reads from the app repo.
func fakeApp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"README.md":              "# Araldo\n\nSee [operations](docs/operations.md) and the [decisions](docs/adr/).\n",
		"docs/operations.md":     "# Operating Araldo\n\n| Variable | Default |\n|---|---|\n| `ARALDO_LISTEN` | `:8080` |\n\n<script>alert(1)</script>\n",
		"docs/architecture.md":   "# Architecture\n\nRules live in [rules.go](../internal/platform/rules.go).\n",
		"docs/errors.md":         "# Error codes\n",
		"docs/roadmap.md":        "# Roadmap\n\nSee [ADR 0001](adr/0001-scope.md).\n",
		"CONTRIBUTING.md":        "# Contributing\n",
		"SECURITY.md":            "# Security policy\n",
		"docs/adr/README.md":     "# Decisions\n",
		"docs/adr/template.md":   "# ADR NNNN\n",
		"docs/adr/0001-scope.md": "# ADR 0001: Scope\n\nBack to the [roadmap](../roadmap.md#now).\n",
		"api/openapi.yaml":       "openapi: 3.1.0\n",
	}
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestBuild(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "dist")
	n, err := Build(fakeApp(t), out)
	if err != nil {
		t.Fatal(err)
	}
	if n != 10 { // seven documents, one ADR, the docs index and the landing page
		t.Fatalf("wrote %d pages", n)
	}
	read := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	checks := map[string][]string{
		"docs/overview.html":       {"<title>Overview · Araldo docs</title>", `href="/docs/operations.html"`, `href="/docs/#decisions"`},
		"docs/operations.html":     {"<h1>Operating Araldo</h1>", "<table>", "<code>ARALDO_LISTEN</code>", "<!-- raw HTML omitted -->"},
		"docs/architecture.html":   {`href="` + Repo + `/blob/main/internal/platform/rules.go"`},
		"docs/adr/0001-scope.html": {`href="/docs/roadmap.html#now"`, "Edit this page on GitHub"},
		"docs/index.html":          {`href="/docs/adr/0001-scope.html">ADR 0001: Scope</a>`, `id="decisions"`},
		"index.html":               {"<h1>One API to announce your product everywhere.</h1>"},
		"openapi.yaml":             {"openapi: 3.1.0"},
		"style.css":                {"--accent"},
	}
	for name, wants := range checks {
		got := read(name)
		for _, want := range wants {
			if !strings.Contains(got, want) {
				t.Errorf("%s lacks %s", name, want)
			}
		}
	}
	if strings.Contains(read("docs/operations.html"), "<script>") {
		t.Error("raw HTML in a document reached the page")
	}
	if _, err := os.Stat(filepath.Join(out, "docs", "adr", "template.html")); !os.IsNotExist(err) {
		t.Error("the ADR template was published")
	}
}

func TestBuildNeedsTheAppRepo(t *testing.T) {
	t.Parallel()
	if _, err := Build(t.TempDir(), filepath.Join(t.TempDir(), "dist")); err == nil {
		t.Fatal("built without the app's docs")
	}
}
