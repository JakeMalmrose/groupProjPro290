package structs

import (
	"github.com/google/uuid"
)

type Game struct {
	ID          string   `json:"ID"`
	Title       string   `json:"Title"`
	Description string   `json:"Description"`
	Tags        []string `json:"Tags"`
	Price       float64  `json:"Price"`
	Published   string   `json:"Published"`
	Author      string   `json:"Author"`
	AuthorID    string   `json:"AuthorID"`
}

type Cart struct {
	ID     string `json:"ID"`
	UserID string `json:"UserID"`
	Games  []Game `json:"Games"`
}

type CreateCartRequest struct {
	UserID string `json:"UserID"`
	Game  *Game   `json:"Game"`
}

func (c *CreateCartRequest) CreateCartRequestToCart() Cart {
	return Cart{
		ID:     uuid.New().String(),
		UserID: c.UserID,
		Games:  []Game{*c.Game},
	}
}
