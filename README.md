
# Nodewatcher

This program monitors the status of a Lightning Node and sends an alert if an issue is detected.

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

## Environment variables

Set your own using the example here:

```bash
cp env-sample.sh env.sh
(make applicable changes)
source env.sh
```
## LND credentials

1) MACAROON_HEADER environment variable should be the hex contents of the file
`/lnd/data/chain/bitcoin/regtest/readonly.macaroon`

2) The file `tls.cert` from directory ----tbd---- should be placed in `/certs`

## Build and Run locally

```bash
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

## Build docker image and deploy to AWS

```bash
make deploy
```