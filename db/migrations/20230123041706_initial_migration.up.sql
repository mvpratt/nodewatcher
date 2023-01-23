

CREATE SEQUENCE IF NOT EXISTS nodes_id_seq;

--migration:split
CREATE TABLE "public"."nodes" (
    "id" int4 NOT NULL DEFAULT nextval('nodes_id_seq'::regclass),
    "url" varchar,
    "macaroon" varchar,
    "alias" varchar,
    "pubkey" varchar,
    PRIMARY KEY ("id")
);

--migration:split
ALTER TABLE "nodes" ADD CONSTRAINT unique_node_url UNIQUE ("url");

--migration:split
CREATE SEQUENCE IF NOT EXISTS channels_id_seq;

--migration:split
CREATE TABLE "public"."channels" (
    "id" int4 NOT NULL DEFAULT nextval('channels_id_seq'::regclass),
    "funding_txid" varchar,
    "output_index" int4,
    PRIMARY KEY ("id")
);

--migration:split
ALTER TABLE "channels" ADD CONSTRAINT unique_channel_port UNIQUE ("funding_txid", "output_index");

--migration:split
CREATE SEQUENCE IF NOT EXISTS "channelBackups_id_seq";

--migration:split
CREATE TABLE "public"."channel_backups" (
    "id" int4 NOT NULL DEFAULT nextval('"channelBackups_id_seq"'::regclass),
    "backup" varchar,
    "created_at" timestamp,
    "funding_txid_bytes" varchar,
    "output_index" int4,
    PRIMARY KEY ("id")
);