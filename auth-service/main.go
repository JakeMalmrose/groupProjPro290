package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"golang.org/x/crypto/bcrypt"

	database "github.com/Draupniyr/auth-service/database"
	kafka "github.com/Draupniyr/auth-service/kafka"
	"github.com/dgrijalva/jwt-go"
)

// Claims represents the JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

var jwtSecret = []byte("zachariah-hansen")

var consulClient *api.Client

func main() {
	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/auth/register", registerHandler)

	http.Handle("/auth/update-role", Authorize(http.HandlerFunc(updateUserRoleHandler), "admin"))

	err := registerService()
	if err != nil {
		log.Fatal("Error registering service with Consul:", err)
	}

	http.ListenAndServe(":3000", nil)
}

func init() {
	var err error
	consulConfig := api.DefaultConfig()
	consulConfig.Address = os.Getenv("CONSUL_ADDRESS")
	consulClient, err = api.NewClient(consulConfig)
	if err != nil {
		log.Fatal("Error creating Consul client:", err)
	}

}

func updateUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the user ID and new role
	var request struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Update the user's role in the database
	err = database.UpdateUserRole(request.UserID, request.Role)
	if err != nil {
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	// Return a success response
	response := map[string]string{"message": "User role updated successfully"}
	json.NewEncoder(w).Encode(response)
}

func Authorize(next http.Handler, allowedRoles ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Verify the token signature using the shared secret or public key
			return []byte("zachariah-hansen"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*jwt.StandardClaims)
		userID := claims.Subject
		userRole := claims.Audience

		if !contains(allowedRoles, userRole) {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		// Pass the user information to the next handler
		ctx := context.WithValue(r.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "userRole", userRole)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var user database.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Authenticate user credentials against DynamoDB
	authenticatedUser, err := database.AuthenticateUser(user.Username, user.Password)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else if err.Error() == "invalid password" {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Create JWT claims
	claims := &Claims{
		UserID: authenticatedUser.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
			Audience:  authenticatedUser.Audience,
		},
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the token as a JSON response
	response := map[string]string{"token": tokenString}
	json.NewEncoder(w).Encode(response)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var user database.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate username and password
	if user.Username == "" || user.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Check if the username already exists
	existingUser, _ := database.GetUserByUsername(user.Username)
	if existingUser != nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	// Generate a unique ID for the user
	user.ID = generateUserID()

	// Save the user to DynamoDB
	err = database.SaveUser(user)
	if err != nil {
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}
	//turn user to json
	userJson, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Failed to marshal user", http.StatusInternalServerError)
		return
	}

	userByte := []byte(userJson)
	//send user to kafka
	err = kafka.PushCommentToQueue("user", "registered", userByte)

	// Return a success response
	response := map[string]string{"message": "User registered successfully"}
	json.NewEncoder(w).Encode(response)
}

func generateUserID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func registerService() error {
	serviceName := os.Getenv("SERVICE_NAME")
	serviceID := os.Getenv("SERVICE_ID")
	port, _ := strconv.Atoi(os.Getenv("SERVICE_PORT"))

	tags := []string{
		"TRAEFIK_ENABLE=true",
		"traefik.http.routers.authservice.rule=PathPrefix(`/auth`)",
		"TRAEFIK_HTTP_SERVICES_AUTH_LOADBALANCER_SERVER_PORT=3000",
	}

	service := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: "auth-service",
		Port:    port,
		Tags:    tags,
	}

	return consulClient.Agent().ServiceRegister(service)
}
