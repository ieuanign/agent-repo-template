# Secrets with SOPS and age

Env files reach git only encrypted with [SOPS](https://github.com/getsops/sops), to the
[age](https://github.com/FiloSottile/age) public keys listed in `.sops.yaml`. Anyone holding the
private key for one of those recipients can decrypt them.

Run everything below inside devbox (`devbox shell` or `devbox run -- …`), which provides `sops` and
`age`.

## What is committed

- **Committed:** `secrets/<name>.sops.env`, the encrypted files. Keys stay readable; values are
  encrypted.
- **Never committed:** `secrets/<name>.env`, the decrypted files. `.gitignore` ignores everything
  under `secrets/` except `*.sops.env`.

## Getting access

A new project's first recipient is added by `make bootstrap`, after step 1; everyone after that follows
every step.

1. **Generate your key yourself**, straight to where sops looks for it (see below), never inside the
   repo, where nothing ignores it. Only you do this; never ask anyone, or any agent, to do it for you.

   ```sh
   # macOS
   mkdir -p ~/Library/Application\ Support/sops/age
   age-keygen -o ~/Library/Application\ Support/sops/age/keys.txt
   # Linux
   mkdir -p ~/.config/sops/age && age-keygen -o ~/.config/sops/age/keys.txt
   ```

2. **Share only your public key**, the `age1…` line `age-keygen` prints. Never share the
   `AGE-SECRET-KEY-…` line or the key file.
3. **A maintainer adds it as a recipient** in `.sops.yaml`, one key per line, annotated with your
   GitHub login.
4. **The maintainer re-encrypts and commits.** From the repo root, with their own key available:

   ```sh
   for f in secrets/*.sops.env; do sops updatekeys -y "$f"; done
   ```

   `updatekeys` needs an existing recipient's key, so only someone who already has access can grant
   it. Removing a person is the same: delete their line, run `updatekeys`, commit, and rotate the
   values they could read.

## Where sops finds your key

In order: `SOPS_AGE_KEY_FILE` if set, otherwise `sops/age/keys.txt` under:

- `$XDG_CONFIG_HOME` when it is set;
- `~/Library/Application Support` on macOS;
- `~/.config` on Linux.

## Decrypting

```sh
make secrets-decrypt
```

Decrypts every `secrets/<name>.sops.env` into `secrets/<name>.env` beside it. Safe to re-run. If no
key of yours is a recipient it fails, points here, and leaves no `secrets/<name>.env` behind.

## Editing and adding secrets

- **Edit an existing file** with `sops secrets/<name>.sops.env`. This decrypts it, so it needs a key:
  it is a human step.
- **Create a new file** without any key, since encrypting needs only the public recipients. From the
  repo root, pipe the plaintext in so it never lands on disk:

  ```sh
  printf 'KEY=value\n' | sops encrypt --filename-override secrets/<name>.sops.env \
    > secrets/<name>.sops.env
  ```

## CI

There is no CI recipient yet. When CI needs secrets, its age public key is added to `.sops.yaml` as
one more recipient, followed by `updatekeys` as in step 4.
