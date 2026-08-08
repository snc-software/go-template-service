package writers

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/snc-software/go-template-service/exceptions"
	"github.com/snc-software/go-template-service/persistence"
	"github.com/snc-software/go-template-service/persistence/entities"
)

type TemplateWriter struct{}

func (TemplateWriter) Create(template entities.Template) (entities.Template, error) {
	db := persistence.NewConnection()
	defer db.Close()

	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := db.NamedExec(
		`INSERT INTO "Templates" ("Id", "Name", "Email", "CreatedAt", "UpdatedAt") 
		VALUES (:Id, :Name, :Email, :CreatedAt, :UpdatedAt)`,
		template,
	)
	if err != nil {
		return entities.Template{}, err
	}

	return template, nil
}

func (TemplateWriter) Delete(id uuid.UUID) error {
	db := persistence.NewConnection()
	defer db.Close()

	result, err := db.Exec(`DELETE FROM "Templates" WHERE "Id" = $1`, id)
	if err != nil {
		return exceptions.Internal()
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return exceptions.NotFound(fmt.Sprintf("Template with id '%s' not found", id))
	}

	return nil
}
