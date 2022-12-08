
# Nodewatcher

This is a simple program to monitor the status of a Lightning Node and send SMS
alerts if any issues are detected

## Rationale

If lightning node is offline, channel parter could force close channels and steal the money on that channel.  Routing nodes are also monitored for uptime by their peers and payment routing is deprioritized for less reliable nodes.

## Requirements

Twilio account

## Environment variables

Example

```shell
#!/bin/sh
export LN_NODE_URL=https://abcxyz.io:8080
export MACAROON_HEADER=ABCD01234567
export SMS_ENABLE=TRUE
export TWILIO_ACCOUNT_SID=ABCD
export TWILIO_AUTH_TOKEN=BEEF42
export TWILIO_PHONE_NUMBER=+15556667777
export TO_PHONE_NUMBER=5554443333
```

Set your own

```
cp env-sample.sh env.sh
(make applicable changes)
make env
```

## Build and Run

```
make build
./nw
```

Sample Output

```
$ ./nw

Getting node status ...

SMS sent successfully!

Good news, lightning node "bowline" is fully synced!
Last block received 15m18.211865s minutes ago
```

## Cron

How to set up a cron job to run once an hour on an AWS EC2 instance

```
crontab -e
0 * * * * sh -c "source ~/nodewatcher/env.sh && ~/nodewatcher/nw"
crontab -l
tail /var/spool/mail/ec2-user
```
