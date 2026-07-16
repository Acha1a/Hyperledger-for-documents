package handler

import (
	"log"
	"net/http"
	"fmt"

	"fabric-go-app/pkg/db"
	"fabric-go-app/pkg/service"
	"github.com/gin-gonic/gin"
	"fabric-go-app/pkg/models"
)

var allowedTypes = map[string]bool{
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"image/jpeg": true,
	"image/png":  true,
	"text/plain": true,
}

type VerifyHandler struct {
	Service *service.VerifyService
	DocService *service.DocumentService
}

func NewVerifyHandler(s *service.VerifyService, ds *service.DocumentService) *VerifyHandler {
	return &VerifyHandler{
		Service: s,
		DocService: ds,
	}
}

func (h *VerifyHandler) Verify(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	result, err := h.Service.Verify(file)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

func (h *VerifyHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	owner := c.PostForm("owner")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File required"})
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if !allowedTypes[mimeType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File format unsupported",
		})
		return
	}

	doc, err := h.DocService.UploadDocument(header, owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}


func (h *VerifyHandler) Approve(c *gin.Context) {
	docID := c.Param("id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	var docHash string
	err := db.DB.QueryRow("SELECT hash FROM documents WHERE id = $1", docID).Scan(&docHash)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found in database"})
		return
	}

	signature, err := service.SignHash(docHash, service.MinisterKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Signing failed: " + err.Error()})
		return
	}

	publicKey := service.ExportPublicKey(service.MinisterKey)

	err = h.Service.Fabric.ApproveDocument(docID, signature, publicKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fabric update failed: " + err.Error()})
		return
	}

	updateQuery := `
		UPDATE documents 
		SET status = 'APPROVED', signature = $1, public_key = $2, updated_at = NOW() 
		WHERE id = $3
	`
	_, err = db.DB.Exec(updateQuery, signature, publicKey, docID)
	if err != nil {
		log.Printf("Database update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Document approved",
		"id":        docID,
		"signature": signature,
		"status":    "APPROVED",
	})
}

func (h *VerifyHandler) GetDocuments(c *gin.Context) {
    docs, err := h.DocService.GetAll() 
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, docs)
}

func (h *VerifyHandler) VerifyIntegrity(c *gin.Context) {
	docID := c.Param("id")

	var dbDoc models.Document 

	// 1. Получаем данные из локальной БД
	query := `
    SELECT 
        path, 
        filename, 
        COALESCE(file_size, 0), -- Если NULL, вернет 0
        COALESCE(updated_at, NOW()), -- Защита для дат
        COALESCE(created_at, NOW()) 
    FROM documents 
    WHERE id = $1
	`
    // Используем &dbDoc для записи результатов запроса
    err := db.DB.QueryRow(query, docID).Scan(
        &dbDoc.Path, 
        &dbDoc.Filename, 
        &dbDoc.FileSize, 
        &dbDoc.UpdatedAt, 
        &dbDoc.CreatedAt,
    )
    
    if err != nil {
        log.Printf("DB Scan Error: %v", err)
        c.JSON(http.StatusNotFound, gin.H{"error": "Документ не найден в базе"})
        return
    }

	// 2. Считаем хеш файла, который реально лежит в папке storage
	currentDiskHash, err := service.CalculateHashFromPath(dbDoc.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось прочитать файл на диске"})
		return
	}

	// 3. Запрашиваем оригинал данных из Fabric
	fabricDoc, err := h.Service.Fabric.GetDocument(docID)
	if err != nil {
		log.Printf("Fabric Query Error: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Ошибка получения данных из блокчейна"})
		return
	}

	// 4. СРАВНЕНИЕ: Сверяем хеш на диске с хешем из блокчейна
	isValid := (currentDiskHash == fabricDoc.Hash)

	log.Printf("ВЕРИФИКАЦИЯ ID: %s", docID)
	log.Printf("Хеш на диске: %s", currentDiskHash)
	log.Printf("Хеш в Fabric: %s", fabricDoc.Hash)

	c.JSON(http.StatusOK, gin.H{
		"is_valid":     isValid,
		"current_hash": currentDiskHash,
		"fabric_hash":  fabricDoc.Hash,
		"status":       fabricDoc.Status,
		"owner":        fabricDoc.Owner,
	})
}

func (h *VerifyHandler) VerifyPublic(c *gin.Context) {
    file, _, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"exists": false, "error": "Файл не выбран"})
        return
    }
    defer file.Close()

    // 1. Считаем хеш загруженного файла без сохранения на диск
    hash, err := service.CalculateHash(file) 
    if err != nil {
        c.JSON(500, gin.H{"exists": false, "error": "Ошибка хеширования"})
        return
    }

    // 2. Ищем этот хеш в Fabric
    doc, err := h.Service.Fabric.GetDocumentByHash(hash)
    if err != nil {
        c.JSON(200, gin.H{"exists": false, "hash": hash})
        return
    }

    // 3. Если нашли - возвращаем данные
    c.JSON(200, gin.H{
        "exists":   true,
        "id":       doc.ID,
        "owner":    doc.Owner,
        "status":   doc.Status,
        "hash":     hash,
    })
}


func (h *VerifyHandler) GetHistory(c *gin.Context) {
    // 1. Получаем ID документа из URL
    docID := c.Param("id")
    if docID == "" {
        c.JSON(400, gin.H{"error": "ID документа не указан"})
        return
    }

    // 2. Вызываем метод из FabricService
    history, err := h.Service.Fabric.GetHistory(docID)
    if err != nil {
        c.JSON(500, gin.H{"error": fmt.Sprintf("Ошибка получения истории из Fabric: %v", err)})
        return
    }

    // 3. Возвращаем историю фронтенду в формате JSON
    c.JSON(200, history)
}