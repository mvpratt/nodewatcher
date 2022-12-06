

## Set these environment variables

```
export LN_NODE_URL=xxxx
export MACAROON_HEADER=xxxx

export SMS_ENABLE=TRUE
export TWILIO_ACCOUNT_SID=xxxxxxxxx
export TWILIO_AUTH_TOKEN=xxxxxxxxx
export TWILIO_PHONE_NUMBER=xxxxxxxxx
export TO_PHONE_NUMBER=xxxxxxxxx
```

Example:
```
export LN_NODE_URL=https://abcxyz.io:8080
```

Set SMS_ENABLE to 'FALSE' to disable text message updates

Run once an hour:
```
crontab -e
0 * * * * sh -c "source ~/nodewatcher/nw-env.sh && ~/nodewatcher/nw"
crontab -l
tail /var/spool/mail/ec2-user
```