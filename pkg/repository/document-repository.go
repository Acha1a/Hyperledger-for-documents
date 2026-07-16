package repository

import (
    "database/sql"
    "fabric-go-app/pkg/db"
    "fabric-go-app/pkg/models"
    "fmt"
    "time"
    "log"
)

func SaveDocument(doc *models.Document) error {
    if db.DB == nil {
        return fmt.Errorf("база данных не подключена")
    }

    log.Printf("DEBUG [Repository]: Saving doc ID=%s, Filename=%s", doc.ID, doc.Filename)

    query := `
        INSERT INTO documents (
            id, filename, hash, path, owner, mime_type, status, created_at, updated_at, file_size
        ) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (id) DO UPDATE SET 
            path = EXCLUDED.path, 
            filename = EXCLUDED.filename,
            file_size = EXCLUDED.file_size,
            updated_at = NOW()
    `
    
    log.Printf("SQL Execute: ID=%s", doc.ID) 

    // Статус устанавливаем как 'PENDING' при первой загрузке
    status := "PENDING"
    if doc.Status != "" {
        status = doc.Status
    }

    _, err := db.DB.Exec(query, 
        doc.ID, 
        doc.Filename, 
        doc.Hash, 
        doc.Path, 
        doc.Owner, 
        doc.MimeType, 
        status,     
        time.Now(), 
        time.Now(), 
        doc.FileSize, 
    )
    
    if err != nil {
        log.Printf("Критическая ошибка SQL: %v", err)
    }
    
    return err
}

func GetAllDocuments() ([]models.Document, error) {
    query := `SELECT id, filename, path, hash, owner, mime_type, created_at FROM documents`

    rows, err := db.DB.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var docs []models.Document
    for rows.Next() {
        var d models.Document
        err := rows.Scan(&d.ID, &d.Filename, &d.Path, &d.Hash, &d.Owner, &d.MimeType, &d.CreatedAt)
        if err != nil {
            return nil, err
        }
        docs = append(docs, d)
    }
    return docs, nil
}

func GetByHash(hash string) (*models.Document, error) {
	doc := &models.Document{}
	
	query := `
		SELECT 
			id, 
			COALESCE(filename, ''), 
			COALESCE(path, ''), 
			COALESCE(hash, ''), 
			COALESCE(owner, ''), 
			COALESCE(mime_type, ''), 
			COALESCE(signature, ''), 
			COALESCE(public_key, ''),  
			COALESCE(status, ''),      
			created_at 
		FROM documents 
		WHERE hash = $1
	`
	
	err := db.DB.QueryRow(query, hash).Scan(
		&doc.ID, &doc.Filename, &doc.Path, &doc.Hash, &doc.Owner, &doc.MimeType,
		&doc.Signature, &doc.PublicKey, &doc.Status, &doc.CreatedAt,    
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil 
		}
		return nil, err
	}
	
	return doc, nil
}

func DeleteDocument(id string) error {
	if db.DB == nil {
		return fmt.Errorf("db wasn't connected")
	}
	_, err := db.DB.Exec("DELETE FROM documents WHERE id = $1", id)
	return err
}