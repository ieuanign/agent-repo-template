---
paths:
  - "backend/Dockerfile"
  - "backend/.dockerignore"
  - "backend/.air.toml"
  - "backend/go.mod"
  - "backend/internal/platform/buildinfo/**"
---

# backend: the image

`Dockerfile`, built from `backend/`:

- Build stage: `--platform=$BUILDPLATFORM golang:1.27-alpine` through the public ECR mirror
  (`public.ecr.aws/docker/library/`), which avoids Docker Hub's pull limits; `CGO_ENABLED=0`, so no C
  library is needed. `-ldflags -X` stamps `VERSION` and `GIT_SHA` into `internal/platform/buildinfo`
  (both default to `dev`).
- `dev` target: air over the bind-mounted source, plus the `seed` binary.
- Production target: `gcr.io/distroless/static-debian12:nonroot` with `server` and `migrate`.
- Every image reference carries its tag and its multi-arch index digest, so a rebuild never changes
  base silently. A bump changes both together, the digest read from
  `docker buildx imagetools inspect <image:tag>`.
