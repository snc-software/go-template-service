-- +goose Up
CREATE TABLE "Templates" (
    "Id"        UUID PRIMARY KEY,
    "Name"      VARCHAR(255) NOT NULL,
    "Email"     VARCHAR(255) NOT NULL UNIQUE,
    "CreatedAt" TIMESTAMPTZ NOT NULL,
    "UpdatedAt" TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE "Templates";