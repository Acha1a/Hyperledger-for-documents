package service

import (
	"fabric-go-app/internal/model"
	"fabric-go-app/internal/repository"
)

type DocumentService struct {
	Repo    *repository.DocumentRepo
	Fabric  *FabricService
}

func (s *DocumentService) Create(doc model.Document) error {
	// 1. Сохраняем в БД
	err := s.Repo.Save(doc)
	if err != nil {
		return err
	}

	// 2. Пишем в блокчейн
	return s.Fabric.CreateDocument(doc.ID, doc.Hash, doc.Owner)
}