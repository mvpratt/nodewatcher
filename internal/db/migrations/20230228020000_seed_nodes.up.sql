
-- add an initial node to the database
INSERT INTO "public"."nodes" ("id", "url", "macaroon", "tls_cert", "alias", "pubkey", "user_id") VALUES
(1,
'required: put your lightning node url here, including the port number for grpc',
'required: put your hex encoded macaroon here',
'optional: put node tls cert here',
'optional: put node alias here',
'optional: put node pubkey here',
1
);
