CREATE TABLE "storage" (
  "id" bigserial PRIMARY KEY,
  "access_token" varchar NOT NULL,
  "refresh_token" varchar NOT NULL,
  "is_refresh_token_expired" boolean NOT NULL DEFAULT (false),
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);