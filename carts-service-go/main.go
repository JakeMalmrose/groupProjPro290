package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type Cart struct {
	ID    string `json:"ID"`
	Games []Game `json:"Games"`
}

type Game struct {
	GameID string `json:"GameID"`
	Title  string `json:"Title"`
	Price  float64 `json:"Price"`
	Owned  bool   `json:"Owned"`
}

type DynamoCart struct {
	ID    string                     `json:"ID"`
	Games []*dynamodb.AttributeValue `json:"Games"`
}

func CartToDynamoDBItem(cart Cart) map[string]*dynamodb.AttributeValue {
	// Convert the Cart struct to a DynamoDB item
	item, err := dynamodbattribute.MarshalMap(cart)
	if err != nil {
		log.Println("Error marshalling Cart:", err)
		return nil
	}

	// Convert the []Game slice to a DynamoDB list
	var games []*dynamodb.AttributeValue
	for _, game := range cart.Games {
		gameItem, err := dynamodbattribute.MarshalMap(game)
		if err != nil {
			log.Println("Error marshalling Game:", err)
			return nil
		}
		games = append(games, &dynamodb.AttributeValue{M: gameItem})
	}
	item["Games"] = &dynamodb.AttributeValue{L: games}

	return item
}


type CreateCartRequest struct {
	Games []Game `json:"Games"`
}

var dynamodbClient *dynamodb.DynamoDB

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
}

func main() {
	http.HandleFunc("/", CartsHandler)
	http.HandleFunc("/{ID}", CartsHandlerID)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func CartsHandlerID(w http.ResponseWriter, r *http.Request) {
	// Retrieve carts from DynamoDB
	switch r.Method {
	case http.MethodGet:
		getCartsID(w, r)
	case http.MethodDelete:
		deleteCartID(w, r)
	case http.MethodPut:
		updateCartID(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func CartsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve carts from DynamoDB
	switch r.Method {
	case http.MethodGet: // ADMIN
		getCarts(w, r)
	case http.MethodPost:
		createCart(w, r)
	case http.MethodDelete: // ADMIN
		deleteCart(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getCartsID(w http.ResponseWriter, r *http.Request) {
	//get ID from URL
	id := r.FormValue("ID")
	// Query the Carts table for the cart with the specified ID
	carts, err := dynamodbClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Carts"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		log.Println("Error getting item from Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render the template with the retrieved carts data
	renderTemplate(w, "cart.html", map[string]interface{}{
		"Carts": carts,
	})
}

func getCarts(w http.ResponseWriter, r *http.Request) {
	// Query the Carts table for all carts
	carts, err := dynamodbClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("Carts"),
	})
	if err != nil {
		log.Println("Error scanning Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "cart.html", map[string]interface{}{
		"Carts": carts,
	})
}


func createCart(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var createRequest CreateCartRequest
	err := json.NewDecoder(r.Body).Decode(&createRequest.Games)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Generate a new UUID for the cart ID
	cartID := uuid.New().String()

	newCart := Cart{
		ID:    cartID,
		Games: createRequest.Games,
	}

	_, err = dynamodbClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("Carts"),
		// https://gist.github.com/wliao008/e0dba6a3cf089d46932d39b90f9d838f
		// maybe the answer for this v
		Item: CartToDynamoDBItem(newCart),
	})
	if err != nil {
		log.Println("Error putting item into Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func deleteCartID(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("ID")
	_, err := dynamodbClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Carts"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		log.Println("Error deleting item from Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func deleteCart(w http.ResponseWriter, r *http.Request) {
	// Delete all items from the Carts table
	_, err := dynamodbClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Carts"),
	})
	if err != nil {
		log.Println("Error deleting item from Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func updateCartID(w http.ResponseWriter, r *http.Request) {
	// get the cart ID from the URL
	id := r.FormValue("ID")
	// Parse the request body = []Game
	var updateRequest []Game
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	updatedCart := Cart{
		ID:    id,
		Games: updateRequest,
	}

	// update the cart with the new games
	_, err = dynamodbClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Carts"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(updatedCart.ID),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":Games": {
				L: CartToDynamoDBItem(updatedCart)["Games"].L,
			},
		},
	})
	if err != nil {
		log.Println("Error updating item in Carts table:", err)
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
