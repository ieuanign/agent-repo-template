# Taking changes from the template

This project was made from [agent-repo-template](https://github.com/ieuanign/agent-repo-template) and
takes the template's later changes **by cherry-pick, one commit at a time, never by merge**. A project
made from the template shares no history with it and is expected to diverge: it renames, adds its own
domain, and may replace a whole part of the stack. A merge would bring in every template change,
wanted or not; a cherry-pick brings in only the commit someone chose.

`devbox shell` adds the template as the git remote `template`. Inside the template itself it adds
nothing.

## Porting a commit

1. `git fetch template`, then compare `git log --oneline template/main` with the ledger below.
2. On a branch, `git cherry-pick -x <commit>`. Expect conflicts wherever the commit touches a line this
   project renamed or changed.
3. Resolve each conflict to this project's version of the line, as **Renames** and **Kept
   differences** say. Drop what touches a part of the stack this project replaced.
4. Run `.githooks/pre-commit`, then `make check`, and open the pull request as usual.
5. Add the commit to the ledger in the same pull request.

A commit that does not fit is still added to the ledger, as skipped and with the reason, so nobody
weighs it twice.

## Renames

What this project calls the things the template names otherwise. The README's rename commands cover
the project name; anything renamed after that is added here.

| Template              | This project |
| --------------------- | ------------ |
| `agent-repo-template` |              |

## Kept differences

What this project keeps on purpose where the template differs, so a port never reverts it.

## Ledger

Started from: the template commit whose tree matches this project's first commit, found with
`git log template/main --format='%h %T' | grep "$(git rev-parse "$(git rev-list --max-parents=0 HEAD)^{tree}")"`.

Every template commit after it, newest first:

| Template commit | Subject | Outcome |
| --------------- | ------- | ------- |
