# worktree repo profile

Repository facts any skill provisioning a worktree reads — `/dev-loop` and `/pr-comments` alike;
answers persisted here are never re-asked. What each key means and when a run reads it is specified in
`/dev-loop`'s `acts/act-0.md` and is not restated here.

## Setup command

```bash
devbox install && { [ ! -f package-lock.json ] || devbox run -- npm ci; }
```

## Full-suite command

```bash
devbox run -- make check
```

## Fix cycles

`2`
