package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/consul/api"

	database "github.com/Draupniyr/carts-service/database"
	structs "github.com/Draupniyr/carts-service/structs"
)

var db database.Database
var consulClient *api.Client

func init() {
	db.Init() // Initialize the database connection

	var err error
	consulConfig := api.DefaultConfig()
	consulConfig.Address = os.Getenv("CONSUL_ADDRESS")
	consulClient, err = api.NewClient(consulConfig)
	if err != nil {
		log.Fatal("Error creating Consul client:", err)
	}
}

func main() {
	http.HandleFunc("/carts", CartsHandler)
	http.HandleFunc("/carts/{ID}", CartsHandlerID)
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
	cart, err := db.GetCart(id)
	if err != nil {
		log.Println("Error getting item from Carts table:", err)
		http.Error(w, "Cart not found", http.StatusNotFound)
		return
	}

	// Render the template with the retrieved carts data
	renderTemplate(w, "cart.html", map[string]interface{}{
		"Carts": cart,
	})
}

func getCarts(w http.ResponseWriter, r *http.Request) {
	// Query the Carts table for all carts
	carts, err := db.GetAllCarts()
	if err != nil {
		log.Println("Error getting items from Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "cart.html", map[string]interface{}{
		"Carts": carts,
	})
}

func createCart(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var createRequest structs.CreateCartRequest
	err := json.NewDecoder(r.Body).Decode(&createRequest.Games)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	db.CreateAndUpdateCart(createRequest.CreateCartRequestToCart())
}

func deleteCartID(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("ID")
	db.DeleteCart(id)
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
	var updateRequest []structs.Game
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	db.AddOrRemoveFromCart(id, updateRequest)
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
