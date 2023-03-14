
# Nodewatcher

This program monitors the status of a Lightning Node and sends an alert if the node is unsynced or another issue is detected. It also and makes static channel backups.

## Rationale

If lightning node is offline, channel parter could force close channels and steal the money on that channel.  Routing nodes are also monitored for uptime by their peers and payment routing is deprioritized for less reliable nodes.

## Requirements

- Twilio account

## Build and Run locally

1. Set environment variables

```bash
cp env-example.sh env.sh
(make applicable changes)
source env.sh
```

2. Seed the database by adding your node and user details here:

```
/internal/db/migrations/20230228010000_seed_users.up.sql
/internal/db/migrations/20230228010000_seed_nodes.up.sql
```

3. Buld and run

```bash
docker compose up postgres
make build
make run
```

Sample Output

```
Getting node status ...

SMS sent successfully!

Good news, lightning node "abcxyz" is fully synced!
Last block received 15m18.211865s minutes ago
```
