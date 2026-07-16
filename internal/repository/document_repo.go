package repository

import (
	"database/sql"
	"fabric-go-app/internal/model"
)

type DocumentRepo struct {
	DB *sql.DB
}

func (r *DocumentRepo) Save(doc model.Document) error {
	log.Printf("SQL: Попытка сохранить документ %s (ID: %s)", doc.Filename, doc.ID)
	
	query := `
		INSERT INTO documents (id, hash, owner, filename, status, created_at) 
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	
	_, err := r.DB.Exec(
		query,
		doc.ID,       // $1 -> id 
		doc.Hash,     // $2 -> hash
		doc.Owner,    // $3 -> owner
		doc.Filename, // $4 -> filename
		"ACTIVE",     // $5 -> status
	)

	if err != nil {
		log.Printf("Ошибка SQL при вставке: %v", err)
		return err
	}

	log.Println("Запись в Postgres успешно создана")
	return nil
}