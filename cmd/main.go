package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"fabric-go-app/pkg/db"  
	"fabric-go-app/pkg/fabric"
	"fabric-go-app/pkg/handler"
	"fabric-go-app/pkg/service"
)

func main() {
	var err error
    // Генерация ключей министра
    service.MinisterKey, err = service.GenerateMinisterKeys()
    if err != nil {
        log.Fatal("Failed to generate keys")
    }

	log.Println("backend initialization...")

	// Подключение к базе данных
	db.Connect()

	// Инициализация Fabric
	contract, err := fabric.InitFabric()
	if err != nil {
		log.Fatalf("Error Fabric: %v", err)
	}
	log.Println("Fabric connected")

	// 1. Создаем FabricService
	fabricService := fabric.NewFabricService(contract)

	// 2. Создаем DocumentService 
	docService := &service.DocumentService{Fabric: fabricService}

	// 3. Создаем VerifyService
	verifyService := service.NewVerifyService(fabricService)

	// 4. Передаем ОБА сервиса в хендлер 
	verifyHandler := handler.NewVerifyHandler(verifyService, docService)

	r := gin.Default()
	r.MaxMultipartMemory = 128 << 20  

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	// Это позволит открывать файлы по ссылке http://localhost:8080/files/1.pdf
	r.Static("/files", "./storage")
	// Роуты
	r.GET("/documents", verifyHandler.GetDocuments)
	r.POST("/verify", verifyHandler.Verify)
	r.GET("/verify/:id", verifyHandler.VerifyIntegrity)
	r.POST("/upload", verifyHandler.Upload)
	r.POST("/approve/:id", verifyHandler.Approve)
	r.POST("/verify-public", verifyHandler.VerifyPublic) 
	r.GET("/history/:id", verifyHandler.GetHistory)     
	

	log.Println("🌐 Server is up on http://localhost:8080")
	r.Run(":8080")
}