-- SQL dump generated using DBML (dbml-lang.org)
-- Database: PostgreSQL
-- Generated at: 2023-05-25T03:19:17.868Z

CREATE TABLE "storage" (
  "id" bigserial PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "access_token" varchar NOT NULL,
  "refresh_token" varchar NOT NULL,
  "is_refresh_token_expired" boolean NOT NULL DEFAULT (false),
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);
