FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Templates and static files are embedded into the binary (frontend/embed.go).
# static/css/app.css must already be built: run `make css` in frontend/ before building the image.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o ftools .

FROM alpine:3.24
RUN adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=builder /app/ftools .
COPY --from=builder /app/config/config.yaml ./config/config.yaml
USER app
ENV GIN_MODE=release
EXPOSE 8080
CMD ["./ftools"]
