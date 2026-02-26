# StackSnap

StackSnap is a lightweight compose snapshot tool for Docker Compose stacks.

It automatically:

- Pulls the original compose YAML from Portainer
- Falls back to Docker label reconstruction if a stack is not in Portainer
- Stores timestamped versions
- Generates unified diffs (only when something changed)
- Rotates old versions
- Runs once - designed for cron or systemd timers

> **Note:** This is a snapshot tool, not a full backup solution. It captures your compose definitions so you have a history of what changed and when.

---

## Quick Start

### Step 1 - Download

**Linux (x86_64):**

```sh
wget https://github.com/mortaljinx/stacksnap/releases/latest/download/stacksnap-linux-amd64
chmod +x stacksnap-linux-amd64
sudo mv stacksnap-linux-amd64 /usr/local/bin/stacksnap
```

**Linux (ARM64 / Raspberry Pi):**

```sh
wget https://github.com/mortaljinx/stacksnap/releases/latest/download/stacksnap-linux-arm64
chmod +x stacksnap-linux-arm64
sudo mv stacksnap-linux-arm64 /usr/local/bin/stacksnap
```

---

### Step 2 - Create a Portainer Access Token

In Portainer:

1. Click your profile (top right)
2. Go to **My Account > Access Tokens**
3. Create a new token and copy it (you will not see it again)

> Portainer is optional. If you do not use it, StackSnap will fall back to reconstructing compose files from running container labels.

---

### Step 3 - Run StackSnap

Prefer passing your token via environment variable - it will not appear in `ps` output or shell history:

```sh
export STACKSNAP_PORTAINER_URL=http://localhost:9000
export STACKSNAP_PORTAINER_TOKEN=YOUR_TOKEN

stacksnap --output /path/to/snapshots
```

Or pass everything inline (fine for testing):

```sh
stacksnap \
  --output /mnt/snapshots/stacks \
  --portainer-url http://localhost:9000 \
  --portainer-token YOUR_TOKEN
```

**All flags:**

| Flag | Default | Description |
|---|---|---|
| `--output` | `./snapshots` | Where to store snapshots |
| `--keep` | `5` | Versions to keep per stack |
| `--portainer-url` | (none) | Portainer base URL |
| `--portainer-token` | (none) | Portainer API token |

**Environment variable equivalents:**

| Variable | Equivalent flag |
|---|---|
| `STACKSNAP_PORTAINER_URL` | `--portainer-url` |
| `STACKSNAP_PORTAINER_TOKEN` | `--portainer-token` |

---

## What It Creates

```
snapshots/
  jellyfin/
    latest.yml              # Current compose
    2026-02-25_0145.yml     # Historical version
    2026-02-25_0145.diff    # What changed (only written if something changed)
    .hash                   # Change detection (SHA-256 of latest content)
  homeassistant/
    latest.yml
    ...
```

StackSnap only writes a new snapshot if the compose content has actually changed. If nothing changed, it exits cleanly without touching anything.

---

## Run Daily with Cron

```sh
crontab -e
```

Add:

```
0 3 * * * STACKSNAP_PORTAINER_URL=http://localhost:9000 STACKSNAP_PORTAINER_TOKEN=YOUR_TOKEN /usr/local/bin/stacksnap --output /mnt/snapshots/stacks
```

Runs every night at 3 AM.

---

## Docker Version

```sh
docker run --rm \
  --network host \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /mnt/snapshots/stacks:/snapshots \
  -e STACKSNAP_PORTAINER_URL=http://localhost:9000 \
  -e STACKSNAP_PORTAINER_TOKEN=YOUR_TOKEN \
  mortaljinx/stacksnap:latest \
  --output /snapshots
```

> The container runs as root to ensure it can access your Docker socket and snapshot directory without permission issues on typical self-hosted setups.

---

## Docker Fallback

**For Portainer-managed stacks**, StackSnap fetches the original compose YAML directly from Portainer and saves it verbatim - named volumes, networks, build config, labels, everything is preserved exactly as you wrote it.

**For stacks not in Portainer** (e.g. you ran `docker compose up` manually), StackSnap reconstructs a compose file from the running container's labels and inspect data. This reconstruction captures:

- Image
- Ports
- Volumes (bind mounts only - named volumes are not visible via inspect)
- Environment variables (user-defined only)
- Restart policy

The reconstructed file includes a comment noting it was auto-generated and should be reviewed before use. Some settings not reflected in container inspect output (such as network aliases or build config) will not be present. For full fidelity, manage your stacks through Portainer.

---

## Security Notes

- **Use environment variables for your token**, not CLI flags. CLI flags are visible in `ps aux`.
- StackSnap makes read-only API calls to Portainer - it never modifies stacks.
- HTTPS connections to Portainer require a valid certificate. Use a local CA or a valid cert.

---

## Requirements

- Linux (amd64 or arm64)
- Docker (for the Docker fallback - not required if you only use Portainer)
- Portainer CE (optional but recommended)

---

## License

MIT
