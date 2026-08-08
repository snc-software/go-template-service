package readers

import (
	"github.com/google/uuid"

	"github.com/snc-software/go-template-service/persistence"
	"github.com/snc-software/go-template-service/persistence/entities"
)

type TemplateReader struct{}

func (TemplateReader) GetByID(id uuid.UUID) (entities.Template, error) {
	db := persistence.NewConnection()
	defer db.Close()

	var template entities.Template
	err := db.Get(&template, `SELECT "Id", "Name", "Email", "CreatedAt", "UpdatedAt" FROM "Templates" WHERE "Id" = $1`, id)
	if err != nil {
		return entities.Template{}, err
	}

	return template, nil
}

func (TemplateReader) GetPage(page, size int) ([]entities.Template, int, error) {
	db := persistence.NewConnection()
	defer db.Close()

	var total int
	err := db.Get(&total, `SELECT COUNT(*) FROM "Templates"`)
	if err != nil {
		return nil, 0, err
	}

	var templates []entities.Template
	offset := (page - 1) * size
	err = db.Select(&templates, `SELECT "Id", "Name", "Email", "CreatedAt", "UpdatedAt" FROM "Templates" ORDER BY "CreatedAt" DESC LIMIT $1 OFFSET $2`, size, offset)
	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}
