# just: command runner (casey/just) — CI wrapper for nixbump
# Windows note (avoid go-task/gosh coreutils gap):
#   set windows-shell := ["powershell.exe", "-NoProfile", "-Command"]
#   (recipes here only call ./nixbump, no rm/cp, so default is safe)

set shell := ["bash", "-c"]

# Validate every package config against the kcl schema
validate:
    for f in pkgs/*/nixbump.yaml; do kcl vet schema.nixbump.k "$f"; done

build:
    go build -o nixbump .

test:
    go test ./...

# List discoverable packages
list: build
    ./nixbump -list

# Dry-run update everything
check: build
    ./nixbump -dry-run

# Update one package
update pkg: build
    ./nixbump -package {{pkg}}

# Update all with PRs
pr: build
    ./nixbump -pr
