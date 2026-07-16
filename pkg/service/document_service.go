package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"

	"fabric-go-app/pkg/models"
	"fabric-go-app/pkg/repository"
	"fabric-go-app/pkg/fabric"
    "fabric-go-app/pkg/db"

	"github.com/google/uuid" //
)

type DocumentService struct {
    Fabric *fabric.FabricService
}

func (s *DocumentService) UploadDocument(file *multipart.FileHeader, owner string) (*models.Document, error) {
    docID := uuid.New().String()
    dstPath := "/home/light_soul/hyperledger/backend/fabric-go-app/storage/" + file.Filename    

    log.Printf("🛠 [Service] Подготовка к сохранению: %s", dstPath)

    // Сохраняем на диск
    err := saveFile(file, dstPath)
    if err != nil {
        return nil, fmt.Errorf("ошибка сохранения файла: %w", err)
    }

    // Считаем хеш
    hash, err := CalculateHashFromPath(dstPath)
    if err != nil {
        return nil, fmt.Errorf("ошибка вычисления хеша: %w", err)
    }

    contentType := file.Header.Get("Content-Type")
    
    // Получаем размер файла напрямую из заголовка мультипарта
    fileSize := file.Size 

    // Регистрация в Fabric
    err = s.Fabric.RegisterDocument(docID, hash, owner, contentType) 
    if err != nil {
        return nil, fmt.Errorf("ошибка Fabric: %w", err)
    }

    now := time.Now()
    doc := &models.Document{
        ID:        docID,
        Filename:  file.Filename,
        Path:      dstPath,
        Hash:      hash,
        FileSize:  fileSize,    
        Owner:     owner,
        MimeType:  contentType,
        CreatedAt: now,         
        UpdatedAt: now,        
    }

    log.Printf("[Service] Отправка в репозиторий: ID=%s, Size=%d bytes", doc.ID, doc.FileSize)
    
    err = repository.SaveDocument(doc)
    if err != nil {
        return nil, fmt.Errorf("ошибка сохранения в БД: %w", err)
    }

    return doc, nil
}

func saveFile(file *multipart.FileHeader, path string) error {
    src, err := file.Open()
    if err != nil {
        return fmt.Errorf("error in opening file: %w", err)
    }
    defer src.Close()

    dst, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("failed to create file at disk (%s): %w", path, err)
    }
    defer dst.Close()

    written, err := io.Copy(dst, src) 
    if err != nil {
        return fmt.Errorf("fail in copy: %w", err)
    }

    log.Printf("File saved: %s (volume: %d bytes)", path, written)
    return nil
}

func CalculateHash(reader io.ReadSeeker) (string, error) {
	_, err := reader.Seek(0, 0)
	if err != nil {
		return "", err
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}

	reader.Seek(0, 0)

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func GetAllDocuments() ([]models.Document, error) {
	return repository.GetAllDocuments()
}

func CalculateHashFromPath(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return CalculateHash(file)
}

func VerifyDocument(file *multipart.FileHeader) (bool, string, string, error) {
	tempPath := "./storage/tmp_" + file.Filename

	err := saveFile(file, tempPath)
	if err != nil {
		return false, "", "", err
	}
	defer os.Remove(tempPath)

	hash, err := CalculateHashFromPath(tempPath)
	if err != nil {
		return false, "", "", err
	}

	blockchainHash, err := fabric.GetDocumentHash(hash)
	if err != nil {
		return false, hash, "", err
	}

	isValid := hash == blockchainHash

	return isValid, hash, blockchainHash, nil
}

func (s *DocumentService) GetAll() ([]models.Document, error) {
    query := `
        SELECT 
            id, 
            filename, 
            path, 
            hash, 
            COALESCE(file_size, 0), -- Новое поле
            COALESCE(status, 'PENDING'), 
            COALESCE(owner, 'Unknown'), 
            COALESCE(updated_at, NOW()), -- Защита для дат
            COALESCE(created_at, NOW()),  
            COALESCE(mime_type, ''), 
            COALESCE(signature, ''), 
            COALESCE(public_key, '')
        FROM documents 
        ORDER BY created_at DESC
    `
    rows, err := db.DB.Query(query)
    if err != nil {
        log.Printf("Database Query Error: %v", err)
        return nil, err
    }
    defer rows.Close()

    var docs []models.Document
    for rows.Next() {
        var d models.Document
        err := rows.Scan(
            &d.ID,
            &d.Filename,
            &d.Path,
            &d.Hash,
            &d.FileSize,   
            &d.Status,
            &d.Owner,
            &d.CreatedAt,
            &d.UpdatedAt, 
            &d.MimeType,
            &d.Signature,
            &d.PublicKey,
        )
        if err != nil {
            log.Printf("Database Scan Error: %v", err)
            return nil, err
        }
        docs = append(docs, d)
    }

    return docs, nil
}