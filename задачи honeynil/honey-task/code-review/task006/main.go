package main

import (
	"context"

	"go.uber.org/zap"
	"honnef.co/go/tools/lintcmd/cache"
)

// Код-ревью. Скажи что не нравится и как бы ты это исправил:
type Service struct {
	db     repository.repository
	logger *zap.Logger
	cache  cache.Cache
}

func New(db repository.Repository, logger zap.Logger, cache cache.Cache) Service {
	return &Service{
		db:     db,
		logger: logger,
		cache:  cache,
	}
}

// GetProduct возвращает сущность по её ID
func (s Service) GetProduct(_ context.Context, id int64) (models.Product, error) {
	product, err := s.cache.GetProduct(id)
	if err != nil {
		return nil, err
	}

	if product != nil {
		return product, nil
	}

	product, err := s.db.GetProduct(context.Background(), id)
	if err != nil {
		return nil, err
	}

	return product, nil
}

// Update обновляет информацию о сущности
func (s Service) CreateProduct(c context.Context, p models.Product) error {
	_, err := s.db.CreateProduct(c, p)
	if err != nil {
		return err
	}

	return nil
}
