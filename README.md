# ftools

FTools, just a fast way to use everyday tools.

## Running

```sh
make -C frontend css          # build the CSS (templates are embedded in the binary)
go run .                      # http://localhost:8080
```

With Docker:

```sh
make -C frontend css
cp .env.example .env          # optional, to change the port
docker compose up -d --build
```

Config lives in `config/config.yaml`; set `CONFIG_PATH` to use another file.

## Next steps

### Don't ship stale CSS

The Docker image embeds `frontend/static/css/app.css` as it is on disk. If templates change and `make css` isn't run, new Tailwind classes are missing in production.

- [ ] Pick one:
  - A root `Makefile` with `docker: ; $(MAKE) -C frontend css && docker compose build`, and only build through it.
  - A CSS stage in the `Dockerfile` that downloads the Tailwind CLI and daisyUI (`make -C frontend setup css`) so the image always builds its own CSS. Slower builds, nothing to remember.
- [ ] Optionally check in CI that `make -C frontend css` leaves no diff.

### Upload limit

Without nginx nothing limits request size yet. The UI promises 20 MB (`frontend/templates/components/dropzone.html`).

- [ ] Add `max_upload_size` to `config.yaml` and `config.Server`.
- [ ] Cap the body before parsing: `ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, limit)` (middleware on the upload routes), and set `r.MaxMultipartMemory` so large files spill to disk instead of RAM.
- [ ] On `*http.MaxBytesError`, respond `413` with `partials/error`: "This file is larger than 20 MB. Choose a smaller file."
- [ ] Keep the "up to 20 MB" text in `dropzone.html` in sync with the config.
- [ ] If nginx is added later, set `client_max_body_size` to match.

### Service layer

- [ ] `internal/service`: `Service` interface + unexported struct, `fx.Module` with `fx.Provide(New)`, added to `main.go`.
- [ ] Request/result types in `entity` (e.g. `QRCodeRequest`, `CompressImageRequest`, `CompressResult`), with no HTTP or template types in them.
- [ ] Web controller gets `service.Service` through `Params`; handlers bind the form, call the service, render a partial.
- [ ] Shared helper to render `partials/error` with a status code and a clear message.

### QR code generator

- [ ] `POST /qrcode/generate`. Form: `content`, `size` (256/512/1024), `level` (low/medium/quartile/high).
- [ ] Generate a PNG (candidate: `github.com/skip2/go-qrcode`).
- [ ] Render `partials/qrcode_result` with `DataURL` as `template.URL` (`data:image/png;base64,...`), `Content`, `Size`. A plain string gets replaced with `#ZgotmplZ`.
- [ ] Validate: empty content, length over 2000, unknown size/level → `422` with `partials/error`.

### Image compression and resizing

- [ ] `POST /image/compress` (multipart). Form: `file`, `quality` (10–100), `max_width` (empty = keep), `format` (empty = same, `jpeg`, `png`, `webp`).
- [ ] Decode JPG/PNG/WebP, resize keeping aspect ratio (candidate: `github.com/disintegration/imaging`), re-encode.
- [ ] Decide on WebP output: the Go standard library and `x/image` only decode WebP. Encoding needs a library (some need cgo, which changes the Dockerfile) or dropping the WebP option.
- [ ] Check the real content type (`http.DetectContentType`), not just the extension.
- [ ] Render `partials/file_result`: `Filename`, `Details`, `OriginalSize`, `ResultSize`, `Saved` (int percent), `DownloadURL`, `DownloadName`, `PreviewURL`.

### PDF compression

- [ ] `POST /pdf/compress` (multipart). Form: `file`, `level` (low/medium/high).
- [ ] Choose the engine:
  - `pdfcpu` (pure Go): removes duplicate objects and optimizes structure, but doesn't recompress images, so savings on scanned PDFs are small.
  - Ghostscript (`gs -dPDFSETTINGS=/printer|/ebook|/screen`): much better compression and maps to the three levels, but it's an external binary (`apk add ghostscript` in the runtime image) and must be run with a timeout on untrusted files.
- [ ] Render `partials/file_result` (no `PreviewURL`).

### Downloads and cleanup

The site says "Files are deleted as soon as your result is ready", so this has to be true.

- [ ] Decide how results are downloaded:
  - Data URL in the response (simplest, no storage, but heavy for multi-MB PDFs).
  - Temp file + `GET /download/:id` with a random id, deleted after download or after a short TTL.
- [ ] Never keep uploads: process from memory or a temp file and remove it with `defer os.Remove`.
- [ ] Background cleanup for anything left in the temp dir (fx `OnStart`/`OnStop` for the ticker).

### Later

- [ ] Server timeouts in `internal/controller/server.go`. Without them, slow or stalled clients can hold connections open forever. At minimum `ReadHeaderTimeout: 10 * time.Second`; add `ReadTimeout`/`WriteTimeout` once uploads exist (long enough for a 20 MB upload on a slow connection). Good candidates for `config.yaml`.
- [ ] 404 page: `r.NoRoute(...)` rendering a page with the layout, instead of Gin's plain "404 page not found".
- [ ] Log server failures:
  - `go srv.Serve(ln)` drops its error, so a server that dies after startup leaves no trace. Log anything that isn't `http.ErrServerClosed`.
  - Replace `p.Logger.Fatal(...)` in `New` with returning `(*server, error)`. `Fatal` calls `os.Exit` and skips fx's shutdown.
- [ ] Request logs: `gin.New()` has no logging middleware. Add `gin.Logger()` or a small zap middleware (method, path, status, duration).
- [ ] Rotate container logs. Docker's default json-file driver never rotates. In `docker-compose.yml`:
  ```yaml
  logging:
    options:
      max-size: "10m"
      max-file: "3"
  ```
- [ ] Processing timeout per request (`context.WithTimeout` around image/PDF work).
- [ ] Rate limiting per IP for the upload routes.
- [ ] Healthcheck endpoint and `healthcheck:` in `docker-compose.yml`.
- [ ] nginx in front (static files, TLS, `client_max_body_size`) once deploying.
- [ ] CI: `go vet`, tests, CSS check, Docker build.
