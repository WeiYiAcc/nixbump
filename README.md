# nixbump

Auto-discover and update custom Nix packages (npm / GitHub / GitLab / nix-update) — written in Go with [cobra](https://github.com/spf13/cobra).

`nixbump` scans a `pkgs/` tree for `nixbump.yaml` manifests, fetches the latest version from the configured source, and rewrites `version` + hash fields in the package derivation **in-place**. Configs may be [sops](https://github.com/getsops/sops)-encrypted YAML.

## Install

```bash
# from source
go install github.com/WeiYiAcc/nixbump@latest

# or Nix (see nix/package.nix)
nix build github:WeiYiAcc/nixbump
```

## Quick start

```
pkgs/
├── atomic/            # npm-tracked package
│   ├── nixbump.yaml   # manifest (may be sops-encrypted)
│   └── package.nix    # derivation (rewritten by nixbump)
└── pi-coding-agent/   # nix-update-tracked (follows nixpkgs)
```

```bash
nixbump list              # list all discoverable packages
nixbump check             # dry-run: show available updates
nixbump check atomic      # check a single package
nixbump update atomic     # update one package in-place
nixbump update            # update all outdated packages
nixbump pr atomic         # update + create a PR (git worktree + gh)
```

## Manifest

`pkgs/<name>/nixbump.yaml` — four sources:

```yaml
# npm registry
source: npm
package: atomic

# GitHub releases
source: github
owner: WeiYiAcc
repo: my-tool

# GitLab
source: gitlab
owner: group
repo: my-tool

# delegate version resolution to nix-update
source: nix-update
```

Optional fields: `host`, `strip_prefix`, `version_regex`, `extract_hash`, `args`.

Secrets (tokens) can be sops-encrypted — `nixbump` decrypts transparently when the age identity is available.

## MTP: `--mtp-describe`

`nixbump` implements the [Model Tools Protocol](https://github.com/modeltoolsprotocol/modeltoolsprotocol): it responds to `--mtp-describe` with a JSON schema of its commands, args, and examples, so LLM agents can discover and call it without training-data knowledge of the tool.

```bash
$ nixbump --mtp-describe
{
  "version": "1.0",
  "name": "nixbump",
  "commands": [
    { "name": "list",  "description": "list all discoverable packages" },
    { "name": "check", "description": "dry-run: show available updates without changing files",
      "examples": ["nixbump check", "nixbump check atomic"] },
    { "name": "update", "description": "update one package (or all) in-place",
      "args": ["[pkg]"], "examples": ["nixbump update atomic"] },
    { "name": "pr", "description": "update each outdated package and create a PR via git worktree + gh" }
  ]
}
```

## KCL schema

`schema.nixbump.k` provides a [KCL](https://www.kcl-lang.io/) schema for manifest validation:

```bash
kcl vet pkgs/atomic/nixbump.yaml schema.nixbump.k
```

## License

MIT
