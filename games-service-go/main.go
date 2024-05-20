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

	database "github.com/Draupniyr/games-service/database"
	structs "github.com/Draupniyr/games-service/structs"
	auth "github.com/Draupniyr/games-service/auth"
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
	port := 3000

	err := registerService()
	if err != nil {
		log.Fatal("Error registering service with Consul:", err)
	}


	http.HandleFunc("/games/getform", GamesFormHandler)
	http.HandleFunc("/games/{id}", GamesHandlerID)
	//http.HandleFunc("/{gameID}/{updateID}", GameUpdateHandler)

	http.HandleFunc("/games", GamesHandler)

	// Developer endpoints
	//http.Handle("/developer/games", auth.Authorize(http.HandlerFunc(getDeveloperGames)))
	http.Handle("/developer/games/create", auth.Authorize(http.HandlerFunc(createGame)))
	http.Handle("/developer/games/delete/{id}", auth.Authorize(http.HandlerFunc(deleteGameID)))
	http.Handle("/developer/games/update/{id}", auth.Authorize(http.HandlerFunc(updateGameID)))

	// Admin endpoints
	//http.Handle("/admin/games", auth.Authorize(http.HandlerFunc(getAllGames)))
	http.Handle("/admin/games/delete/{id}", auth.Authorize(http.HandlerFunc(deleteGameID)))
	//http.Handle("/admin/games/approve/{id}", auth.Authorize(http.HandlerFunc(approveGameID)))

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

func getGamesID(w http.ResponseWriter, r *http.Request) {
	id := getIDfromURL(r)
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
	// get search string from body
	search := r.FormValue("search")

	// Filter the Games based on the search string
	GamesToDisplay, err := db.SearchGames(search)
	if err != nil {
		log.Println("Error getting Game from database:", err)
		http.Error(w, "Internal Server Error", http.StatusNotFound)
		return
	}
	log.Println("GamesToDisplay:", GamesToDisplay)

	renderTemplate(w, "gameslist2.html", map[string]interface{}{
		"Games": GamesToDisplay,
	})
}

func createGame(w http.ResponseWriter, r *http.Request) {
    // Get the developer's ID from the request context
    developerID := r.Context().Value("userID").(string)

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

    // Set the developer ID in the create request
    createRequest.AuthorID = developerID

    db.CreateGame(createRequest.GamePostRequestToGame())
}

func deleteGameID(w http.ResponseWriter, r *http.Request) {
	id := getIDfromURL(r)
	db.DeleteGame(id)
}

func deleteAllGame(w http.ResponseWriter, r *http.Request) {
	// Delete all items from the Games table
	db.DeleteAll()
}

func updateGameID(w http.ResponseWriter, r *http.Request) {
	id := getIDfromURL(r)
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
	// Parse the request body
	var updateRequest structs.UpdatePostObject
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	db.CreateUpdate(gameID, updateRequest.UpdatePostObjectToUpdate())
}

func deleteUpdate(w http.ResponseWriter, r *http.Request) {
	gameID, updateID := getTwoIDsfromURL(r)
	db.DeleteUpdate(gameID, updateID)
}

func updateUpdate(w http.ResponseWriter, r *http.Request) {
	gameID, updateID := getTwoIDsfromURL(r)
	// Parse the request body
	var updateRequest structs.UpdatePostObject
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	db.UpdateUpdate(gameID, updateID, updateRequest)
}

func getUpdate(w http.ResponseWriter, r *http.Request) {
	gameID, updateID := getTwoIDsfromURL(r)
	update, err := db.GetUpdate(gameID, updateID)
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
