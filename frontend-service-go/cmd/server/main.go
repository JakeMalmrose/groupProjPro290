package main

import (
	"bytes"
	"html/template"
	"net/http"
	"log"
	"github.com/hashicorp/consul/api"
	"os"
	"strconv"
)

var consulClient *api.Client

func init() {
    consulConfig := api.DefaultConfig()
    consulConfig.Address = os.Getenv("CONSUL_ADDRESS")
    var err error
    consulClient, err = api.NewClient(consulConfig)
    if err != nil {
        log.Fatal("Error creating Consul client:", err)
    }
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/games", handleGames)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/store", handleStore)
	http.HandleFunc("/library", handleLibrary)

	
	err := registerService()
    if err != nil {
        log.Fatal("Error registering service with Consul:", err)
    }

	http.ListenAndServe(":3000", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

type Game struct {
    ID     string
    Title  string
    Price  float64
    Owned  bool
}


func handleGames(w http.ResponseWriter, r *http.Request) {
    // Create a slice of Game objects (you can replace this with data from DynamoDB later)
    games := []Game{
        {ID: "1", Title: "Game 1", Price: 9.99, Owned: true},
        {ID: "2", Title: "Game 2", Price: 19.99, Owned: false},
        {ID: "3", Title: "Game 3", Price: 14.99, Owned: true},
    }

    // Pass the games data to the template
    renderTemplate(w, "gameslist.html", map[string]interface{}{
        "Games": games,
    })
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "index.html", map[string]interface{}{
			"Content": template.HTML(getContentHTML("login.html")),
		})
	} else {
		// Handle the login form submission
		// ...
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", map[string]interface{}{
		"Content": template.HTML(getContentHTML("register.html")),
	})
}

func handleStore(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", map[string]interface{}{
		"Content": template.HTML(getContentHTML("store.html")),
	})
}

func handleLibrary(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", map[string]interface{}{
		"Content": template.HTML(getContentHTML("library.html")),
	})
}

func getContentHTML(tmpl string) string {
	t, err := template.ParseFiles("templates/" + tmpl)
	if err != nil {
		return ""
	}

	var contentBuf bytes.Buffer
	err = t.Execute(&contentBuf, nil)
	if err != nil {
		return ""
	}

	return contentBuf.String()
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

func registerService() error {
    serviceName := os.Getenv("SERVICE_NAME")
    serviceID := os.Getenv("SERVICE_ID")
    port, _ := strconv.Atoi(os.Getenv("SERVICE_PORT"))

    tags := []string{
		"TRAEFIK_ENABLE=true",
		"traefik.http.routers.frontendservice.rule=PathPrefix(`/`)",
		"TRAEFIK_HTTP_SERVICES_FRONTEND_LOADBALANCER_SERVER_PORT=3000",
	}

    service := &api.AgentServiceRegistration{
        ID:      serviceID,
        Name:    serviceName,
        Address: "frontend-service",
        Port:    port,
        Tags:    tags,
    }

    return consulClient.Agent().ServiceRegister(service)
}
