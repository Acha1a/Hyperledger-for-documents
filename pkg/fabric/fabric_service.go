package fabric

import (
	"fmt"
	"encoding/json" 
	//"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"fabric-go-app/pkg/models"
)

type FabricService struct {
	Contract *client.Contract
}

func NewFabricService(contract *client.Contract) *FabricService {
	return &FabricService{Contract: contract}
}

func (f *FabricService) GetDocument(id string) (*models.Document, error) {
	// Вызываем существующую функцию смарт-контракта
	result, err := f.Contract.EvaluateTransaction("GetDocument", id)
	if err != nil {
		return nil, fmt.Errorf("fabric error: %w", err)
	}

	var doc models.Document
	if err := json.Unmarshal(result, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	return &doc, nil
}

func (s *FabricService) GetDocumentByHash(targetHash string) (*models.Document, error) {
	result, err := s.Contract.EvaluateTransaction("GetDocumentByHash", targetHash)
	if err != nil {
		return nil, fmt.Errorf("документ с таким хешем не найден в блокчейне: %w", err)
	}

	var doc models.Document
	if err := json.Unmarshal(result, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return &doc, nil
}

func (f *FabricService) RegisterDocument(id string, hash string, owner string, mimeType string) error {
    _, err := f.Contract.SubmitTransaction("CreateDocument", id, hash, owner, mimeType)
    if err != nil {
        return fmt.Errorf("failed to register document in blockchain: %w", err)
    }
    return nil
}

func (f *FabricService) VerifyByHash(hash string) (string, error) {

    result, err := f.Contract.EvaluateTransaction("GetDocumentByHash", hash)
    if err != nil {
        return "", fmt.Errorf("fabric error: %w", err)
    }

    return string(result), nil
}


func (s *FabricService) ApproveDocument(id string, signature string, publicKey string) error {
	if s.Contract == nil {
		return fmt.Errorf("contract not initialized")
	}

	_, err := s.Contract.SubmitTransaction("ApproveDocument", id, signature, publicKey)
	return err
}

func (s *FabricService) GetHistory(docID string) ([]map[string]interface{}, error) {
	result, err := s.Contract.EvaluateTransaction("GetHistory", docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	var history []map[string]interface{}
	if err := json.Unmarshal(result, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return history, nil
}