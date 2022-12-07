
# Nodewatcher

This is a simple script to monitor the status of a Lightning Node and send SMS
alerts if any issues are detected. Uses Twilio to send texts.

## Requirements

Twilio account

## Environment variables

Set SMS_ENABLE to 'TRUE' to enable text message updates

`env-example.sh`

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

## Build

```
make build
```

## Run

Cron job to run once an hour

```
crontab -e
0 * * * * sh -c "source ~/nodewatcher/nw-env.sh && ~/nodewatcher/nw"
crontab -l
tail /var/spool/mail/ec2-user
```