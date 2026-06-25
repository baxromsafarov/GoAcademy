# --- build the api + migrate binaries (migrations are embedded via go:embed) ---
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -o /out/migrate ./cmd/migrate

# --- minimal runtime ---
# NB: this image runs as a non-root user and contains no Go toolchain or Docker
# client, so the code SANDBOX/JUDGE (SANDBOX_ENABLED) is NOT usable here — those
# need a host with the Go toolchain + Docker access (see README).
FROM alpine:3.20
RUN adduser -D -u 10001 app \
 && mkdir -p /data/storage \
 && chown -R app:app /data
COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/migrate /usr/local/bin/migrate
USER app
WORKDIR /data
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
