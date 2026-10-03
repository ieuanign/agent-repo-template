FROM public.ecr.aws/docker/library/python:3.14-slim@sha256:51dafde81dbdb6ebde285137a295cf18a47ca95234fe388a343719cb97305b3d AS build
# Pinned to devbox.lock's uv so the host and the image read uv.lock alike.
COPY --from=ghcr.io/astral-sh/uv:0.12.5@sha256:e85be844203885286c60ffad8a858d48afb6c5a5c237ca0e67f12e74b8f174b1 /uv /usr/local/bin/uv
# Outside WORKDIR, which Compose bind-mounts over; production copies it without uv, so no downloaded Python.
ENV UV_PROJECT_ENVIRONMENT=/opt/venv \
    UV_PYTHON_DOWNLOADS=never \
    PATH=/opt/venv/bin:$PATH
WORKDIR /app
COPY . .
ARG VERSION=dev
ARG GIT_SHA=dev
RUN printf 'VERSION = "%s"\nCOMMIT = "%s"\n' "$VERSION" "$GIT_SHA" > src/aisvc/platform/_buildinfo.py \
 && uv sync --locked --no-dev --no-editable

FROM build AS dev
# Editable, so the bind-mounted source is what runs.
RUN uv sync --locked
CMD ["watchfiles", "--filter", "python", "--sigint-timeout", "20", "python -m aisvc serve", "src"]

FROM public.ecr.aws/docker/library/python:3.14-slim@sha256:51dafde81dbdb6ebde285137a295cf18a47ca95234fe388a343719cb97305b3d
COPY --from=build /opt/venv /opt/venv
ENV PATH=/opt/venv/bin:$PATH
RUN useradd --system --no-create-home aisvc
USER aisvc
HEALTHCHECK --interval=5s --timeout=5s --start-period=30s --start-interval=1s --retries=3 \
  CMD ["python", "-m", "aisvc", "healthcheck"]
CMD ["python", "-m", "aisvc", "serve"]
