# ftools

FTools, just a fast way to use everyday tools.

## Running

```sh
make -C frontend css          # build the CSS (templates are embedded in the binary)
go run .                      # http://localhost:8080
```

Requires Ghostscript (`gs`) for PDF compression; the server won't start without it (`apt install ghostscript` / `brew install ghostscript`).

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

`server.max_upload_size` (50 MB) is enforced by `maxUploadSizeMiddleware`, registered globally with `r.Use`.

- [ ] Attach the middleware per route instead of globally, with a much smaller limit for `qrcode/generate` (~16 KiB).
- [ ] Set `r.MaxMultipartMemory` so large files spill to disk instead of being held in RAM twice (multipart buffer + `readFormFile`).
- [ ] On `*http.MaxBytesError`, respond `413` with `partials/error`: "This file is larger than 50 MB. Choose a smaller file." (today it's a 422 "file exceeds N bytes").
- [ ] Keep the "up to 50 MB" text in `dropzone.html` in sync with the config.
- [ ] If nginx is added later, set `client_max_body_size` to match.

### QR code generator

- [ ] Tests with mocks:
  - Service: each level produces a PNG; unknown level returns `ErrQRCodeInvalidLevel`.
  - `QRCodeRequest.Validate`: empty or whitespace-only content, invalid size/level, over capacity, bytes vs characters (`"é"` is 2 bytes).
  - The `qrCodeCapacity` table matches the library: content at each limit encodes, one byte more is rejected.
  - Handler with a mocked `service.Service`: success renders the image, a known error returns 422 with its message, an unknown error returns 500 without leaking details.

### Image compression and resizing

- [ ] `POST /image/compress` (multipart). Form: `file`, `quality` (10–100), `max_width` (empty = keep), `format` (empty = same, `jpeg`, `png`, `webp`).
- [ ] Decode JPG/PNG/WebP, resize keeping aspect ratio (candidate: `github.com/disintegration/imaging`), re-encode.
- [ ] Decide on WebP output: the Go standard library and `x/image` only decode WebP. Encoding needs a library (some need cgo, which changes the Dockerfile) or dropping the WebP option.
- [ ] Check the real content type (`http.DetectContentType`), not just the extension.
- [ ] Render `partials/file_result`: `Filename`, `Details`, `OriginalSize`, `ResultSize`, `Saved` (int percent), `DownloadURL`, `DownloadName`, `PreviewURL`.

### PDF compression (Ghostscript)

`POST /pdf/compress` is implemented. Known limitations:

- The download is a base64 data URL in the response (see "Downloads and cleanup").
- `-dPDFSTOPONERROR` is needed: without it gs exits 0 on a password-protected file and writes an empty PDF. Downside: damaged files gs could have repaired are rejected. Revisit if users report valid-looking PDFs failing.

- [ ] Test the presets on real files (scans, text-only, with links). Ghostscript rewrites the PDF and may drop hyperlinks. If it does, mention it on the page.
- [ ] Tests:
  - Integration test behind a build tag or `exec.LookPath` skip: a small fixture PDF compresses, a non-PDF is rejected, a password-protected PDF returns 422.
  - Handler with a mocked service: 422 for validation errors, 500 without details for internal errors.

### PDF → Images (evaluating)

Ghostscript, which compression already needs, can render pages to images (`-sDEVICE=png16m|pnggray|jpeg -r<dpi> -dFirstPage/-dLastPage`), so it needs nothing new installed.

- [ ] Decide where it lives:
  - **New page** (`/pdf/images`): its own tool in the sidebar, clearer for users looking for "PDF to JPG". Duplicates the dropzone and page shell.
  - **Option on the existing PDF page**: an "Action" choice (Compress / Convert to images) that swaps the form fields. One upload area, but the page does two unrelated jobs.
- [ ] Options to offer: format (PNG / JPG), resolution (cap around 200 dpi), page range.
- [ ] Output: one image directly, several pages as a zip (streamed, not built in memory).
- [ ] Guardrails: everything from PDF compression, plus a page cap (`-dLastPage`) and a DPI cap. An A4 page at 300 dpi is about 26 MB of RGB, so unbounded input can exhaust memory and disk.
- [ ] Related: render page 1 of the compressed PDF as a small thumbnail (`-r40`) for the compression result's `PreviewURL`.

### Downloads and cleanup

The site says "Files are deleted as soon as your result is ready", so this has to be true.

- [ ] Decide how results are downloaded:
  - Data URL in the response (simplest, no storage, but heavy for multi-MB PDFs).
  - Temp file + `GET /download/:id` with a random id, deleted after download or after a short TTL.
- [ ] Never keep uploads: process from memory or a temp file and remove it with `defer os.Remove`.
- [ ] Background cleanup for anything left in the temp dir (fx `OnStart`/`OnStop` for the ticker).

### Later

- [ ] Revisit error handling in the web controller. Right now handlers return `422` with the validation error's text (e.g. "qr code: content is empty") and `500` with a generic message for service errors. Consider one shared way to map known errors to a status and a user-facing message, so wording isn't the raw error string and isn't repeated per handler.
- [ ] Server timeouts in `internal/controller/server.go`. Without them, slow or stalled clients can hold connections open forever. At minimum `ReadHeaderTimeout: 10 * time.Second`; add `ReadTimeout`/`WriteTimeout` once uploads exist (long enough for a 50 MB upload on a slow connection). Good candidates for `config.yaml`.
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
- [ ] Processing timeout per request (`context.WithTimeout`) for image work. PDF already has `pdf.timeout`.
- [ ] Rate limiting per IP for the upload routes.
- [ ] Healthcheck endpoint and `healthcheck:` in `docker-compose.yml`.
- [ ] nginx in front (static files, TLS, `client_max_body_size`) once deploying.
- [ ] CI: `go vet`, tests, CSS check, Docker build.
