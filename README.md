
# Nodewatcher

This is a simple program to monitor the status of a Lightning Node and send SMS
alerts if any issues are detected

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

MACAROON_HEADER should be readonly.macaroon from
/lnd/data/chain/bitcoin/regtest/readonly.macaroon

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