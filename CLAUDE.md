# CLAUDE.md

## Working principles

These bias toward caution over speed; for a trivial task, use judgment.

- **State assumptions.** Before implementing, say what you assume and surface the tradeoffs. Present competing readings instead of picking one silently; where something is unclear, stop, name it and ask; say when a simpler approach exists, and push back when warranted.
- **Simplicity first.** Write the minimum code that solves the problem: no feature or configurability nobody asked for, no abstraction for single-use code, no handling of cases that cannot happen.
- **Surgical changes.** Every changed line traces to the request. Match the existing style, remove what your own change orphaned and nothing else, and mention unrelated dead code instead of deleting it.
- **Goal-driven execution.** Turn the task into a check that can fail, such as a test that reproduces the bug or tests that pass before and after a refactor, and loop until it passes. For a multi-step task, state a brief plan as steps, each with its check.

## Agent skills

### Issue tracker

Issues live as GitHub issues in this repository's own GitHub repo, managed with the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default five-role triage vocabulary plus three `/dev-loop auto` workflow roles (`in-progress`, `awaiting-human`, `failed`); each label string equals its canonical role name. See `docs/agents/triage-labels.md`.

### Domain docs

Multi-context: root `CONTEXT-MAP.md` pointing at one `CONTEXT.md` per service. See `docs/agents/domain.md`.
