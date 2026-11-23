# Our build of forgejo-mcp

This fork of [goern/forgejo-mcp](https://github.com/goern/forgejo-mcp) is the Forgejo MCP server of The Black Cat
Company's agents. It runs in the homelab cluster behind the agentgateway, which authenticates each agent, adds that
agent's own Forgejo token to each request and allows only that agent's tools (app-def.talos:
`manifests/forgejo-mcp`). Everything outside `bcc/` and `.forgejo/` is upstream's, unchanged.

## Release

Copied to Forgejo as `github/forgejo-mcp`. A tag like `3.5.0-bcc.1` (upstream's version, then `-bcc.N`), which only
gerry may create there, runs `.forgejo/workflows/release.yaml`: it checks the code (`go mod verify`, `go vet`, the
tests, `govulncheck`), builds `bcc/build-image.sh` and pushes `zot.zot.svc/forgejo-mcp/forgejo-mcp:<tag>`. The release
on Forgejo names the digest; pin it in app-def.talos.

We build our own image rather than upstream's: its registry (git.b4mad.industries) wasn't reachable, its builder
installs unpinned packages, and its runtime image has a shell and a package manager. Ours is a static binary on
`distroless/static:nonroot`.

## Updating from upstream

Sync the fork on GitHub, then read what changed before you tag: the review of 2026-11-23 (commit `07f9252`) covered
the authentication path, the agents' tools and the supply chain. Look again at `operation/operation.go`,
`pkg/forgejo/`, `cmd/`, the tools an agent may use, and `go.mod`.

## Don't open this repo as an agent's workspace

Upstream keeps configuration for AI agents here (`.claude/`, `.mcp.json`, `.agents/`, `.pi/`, `.beads/`, `.castra/`,
`.husky/`): hooks that run commands and MCP servers that point at other forges. None of it matters for the build, but
an agent working in this repo would pick it up.
