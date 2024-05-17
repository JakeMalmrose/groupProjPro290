package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	port := 8080

	err := registerService()
	if err != nil {
		log.Fatal("Error registering service with Consul:", err)
	}

	http.HandleFunc("/carts", CartsHandler)
	http.HandleFunc("/carts/{ID}", CartsHandlerID)
	log.Fatal(http.ListenAndServe(":8080", nil))

	log.Printf("Carts service listening on port %d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

func registerService() error {
	serviceName := os.Getenv("SERVICE_NAME")
	serviceID := os.Getenv("SERVICE_ID")
	port, _ := strconv.Atoi(os.Getenv("SERVICE_PORT"))

	tags := []string{
		"TRAEFIK_ENABLE=true",
		"traefik.http.routers.cartsservice.rule=PathPrefix(`/carts`)",
		"TRAEFIK_HTTP_SERVICES_CARTS_LOADBALANCER_SERVER_PORT=3000",
	}

	service := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: "carts-service",
		Port:    port,
		Tags:    tags,
	}

	return consulClient.Agent().ServiceRegister(service)
}

func CartsHandlerID(w http.ResponseWriter, r *http.Request) {
	// Retrieve carts from DynamoDB
	switch r.Method {
	case http.MethodGet:
		getCartsID(w, r)
	case http.MethodDelete:
		deleteCartID(w, r)
	case http.MethodPatch:
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
	id := getIDfromURL(r)
	// Query the Carts table for the cart with the specified ID
	cart, err := db.GetCart(id)
	if err != nil {
		log.Println("Error getting item from Carts table:", err)
		http.Error(w, "Cart not found", http.StatusNotFound)
		return
	}

	// Render the template with the retrieved carts data
	renderTemplate(w, "cart.html", map[string]interface{}{
        "Carts": []structs.Cart{cart},
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
	log.Println("Create Cart Endpoint Hit")
	var cartRequest structs.CreateCartRequest
	err := json.NewDecoder(r.Body).Decode(&cartRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	log.Println("Create Request: ", cartRequest)

	db.CreateAndUpdateCart(cartRequest.CreateCartRequestToCart())
}

func deleteCartID(w http.ResponseWriter, r *http.Request) {
	log.Println("Delete Cart Endpoint Hit")
	id := getIDfromURL(r)
	log.Println("ID: ", id)
	db.DeleteCart(id)
}

func deleteCart(w http.ResponseWriter, r *http.Request) {
	// Delete all items from the Carts table
	err := db.DeleteAll()
	if err != nil {
		log.Println("Error deleting items from Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func updateCartID(w http.ResponseWriter, r *http.Request) {
	// get the cart ID from the URL
	id := getIDfromURL(r)
	// Parse the request body = []Game
	var game structs.Game
	err := json.NewDecoder(r.Body).Decode(&game)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	log.Println("Update Request: ", game.ID)

	db.AddOrRemoveFromCart(id, game)
}

func getIDfromURL(r *http.Request) string {
	url := r.URL.Path
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
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
