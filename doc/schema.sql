-- SQL dump generated using DBML (dbml-lang.org)
-- Database: PostgreSQL
-- Generated at: 2023-06-10T08:01:28.032Z

CREATE TABLE "storage" (
  "id" bigserial PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "access_token" varchar NOT NULL,
  "access_token_expires_in" int,
  "refresh_token" varchar NOT NULL,
  "is_refresh_token_expired" boolean NOT NULL DEFAULT (false),
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);
