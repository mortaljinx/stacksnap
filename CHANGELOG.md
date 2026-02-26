# Changelog

## v0.5.0

- Hardened Portainer client with 15s timeout and TLS 1.2 minimum
- Specific auth error messages for HTTP 401/403
- Batched `docker inspect` (single call vs per-container loop)
- Docker fallback now captures ports, bind mounts, environment variables, restart policy
- Snapshot engine extracted into own package
- Atomic file writes (temp-then-rename) to prevent truncated files on interruption
- Stale `.tmp` file cleanup on startup
- Empty diff fix — diffs only written when meaningful previous version exists
- Stack name sanitisation to prevent path traversal
- Fixed build path in Dockerfile (was pointing to non-existent `./cmd/stacksnap`)
- Dockerfile now multi-arch via `TARGETARCH`/`TARGETOS` build args
- Removed dead `config` package
- `--keep` now validates minimum of 1

---

## v0.4.0

- Portainer API integration
- Exact compose YAML backup
- Timestamped version history
- Timestamped unified diffs
- Rotation control
- Fail-fast on Portainer auth
- Hybrid Docker fallback support

---

## v0.3.x

- Docker reconstruction support
- Hash-based change detection
- Basic snapshot engine

---

## v0.1.x

- Initial proof-of-concept
- Container discovery
- Minimal compose generation
