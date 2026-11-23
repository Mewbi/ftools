# frontend

Server-rendered HTML templates and static assets.
Stack: Go `html/template`, HTMX, Tailwind CSS v4 + daisyUI v5, Alpine.js (only where local state is needed), Lucide icons as an SVG sprite, Geist font. No Node: Tailwind runs as the standalone CLI.

```text
frontend/
├── embed.go              go:embed for templates and static (package frontend)
├── Makefile              setup, css, watch, icons, vendor
├── styles/app.css        Tailwind + daisyUI entry point and the two themes
├── scripts/build-icons.py
├── templates/
│   ├── layouts/base.html       page shell: "layouts/start" and "layouts/end"
│   ├── components/             sidebar, icon, logo, tool_header, result_empty, dropzone
│   ├── pages/home.html         "pages/home"
│   ├── pages/tools/*.html      "pages/tools/qrcode", "pages/tools/image", "pages/tools/pdf"
│   └── partials/results.html   HTMX fragments: qrcode_result, file_result, error
└── static/                     served at /static/ (by nginx in production)
    ├── css/app.css             built, don't edit
    ├── js/                     htmx, alpine (vendored), app.js
    ├── fonts/  images/  icons.svg
```

## Building CSS

```sh
cd frontend
make setup   # once: downloads the Tailwind CLI and daisyUI into bin/ and styles/vendor/ (gitignored)
make watch   # while editing templates
make css     # minified build; commit static/css/app.css
```

Tailwind only generates classes it finds in `templates/` and `static/js/app.js`, so rebuild after changing templates.

## Using it from Go

Every template has a unique name (`{{define "pages/home"}}`), so all of them parse into one set:

```go
tpl := template.Must(template.ParseFS(frontend.Templates, frontend.TemplatePatterns...))
r.SetHTMLTemplate(tpl)
r.StaticFS("/static", http.FS(frontend.Static()))
```

| Route | Template | Data |
|---|---|---|
| `GET /` | `pages/home` | `Title`, `Path` |
| `GET /qrcode` | `pages/tools/qrcode` | `Title`, `Path` |
| `GET /image` | `pages/tools/image` | `Title`, `Path` |
| `GET /pdf` | `pages/tools/pdf` | `Title`, `Path` |
| `POST /qrcode/generate` | `partials/qrcode_result` | see the comment in `partials/results.html` |
| `POST /image/compress` | `partials/file_result` | same |
| `POST /pdf/compress` | `partials/file_result` | same |
| any failed POST | `partials/error` | `Message`, with a 4xx/5xx status |

`Path` is the request path (`c.Request.URL.Path`) and highlights the current page in the sidebar.

Form fields the handlers receive:

- QR code: `content`, `size` (256, 512, 1024), `level` (low, medium, quartile, high)
- Image (multipart): `file`, `quality` (10–100), `max_width` (empty = keep), `format` (empty = same, jpeg, png, webp)
- PDF (multipart): `file`, `level` (low, medium, high)

## Adding a tool

1. Copy a page in `templates/pages/tools/` and change the form fields and `hx-post` URL.
2. Add a link in `components/sidebar.html` and a row in `pages/home.html`.
3. Reuse `partials/file_result` for anything that turns one file into a smaller one; otherwise add a fragment to `partials/`.
4. Need an icon? Add its Lucide name to `scripts/build-icons.py` and run `make icons`.
5. `make css`.
