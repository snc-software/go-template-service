-- +goose Up
CREATE TABLE "Templates" (
    "Id"        UUID PRIMARY KEY,
    "Name"      VARCHAR(255) NOT NULL,
    "Email"     VARCHAR(255) NOT NULL UNIQUE,
    "CreatedAt" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "UpdatedAt" TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX "IX_Templates_CreatedAt_Id" ON "Templates" ("CreatedAt" DESC, "Id" DESC);

-- +goose Down
DROP TABLE "Templates";
