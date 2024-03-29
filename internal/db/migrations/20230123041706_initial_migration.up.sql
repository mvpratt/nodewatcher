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
ALTER TABLE "nodes" ADD CONSTRAINT unique_node_pubkey UNIQUE ("pubkey");

--migration:split
CREATE SEQUENCE IF NOT EXISTS channels_id_seq;

--migration:split
CREATE TABLE "public"."channels" (
    "id" int4 NOT NULL DEFAULT nextval('channels_id_seq'::regclass),
    "funding_txid" varchar,
    "output_index" int4,
    "node_id" int4,
    PRIMARY KEY ("id")
);

--migration:split
ALTER TABLE "channels" ADD CONSTRAINT fk_channel_to_node FOREIGN KEY ("node_id") REFERENCES "nodes" ("id");

--migration:split
ALTER TABLE "channels" ADD CONSTRAINT unique_channel_port UNIQUE ("funding_txid", "output_index");

--migration:split
CREATE SEQUENCE IF NOT EXISTS "channel_backups_id_seq";

--migration:split
CREATE TABLE "public"."channel_backups" (
    "id" int4 NOT NULL DEFAULT nextval('"channel_backups_id_seq"'::regclass),
    "created_at" timestamp,
    "backup" varchar,
    "funding_txid_bytes" varchar,
    "output_index" int4,
    PRIMARY KEY ("id")
);

--migration:split
CREATE SEQUENCE IF NOT EXISTS "multi_channel_backups_id_seq";

--migration:split
CREATE TABLE "public"."multi_channel_backups" (
    "id" int4 NOT NULL DEFAULT nextval('"multi_channel_backups_id_seq"'::regclass),
    "created_at" timestamp,
    "backup" varchar,
    "node_id" int4,
    PRIMARY KEY ("id")
);

--migration:split
ALTER TABLE "multi_channel_backups" ADD CONSTRAINT fk_multi_backup_to_node FOREIGN KEY ("node_id") REFERENCES "nodes" ("id");