package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	database "github.com/Draupniyr/auth-service/database"
	"github.com/dgrijalva/jwt-go"
)

// Claims represents the JWT claims
type Claims struct {
    UserID string `json:"user_id"`
    jwt.StandardClaims
}

var jwtSecret = []byte("your-secret-key")

func main() {
    http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
    http.ListenAndServe(":8080", nil)
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

