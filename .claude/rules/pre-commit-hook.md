# The pre-commit hook

**Before `git commit`, run `.githooks/pre-commit` from the repository root yourself, and commit once it
passes — never with `--no-verify`.** When it rewrites files, or its lint or tests fail in a workspace
the commit touches, that is yours to fix as part of the change: for `/dev-loop`'s code-writer, never a
reason to stop or return FAILED. `.claude/reference/pre-commit-hook.md` has the steps.
