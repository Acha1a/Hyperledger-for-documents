package service
import (
    "crypto/sha256"
    "encoding/hex"
    "io"
    "log"

    "encoding/json" 
    "fabric-go-app/pkg/fabric" 
    "fabric-go-app/pkg/models"
    "fabric-go-app/pkg/repository"
)
type VerifyService struct {
    Fabric *fabric.FabricService 
}
func NewVerifyService(f *fabric.FabricService) *VerifyService {
    return &VerifyService{Fabric: f}
}
func (s *VerifyService) Verify(file io.Reader) (map[string]interface{}, error) {
	hashBytes := sha256.New()
	if _, err := io.Copy(hashBytes, file); err != nil {
		return nil, err
	}
	hash := hex.EncodeToString(hashBytes.Sum(nil))

	log.Println("Checking document with hash:", hash)
	localDoc, err := repository.GetByHash(hash)
	if err != nil {
		log.Printf("Document not found in local DB: %v", err)
	}
    if localDoc != nil {
        log.Printf("DEBUG: path from DB is: %s", localDoc.Path)
    } else {
        log.Printf("DEBUG: localDoc is nil, skipping path log")
    }
	resultStr, err := s.Fabric.VerifyByHash(hash)
	if err != nil {
		log.Println("Fabric error:", err)
		return nil, err
	}

	if resultStr == "" || resultStr == "null" || resultStr == "[]" {
		return map[string]interface{}{
			"status": "NOT_FOUND",
			"hash":   hash,
		}, nil
	}
	var blockchainDoc models.Document
	if err := json.Unmarshal([]byte(resultStr), &blockchainDoc); err != nil {
		log.Printf("JSON Parse error from Fabric: %v", err)
		blockchainDoc = models.Document{}
	}

    log.Printf("DEBUG: path from DB is: %s", localDoc.Path)
	if localDoc != nil {
		blockchainDoc.ID = localDoc.ID
		blockchainDoc.Filename = localDoc.Filename
		blockchainDoc.Path = localDoc.Path
		blockchainDoc.Owner = localDoc.Owner
				if blockchainDoc.MimeType == "" {
			blockchainDoc.MimeType = localDoc.MimeType
		}
	}
	return map[string]interface{}{
		"status":   "VALID",
		"hash":     hash,
		"document": blockchainDoc, 
	}, nil
}

