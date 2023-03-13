
# Nodewatcher

This program monitors the status of a Lightning Node and sends an alert if an issue is detected. and makes backups. graphql interface for inspection

### Features
Health Monitor
- Sends an SMS alert if node is offline

Static Channel Backups
- Backs up channel state to a postgres database

### Future work
- Restore backups
- Telegram, Slack, Discord integration

## Rationale

If lightning node is offline, channel parter could force close channels and steal the money on that channel.  Routing nodes are also monitored for uptime by their peers and payment routing is deprioritized for less reliable nodes.

## Requirements

Twilio account

## Build and Run locally

Set environment variables

```bash
cp env-example.sh env.sh
(make applicable changes)
source env.sh
```

Seed the database by adding your node connection details here:

`/internal/db/migrations/20230228010000_seed_nodes.up.sql`

Macaroon should be `readonly.macaroon` from
`/lnd/data/chain/bitcoin/regtest/readonly.macaroon`

Buld and run

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
