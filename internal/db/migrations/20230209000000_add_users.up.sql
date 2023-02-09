CREATE SEQUENCE IF NOT EXISTS "users_id_seq";

--migration:split
CREATE TABLE "public"."users" (
    "id" int4 NOT NULL DEFAULT nextval('users_id_seq'::regclass),
    "email" varchar,
    "password" varchar,
    PRIMARY KEY ("id")
);

--migration:split
ALTER TABLE "users" ADD CONSTRAINT unique_user_name UNIQUE ("email");
