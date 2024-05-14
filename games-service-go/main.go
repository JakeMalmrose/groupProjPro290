package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

/*
Games object
{
ID: int
Title: String
Description: String
Tags: String[]
Price: float
Updates[]: ID[] to update objects
Published: Date
}
*/

type UpdatePostObject struct {
	Title   string `json:"Title"`
	Content string `json:"Content"`
}

type Update struct {
	ID      int    `json:"ID"`
	Title   string `json:"Title"`
	Content string `json:"Content"`
	Date    string `json:"Date"`
}

type GamePostRequest struct {
	Title       string   `json:"Title"`
	Description string   `json:"Description"`
	Tags        []string `json:"Tags"`
	Price       float64  `json:"price"`
}

type Game struct {
	ID          string   `json:"ID"`
	Title       string   `json:"Title"`
	Description string   `json:"Description"`
	Tags        []string `json:"Tags"`
	Price       float64  `json:"Price"`
	Updates     []Update `json:"Updates"`
	Published   string   `json:"Published"`
}

func GameToDynamoDBItem(game Game) map[string]*dynamodb.AttributeValue {
	item := map[string]*dynamodb.AttributeValue{
		"ID": {
			S: aws.String(game.ID),
		},
		"Title": {
			S: aws.String(game.Title),
		},
		"Description": {
			S: aws.String(game.Description),
		},
		"Tags": {
			SS: aws.StringSlice(game.Tags),
		},
		"Price": {
			N: aws.String(strconv.FormatFloat(game.Price, 'f', -1, 64)),
		},
		"Updates": {
			L: []*dynamodb.AttributeValue{},
		},
		"Published": {
			S: aws.String(game.Published),
		},
	}

	var updates []*dynamodb.AttributeValue
	for _, update := range game.Updates {
		updateAttributeValue, err := dynamodbattribute.MarshalMap(update)
		if err != nil {
			log.Println("Error marshaling update:", err)
			return nil
		}
		updates = append(updates, &dynamodb.AttributeValue{M: updateAttributeValue})
	}
	item["Updates"] = &dynamodb.AttributeValue{L: updates}

	return item
}

var dynamodbClient *dynamodb.DynamoDB
var consulClient *api.Client

func init() {
	region := os.Getenv("AWS_REGION")
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(region),
		Endpoint: aws.String(endpoint),
	})
	if err != nil {
		log.Fatal(err)
	}

	dynamodbClient = dynamodb.New(sess)

	consulConfig := api.DefaultConfig()
	consulConfig.Address = os.Getenv("CONSUL_ADDRESS")
	consulClient, err = api.NewClient(consulConfig)
	if err != nil {
		log.Fatal("Error creating Consul client:", err)
	}
}

func main() {
	port := 8080

	err := registerService()
	if err != nil {
		log.Fatal("Error registering service with Consul:", err)
	}

	http.HandleFunc("/", GamesHandler)
	http.HandleFunc("/{id}", GamesHandlerID)
	log.Printf("Games service listening on port %d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

func registerService() error {
	serviceName := os.Getenv("SERVICE_NAME")
	serviceID := os.Getenv("SERVICE_ID")
	port, _ := strconv.Atoi(os.Getenv("SERVICE_PORT"))

	tags := []string{
		"TRAEFIK_ENABLE=true",
		"traefik.http.routers.gamesservice.rule=PathPrefix(`/games`)",
		"TRAEFIK_HTTP_SERVICES_GAMES_LOADBALANCER_SERVER_PORT=3000",
	}

	service := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: "games-service",
		Port:    port,
		Tags:    tags,
	}

	return consulClient.Agent().ServiceRegister(service)
}

func GamesHandlerID(w http.ResponseWriter, r *http.Request) {
	// Retrieve Games from DynamoDB
	switch r.Method {
	case http.MethodGet:
		getGamesID(w, r)
	case http.MethodDelete:
		deleteGameID(w, r)
	case http.MethodPut:
		updateGameID(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GamesHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve Games from DynamoDB
	switch r.Method {
	case http.MethodGet:
		getGames(w, r)
	case http.MethodPost: //DEV
		createGame(w, r)
	case http.MethodDelete: // DEV
		deleteGame(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getGamesID(w http.ResponseWriter, r *http.Request) {
	//get ID from URL
	id := r.FormValue("ID")
	// Query the Games table for the Game with the specified ID
	Games, err := dynamodbClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Games"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		log.Println("Error getting item from Games table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render the template with the retrieved Games data
	renderTemplate(w, "Game.html", map[string]interface{}{
		"Games": Games,
	})
}

func getGames(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("Search")

	// Query the Games table for all Games and if search is not empty check if title or description contains the search string
	Games, err := dynamodbClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("Games"),
	})
	if err != nil {
		log.Println("Error scanning Games table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	RealGames := []Game{}
	err = dynamodbattribute.UnmarshalListOfMaps(Games.Items, &RealGames)
	if err != nil {
		log.Println("Error unmarshalling Game:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Filter the Games based on the search string
	GamesToDisplay := []Game{}
	for _, Game := range RealGames {
		if search == "" || strings.Contains(Game.Title, search) || strings.Contains(Game.Description, search) {
			GamesToDisplay = append(GamesToDisplay, Game)
		}
	}

	renderTemplate(w, "Game.html", map[string]interface{}{
		"Games": GamesToDisplay,
	})
}

func createGame(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var createRequest GamePostRequest
	err := json.NewDecoder(r.Body).Decode(&createRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Generate a new UUID for the Game ID
	GameID := uuid.New().String()

	newGame := Game{
		ID:          GameID,
		Title:       createRequest.Title,
		Description: createRequest.Description,
		Tags:        createRequest.Tags,
		Price:       createRequest.Price,
		Updates:     []Update{},
		Published:   time.Now().String(),
	}

	_, err = dynamodbClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("Games"),
		Item:      GameToDynamoDBItem(newGame),
	})
	if err != nil {
		log.Println("Error putting item into Games table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func deleteGameID(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("ID")
	_, err := dynamodbClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Games"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		log.Println("Error deleting item from Games table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func deleteGame(w http.ResponseWriter, r *http.Request) {
	// Delete all items from the Games table
	_, err := dynamodbClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Games"),
	})
	if err != nil {
		log.Println("Error deleting item from Games table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func updateGameID(w http.ResponseWriter, r *http.Request) {
	// get the Game ID from the URL
	id := r.FormValue("ID")
	// Parse the request body = []Game
	var updateRequest Game
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// get the current Game
	currentGame, err := dynamodbClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Games"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		log.Println("Error getting item from Games table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var RealCurrentGame Game
	err = dynamodbattribute.UnmarshalMap(currentGame.Item, &RealCurrentGame)
	if err != nil {
		log.Println("Error unmarshalling Game:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	updatedGame := Game{
		ID:          RealCurrentGame.ID,
		Title:       updateRequest.Title,
		Description: updateRequest.Description,
		Tags:        updateRequest.Tags,
		Price:       updateRequest.Price,
		Updates:     RealCurrentGame.Updates,
		Published:   RealCurrentGame.Published,
	}

	if updatedGame.Title == "" {
		updatedGame.Title = RealCurrentGame.Title
	}
	if updatedGame.Description == "" {
		updatedGame.Description = RealCurrentGame.Description
	}
	if updatedGame.Tags == nil {
		updatedGame.Tags = RealCurrentGame.Tags
	}
	if updatedGame.Price == 0 {
		updatedGame.Price = RealCurrentGame.Price
	}

	// update the Game with the new games
	_, err = dynamodbClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Games"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(updatedGame.ID),
			},
		},
		UpdateExpression: aws.String("SET Title = :title, Description = :description, Tags = :tags, Price = :price"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":title": {
				S: aws.String(updatedGame.Title),
			},
			":description": {
				S: aws.String(updatedGame.Description),
			},
			":tags": {
				SS: aws.StringSlice(updatedGame.Tags),
			},
			":price": {
				N: aws.String(strconv.FormatFloat(updatedGame.Price, 'f', -1, 64)),
			},
		},
	})
	if err != nil {
		log.Println("Error updating item in Games table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}


func renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	t, err := template.ParseFiles("templates/" + templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
