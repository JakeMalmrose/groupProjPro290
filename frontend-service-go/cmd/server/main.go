package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/hashicorp/consul/api"

	auth "github.com/Draupniyr/frontend-service/auth"
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

	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/store", handleStore)
	http.HandleFunc("/library", handleLibrary)
	http.Handle("/dev", auth.Authorize(http.HandlerFunc(handleDev), http.HandlerFunc(handleUnauthorized), "dev", "admin"))

	http.Handle("/admin", auth.Authorize(http.HandlerFunc(handleAdmin), http.HandlerFunc(handleUnauthorized), "admin"))

	http.HandleFunc("/", handleIndex) // only file with headers/footers

	err := registerService()
	if err != nil {
		log.Fatal("Error registering service with Consul:", err)
	}

	http.ListenAndServe(":3000", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin.html", nil)
}

func handleDev(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "dev.html", nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login.html", nil)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "register.html", nil)
}

func handleStore(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "store.html", nil)
}

func handleLibrary(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "library.html", nil)
}

func handleUnauthorized(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "unauthorized.html", nil)
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
