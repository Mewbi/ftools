// Package frontend embeds the HTML templates and static assets so the Go
// binary can run on its own. In production nginx serves /static/ directly.
package frontend

import (
	"embed"
	"io/fs"
)

//go:embed templates
var Templates embed.FS

// TemplatePatterns are the globs that match every template file, for
// template.ParseFS(frontend.Templates, frontend.TemplatePatterns...).
var TemplatePatterns = []string{
	"templates/*/*.html",
	"templates/*/*/*.html",
}

//go:embed static
var static embed.FS

// Static returns the static assets rooted at static/, to be served under /static/.
func Static() fs.FS {
	sub, err := fs.Sub(static, "static")
	if err != nil {
		panic(err) // only fails if the embed directive above is wrong
	}
	return sub
}
