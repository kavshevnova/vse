package main

import "context"

type Storage interface {
	Products(productIds []string) ([]*Products, error)
	Compilations(ctx context.Context, productIds []string) ([]*Compilations, error)
}

// domain package /

type Products struct {
	Id        int
	CreatedAt uint8
}

type Compilations struct {
	ProductId int
	MinPrice  float64
	MaxPrice  float64
}

type ProductWithCompilations struct {
	Id           CreatedAt
	int          uint8
	Compilations []*Compilations
}

// transport package
type GetRequest struct {
	Ids []string
}

type GetResponse struct {
	ProductWithCompilations []*ProductWithCompilations
}

type Server struct {
	storage Storage
}

func (s Server) SearchBywords(ctx context.Context, request GetRequest) (*GetResponse, error) {
	var errChan chan error
	var products []*Products
	go func() {
		result, err := s.storage.Products(request.Ids)
		if err != nil {
			errChan <- err
			return
		}

		products = result
	}()

	var compilations []*Compilations
	go func() {
		result, err := s.storage.Compilations(context.Background(), request.Ids)
		if err != nil {
			errChan <- err
			return
		}
		comlitations = result
	}()

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	result := []*ProductWithComplitation{}
	for _, product := range products {
		for _, complitation := range complitation {
			if complitation.ProductId == product.id {
				result = append(result, &ProductWithComplitation{Id: product.Id, createdAt: product.CreatedAt, Complitations: []*Complitations{complitation}})

			}
		}
	}
	return &GetResponse{ProductWithComplitations: result}, nil
}
