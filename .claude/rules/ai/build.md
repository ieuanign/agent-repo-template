---
paths:
  - "ai/Dockerfile"
  - "ai/.dockerignore"
  - "ai/pyproject.toml"
  - "ai/src/aisvc/platform/buildinfo.py"
---

# ai: the image

`Dockerfile`, built from `ai/`, every stage `FROM python:3.14-slim` through the public ECR mirror
(`public.ecr.aws/docker/library/`), which avoids Docker Hub's pull limits:

- Every image reference carries its tag and its multi-arch index digest, so a rebuild never changes
  base silently. A bump changes both together, the digest read from
  `docker buildx imagetools inspect <image:tag>`.
- Build stage: uv copied from `ghcr.io/astral-sh/uv`, pinned to the version `devbox.lock` resolves, so
  the host and the image read `uv.lock` alike; the pin moves whenever `devbox.lock`'s uv does. It
  writes `src/aisvc/platform/_buildinfo.py` from the `VERSION` and `GIT_SHA` args (both default to
  `dev`), then `uv sync --locked --no-dev --no-editable` into `/opt/venv`. uv never downloads a
  Python.
- `_buildinfo.py` is gitignored and listed as a hatch wheel `artifacts` entry; hatchling drops
  VCS-ignored files from wheels otherwise, and production would report `dev`.
- `dev` target: the dev group with editable installs, and watchfiles restarting `python -m aisvc
  serve` over the bind-mounted `src`. The mount hides the image's `_buildinfo.py`, so dev reports
  `dev`.
- Production target: the base image plus `/opt/venv` only, with no uv and no source tree, run as a
  non-root user.
