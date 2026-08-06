-- Create "auth_token" table
CREATE TABLE "public"."auth_token" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "prefix" character varying(48) NOT NULL,
  "secret" character varying(127) NOT NULL,
  "refresh_secret" character varying(127) NULL,
  "expires_at" timestamptz NOT NULL,
  "refreshed_at" timestamptz NULL,
  "refresh_expires_at" timestamptz NULL,
  "is_revoked" boolean NOT NULL DEFAULT false,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "created_by" bigint NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_auth_token_prefix" to table: "auth_token"
CREATE UNIQUE INDEX "idx_auth_token_prefix" ON "public"."auth_token" ("prefix");
-- Create index "idx_auth_token_user_id" to table: "auth_token"
CREATE INDEX "idx_auth_token_user_id" ON "public"."auth_token" ("user_id");

-- Create "user_token" table
CREATE TABLE "public"."user_token" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "prefix" character varying(48) NOT NULL,
  "secret" character varying(127) NOT NULL,
  "refresh_secret" character varying(127) NULL,
  "type" character varying(50) NOT NULL,
  "data" character varying(255) NULL,
  "expires_at" timestamptz NOT NULL,
  "refreshed_at" timestamptz NULL,
  "refresh_expires_at" timestamptz NULL,
  "is_used" boolean NULL DEFAULT false,
  "used_at" timestamptz NULL,
  "is_revoked" boolean NOT NULL DEFAULT false,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "created_by" bigint NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_token_is_used" to table: "user_token"
CREATE INDEX "idx_user_token_is_used" ON "public"."user_token" ("is_used");
-- Create index "idx_user_token_prefix" to table: "user_token"
CREATE UNIQUE INDEX "idx_user_token_prefix" ON "public"."user_token" ("prefix");
-- Create index "idx_user_token_type" to table: "user_token"
CREATE INDEX "idx_user_token_type" ON "public"."user_token" ("type");
-- Create index "idx_user_token_user_id" to table: "user_token"
CREATE INDEX "idx_user_token_user_id" ON "public"."user_token" ("user_id");

-- Create "user_user" table
CREATE TABLE "public"."user_user" (
  "id" bigserial NOT NULL,
  "email" character varying(255) NOT NULL,
  "username" character varying(255) NOT NULL,
  "password" text NULL,
  "name" character varying(127) NULL,
  "avatar" character varying(256) NULL,
  "active" boolean NULL,
  "created_by" bigint NULL,
  "updated_by" bigint NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_user_email" to table: "user_user"
CREATE UNIQUE INDEX "idx_user_user_email" ON "public"."user_user" ("email");
-- Create index "idx_user_user_username" to table: "user_user"
CREATE UNIQUE INDEX "idx_user_user_username" ON "public"."user_user" ("username");

-- Create "user_auth_providers" table
CREATE TABLE "public"."user_auth_providers" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "provider" text NOT NULL,
  "provider_user_id" text NOT NULL,
  "email" text NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_auth_providers_user" FOREIGN KEY ("user_id") REFERENCES "public"."user_user" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_user_auth_providers_provider" CHECK (provider = ANY (ARRAY['google'::text, 'github'::text]))
);
-- Create index "idx_user_auth_providers_user_id" to table: "user_auth_providers"
CREATE INDEX "idx_user_auth_providers_user_id" ON "public"."user_auth_providers" ("user_id");
