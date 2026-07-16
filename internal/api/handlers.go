package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"fabric-go-app/internal/service" 
)

// Verify проверяет существование документа в блокчейне только по его хешу
func (h *Handler) Verify(c *gin.Context) {
	// 1. Извлекаем файл из POST-запроса 
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Ошибка получения файла: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден в запросе"})
		return
	}
	defer file.Close()

	// 2. Вычисляем хеш загруженного файла (SHA-256)
	hash, err := service.CalculateHash(file) 
	
	log.Printf("--- ДИАГНОСТИКА ---")
	log.Printf("Хеш от бэкенда: %s", hash)
	log.Printf("Ожидаемый хеш:  3c916e919754458ef00aff6df68f3ef13364fb569a0495fe292546bfc77b794c")
	log.Printf("-------------------")

	if err != nil {
		log.Printf("Ошибка при вычислении хеша: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки файла"})
		return
	}

	log.Printf("Запрос в Fabric для хеша: %s", hash)

	// 3. ПРЯМОЙ ЗАПРОС В FABRIC
	// Полностью игнорируем Postgres, идем сразу в блокчейн
	result, err := h.Service.Fabric.VerifyByHash(hash)

	// Логируем ответ от чейнкода
	log.Printf("Ответ от Fabric: '%s' | Ошибка: %v", result, err)

	// 4. Анализ результата
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "ERROR",
			"hash":   hash,
			"error":  "Ошибка связи с блокчейном",
		})
		return
	}

	// Если Fabric вернул пустоту, null или пустой JSON-массив
	if result == "" || result == "null" || result == "{}" || result == "[]" {
		log.Printf("Документ НЕ НАЙДЕН в блокчейне")
		c.JSON(http.StatusOK, gin.H{
			"status": "NOT_FOUND",
			"hash":   hash,
		})
		return
	}

	// 5. УСПЕХ: Документ найден
	log.Printf("Документ подтвержден в Fabric!")
	c.JSON(http.StatusOK, gin.H{
		"status":   "FOUND",
		"hash":     hash,
		"document": result, 
	})
}