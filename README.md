
# Nodewatcher

This is a simple program to monitor the status of a Lightning Node and send SMS
alerts if any issues are detected

## Rationale

If lightning node is offline, channel parter could force close channels and steal the money on that channel.  Routing nodes are also monitored for uptime by their peers and payment routing is deprioritized for less reliable nodes.

## Requirements

Twilio account

## Environment variables

Example

```bash
#!/bin/sh
export LN_NODE_URL=abcxyz.io:10009
export MACAROON_HEADER=ABCD01234567
export SMS_ENABLE=TRUE
export TWILIO_ACCOUNT_SID=ABCD
export TWILIO_AUTH_TOKEN=BEEF42
export TWILIO_PHONE_NUMBER=+15556667777
export TO_PHONE_NUMBER=5554443333
export DOCKER_REPO=11111111.dkr.ecr.us-east-1.amazonaws.com/nodewatcher
```

Set your own

```bash
cp env-sample.sh env.sh
(make applicable changes)
source env.sh
```

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