package structs

import (
	"log"
	"strconv"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type UpdatePostObject struct {
	Title   string `json:"Title"`
	Content string `json:"Content"`
}

func (u *UpdatePostObject) UpdatePostObjectToUpdate() Update {
	return Update{
		ID:      uuid.New().String(),
		Title:   u.Title,
		Content: u.Content,
		Date:    time.Now().Format(time.RFC3339),
	}
}

type Update struct {
	ID      string `json:"ID"`
	Title   string `json:"Title"`
	Content string `json:"Content"`
	Date    string `json:"Date"`
}

func UpdateToDynamoDBItem(updates []Update) []*dynamodb.AttributeValue {
	var updatesAttributeValue []*dynamodb.AttributeValue
	for _, update := range updates {
		updateAttributeValue, err := dynamodbattribute.MarshalMap(update)
		if err != nil {
			log.Println("Error marshaling update:", err)
			return nil}
		updatesAttributeValue = append(updatesAttributeValue, &dynamodb.AttributeValue{M: updateAttributeValue})
	}
	return updatesAttributeValue
}

type GamePostRequest struct {
	Title       string   `json:"Title"`
	Description string   `json:"Description"`
	Tags        []string `json:"Tags"`
	Price       float64  `json:"price"`
	Author      string   `json:"Author"`
	AuthorID    string   `json:"AuthorID"`
}

func (g *GamePostRequest) GamePostRequestToGame() Game {

	game := Game{
		ID:          uuid.New().String(),
		Published:   time.Now().Format(time.RFC3339),
		Title:       g.Title,
		Description: g.Description,
		Tags:        g.Tags,
		Price:       g.Price,
		Updates:    []Update{},
		Author: 	g.Author,
		AuthorID: 	g.AuthorID,
	}
	log.Println("ID: ", game.ID)
	log.Println("Published: ", game.Published)
	return game
}

type Game struct {
	ID          string   `json:"ID"`
	Title       string   `json:"Title"`
	Description string   `json:"Description"`
	Tags        []string `json:"Tags"`
	Price       float64  `json:"Price"`
	Updates     []Update `json:"Updates"`
	Published   string   `json:"Published"`
	Author      string   `json:"Author"`
	AuthorID    string   `json:"AuthorID"`
}

func (g* Game) GameToDynamoDBItem() map[string]*dynamodb.AttributeValue{
	ExpressionAttributeValues:= map[string]*dynamodb.AttributeValue{}
	if g.Title != "" {
		ExpressionAttributeValues[":title"] = &dynamodb.AttributeValue{
			S: aws.String(g.Title),
		}
	}
	if g.Description != "" {
		ExpressionAttributeValues[":description"] = &dynamodb.AttributeValue{
			S: aws.String(g.Description),
		}
	}
	if len(g.Tags) != 0 {
		ExpressionAttributeValues[":tags"] = &dynamodb.AttributeValue{
			SS: aws.StringSlice(g.Tags),
		}
	}
	if g.Price != 0 {
		ExpressionAttributeValues[":price"] = &dynamodb.AttributeValue{
			N: aws.String(strconv.FormatFloat(g.Price, 'f', -1, 64)),
		}
	}
	// if len(g.Updates) != 0 {
	// 	ExpressionAttributeValues[":updates"] = &dynamodb.AttributeValue{
	// 		L: UpdateToDynamoDBItem(g.Updates),
	// 	}
	// }
	if g.Published != "" {
		ExpressionAttributeValues[":published"] = &dynamodb.AttributeValue{
			S: aws.String(g.Published),
		}
	}
	if g.Author != "" {
		ExpressionAttributeValues[":author"] = &dynamodb.AttributeValue{
			S: aws.String(g.Author),
		}
	}
	if g.AuthorID != "" {
		ExpressionAttributeValues[":authorID"] = &dynamodb.AttributeValue{
			S: aws.String(g.AuthorID),
		}
	}

	return ExpressionAttributeValues
}

func (g* Game) GameToDynamoDBUpdateItem() map[string]*dynamodb.AttributeValue{
	ExpressionAttributeValues:= map[string]*dynamodb.AttributeValue{}
	if g.Title != "" {
		ExpressionAttributeValues[":title"] = &dynamodb.AttributeValue{
			S: aws.String(g.Title),
		}
	}
	if g.Description != "" {
		ExpressionAttributeValues[":description"] = &dynamodb.AttributeValue{
			S: aws.String(g.Description),
		}
	}
	if len(g.Tags) != 0 {
		ExpressionAttributeValues[":tags"] = &dynamodb.AttributeValue{
			SS: aws.StringSlice(g.Tags),
		}
	}
	if g.Price != 0 {
		ExpressionAttributeValues[":price"] = &dynamodb.AttributeValue{
			N: aws.String(strconv.FormatFloat(g.Price, 'f', -1, 64)),
		}
	}
	// if len(g.Updates) != 0 {
	// 	ExpressionAttributeValues[":updates"] = &dynamodb.AttributeValue{
	// 		L: UpdateToDynamoDBItem(g.Updates),
	// 	}
	// }
	if g.Author != "" {
		ExpressionAttributeValues[":author"] = &dynamodb.AttributeValue{
			S: aws.String(g.Author),
		}
	}


	return ExpressionAttributeValues
}


// Something like: "set Title = :title, Description = :description, Tags = :tags, Price = :price, Updates = :updates, Published = :published"
func (g* Game) GameToUpdateString() string{
	FinalString := "set "
	if g.Title != "" {
		FinalString += "Title = :title, "
	}
	if g.Description != "" {
		FinalString += "Description = :description, "
	}
	if len(g.Tags) != 0 {
		FinalString += "Tags = :tags, "
	}
	if g.Price != 0 {
		FinalString += "Price = :price, "
	}
	// if len(g.Updates) != 0 {
	// 	FinalString += "Updates = :updates, "
	// }

	FinalString = FinalString[:len(FinalString)-2]

	return FinalString
}
