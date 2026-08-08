-- +goose Up
ALTER TABLE "Templates"
    ALTER COLUMN "CreatedAt" SET DEFAULT now(),
    ALTER COLUMN "UpdatedAt" SET DEFAULT now();

CREATE INDEX "IX_Templates_CreatedAt_Id" ON "Templates" ("CreatedAt" DESC, "Id" DESC);

-- +goose Down
DROP INDEX "IX_Templates_CreatedAt_Id";

ALTER TABLE "Templates"
    ALTER COLUMN "CreatedAt" DROP DEFAULT,
    ALTER COLUMN "UpdatedAt" DROP DEFAULT;
