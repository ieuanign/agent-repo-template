# The `make check` acceptance criterion

An issue's criterion "`make check` passes" is judged **`met`**, never `partial` merely because the
reviewer did not run the suite itself. The evidence line says it is settled by `/dev-loop`'s suite
gate, plus any scoped checks the reviewer did run.

`make check` is this repository's Full-suite command (`docs/agents/worktree.md`), and `/dev-loop` runs
it on every branch after review. That run is the evidence for this criterion. A red suite drafts the
pull request by itself, whatever this criterion's verdict says, so a `partial` here only adds a second
draft reason saying the same thing.

**Still `not-met`** when the review itself saw something that would fail it, such as a lint or type
error in the diff.

**Only this criterion.** A criterion that needs the running stack — `make dev`, Swagger UI, backend up
or stopped, any manual check — is not covered by `make check`. It stays `partial` until someone runs
it, and the draft it causes is deliberate.

This holds only while `docs/agents/worktree.md` names `make check` as the Full-suite command.
