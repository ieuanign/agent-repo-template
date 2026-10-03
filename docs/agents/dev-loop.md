# dev-loop repo profile

Per-repo configuration for the `/dev-loop` pipeline's own artifacts — the branches it names and the
pull requests it writes; what any skill provisioning a worktree reads is `docs/agents/worktree.md`.

## Branch template

`<type>/<issue-number>-<short_title>`, sub-lanes `<type>/<issue-number>-<short_title>-<area>`

`<type>` is the conventional-commit type; `<short_title>` is kebab-case (e.g. `chore/3-root-npm-workspaces`).

## PR title format

`<type>(<scope>): #<issue> - <title>`

## PR body template

```markdown
> **Draft:** {one line per reason — omit on a ready PR}

## PR Description

{concise summary, bullets if needed — {n} planned, {m} made}

## Issue Link

- Closes #{issue}   ← first sub-lane only; later sub-lanes: Part of #{issue}
- {other related issues, one bullet each}

## Additional Information

- Suite (`devbox run -- make check`): {passed | failed: <tests> | not run: <why>}
- {screenshots or video of the test, if any — from the e2e test when the ticket is an e2e-test ticket}

## Severity

- [ ] Low
- [ ] Medium
- [ ] High

{AI ticks exactly one}

## Notes

- {unmet acceptance criterion — why, one line each}
- {fixed or reviewer item the human needs to know, one line each; won't-fix with reason}
- {stacked on #<A>'s PR — rebase onto main after it merges}
- {local-only artifacts · defaults taken — only when present}

<details><summary>Pipeline record</summary>

- Acceptance criteria: {met|partial|not-met} · {criterion} · {evidence}, one line each
- {whole-issue roll-up — last sub-lane only}
- Review: {fixed count}; reviewer notes verbatim; {trajectory if a bound was hit}
- {attempt log and run handle — ended sub-lanes only}

</details>

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

## Constraints

Never run two `make check` at once: its local-stack smoke step binds host port 4008 for Traefik.
