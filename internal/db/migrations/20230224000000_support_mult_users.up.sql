
ALTER TABLE "users" ADD COLUMN "phone_number" varchar;

--migration:split
ALTER TABLE "users" ADD COLUMN "sms_enabled" boolean;

--migration:split
ALTER TABLE "users" ADD COLUMN "sms_notify_time" timestamp;

--migration:split
ALTER TABLE "users" ADD COLUMN "sms_last_sent" timestamp;

--migration:split
ALTER TABLE "nodes" ADD COLUMN "user_id" int4;

--migration:split
ALTER TABLE "nodes" ADD CONSTRAINT fk_node_to_user FOREIGN KEY ("user_id") REFERENCES "users" ("id");
