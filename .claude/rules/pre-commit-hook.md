# The pre-commit hook

**Before `git commit`, run `.githooks/pre-commit` from the repository root yourself.** It is the same
script the commit runs. It re-runs itself through devbox when it needs to, and once it passes, the
commit's own run finds nothing to fix and replays the checks from turbo's cache.

**A non-zero exit that lists fixed files is a fix-abort.** The hook's fixers rewrote those files on
disk and staged nothing. Review them with `git diff`, stage them, run `.githooks/pre-commit` again,
then commit.

**A fix-abort is a lint failure inside your own scope.** For `/dev-loop`'s code-writer it is never a
reason to stop or return FAILED: review, stage, re-run, commit. The same holds for a failure of the
hook's lint or test checks in a workspace the commit touches — fix it as part of the change.

**`--no-verify` is never the answer.** Skipping the hook commits exactly what it exists to stop.
