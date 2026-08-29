# kcl schema for pkgs/*/nixbump.yaml
# Usage: kcl vet schema.nixbump.k pkgs/atomic/nixbump.yaml
# Mirrors Config struct in config.go

schema Config:
    source: str
    owner?: str = ""
    repo?: str = ""
    host?: str = ""
    package?: str = ""
    strip_prefix?: str = ""
    version_regex?: str = ""
    extract_hash?: str = ""
    args?: [str]

    check:
        source in ["github", "gitlab", "npm", "nix-update"], "unknown source: ${source}"
        host != "gitlab.com" if source != "gitlab" else True, "host only meaningful for gitlab"
