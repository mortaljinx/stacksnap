<p align="center">
  <img src="logo.png" width="600" />
</p>

<h1 align="center">StackSnap</h1>

<p align="center">
  Snapshot version history for your Portainer stacks.
</p>

# StackSnap

StackSnap is a lightweight CLI tool to snapshot Docker Compose stacks
deployed via Portainer CE.

Portainer CE does not keep version history of stack YAML --- StackSnap
fills that gap.

------------------------------------------------------------------------

## Quick Start

Download the latest release:

``` bash
wget https://github.com/mortaljinx/stacksnap/releases/latest/download/stacksnap-linux-amd64
chmod +x stacksnap-linux-amd64
sudo mv stacksnap-linux-amd64 /usr/local/bin/stacksnap
```

Run it:

``` bash
export STACKSNAP_PORTAINER_URL=http://localhost:9000
export STACKSNAP_PORTAINER_TOKEN=your_token

stacksnap --output ./snapshots
```

------------------------------------------------------------------------

## Features

-   Pulls original compose YAML via Portainer API
-   Timestamped snapshot history
-   Unified diffs when stacks change
-   Configurable rotation (--keep)
-   Safer writes (temp + rename)
-   Hardened Portainer client (timeouts, proper error handling)
-   Docker fallback reconstruction for non-Portainer stacks
    (best-effort)

------------------------------------------------------------------------

## Important Notes

-   This is a snapshot tool, not a data backup solution.
-   Docker fallback reconstruction captures image, ports, volumes,
    environment variables and restart policy (best-effort).
-   Designed for run-once usage (cron / systemd timer friendly).
-   Docker image publishing coming soon. No official Docker image
    available yet.

------------------------------------------------------------------------

MIT Licensed.
