
-- add an initial user to the database
INSERT INTO "public"."users" ("id", "email", "password", "phone_number", "sms_enabled", "sms_last_sent", "sms_notify_time") VALUES
(1,
'user@user.com',
'password',
'5555555555',
't',
'2021-01-01 00:00:00',
'2021-01-01 01:00:00')
;
