# proto

The buf module holding the backend → ai gRPC contract. Package `ai.v1` (`ai/v1/info.proto`) defines
`InfoService.GetInfo`, which returns ai's service name and version.

## Generating

```sh
devbox run -- npm run generate -w proto
```

buf generates with remote plugins pinned in `buf.gen.yaml` and writes committed output to:

- `backend/gen/ai/v1/` — Go package `aiv1`
- `ai/gen/ai/v1/` — Python modules `ai.v1.info_pb2` and `ai.v1.info_pb2_grpc`, plus the type stub
  `info_pb2.pyi`

Never edit that output by hand. `ai/gen/pyproject.toml` and `ai/gen/ai/v1/py.typed` are the two
hand-written files there, and regenerate-and-diff exempts both: `pyproject.toml` makes the output
installable, and `ai/pyproject.toml` depends on it by path; `py.typed` is the
[PEP 561](https://peps.python.org/pep-0561/) marker, without which a type checker ignores the
installed stubs.

Bumping a plugin version means bumping the matching runtime too — `google.golang.org/protobuf` /
`google.golang.org/grpc` in `backend/go.mod`, `protobuf` / `grpcio` in `ai/uv.lock` — since generated
code refuses to compile or import against an older runtime. `buf.build/protocolbuffers/python` and
`buf.build/protocolbuffers/pyi` always move to the same tag, so the stubs describe the modules beside
them.

## Checks

`npm run lint -w proto` (part of `turbo run lint`, so `make check`) runs `scripts/lint.sh`:

1. `buf lint` (`STANDARD`).
2. `buf breaking` (`FILE`) against the local `main` branch's `proto/`; skipped while `main` has no
   `proto/buf.yaml`.
3. Regenerate into `.scratch/` and `diff -r` against the committed `backend/gen` and `ai/gen`, so a
   hand-edit, a missing or extra file, or an unregenerated `.proto` change all fail.
   `backend/gen/openapi` is excluded: oapi-codegen writes it, and `backend#lint` diffs it.

## Moving the generated code

`backend/gen` and `ai/gen` are placeholders the backend and ai structure work may move. Moving either
means changing, together:

- `out` in `buf.gen.yaml`, and `go_package_prefix` there for the Go side;
- the paths diffed in `scripts/lint.sh`;
- the `path` of `ai-gen` under `[tool.uv.sources]` in `ai/pyproject.toml`.

Package `ai.v1` claims the Python top-level name `ai`. ai's own code must either live under another
top-level package or absorb `ai/v1` into its own `ai` package.
