FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS build
# Set here so the dev stage builds without cgo as well.
ENV CGO_ENABLED=0
WORKDIR /src
# No `go mod download`: go.mod also lists every driver SDK the migrate CLI tool requires, and the build fetches only what it imports.
COPY . .
ARG TARGETOS TARGETARCH
ARG VERSION=dev
ARG GIT_SHA=dev
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
      -ldflags "-s -w -X github.com/ieuanign/agent-repo-template/backend/internal/platform/buildinfo.Version=$VERSION -X github.com/ieuanign/agent-repo-template/backend/internal/platform/buildinfo.Commit=$GIT_SHA" \
      -o /out/ ./cmd/server ./cmd/migrate

FROM build AS dev
WORKDIR /src
# Outside WORKDIR, which Compose bind-mounts over; dev-only so production never holds it.
RUN go build -trimpath -o /usr/local/bin/seed ./cmd/seed
# Warms air's build; the server's dependencies are already cached by the build stage.
RUN go tool air -v
CMD ["go", "tool", "air"]

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /out/server /out/migrate /
# Exec form: distroless has no shell.
HEALTHCHECK --interval=5s --timeout=5s --start-period=30s --start-interval=1s --retries=3 \
  CMD ["/server", "healthcheck"]
ENTRYPOINT ["/server"]
