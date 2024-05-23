package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/hashicorp/consul/api"

	auth "github.com/Draupniyr/carts-service/auth"
	database "github.com/Draupniyr/carts-service/database"
	logic "github.com/Draupniyr/carts-service/logic"
	structs "github.com/Draupniyr/carts-service/structs"
	kafkaProducer "github.com/Draupniyr/carts-service/kafka"
)

var db database.Database
var consulClient *api.Client
var kafka kafkaProducer.KafkaProducer

func init() {
	err := db.Init("Carts", "ID")
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	log.Println("Database initialized")

	err =kafka.InitKafkaProducer()
	for err != nil {
		err = kafka.InitKafkaProducer()
		log.Println("Error initializing Kafka producer:", err)
	}
	log.Println("Kafka producer initialized")

	consulConfig := api.DefaultConfig()
	consulConfig.Address = os.Getenv("CONSUL_ADDRESS")
	consulClient, err = api.NewClient(consulConfig)
	if err != nil {
		log.Fatal("Error creating Consul client:", err)
	}
	log.Println("Consul client created")

}
func main() {
	port := 3000

	err := registerService()
	if err != nil {
		log.Fatal("Error registering service with Consul:", err)
	}
	// http.Handle("/games/dev/create", auth.Authorize(http.HandlerFunc(createGame)))

	http.Handle("/carts/all", auth.Authorize(http.HandlerFunc(CartsHandler)))
	http.Handle("/carts", auth.Authorize(http.HandlerFunc(CartsHandlerID)))
	http.Handle("/carts/checkout", auth.Authorize(http.HandlerFunc(checkout)))
	log.Fatal(http.ListenAndServe(":3000", nil))

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
	case http.MethodPost:
		createCart(w, r)
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
	case http.MethodDelete: // ADMIN
		deleteCart(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getCartsID(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /carts hit")
	id := r.Context().Value("userID").(string)

	cart, err := logic.GetCart(id, db)
	if err != nil {
		log.Println("Error getting item from Carts table:", err)
		http.Error(w, "Cart not found", http.StatusNotFound)
		return
	}

	renderTemplate(w, "cart.html", map[string]interface{}{
		"Cart": cart,
	})
}

func getCarts(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /carts/all hit")
	// Query the Carts table for all carts
	carts, err := logic.GetAllCarts(db)
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
	// Create a new cart
	log.Println("POST to /carts hit")
	id := r.Context().Value("userID").(string)

	var game structs.Game
	err := json.NewDecoder(r.Body).Decode(&game)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err =logic.CreateORUpdateCart(id, game, db)
	if err != nil {
		log.Println("Error creating item in Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func deleteCartID(w http.ResponseWriter, r *http.Request) {
	// Delete the cart with the given ID
	log.Println("DELETE /carts hit")
	id := r.Context().Value("userID").(string)
	err := logic.DeleteCart(id, db)
	if err != nil {
		log.Println("Error deleting item from Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func deleteCart(w http.ResponseWriter, r *http.Request) {
	// Delete all items from the Carts table
	log.Println("DELETE /carts/all hit")

	err := logic.DeleteAll(db)
	if err != nil {
		log.Println("Error deleting items from Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func updateCartID(w http.ResponseWriter, r *http.Request) {
	// Update the cart with the given ID
	log.Println("PATCH /carts hit")

	userID := r.Context().Value("userID").(string)

	var game structs.Game
	err := json.NewDecoder(r.Body).Decode(&game)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	
	err = logic.AddOrRemoveFromCart(userID, game, db)
	if err != nil {
		log.Println("Error adding or removing game from cart:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func checkout(w http.ResponseWriter, r *http.Request) {
	// Checkout the cart
	log.Println("POST /carts/checkout hit")

	id := r.Context().Value("userID").(string)
	err := logic.Checkout(id, db, kafka)
	if err != nil {
		log.Println("Error checking out cart:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "cart.html", map[string]interface{}{
		"Cart": structs.Cart{},
	})
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
