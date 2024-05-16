package structs

import (
	"github.com/google/uuid"
)

type Cart struct {
	ID    string `json:"ID"`
	Games []Game `json:"Games"`
}

type Game struct {
	GameID string  `json:"GameID"`
	Title  string  `json:"Title"`
	Price  float64 `json:"Price"`
	Owned  bool    `json:"Owned"`
}

type CreateCartRequest struct {
	Games []Game `json:"Games"`
}

func (c *CreateCartRequest) CreateCartRequestToCart() Cart {
	return Cart{
		ID:    uuid.New().String(),
		Games: c.Games,
	}
}
