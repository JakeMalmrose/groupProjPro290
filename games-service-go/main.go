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

	auth "github.com/Draupniyr/games-service/auth"
	database "github.com/Draupniyr/games-service/database"
	structs "github.com/Draupniyr/games-service/structs"
	logic "github.com/Draupniyr/games-service/logic"
)

var db database.Database
var consulClient *api.Client

func init() {

	err := db.Init("Games", "ID")
	if err != nil {
		log.Fatal("Error initializing database:", err)
	} // Initialize the database connection   Hopefully

	consulConfig := api.DefaultConfig()
	consulConfig.Address = os.Getenv("CONSUL_ADDRESS")
	consulClient, err = api.NewClient(consulConfig)
	if err != nil {
		log.Fatal("Error creating Consul client:", err)
	}
}

func main() {
	port := 3000

	err := registerService()
	if err != nil {
		log.Fatal("Error registering service with Consul:", err)
	}

	http.HandleFunc("/games/getform", GamesFormHandler)
	http.HandleFunc("/games/{id}", GamesHandlerID)
	http.HandleFunc("/games/search/{search}", getGamesBySearch)
	http.HandleFunc("/games/author/{id}", getGamesByAuthor)

	//http.HandleFunc("/{gameID}/{updateID}", GameUpdateHandler)

	http.HandleFunc("/games", GamesHandler)

	http.Handle("/games/library", auth.Authorize(http.HandlerFunc(getGamesByUserOwned)))

	// Developer endpoints
	//http.Handle("/developer/games", auth.Authorize(http.HandlerFunc(getDeveloperGames)))
	http.Handle("/games/dev/create", auth.Authorize(http.HandlerFunc(createGame)))
	http.Handle("/games/dev/delete/{id}", auth.Authorize(http.HandlerFunc(deleteGameID)))
	http.Handle("/games/dev/update/{id}", auth.Authorize(http.HandlerFunc(updateGameID)))

	// Admin endpoints
	http.Handle("/games/admin", auth.Authorize(http.HandlerFunc(getGamesAdmin), "admin"))
	http.Handle("/games/admin/delete/{id}", auth.Authorize(http.HandlerFunc(deleteGameID), "admin"))
	http.Handle("/games/admin/approve/{id}", auth.Authorize(http.HandlerFunc(approveGameID), "admin"))

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

func GamesFormHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "submitgameform.html", nil)
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

func approveGameID(w http.ResponseWriter, r *http.Request) {
	logic.UpdateGameField("Published", "Approved", "true", db)
}

func getGamesID(w http.ResponseWriter, r *http.Request) {
	id := getIDfromURL(r)
	games, err := logic.GetGame(id, &db)
	if err != nil {
		log.Println("Error getting Game from database:", err)
		http.Error(w, "Internal Server Error", http.StatusNotFound)
		return
	}
	// Render the template with the retrieved Games data
	renderTemplate(w, "gameslist2.html", map[string]interface{}{
		"Games": games,
	})
}

func getGames(w http.ResponseWriter, r *http.Request) {


	// Filter the Games based on the search string
	GamesToDisplay, err := logic.GetAllGames(&db)
	if err != nil {
		log.Println("Error getting Game from database:", err)
		http.Error(w, "Internal Server Error", http.StatusNotFound)
		return
	}

	renderTemplate(w, "gameslist2.html", map[string]interface{}{
		"Games": GamesToDisplay,
	})
}

func getGamesByUserOwned(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request context
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		log.Println("User ID not found in the request context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		log.Println("User ID is not of type string")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Print(userID)

	// Get the Games owned by the user
	// DOES NOT WORK BECAUSE USERS ARENT REALLY A THING THAT EXISTS
	//GamesToDisplay, err := db.GetGamesByUser(userID)
	//if err != nil {
	//	log.Println("Error getting Game from database:", err)
	//	http.Error(w, "Internal Server Error", http.StatusNotFound)
	//	return
	//}

	renderTemplate(w, "gameslist2.html", map[string]interface{}{
		"Games": nil,
	})
}

func getGamesAdmin(w http.ResponseWriter, r *http.Request) {
	// Get all Games from the database
	GamesToDisplay, err := logic.GetAllGames(&db)
	if err != nil {
		log.Println("Error getting Game from database:", err)
		http.Error(w, "Internal Server Error", http.StatusNotFound)
		return
	}

	renderTemplate(w, "admingameslist.html", map[string]interface{}{
		"Games": GamesToDisplay,
	})
}

func getGamesByAuthor(w http.ResponseWriter, r *http.Request) {
	// Get the author ID from the URL
	authorID := getIDfromURL(r)

	// Get the Games by the author
	GamesToDisplay, err := logic.GetGamesByAuthor(authorID, &db)
	if err != nil {
		log.Println("Error getting Game from database:", err)
		http.Error(w, "Internal Server Error", http.StatusNotFound)
		return
	}

	renderTemplate(w, "gameslist2.html", map[string]interface{}{
		"Games": GamesToDisplay,
	})
}

func getGamesBySearch(w http.ResponseWriter, r *http.Request) {
	// Get the search string from the URL
	search := getIDfromURL(r)

	// Get the Games by the search string
	GamesToDisplay, err := logic.SearchGames(search, &db)
	if err != nil {
		log.Println("Error getting Game from database:", err)
		http.Error(w, "Internal Server Error", http.StatusNotFound)
		return
	}

	renderTemplate(w, "gameslist2.html", map[string]interface{}{
		"Games": GamesToDisplay,
	})
}

func createGame(w http.ResponseWriter, r *http.Request) {
	// Get the developer's ID from the request context
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		log.Println("User ID not found in the request context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the request body
	var createRequest structs.GamePostRequest
	err := json.NewDecoder(r.Body).Decode(&createRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Set the developer ID in the create request
	createRequest.AuthorID = userID

	log.Println("Author ID: ", createRequest.AuthorID)
	log.Println("Author: ", createRequest.Author)
	log.Println("Title: ", createRequest.Title)
	log.Println("Description: ", createRequest.Description)
	log.Println("Tags: ", createRequest.Tags)
	log.Println("Price: ", createRequest.Price)

	err = logic.CreateGame(createRequest.GamePostRequestToGame(), &db)
	if err != nil {
		log.Println("Error creating Game in database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func deleteGameID(w http.ResponseWriter, r *http.Request) {
	id := getIDfromURL(r)
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		log.Println("User ID not found in the request context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logic.DeleteGame(id, userID, &db)
}

func deleteAllGame(w http.ResponseWriter, r *http.Request) {
	// Delete all items from the Games table
	db.DeleteAll()
}

func updateGameID(w http.ResponseWriter, r *http.Request) {
	id := getIDfromURL(r)
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		log.Println("User ID not found in the request context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the request body
	var updateRequest structs.GamePostRequest
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	//       v The new Id and Publish are igored here, they should never be updated
	logic.UpdateGame(id, userID, updateRequest.GamePostRequestToGame(), &db)
}

func GameUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// get the Game ID from the URL

	switch r.Method {
	case http.MethodPost:
		createUpdate(w, r)
	case http.MethodDelete:
		deleteUpdate(w, r)
	case http.MethodPut:
		updateUpdate(w, r)
	case http.MethodGet:
		getUpdate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createUpdate(w http.ResponseWriter, r *http.Request) {
	gameID := getIDfromURL(r)
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		log.Println("User ID not found in the request context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Parse the request body
	var updateRequest structs.UpdatePostObject
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	logic.CreateUpdate(gameID, userID, updateRequest.UpdatePostObjectToUpdate(), &db)
}

func deleteUpdate(w http.ResponseWriter, r *http.Request) {
	gameID, updateID := getTwoIDsfromURL(r)
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		log.Println("User ID not found in the request context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	logic.DeleteUpdate(gameID, userID, updateID, &db)
}

func updateUpdate(w http.ResponseWriter, r *http.Request) {
	gameID, updateID := getTwoIDsfromURL(r)
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		log.Println("User ID not found in the request context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Parse the request body
	var updateRequest structs.UpdatePostObject
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	logic.UpdateUpdate(gameID, userID, updateID, updateRequest, &db)
}

func getUpdate(w http.ResponseWriter, r *http.Request) {
	gameID, updateID := getTwoIDsfromURL(r)
	update, err := logic.GetUpdate(gameID, updateID, &db)
	if err != nil {
		log.Println("Error getting Update from database:", err)
		http.Error(w, "Internal Server Error", http.StatusNotFound)
		return
	}
	// todo: Render the template with the retrieved Update data
	renderTemplate(w, "Update.html", map[string]interface{}{
		"Update": update,
	})
}

func getIDfromURL(r *http.Request) string {
	url := r.URL.Path
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func getTwoIDsfromURL(r *http.Request) (string, string) {
	url := r.URL.Path
	parts := strings.Split(url, "/")
	return parts[len(parts)-2], parts[len(parts)-1]
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
