package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/hashicorp/consul/api"

	database "github.com/Draupniyr/games-service/database"
	structs "github.com/Draupniyr/games-service/structs"
)

var db database.Database
var consulClient *api.Client

func init() {

	db.Init() // Initialize the database connection   Hopefully

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

	http.HandleFunc("/", GamesHandler)
	http.HandleFunc("/{id}", GamesHandlerID)
	//http.HandleFunc("/{gameID}/{updateID}", GameUpdateHandler)
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
	case http.MethodDelete: // Dev
		deleteGameID(w, r)
	case http.MethodPatch: // Dev
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
	case http.MethodDelete: // ADMIN
		deleteAllGame(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getGamesID(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("ID")
	games, err := db.GetGame(id)
	if err != nil {
		log.Println("Error getting Game from database:", err)
		http.Error(w, "Internal Server Error", http.StatusNotFound)
		return
	}
	// Render the template with the retrieved Games data
	renderTemplate(w, "gameslist.html", map[string]interface{}{
		"Games": games,
	})
}

func getGames(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("Search")

	// Filter the Games based on the search string
	GamesToDisplay, err := db.SearchGames(search)
	if err != nil {
		log.Println("Error getting Game from database:", err)
		http.Error(w, "Internal Server Error", http.StatusNotFound)
		return
	}

	renderTemplate(w, "gameslist.html", map[string]interface{}{
		"Games": GamesToDisplay,
	})
}

func createGame(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var createRequest structs.GamePostRequest
	err := json.NewDecoder(r.Body).Decode(&createRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	log.Println("Create Game Request:")
	log.Println("Title:", createRequest.Title)
	log.Println("Description:", createRequest.Description)
	log.Println("Tags:", createRequest.Tags)
	log.Println("Price:", createRequest.Price)
	log.Println("Author:", createRequest.Author)
	log.Println("AuthorID:", createRequest.AuthorID)

	db.CreateGame(createRequest.GamePostRequestToGame())
}

func deleteGameID(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("ID")
	db.DeleteGame(id)
}

func deleteAllGame(w http.ResponseWriter, r *http.Request) {
	// Delete all items from the Games table
	db.DeleteAll()
}

func updateGameID(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("ID")
	// Parse the request body
	var updateRequest structs.GamePostRequest
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	//       v The new Id and Publish are igored here, they should never be updated
	db.UpdateGame(id, updateRequest.GamePostRequestToGame())
}

// func GameUpdateHandler(w http.ResponseWriter, r *http.Request) {
// 	// get the Game ID from the URL

// 	switch r.Method {
// 	case http.MethodPost:
// 		createUpdate(w, r)
// 	case http.MethodDelete:
// 		deleteUpdate(w, r)
// 	case http.MethodPut:
// 		updateUpdate(w, r)
// 	case http.MethodGet:
// 		getUpdate(w, r)
// 	default:
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 	}
// }

// func createUpdate(w http.ResponseWriter, r *http.Request) {
// 	gameID := r.FormValue("gameID")
// 	// Parse the request body
// 	var updateRequest structs.UpdatePostObject
// 	err := json.NewDecoder(r.Body).Decode(&updateRequest)
// 	if err != nil {
// 		log.Println("Error decoding request body:", err)
// 		http.Error(w, "Bad Request", http.StatusBadRequest)
// 		return
// 	}
// 	db.CreateUpdate(gameID, updateRequest.UpdatePostObjectToUpdate())
// }

// func deleteUpdate(w http.ResponseWriter, r *http.Request) {
// 	gameID := r.FormValue("gameID")
// 	updateID := r.FormValue("updateID")
// 	db.DeleteUpdate(gameID, updateID)
// }

// func updateUpdate(w http.ResponseWriter, r *http.Request) {
// 	gameID := r.FormValue("gameID")
// 	updateID := r.FormValue("updateID")
// 	// Parse the request body
// 	var updateRequest structs.UpdatePostObject
// 	err := json.NewDecoder(r.Body).Decode(&updateRequest)
// 	if err != nil {
// 		log.Println("Error decoding request body:", err)
// 		http.Error(w, "Bad Request", http.StatusBadRequest)
// 		return
// 	}
// 	db.UpdateUpdate(gameID, updateID, updateRequest)
// }

// func getUpdate(w http.ResponseWriter, r *http.Request) {
// 	gameID := r.FormValue("gameID")
// 	updateID := r.FormValue("updateID")
// 	update, err := db.GetUpdate(gameID, updateID)
// 	if err != nil {
// 		log.Println("Error getting Update from database:", err)
// 		http.Error(w, "Internal Server Error", http.StatusNotFound)
// 		return
// 	}
// 	// todo: Render the template with the retrieved Update data
// 	renderTemplate(w, "Update.html", map[string]interface{}{
// 		"Update": update,
// 	})
// }

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
